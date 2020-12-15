package compiler_test

import (
	"testing"

	"github.com/neophora/neo2go/pkg/vm"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoragePutGet(t *testing.T) {
	src := `
		package foo

		import "github.com/neophora/neo2go/pkg/interop/storage"

		func Main() string {
			ctx := storage.GetContext()
			key := []byte("token")
			storage.Put(ctx, key, []byte("foo"))
			x := storage.Get(ctx, key)
			return x.(string)
		}
	`
	eval(t, src, []byte("foo"))
}

func TestNotify(t *testing.T) {
	src := `package foo
	import "github.com/neophora/neo2go/pkg/interop/runtime"
	func Main(arg int) {
		runtime.Notify(arg, "sum", arg+1)
		runtime.Notify()
		runtime.Notify("single")
	}`

	v, s := vmAndCompileInterop(t, src)
	v.Estack().PushVal(11)

	require.NoError(t, v.Run())
	require.Equal(t, 3, len(s.events))

	exp0 := []vm.StackItem{vm.NewBigIntegerItem(11), vm.NewByteArrayItem([]byte("sum")), vm.NewBigIntegerItem(12)}
	assert.Equal(t, exp0, s.events[0].Value())
	assert.Equal(t, []vm.StackItem{}, s.events[1].Value())
	assert.Equal(t, []vm.StackItem{vm.NewByteArrayItem([]byte("single"))}, s.events[2].Value())
}
