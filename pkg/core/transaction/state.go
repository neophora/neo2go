package transaction

import (
	"github.com/neophora/neo2go/pkg/io"
)

// StateTX represents a state transaction.
type StateTX struct {
	Descriptors []*StateDescriptor
}

// DecodeBinary implements Serializable interface.
func (tx *StateTX) DecodeBinary(r *io.BinReader) {
	r.ReadArray(&tx.Descriptors)
}

// EncodeBinary implements Serializable interface.
func (tx *StateTX) EncodeBinary(w *io.BinWriter) {
	w.WriteArray(tx.Descriptors)
}
