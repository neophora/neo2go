package core

import "github.com/neophora/neo2go/pkg/core/transaction"

// spentCoin represents the state of a single spent coin output.
type spentCoin struct {
	Output      *transaction.Output
	StartHeight uint32
	EndHeight   uint32
}
