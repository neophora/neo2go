package transaction

import (
	"encoding/hex"
	"encoding/json"

	"github.com/neophora/neo2go/pkg/crypto/hash"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/util"
)

// Witness contains 2 scripts.
type Witness struct {
	InvocationScript   []byte `json:"invocation"`
	VerificationScript []byte `json:"verification"`
}

// DecodeBinary implements Serializable interface.
func (w *Witness) DecodeBinary(br *io.BinReader) {
	w.InvocationScript = br.ReadVarBytes()
	w.VerificationScript = br.ReadVarBytes()
}

// EncodeBinary implements Serializable interface.
func (w *Witness) EncodeBinary(bw *io.BinWriter) {
	bw.WriteVarBytes(w.InvocationScript)
	bw.WriteVarBytes(w.VerificationScript)
}

// MarshalJSON implements the json marshaller interface.
func (w Witness) MarshalJSON() ([]byte, error) {
	data := map[string]string{
		"invocation":   hex.EncodeToString(w.InvocationScript),
		"verification": hex.EncodeToString(w.VerificationScript),
	}

	return json.Marshal(data)
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (w *Witness) UnmarshalJSON(data []byte) error {
	m := map[string]string{}
	err := json.Unmarshal(data, &m)
	if err != nil {
		return err
	}
	if w.InvocationScript, err = hex.DecodeString(m["invocation"]); err != nil {
		return err
	}
	w.VerificationScript, err = hex.DecodeString(m["verification"])
	return err
}

// ScriptHash returns the hash of the VerificationScript.
func (w Witness) ScriptHash() util.Uint160 {
	return hash.Hash160(w.VerificationScript)
}
