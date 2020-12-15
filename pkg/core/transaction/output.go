package transaction

import (
	"encoding/json"

	"github.com/neophora/neo2go/pkg/encoding/address"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/util"
)

// Output represents a Transaction output.
type Output struct {
	// The NEO asset id used in the transaction.
	AssetID util.Uint256 `json:"asset"`

	// Amount of AssetType send or received.
	Amount util.Fixed8 `json:"value"`

	// The address of the recipient.
	ScriptHash util.Uint160 `json:"address"`

	// The position of the Output in slice []Output. This is actually set in NewTransactionOutputRaw
	// and used for displaying purposes.
	Position int `json:"n"`
}

type outputAux struct {
	AssetID    util.Uint256 `json:"asset"`
	Amount     util.Fixed8  `json:"value"`
	ScriptHash string       `json:"address"`
	Position   int          `json:"n"`
}

// NewOutput returns a new transaction output.
func NewOutput(assetID util.Uint256, amount util.Fixed8, scriptHash util.Uint160) *Output {
	return &Output{
		AssetID:    assetID,
		Amount:     amount,
		ScriptHash: scriptHash,
	}
}

// DecodeBinary implements Serializable interface.
func (out *Output) DecodeBinary(br *io.BinReader) {
	br.ReadBytes(out.AssetID[:])
	out.Amount.DecodeBinary(br)
	br.ReadBytes(out.ScriptHash[:])
}

// EncodeBinary implements Serializable interface.
func (out *Output) EncodeBinary(bw *io.BinWriter) {
	bw.WriteBytes(out.AssetID[:])
	out.Amount.EncodeBinary(bw)
	bw.WriteBytes(out.ScriptHash[:])
}

// MarshalJSON implements the Marshaler interface.
func (out *Output) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}{
		"asset":   out.AssetID,
		"value":   out.Amount,
		"address": address.Uint160ToString(out.ScriptHash),
		"n":       out.Position,
	})
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (out *Output) UnmarshalJSON(data []byte) error {
	var outAux outputAux
	err := json.Unmarshal(data, &outAux)
	if err != nil {
		return err
	}
	out.ScriptHash, err = address.StringToUint160(outAux.ScriptHash)
	if err != nil {
		return err
	}
	out.Amount = outAux.Amount
	out.AssetID = outAux.AssetID
	out.Position = outAux.Position
	return nil
}
