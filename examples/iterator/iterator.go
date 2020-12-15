package iteratorcontract

import (
	"github.com/neophora/neo2go/pkg/interop/iterator"
	"github.com/neophora/neo2go/pkg/interop/runtime"
	"github.com/neophora/neo2go/pkg/interop/storage"
)

// Main is Main(), really.
func Main() bool {
	iter := storage.Find(storage.GetContext(), []byte("foo"))
	values := iterator.Values(iter)
	keys := iterator.Keys(iter)

	runtime.Notify("found storage values", values)
	runtime.Notify("found storage keys", keys)

	return true
}
