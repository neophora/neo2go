package mpt

import (
	"errors"
	"fmt"

	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/util"
)

// HashNode represents MPT's hash node.
type HashNode struct {
	BaseNode
}

var _ Node = (*HashNode)(nil)

// NewHashNode returns hash node with the specified hash.
func NewHashNode(h util.Uint256) *HashNode {
	return &HashNode{
		BaseNode: BaseNode{
			hash:      h,
			hashValid: true,
		},
	}
}

// Type implements Node interface.
func (h *HashNode) Type() NodeType { return HashT }

// Hash implements Node interface.
func (h *HashNode) Hash() util.Uint256 {
	if !h.hashValid {
		panic("can't get hash of an empty HashNode")
	}
	return h.hash
}

// IsEmpty returns true if h is an empty node i.e. contains no hash.
func (h *HashNode) IsEmpty() bool { return !h.hashValid }

// Bytes returns serialized HashNode.
func (h *HashNode) Bytes() []byte {
	return h.getBytes(h)
}

// DecodeBinary implements io.Serializable.
func (h *HashNode) DecodeBinary(r *io.BinReader) {
	sz := r.ReadVarUint()
	switch sz {
	case 0:
		h.hashValid = false
	case util.Uint256Size:
		h.hashValid = true
		r.ReadBytes(h.hash[:])
	default:
		r.Err = fmt.Errorf("invalid hash node size: %d", sz)
	}
}

// EncodeBinary implements io.Serializable.
func (h HashNode) EncodeBinary(w *io.BinWriter) {
	if !h.hashValid {
		w.WriteVarUint(0)
		return
	}
	w.WriteVarBytes(h.hash[:])
}

// MarshalJSON implements json.Marshaler.
func (h *HashNode) MarshalJSON() ([]byte, error) {
	if !h.hashValid {
		return []byte(`{}`), nil
	}
	return []byte(`{"hash":"` + h.hash.StringLE() + `"}`), nil
}

// UnmarshalJSON implements json.Unmarshaler.
func (h *HashNode) UnmarshalJSON(data []byte) error {
	var obj NodeObject
	if err := obj.UnmarshalJSON(data); err != nil {
		return err
	} else if u, ok := obj.Node.(*HashNode); ok {
		*h = *u
		return nil
	}
	return errors.New("expected hash node")
}
