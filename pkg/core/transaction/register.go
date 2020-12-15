package transaction

import (
	"encoding/json"

	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/util"
)

// RegisterTX represents a register transaction.
// NOTE: This is deprecated.
type RegisterTX struct {
	// The type of the asset being registered.
	AssetType AssetType

	// Name of the asset being registered.
	Name string

	// Amount registered.
	// Unlimited mode -0.00000001.
	Amount util.Fixed8

	// Decimals.
	Precision uint8

	// Public key of the owner.
	Owner keys.PublicKey

	Admin util.Uint160
}

// DecodeBinary implements Serializable interface.
func (tx *RegisterTX) DecodeBinary(br *io.BinReader) {
	tx.AssetType = AssetType(br.ReadB())

	tx.Name = br.ReadString()

	tx.Amount.DecodeBinary(br)
	tx.Precision = uint8(br.ReadB())

	tx.Owner.DecodeBinary(br)

	br.ReadBytes(tx.Admin[:])
}

// EncodeBinary implements Serializable interface.
func (tx *RegisterTX) EncodeBinary(bw *io.BinWriter) {
	bw.WriteB(byte(tx.AssetType))
	bw.WriteString(tx.Name)
	tx.Amount.EncodeBinary(bw)
	bw.WriteB(byte(tx.Precision))
	bw.WriteBytes(tx.Owner.Bytes())
	bw.WriteBytes(tx.Admin[:])
}

// registeredAsset is a wrapper for RegisterTransaction
type registeredAsset struct {
	AssetType AssetType       `json:"type,omitempty"`
	Name      json.RawMessage `json:"name,omitempty"`
	Amount    util.Fixed8     `json:"amount,omitempty"`
	Precision uint8           `json:"precision,omitempty"`
	Owner     keys.PublicKey  `json:"owner,omitempty"`
	Admin     string          `json:"admin,omitempty"`
}
