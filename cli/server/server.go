package server

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/neophora/neo2go/pkg/config"
	"github.com/neophora/neo2go/pkg/core"
	"github.com/neophora/neo2go/pkg/core/block"
	"github.com/neophora/neo2go/pkg/core/state"
	"github.com/neophora/neo2go/pkg/core/storage"
	"github.com/neophora/neo2go/pkg/encoding/address"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/network"
	"github.com/neophora/neo2go/pkg/network/metrics"
	"github.com/neophora/neo2go/pkg/rpc/server"
	"github.com/pkg/errors"
	"github.com/urfave/cli"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// NewCommands returns 'node' command.
func NewCommands() []cli.Command {
	var cfgFlags = []cli.Flag{
		cli.StringFlag{Name: "config-path"},
		cli.BoolFlag{Name: "privnet, p"},
		cli.BoolFlag{Name: "mainnet, m"},
		cli.BoolFlag{Name: "testnet, t"},
		cli.BoolFlag{Name: "debug, d"},
	}
	var cfgWithCountFlags = make([]cli.Flag, len(cfgFlags))
	copy(cfgWithCountFlags, cfgFlags)
	cfgWithCountFlags = append(cfgWithCountFlags,
		cli.UintFlag{
			Name:  "count, c",
			Usage: "number of blocks to be processed (default or 0: all chain)",
		},
		cli.UintFlag{
			Name:  "start, s",
			Usage: "block number to start from (default: 0)",
		},
	)
	var cfgCountOutFlags = make([]cli.Flag, len(cfgWithCountFlags))
	copy(cfgCountOutFlags, cfgWithCountFlags)
	cfgCountOutFlags = append(cfgCountOutFlags,
		cli.StringFlag{
			Name:  "out, o",
			Usage: "Output file (stdout if not given)",
		},
		cli.StringFlag{
			Name:  "state, r",
			Usage: "File to export state roots to",
		},
	)
	var cfgCountInFlags = make([]cli.Flag, len(cfgWithCountFlags))
	copy(cfgCountInFlags, cfgWithCountFlags)
	cfgCountInFlags = append(cfgCountInFlags,
		cli.StringFlag{
			Name:  "in, i",
			Usage: "Input file (stdin if not given)",
		},
		cli.StringFlag{
			Name:  "dump",
			Usage: "directory for storing JSON dumps",
		},
		cli.BoolFlag{
			Name:  "diff, k",
			Usage: "Use if DB is restore from diff and not full dump",
		},
		cli.StringFlag{
			Name:  "state, r",
			Usage: "File to import state roots from",
		},
	)
	return []cli.Command{
		{
			Name:   "node",
			Usage:  "start a NEO node",
			Action: startServer,
			Flags:  cfgFlags,
		},
		{
			Name:  "db",
			Usage: "database manipulations",
			Subcommands: []cli.Command{
				{
					Name:  "dump",
					Usage: "dump blocks (starting with block #1) to the file",
					UsageText: "When --start option is provided format is different because " +
						"index of the first block is written first.",
					Action: dumpDB,
					Flags:  cfgCountOutFlags,
				},
				{
					Name:   "restore",
					Usage:  "restore blocks from the file",
					Action: restoreDB,
					Flags:  cfgCountInFlags,
				},
			},
		},
	}
}

func newGraceContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	go func() {
		<-stop
		cancel()
	}()
	return ctx
}

// getConfigFromContext looks at path and mode flags in the given config and
// returns appropriate config.
func getConfigFromContext(ctx *cli.Context) (config.Config, error) {
	var net = config.ModePrivNet
	if ctx.Bool("testnet") {
		net = config.ModeTestNet
	}
	if ctx.Bool("mainnet") {
		net = config.ModeMainNet
	}
	configPath := "./config"
	if argCp := ctx.String("config-path"); argCp != "" {
		configPath = argCp
	}
	return config.Load(configPath, net)
}

