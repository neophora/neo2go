package client

import (
	"encoding/hex"

	"github.com/neophora/neo2go/pkg/core"
	"github.com/neophora/neo2go/pkg/core/block"
	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/encoding/address"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/rpc/request"
	"github.com/neophora/neo2go/pkg/rpc/response/result"
	"github.com/neophora/neo2go/pkg/smartcontract"
	"github.com/neophora/neo2go/pkg/util"
	"github.com/neophora/neo2go/pkg/wallet"
	"github.com/pkg/errors"
)

// GetAccountState returns detailed information about a NEO account.
func (c *Client) GetAccountState(address string) (*result.AccountState, error) {
	var (
		params = request.NewRawParams(address)
		resp   = &result.AccountState{}
	)
	if err := c.performRequest("getaccountstate", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetAllTransferTx returns all transfer transactions for a given account within
// specified timestamps (by block time) with specified output limits and page. It
// only works with neo-go 0.78.0+ servers.
func (c *Client) GetAllTransferTx(acc util.Uint160, start, end uint32, limit, page int) ([]result.TransferTx, error) {
	var (
		params = request.NewRawParams(acc.StringLE(), start, end, limit, page)
		resp   = new([]result.TransferTx)
	)
	if err := c.performRequest("getalltransfertx", params, resp); err != nil {
		return nil, err
	}
	return *resp, nil
}

// GetApplicationLog returns the contract log based on the specified txid.
func (c *Client) GetApplicationLog(hash util.Uint256) (*result.ApplicationLog, error) {
	var (
		params = request.NewRawParams(hash.StringLE())
		resp   = &result.ApplicationLog{}
	)
	if err := c.performRequest("getapplicationlog", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetAssetState queries the asset information, based on the specified asset number.
func (c *Client) GetAssetState(hash util.Uint256) (*result.AssetState, error) {
	var (
		params = request.NewRawParams(hash.StringLE())
		resp   = &result.AssetState{}
	)
	if err := c.performRequest("getassetstate", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetBestBlockHash returns the hash of the tallest block in the main chain.
func (c *Client) GetBestBlockHash() (util.Uint256, error) {
	var resp = util.Uint256{}
	if err := c.performRequest("getbestblockhash", request.NewRawParams(), &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetBlockCount returns the number of blocks in the main chain.
func (c *Client) GetBlockCount() (uint32, error) {
	var resp uint32
	if err := c.performRequest("getblockcount", request.NewRawParams(), &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetBlockByIndex returns a block by its height.
func (c *Client) GetBlockByIndex(index uint32) (*block.Block, error) {
	return c.getBlock(request.NewRawParams(index))
}

// GetBlockByHash returns a block by its hash.
func (c *Client) GetBlockByHash(hash util.Uint256) (*block.Block, error) {
	return c.getBlock(request.NewRawParams(hash.StringLE()))
}

func (c *Client) getBlock(params request.RawParams) (*block.Block, error) {
	var (
		resp string
		err  error
		b    *block.Block
	)
	if err = c.performRequest("getblock", params, &resp); err != nil {
		return nil, err
	}
	blockBytes, err := hex.DecodeString(resp)
	if err != nil {
		return nil, err
	}
	r := io.NewBinReaderFromBuf(blockBytes)
	b = new(block.Block)
	b.DecodeBinary(r)
	if r.Err != nil {
		return nil, r.Err
	}
	return b, nil
}

// GetBlockByIndexVerbose returns a block wrapper with additional metadata by
// its height.
// NOTE: to get transaction.ID and transaction.Size, use t.Hash() and io.GetVarSize(t) respectively.
func (c *Client) GetBlockByIndexVerbose(index uint32) (*result.Block, error) {
	return c.getBlockVerbose(request.NewRawParams(index, 1))
}

// GetBlockByHashVerbose returns a block wrapper with additional metadata by
// its hash.
func (c *Client) GetBlockByHashVerbose(hash util.Uint256) (*result.Block, error) {
	return c.getBlockVerbose(request.NewRawParams(hash.StringLE(), 1))
}

func (c *Client) getBlockVerbose(params request.RawParams) (*result.Block, error) {
	var (
		resp = &result.Block{}
		err  error
	)
	if err = c.performRequest("getblock", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetBlockHash returns the hash value of the corresponding block, based on the specified index.
func (c *Client) GetBlockHash(index uint32) (util.Uint256, error) {
	var (
		params = request.NewRawParams(index)
		resp   = util.Uint256{}
	)
	if err := c.performRequest("getblockhash", params, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetBlockHeader returns the corresponding block header information from serialized hex string
// according to the specified script hash.
func (c *Client) GetBlockHeader(hash util.Uint256) (*block.Header, error) {
	var (
		params = request.NewRawParams(hash.StringLE())
		resp   string
		h      *block.Header
	)
	if err := c.performRequest("getblockheader", params, &resp); err != nil {
		return nil, err
	}
	headerBytes, err := hex.DecodeString(resp)
	if err != nil {
		return nil, err
	}
	r := io.NewBinReaderFromBuf(headerBytes)
	h = new(block.Header)
	h.DecodeBinary(r)
	if r.Err != nil {
		return nil, r.Err
	}
	return h, nil
}

// GetBlockHeaderVerbose returns the corresponding block header information from Json format string
// according to the specified script hash.
func (c *Client) GetBlockHeaderVerbose(hash util.Uint256) (*result.Header, error) {
	var (
		params = request.NewRawParams(hash.StringLE(), 1)
		resp   = &result.Header{}
	)
	if err := c.performRequest("getblockheader", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetBlockSysFee returns the system fees of the block, based on the specified index.
func (c *Client) GetBlockSysFee(index uint32) (util.Fixed8, error) {
	var (
		params = request.NewRawParams(index)
		resp   util.Fixed8
	)
	if err := c.performRequest("getblocksysfee", params, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// getBlockTransferTx is an internal version of GetBlockTransferTxByIndex/GetBlockTransferTxByHash.
func (c *Client) getBlockTransferTx(param interface{}) ([]result.TransferTx, error) {
	var (
		params = request.NewRawParams(param)
		resp   = new([]result.TransferTx)
	)
	if err := c.performRequest("getblocktransfertx", params, resp); err != nil {
		return nil, err
	}
	return *resp, nil
}

// GetBlockTransferTxByIndex returns all transfer transactions from a block.
// It only works with neo-go 0.79.0+ servers.
func (c *Client) GetBlockTransferTxByIndex(index uint32) ([]result.TransferTx, error) {
	return c.getBlockTransferTx(index)
}

// GetBlockTransferTxByHash returns all transfer transactions from a block.
// It only works with neo-go 0.79.0+ servers.
func (c *Client) GetBlockTransferTxByHash(hash util.Uint256) ([]result.TransferTx, error) {
	return c.getBlockTransferTx(hash)
}

// GetClaimable returns tx outputs which can be claimed.
func (c *Client) GetClaimable(address string) (*result.ClaimableInfo, error) {
	params := request.NewRawParams(address)
	resp := new(result.ClaimableInfo)
	if err := c.performRequest("getclaimable", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetConnectionCount returns the current number of connections for the node.
func (c *Client) GetConnectionCount() (int, error) {
	var (
		params = request.NewRawParams()
		resp   int
	)
	if err := c.performRequest("getconnectioncount", params, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetContractState queries contract information, according to the contract script hash.
func (c *Client) GetContractState(hash util.Uint160) (*result.ContractState, error) {
	var (
		params = request.NewRawParams(hash.StringLE())
		resp   = &result.ContractState{}
	)
	if err := c.performRequest("getcontractstate", params, resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetNEP5Balances is a wrapper for getnep5balances RPC.
func (c *Client) GetNEP5Balances(address util.Uint160) (*result.NEP5Balances, error) {
	params := request.NewRawParams(address.StringLE())
	resp := new(result.NEP5Balances)
	if err := c.performRequest("getnep5balances", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetNEP5Transfers is a wrapper for getnep5transfers RPC. Address parameter
// is mandatory, while all the others are optional. Start and stop parameters
// are supported since neo-go 0.77.0 and limit and page since neo-go 0.78.0.
// These parameters are positional in the JSON-RPC call, you can't specify limit
// and not specify start/stop for example.
func (c *Client) GetNEP5Transfers(address string, start, stop *uint32, limit, page *int) (*result.NEP5Transfers, error) {
	params := request.NewRawParams(address)
	if start != nil {
		params.Values = append(params.Values, *start)
		if stop != nil {
			params.Values = append(params.Values, *stop)
			if limit != nil {
				params.Values = append(params.Values, *limit)
				if page != nil {
					params.Values = append(params.Values, *page)
				}
			} else if page != nil {
				return nil, errors.New("bad parameters")
			}
		} else if limit != nil || page != nil {
			return nil, errors.New("bad parameters")
		}
	} else if stop != nil || limit != nil || page != nil {
		return nil, errors.New("bad parameters")
	}
	resp := new(result.NEP5Transfers)
	if err := c.performRequest("getnep5transfers", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetPeers returns the list of nodes that the node is currently connected/disconnected from.
func (c *Client) GetPeers() (*result.GetPeers, error) {
	var (
		params = request.NewRawParams()
		resp   = &result.GetPeers{}
	)
	if err := c.performRequest("getpeers", params, resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetRawMemPool returns the list of unconfirmed transactions in memory.
func (c *Client) GetRawMemPool() ([]util.Uint256, error) {
	var (
		params = request.NewRawParams()
		resp   = new([]util.Uint256)
	)
	if err := c.performRequest("getrawmempool", params, resp); err != nil {
		return *resp, err
	}
	return *resp, nil
}

// GetRawTransaction returns a transaction by hash.
func (c *Client) GetRawTransaction(hash util.Uint256) (*transaction.Transaction, error) {
	var (
		params = request.NewRawParams(hash.StringLE())
		resp   string
		err    error
	)
	if err = c.performRequest("getrawtransaction", params, &resp); err != nil {
		return nil, err
	}
	txBytes, err := hex.DecodeString(resp)
	if err != nil {
		return nil, err
	}
	r := io.NewBinReaderFromBuf(txBytes)
	tx := new(transaction.Transaction)
	tx.DecodeBinary(r)
	if r.Err != nil {
		return nil, r.Err
	}
	return tx, nil
}

// GetRawTransactionVerbose returns a transaction wrapper with additional
// metadata by transaction's hash.
// NOTE: to get transaction.ID and transaction.Size, use t.Hash() and io.GetVarSize(t) respectively.
func (c *Client) GetRawTransactionVerbose(hash util.Uint256) (*result.TransactionOutputRaw, error) {
	var (
		params = request.NewRawParams(hash.StringLE(), 1)
		resp   = &result.TransactionOutputRaw{}
		err    error
	)
	if err = c.performRequest("getrawtransaction", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetStorage returns the stored value, according to the contract script hash and the stored key.
func (c *Client) GetStorage(hash util.Uint160, key []byte) ([]byte, error) {
	var (
		params = request.NewRawParams(hash.StringLE(), hex.EncodeToString(key))
		resp   string
	)
	if err := c.performRequest("getstorage", params, &resp); err != nil {
		return nil, err
	}
	res, err := hex.DecodeString(resp)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// GetTransactionHeight returns the block index in which the transaction is found.
func (c *Client) GetTransactionHeight(hash util.Uint256) (uint32, error) {
	var (
		params = request.NewRawParams(hash.StringLE())
		resp   uint32
	)
	if err := c.performRequest("gettransactionheight", params, &resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetTxOut returns the corresponding unspent transaction output information (returned change),
// based on the specified hash and index.
func (c *Client) GetTxOut(hash util.Uint256, num int) (*result.TransactionOutput, error) {
	var (
		params = request.NewRawParams(hash.StringLE(), num)
		resp   = &result.TransactionOutput{}
	)
	if err := c.performRequest("gettxout", params, resp); err != nil {
		return resp, err
	}
	return resp, nil
}

// GetUnclaimed returns unclaimed GAS amount of the specified address.
func (c *Client) GetUnclaimed(address string) (*result.Unclaimed, error) {
	var (
		params = request.NewRawParams(address)
		resp   = &result.Unclaimed{}
	)
	if err := c.performRequest("getunclaimed", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUnspents returns UTXOs for the given NEO account.
func (c *Client) GetUnspents(address string) (*result.Unspents, error) {
	var (
		params = request.NewRawParams(address)
		resp   = &result.Unspents{}
	)
	if err := c.performRequest("getunspents", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetUTXOTransfers is a wrapper for getutxoransfers RPC. Address parameter
// is mandatory, while all the others are optional. It's only supported since
// neo-go 0.77.0 with limit and page parameters only since neo-go 0.78.0.
// These parameters are positional in the JSON-RPC call, you can't specify limit
// and not specify start/stop for example.
func (c *Client) GetUTXOTransfers(address string, start, stop *uint32, limit, page *int) (*result.GetUTXO, error) {
	params := request.NewRawParams(address)
	if start != nil {
		params.Values = append(params.Values, *start)
		if stop != nil {
			params.Values = append(params.Values, *stop)
			if limit != nil {
				params.Values = append(params.Values, *limit)
				if page != nil {
					params.Values = append(params.Values, *page)
				}
			} else if page != nil {
				return nil, errors.New("bad parameters")
			}
		} else if limit != nil || page != nil {
			return nil, errors.New("bad parameters")
		}
	} else if stop != nil || limit != nil || page != nil {
		return nil, errors.New("bad parameters")
	}
	resp := new(result.GetUTXO)
	if err := c.performRequest("getutxotransfers", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// GetValidators returns the current NEO consensus nodes information and voting status.
func (c *Client) GetValidators() ([]result.Validator, error) {
	var (
		params = request.NewRawParams()
		resp   = new([]result.Validator)
	)
	if err := c.performRequest("getvalidators", params, resp); err != nil {
		return nil, err
	}
	return *resp, nil
}

// GetVersion returns the version information about the queried node.
func (c *Client) GetVersion() (*result.Version, error) {
	var (
		params = request.NewRawParams()
		resp   = &result.Version{}
	)
	if err := c.performRequest("getversion", params, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// InvokeScript returns the result of the given script after running it true the VM.
// NOTE: This is a test invoke and will not affect the blockchain.
func (c *Client) InvokeScript(script string, hashesForVerifying []util.Uint160) (*result.Invoke, error) {
	params := request.NewRawParams(script)
	return c.invokeSomething("invokescript", params, hashesForVerifying)
}

// InvokeFunction returns the results after calling the smart contract scripthash
// with the given operation and parameters.
// NOTE: this is test invoke and will not affect the blockchain.
func (c *Client) InvokeFunction(script, operation string, params []smartcontract.Parameter, hashesForVerifying []util.Uint160) (*result.Invoke, error) {
	p := request.NewRawParams(script, operation, params)
	return c.invokeSomething("invokefunction", p, hashesForVerifying)
}

// Invoke returns the results after calling the smart contract scripthash
// with the given parameters.
// NOTE: this is test invoke and will not affect the blockchain.
func (c *Client) Invoke(script string, params []smartcontract.Parameter, hashesForVerifying []util.Uint160) (*result.Invoke, error) {
	p := request.NewRawParams(script, params)
	return c.invokeSomething("invoke", p, hashesForVerifying)
}

// invokeSomething is an inner wrapper for Invoke* functions
func (c *Client) invokeSomething(method string, p request.RawParams, hashesForVerifying []util.Uint160) (*result.Invoke, error) {
	var resp = new(result.Invoke)
	if hashesForVerifying != nil {
		p.Values = append(p.Values, hashesForVerifying)
	}
	if err := c.performRequest(method, p, resp); err != nil {
		return nil, err
	}
	return resp, nil
}

// SendRawTransaction broadcasts a transaction over the NEO network.
// The given hex string needs to be signed with a keypair.
// When the result of the response object is true, the TX has successfully
// been broadcasted to the network.
func (c *Client) SendRawTransaction(rawTX *transaction.Transaction) error {
	var (
		params = request.NewRawParams(hex.EncodeToString(rawTX.Bytes()))
		resp   bool
	)
	if err := c.performRequest("sendrawtransaction", params, &resp); err != nil {
		return err
	}
	if !resp {
		return errors.New("sendrawtransaction returned false")
	}
	return nil
}

// SubmitBlock broadcasts a raw block over the NEO network.
func (c *Client) SubmitBlock(b block.Block) error {
	var (
		params request.RawParams
		resp   bool
	)
	buf := io.NewBufBinWriter()
	b.EncodeBinary(buf.BinWriter)
	if err := buf.Err; err != nil {
		return err
	}
	params = request.NewRawParams(hex.EncodeToString(buf.Bytes()))

	if err := c.performRequest("submitblock", params, &resp); err != nil {
		return err
	}
	if !resp {
		return errors.New("submitblock returned false")
	}
	return nil
}

// TransferAsset sends an amount of specific asset to a given address.
// This call requires open wallet. (`wif` key in client struct.)
// If response.Result is `true` then transaction was formed correctly and was written in blockchain.
func (c *Client) TransferAsset(asset util.Uint256, address string, amount util.Fixed8) (util.Uint256, error) {
	var (
		err      error
		rawTx    *transaction.Transaction
		txParams = request.ContractTxParams{
			AssetID:  asset,
			Address:  address,
			Value:    amount,
			WIF:      c.WIF(),
			Balancer: c.opts.Balancer,
		}
		resp util.Uint256
	)

	if rawTx, err = request.CreateRawContractTransaction(txParams); err != nil {
		return resp, errors.Wrap(err, "failed to create raw transaction")
	}
	if err = c.SendRawTransaction(rawTx); err != nil {
		return resp, errors.Wrap(err, "failed to send raw transaction")
	}
	return rawTx.Hash(), nil
}

// SignAndPushInvocationTx signs and pushes given script as an invocation
// transaction  using given wif to sign it and spending the amount of gas
// specified. It returns a hash of the invocation transaction and an error.
func (c *Client) SignAndPushInvocationTx(script []byte, acc *wallet.Account, sysfee util.Fixed8, netfee util.Fixed8) (util.Uint256, error) {
	var txHash util.Uint256
	var err error

	tx := transaction.NewInvocationTX(script, sysfee)
	gas := sysfee + netfee

	if gas > 0 {
		if err = request.AddInputsAndUnspentsToTx(tx, acc.Address, core.UtilityTokenID(), gas, c); err != nil {
			return txHash, errors.Wrap(err, "failed to add inputs and unspents to transaction")
		}
	} else {
		addr, err := address.StringToUint160(acc.Address)
		if err != nil {
			return txHash, errors.Wrap(err, "failed to get address")
		}
		tx.AddVerificationHash(addr)
	}

	if err = acc.SignTx(tx); err != nil {
		return txHash, errors.Wrap(err, "failed to sign tx")
	}
	txHash = tx.Hash()
	err = c.SendRawTransaction(tx)

	if err != nil {
		return txHash, errors.Wrap(err, "failed sendning tx")
	}
	return txHash, nil
}

// ValidateAddress verifies that the address is a correct NEO address.
func (c *Client) ValidateAddress(address string) error {
	var (
		params = request.NewRawParams(address)
		resp   = &result.ValidateAddress{}
	)

	if err := c.performRequest("validateaddress", params, resp); err != nil {
		return err
	}
	if !resp.IsValid {
		return errors.New("validateaddress returned false")
	}
	return nil
}
