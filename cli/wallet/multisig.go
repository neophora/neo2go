package wallet

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/encoding/address"
	"github.com/neophora/neo2go/pkg/rpc/client"
	"github.com/neophora/neo2go/pkg/smartcontract/context"
	"github.com/urfave/cli"
)

func newMultisigCommands() []cli.Command {
	return []cli.Command{
		{
			Name:      "sign",
			Usage:     "sign a transaction",
			UsageText: "multisig sign --path <path> --addr <addr> --in <file.in> --out <file.out>",
			Action:    signMultisig,
			Flags: []cli.Flag{
				walletPathFlag,
				rpcFlag,
				timeoutFlag,
				outFlag,
				inFlag,
				cli.StringFlag{
					Name:  "addr",
					Usage: "Address to use",
				},
			},
		},
	}
}

func signMultisig(ctx *cli.Context) error {
	wall, err := openWallet(ctx.String("path"))
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	defer wall.Close()

	c, err := readParameterContext(ctx.String("in"))
	if err != nil {
		return cli.NewExitError(err, 1)
	}
	addr := ctx.String("addr")
	sh, err := address.StringToUint160(addr)
	if err != nil {
		return cli.NewExitError(fmt.Errorf("invalid address: %v", err), 1)
	}
	acc := wall.GetAccount(sh)
	if acc == nil {
		return cli.NewExitError(fmt.Errorf("can't find account for the address: %s", addr), 1)
	}

	tx, ok := c.Verifiable.(*transaction.Transaction)
	if !ok {
		return cli.NewExitError("verifiable item is not a transaction", 1)
	}
	printTxInfo(tx)
	fmt.Println("Enter password to unlock wallet and sign the transaction")
	pass, err := readPassword("Password > ")
	if err != nil {
		return cli.NewExitError(err, 1)
	} else if err := acc.Decrypt(pass); err != nil {
		return cli.NewExitError(fmt.Errorf("can't unlock an account: %v", err), 1)
	}

	priv := acc.PrivateKey()
	sign := priv.Sign(tx.GetSignedPart())
	if err := c.AddSignature(acc.Contract, priv.PublicKey(), sign); err != nil {
		return cli.NewExitError(fmt.Errorf("can't add signature: %v", err), 1)
	} else if err := writeParameterContext(c, ctx.String("out")); err != nil {
		return cli.NewExitError(err, 1)
	}
	if endpoint := ctx.String("rpc"); endpoint != "" {
		w, err := c.GetWitness(acc.Contract)
		if err != nil {
			return cli.NewExitError(err, 1)
		}
		tx.Scripts = append(tx.Scripts, *w)

		gctx, cancel := getGoContext(ctx)
		defer cancel()

		c, err := client.New(gctx, ctx.String("rpc"), client.Options{})
		if err != nil {
			return cli.NewExitError(err, 1)
		} else if err := c.SendRawTransaction(tx); err != nil {
			return cli.NewExitError(err, 1)
		}
	}

	fmt.Println(tx.Hash().StringLE())
	return nil
}

func readParameterContext(filename string) (*context.ParameterContext, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("can't read input file: %v", err)
	}

	c := new(context.ParameterContext)
	if err := json.Unmarshal(data, c); err != nil {
		return nil, fmt.Errorf("can't parse transaction: %v", err)
	}
	return c, nil
}

func writeParameterContext(c *context.ParameterContext, filename string) error {
	if data, err := json.Marshal(c); err != nil {
		return fmt.Errorf("can't marshal transaction: %v", err)
	} else if err := ioutil.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("can't write transaction to file: %v", err)
	}
	return nil
}

func printTxInfo(t *transaction.Transaction) {
	fmt.Printf("Hash: %s\n", t.Hash().StringLE())
	for i := range t.Inputs {
		fmt.Printf("Input%02d: [%2d] %s\n", i, t.Inputs[i].PrevIndex, t.Inputs[i].PrevHash.StringLE())
	}
	for i := range t.Outputs {
		fmt.Printf("Output%02d:\n", i)
		fmt.Printf("\tAssetID   : %s\n", t.Outputs[i].AssetID.StringLE())
		fmt.Printf("\tAmount    : %s\n", t.Outputs[i].Amount.String())
		h := t.Outputs[i].ScriptHash
		fmt.Printf("\tScriptHash: %s\n", t.Outputs[i].ScriptHash.StringLE())
		fmt.Printf("\tToAddr    : %s\n", address.Uint160ToString(h))
	}
}