// handleLoggingParams reads logging parameters.
// If user selected debug level -- function enables it.
// If logPath is configured -- function creates dir and file for logging.
func handleLoggingParams(ctx *cli.Context, cfg config.ApplicationConfiguration) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if ctx.Bool("debug") {
		level = zapcore.DebugLevel
	}

	cc := zap.NewProductionConfig()
	cc.DisableCaller = true
	cc.DisableStacktrace = true
	cc.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder
	cc.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cc.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cc.Encoding = "console"
	cc.Level = zap.NewAtomicLevelAt(level)
	cc.Sampling = nil

	if logPath := cfg.LogPath; logPath != "" {
		if err := io.MakeDirForFile(logPath, "logger"); err != nil {
			return nil, err
		}

		cc.OutputPaths = []string{logPath}
	}

	return cc.Build()
}

func initBCWithMetrics(cfg config.Config, log *zap.Logger) (*core.Blockchain, *metrics.Service, *metrics.Service, error) {
	chain, err := initBlockChain(cfg, log)
	if err != nil {
		return nil, nil, nil, cli.NewExitError(err, 1)
	}
	configureAddresses(cfg.ApplicationConfiguration)
	prometheus := metrics.NewPrometheusService(cfg.ApplicationConfiguration.Prometheus, log)
	pprof := metrics.NewPprofService(cfg.ApplicationConfiguration.Pprof, log)

	go chain.Run()
	go prometheus.Start()
	go pprof.Start()

	return chain, prometheus, pprof, nil
}

func dumpDB(ctx *cli.Context) error {
	cfg, err := getConfigFromContext(ctx)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	log, err := handleLoggingParams(ctx, cfg.ApplicationConfiguration)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	count := uint32(ctx.Uint("count"))
	start := uint32(ctx.Uint("start"))

	var outStream = os.Stdout
	if out := ctx.String("out"); out != "" {
		outStream, err = os.Create(out)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}
	defer outStream.Close()
	writer := io.NewBinWriterFromIO(outStream)

	chain, prometheus, pprof, err := initBCWithMetrics(cfg, log)
	if err != nil {
		return err
	}

	chainCount := chain.BlockHeight() + 1
	if start+count > chainCount {
		return cli.NewExitError(fmt.Errorf("chain is not that high (%d) to dump %d blocks starting from %d", chainCount-1, count, start), 1)
	}
	if count == 0 {
		count = chainCount - start
	}

	var rootsWriter *io.BinWriter
	if out := ctx.String("state"); out != "" {
		rootsStream, err := os.Create(out)
		if err != nil {
			return cli.NewExitError(fmt.Errorf("can't create file for state roots: %w", err), 1)
		}
		defer rootsStream.Close()
		rootsWriter = io.NewBinWriterFromIO(rootsStream)
	}
	if start != 0 {
		writer.WriteU32LE(start)
	}
	writer.WriteU32LE(count)

	rootsCount := count
	if rootsWriter != nil {
		rootsWriter.WriteU32LE(start)
		if h := chain.StateHeight() + 1; start+rootsCount > h {
			log.Info("state height is low, state root dump will be cut", zap.Uint32("height", h))
			rootsCount = h - start
		}
		rootsWriter.WriteU32LE(rootsCount)
	}

	for i := start; i < start+count; i++ {
		if rootsWriter != nil && i < start+rootsCount {
			r, err := chain.GetStateRoot(i)
			if err != nil {
				return cli.NewExitError(fmt.Errorf("failed to get stateroot %d: %w", i, err), 1)
			}
			if err := writeSizedItem(&r.MPTRoot, rootsWriter); err != nil {
				return cli.NewExitError(err, 1)
			}
		}
		bh := chain.GetHeaderHash(int(i))
		b, err := chain.GetBlock(bh)
		if err != nil {
			return cli.NewExitError(fmt.Errorf("failed to get block %d: %s", i, err), 1)
		}
		if err := writeSizedItem(b, writer); err != nil {
			return cli.NewExitError(err, 1)
		}
	}
	pprof.ShutDown()
	prometheus.ShutDown()
	chain.Close()
	return nil
}

func writeSizedItem(item io.Serializable, w *io.BinWriter) error {
	buf := io.NewBufBinWriter()
	item.EncodeBinary(buf.BinWriter)
	bytes := buf.Bytes()
	w.WriteU32LE(uint32(len(bytes)))
	w.WriteBytes(bytes)
	return w.Err
}

