/*
Package blockchain provides functions to access various blockchain data.
*/
package blockchain

import (
	"github.com/neophora/neo2go/pkg/interop/account"
	"github.com/neophora/neo2go/pkg/interop/asset"
	"github.com/neophora/neo2go/pkg/interop/block"
	"github.com/neophora/neo2go/pkg/interop/contract"
	"github.com/neophora/neo2go/pkg/interop/header"
	"github.com/neophora/neo2go/pkg/interop/transaction"
)

// GetHeight returns current block height (index of the last accepted block).
// Note that when transaction is being run as a part of new block this block is
// considered as not yet accepted (persisted) and thus you'll get an index of
// the previous (already accepted) block. This function uses
// `Neo.Blockchain.GetHeight` syscall.
func GetHeight() int {
	return 0
}

// GetHeader returns header found by the given hash (256 bit hash in BE format
// represented as a slice of 32 bytes) or index (integer). Refer to the `header`
// package for possible uses of returned structure. This function uses
// `Neo.Blockchain.GetHeader` syscall.
func GetHeader(heightOrHash interface{}) header.Header {
	return header.Header{}
}

// GetBlock returns block found by the given hash or index (with the same
// encoding as for GetHeader). Refer to the `block` package for possible uses
// of returned structure. This function uses `Neo.Blockchain.GetBlock` syscall.
func GetBlock(heightOrHash interface{}) block.Block {
	return block.Block{}
}

// GetTransaction returns transaction found by the given (256 bit in BE format
// represented as a slice of 32 bytes). Refer to the `transaction` package for
// possible uses of returned structure. This function uses
// `Neo.Blockchain.GetTransaction` syscall.
func GetTransaction(hash []byte) transaction.Transaction {
	return transaction.Transaction{}
}

// GetTransactionHeight returns transaction's height (index of the block that
// includes it) by the given ID (256 bit in BE format represented as a slice of
// 32 bytes). This function uses `Neo.Blockchain.GetTransactionHeight` syscall.
func GetTransactionHeight(hash []byte) int {
	return 0
}

// GetContract returns contract found by the given script hash (160 bit in BE
// format represented as a slice of 20 bytes). Refer to the `contract` package
// for details on how to use the returned structure. This function uses
// `Neo.Blockchain.GetContract` syscall.
func GetContract(scriptHash []byte) contract.Contract {
	return contract.Contract{}
}

// GetAccount returns account found by the given script hash (160 bit in BE
// format represented as a slice of 20 bytes). Refer to the `account` package
// for details on how to use the returned structure. This function uses
// `Neo.Blockchain.GetAccount` syscall.
func GetAccount(scriptHash []byte) account.Account {
	return account.Account{}
}

// GetValidators returns a slice of current validators public keys represented
// as a compressed serialized byte slice (33 bytes long). This function uses
// `Neo.Blockchain.GetValidators` syscall.
func GetValidators() [][]byte {
	return nil
}

// GetAsset returns asset found by the given asset ID (256 bit in BE format
// represented as a slice of 32 bytes). Refer to the `asset` package for
// possible uses of returned structure. This function uses
// `Neo.Blockchain.GetAsset` syscall.
func GetAsset(assetID []byte) asset.Asset {
	return asset.Asset{}
}
