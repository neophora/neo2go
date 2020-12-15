package core

import (
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/util"
)

// A HeaderHashList represents a list of header hashes.
// This data structure in not routine safe and should be
// used under some kind of protection against race conditions.
type HeaderHashList struct {
	hashes []util.Uint256
}

// NewHeaderHashListFromBytes returns a new hash list from the given bytes.
func NewHeaderHashListFromBytes(b []byte) (*HeaderHashList, error) {
	return nil, nil
}

// NewHeaderHashList returns a new pointer to a HeaderHashList.
func NewHeaderHashList(hashes ...util.Uint256) *HeaderHashList {
	return &HeaderHashList{
		hashes: hashes,
	}
}

// Add appends the given hash to the list of hashes.
func (l *HeaderHashList) Add(h ...util.Uint256) {
	l.hashes = append(l.hashes, h...)
}

// Len returns the length of the underlying hashes slice.
func (l *HeaderHashList) Len() int {
	return len(l.hashes)
}

// Get returns the hash by the given index.
func (l *HeaderHashList) Get(i int) util.Uint256 {
	if l.Len() <= i {
		return util.Uint256{}
	}
	return l.hashes[i]
}

// Last return the last hash in the HeaderHashList.
func (l *HeaderHashList) Last() util.Uint256 {
	return l.hashes[l.Len()-1]
}

// Slice return a subslice of the underlying hashes.
// Subsliced from start to end.
// Example:
// 	headers := headerList.Slice(0, 2000)
func (l *HeaderHashList) Slice(start, end int) []util.Uint256 {
	return l.hashes[start:end]
}

// WriteTo writes n underlying hashes to the given BinWriter
// starting from start.
func (l *HeaderHashList) Write(bw *io.BinWriter, start, n int) error {
	bw.WriteArray(l.Slice(start, start+n))
	return bw.Err
}