func restoreDB(ctx *cli.Context) error {
	cfg, err := getConfigFromContext(ctx)
	if err != nil {
		return err
	}
	log, err := handleLoggingParams(ctx, cfg.ApplicationConfiguration)
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	count := uint32(ctx.Uint("count"))
	start := uint32(ctx.Uint("start"))

	var inStream = os.Stdin
	if in := ctx.String("in"); in != "" {
		inStream, err = os.Open(in)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}
	defer inStream.Close()
	reader := io.NewBinReaderFromIO(inStream)

	var rootsIn *io.BinReader
	if in := ctx.String("state"); in != "" {
		inStream, err := os.Open(in)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		defer inStream.Close()
		rootsIn = io.NewBinReaderFromIO(inStream)
	}

	dumpDir := ctx.String("dump")
	if dumpDir != "" {
		cfg.ProtocolConfiguration.SaveStorageBatch = true
	}

	chain, prometheus, pprof, err := initBCWithMetrics(cfg, log)
	if err != nil {
		return err
	}
	defer chain.Close()
	defer prometheus.ShutDown()
	defer pprof.ShutDown()

	dumpStart := uint32(0)
	dumpSize := reader.ReadU32LE()
	if ctx.Bool("diff") {
		// in diff first uint32 is the index of the first block
		dumpStart = dumpSize
		dumpSize = reader.ReadU32LE()
	}
	if reader.Err != nil {
		return cli.NewExitError(reader.Err, 1)
	}
	if start < dumpStart {
		return cli.NewExitError(fmt.Errorf("input file start from %d block, can't import %d", dumpStart, start), 1)
	}

	lastBlock := dumpStart + dumpSize
	if start+count > lastBlock {
		return cli.NewExitError(fmt.Errorf("input file has blocks up until %d, can't read %d starting from %d", lastBlock, count, start), 1)
	}
	if count == 0 {
		count = lastBlock - start
	}

	var rootStart, rootSize uint32
	rootCount := count
	if rootsIn != nil {
		rootStart = rootsIn.ReadU32LE()
		rootSize = rootsIn.ReadU32LE()
		if rootsIn.Err != nil {
			return cli.NewExitError(fmt.Errorf("error while reading roots file: %w", rootsIn.Err), 1)
		}
		if start < rootStart {
			return cli.NewExitError(fmt.Errorf("roots file start from %d root, can't import %d", rootStart, start), 1)
		}
		lastRoot := rootStart + rootSize
		if rootStart+rootCount > lastRoot {
			log.Info("state root height is low", zap.Uint32("height", lastRoot))
			rootCount = lastRoot - rootStart
		}
	}

	i := dumpStart
	for ; i < start; i++ {
		_, err := readBytes(reader)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
	}
	if rootsIn != nil {
		for j := rootStart; j < start; j++ {
			_, err := readBytes(rootsIn)
			if err != nil {
				return cli.NewExitError(err, 1)
			}
		}
	}

	gctx := newGraceContext()
	var lastIndex uint32
	dump := newDump()
	defer func() {
		_ = dump.tryPersist(dumpDir, lastIndex)
	}()

	for ; i < start+count; i++ {
		select {
		case <-gctx.Done():
			return cli.NewExitError("cancelled", 1)
		default:
		}
		block := new(block.Block)
		if err := readSizedItem(block, reader); err != nil {
			return cli.NewExitError(err, 1)
		}
		var skipBlock bool
		if block.Index == 0 && i == 0 && start == 0 {
			genesis, err := chain.GetBlock(block.Hash())
			if err == nil && genesis.Index == 0 {
				log.Info("skipped genesis block", zap.String("hash", block.Hash().StringLE()))
				skipBlock = true
			}
		}
		if !skipBlock {
			err = chain.AddBlock(block)
			if err != nil {
				return cli.NewExitError(fmt.Errorf("failed to add block %d: %s", i, err), 1)
			}
			if dumpDir != "" {
				batch := chain.LastBatch()
				dump.add(block.Index, batch)
				lastIndex = block.Index
				if block.Index%1000 == 0 {
					if err := dump.tryPersist(dumpDir, block.Index); err != nil {
						return cli.NewExitError(fmt.Errorf("can't dump storage to file: %v", err), 1)
					}
				}
			}
		}
		if rootsIn != nil && i < rootStart+rootCount {
			sr := new(state.MPTRoot)
			if err := readSizedItem(sr, rootsIn); err != nil {
				return cli.NewExitError(err, 1)
			}
			err = chain.AddStateRoot(sr)
			if err != nil {
				return cli.NewExitError(fmt.Errorf("can't add state root: %w", err), 1)
			}
		}
	}
	return nil
}

