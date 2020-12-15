package context

import (
	"encoding/hex"
	"encoding/json"

	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/smartcontract"
	"github.com/neophora/neo2go/pkg/util"
)

// Item represents a transaction context item.
type Item struct {
	Script     util.Uint160
	Parameters []smartcontract.Parameter
	Signatures map[string][]byte
}

type itemAux struct {
	Script     util.Uint160              `json:"script"`
	Parameters []smartcontract.Parameter `json:"parameters"`
	Signatures map[string]string         `json:"signatures"`
}

// GetSignature returns signature for pub if present.
func (it *Item) GetSignature(pub *keys.PublicKey) []byte {
	return it.Signatures[hex.EncodeToString(pub.Bytes())]
}

// AddSignature adds a signature for pub.
func (it *Item) AddSignature(pub *keys.PublicKey, sig []byte) {
	pubHex := hex.EncodeToString(pub.Bytes())
	it.Signatures[pubHex] = sig
}

// MarshalJSON implements json.Marshaler interface.
func (it Item) MarshalJSON() ([]byte, error) {
	ci := itemAux{
		Script:     it.Script,
		Parameters: it.Parameters,
		Signatures: make(map[string]string, len(it.Signatures)),
	}

	for key, sig := range it.Signatures {
		ci.Signatures[key] = hex.EncodeToString(sig)
	}

	return json.Marshal(ci)
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (it *Item) UnmarshalJSON(data []byte) error {
	ci := new(itemAux)
	if err := json.Unmarshal(data, ci); err != nil {
		return err
	}

	sigs := make(map[string][]byte, len(ci.Signatures))
	for keyHex, sigHex := range ci.Signatures {
		_, err := keys.NewPublicKeyFromString(keyHex)
		if err != nil {
			return err
		}
		sig, err := hex.DecodeString(sigHex)
		if err != nil {
			return err
		}
		sigs[keyHex] = sig
	}

	it.Signatures = sigs
	it.Script = ci.Script
	it.Parameters = ci.Parameters
	return nil
}
