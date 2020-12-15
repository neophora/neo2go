package result

import (
	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/encoding/address"
	"github.com/neophora/neo2go/pkg/util"
)

// TransactionOutput is a wrapper to represent transaction's output.
type TransactionOutput struct {
	N       int         `json:"n"`
	Asset   string      `json:"asset"`
	Value   util.Fixed8 `json:"value"`
	Address string      `json:"address"`
}

// NewTxOutput converts out to a TransactionOutput.
func NewTxOutput(out *transaction.Output) *TransactionOutput {
	addr := address.Uint160ToString(out.ScriptHash)

	return &TransactionOutput{
		N:       out.Position,
		Asset:   "0x" + out.AssetID.String(),
		Value:   out.Amount,
		Address: addr,
	}
}