func readSizedItem(item io.Serializable, r *io.BinReader) error {
	bytes, err := readBytes(r)
	if err != nil {
		return err
	}
	newReader := io.NewBinReaderFromBuf(bytes)
	item.DecodeBinary(newReader)
	return newReader.Err
}

// readBytes performs reading of block size and then bytes with the length equal to that size.
func readBytes(reader *io.BinReader) ([]byte, error) {
	var size = reader.ReadU32LE()
	bytes := make([]byte, size)
	reader.ReadBytes(bytes)
	if reader.Err != nil {
		return nil, reader.Err
	}
	return bytes, nil
}

func startServer(ctx *cli.Context) error {
	cfg, err := getConfigFromContext(ctx)
	if err != nil {
		return err
	}
	log, err := handleLoggingParams(ctx, cfg.ApplicationConfiguration)
	if err != nil {
		return err
	}

	grace, cancel := context.WithCancel(newGraceContext())
	defer cancel()

	serverConfig := network.NewServerConfig(cfg)

	chain, prometheus, pprof, err := initBCWithMetrics(cfg, log)
	if err != nil {
		return err
	}

	serv, err := network.NewServer(serverConfig, chain, log)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("failed to create network server: %v", err), 1)
	}
	rpcServer := server.New(chain, cfg.ApplicationConfiguration.RPC, serv, log)
	errChan := make(chan error)

	go serv.Start(errChan)
	go rpcServer.Start(errChan)

	fmt.Println(logo())
	fmt.Println(serv.UserAgent)
	fmt.Println()

	var shutdownErr error
Main:
	for {
		select {
		case err := <-errChan:
			shutdownErr = errors.Wrap(err, "Error encountered by server")
			cancel()

		case <-grace.Done():
			serv.Shutdown()
			if serverErr := rpcServer.Shutdown(); serverErr != nil {
				shutdownErr = errors.Wrap(serverErr, "Error encountered whilst shutting down server")
			}
			prometheus.ShutDown()
			pprof.ShutDown()
			chain.Close()
			break Main
		}
	}

	if shutdownErr != nil {
		return cli.NewExitError(shutdownErr, 1)
	}

	return nil
}

// configureAddresses sets up addresses for RPC, Prometheus and Pprof depending from the provided config.
// In case RPC or Prometheus or Pprof Address provided each of them will use it.
// In case global Address (of the node) provided and RPC/Prometheus/Pprof don't have configured addresses they will
// use global one. So Node and RPC and Prometheus and Pprof will run on one address.
func configureAddresses(cfg config.ApplicationConfiguration) {
	if cfg.Address != "" {
		if cfg.RPC.Address == "" {
			cfg.RPC.Address = cfg.Address
		}
		if cfg.Prometheus.Address == "" {
			cfg.Prometheus.Address = cfg.Address
		}
		if cfg.Pprof.Address == "" {
			cfg.Pprof.Address = cfg.Address
		}
	}
}

// initBlockChain initializes BlockChain with preselected DB.
func initBlockChain(cfg config.Config, log *zap.Logger) (*core.Blockchain, error) {
	store, err := storage.NewStore(cfg.ApplicationConfiguration.DBConfiguration)
	if err != nil {
		return nil, cli.NewExitError(fmt.Errorf("could not initialize storage: %s", err), 1)
	}

	chain, err := core.NewBlockchain(store, cfg.ProtocolConfiguration, log)
	if err != nil {
		return nil, cli.NewExitError(fmt.Errorf("could not initialize blockchain: %s", err), 1)
	}
	if cfg.ProtocolConfiguration.AddressVersion != 0 {
		address.Prefix = cfg.ProtocolConfiguration.AddressVersion
	}
	return chain, nil
}

func logo() string {
	return `
    _   ____________        __________
   / | / / ____/ __ \      / ____/ __ \
  /  |/ / __/ / / / /_____/ / __/ / / /
 / /|  / /___/ /_/ /_____/ /_/ / /_/ /
/_/ |_/_____/\____/      \____/\____/
`
}
