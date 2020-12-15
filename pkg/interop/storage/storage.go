/*
Package storage provides functions to access and modify contract's storage.
Neo storage's model follows simple key-value DB pattern, this storage is a part
of blockchain state, so you can use it between various invocations of the same
contract.
*/
package storage

import "github.com/neophora/neo2go/pkg/interop/iterator"

// Context represents storage context that is mandatory for Put/Get/Delete
// operations. It's an opaque type that can only be created properly by
// GetContext, GetReadOnlyContext or ConvertContextToReadOnly. It's similar
// to Neo .net framework's StorageContext class.
type Context struct{}

// ConvertContextToReadOnly returns new context from the given one, but with
// writing capability turned off, so that you could only invoke Get and Find
// using this new Context. If Context is already read-only this function is a
// no-op. It uses `Neo.StorageContext.AsReadOnly` syscall.
func ConvertContextToReadOnly(ctx Context) Context { return Context{} }

// GetContext returns current contract's (that invokes this function) storage
// context. It uses `Neo.Storage.GetContext` syscall.
func GetContext() Context { return Context{} }

// GetReadOnlyContext returns current contract's (that invokes this function)
// storage context in read-only mode, you can use this context for Get and Find
// functions, but using it for Put and Delete will fail. It uses
// `Neo.Storage.GetReadOnlyContext` syscall.
func GetReadOnlyContext() Context { return Context{} }

// Put saves given value with given key in the storage using given Context.
// Even though it accepts interface{} for both, you can only pass simple types
// there like string, []byte, int or bool (not structures or slices of more
// complex types). To put more complex types there serialize them first using
// runtime.Serialize. This function uses `Neo.Storage.Put` syscall.
func Put(ctx Context, key interface{}, value interface{}) {}

// Get retrieves value stored for the given key using given Context. See Put
// documentation on possible key and value types. This function uses
// `Neo.Storage.Get` syscall.
func Get(ctx Context, key interface{}) interface{} { return 0 }

// Delete removes key-value pair from storage by the given key using given
// Context. See Put documentation on possible key types. This function uses
// `Neo.Storage.Delete` syscall.
func Delete(ctx Context, key interface{}) {}

// Find returns an iterator.Iterator over key-value pairs in the given Context
// that match the given key (contain it as a prefix). See Put documentation on
// possible key types and iterator package documentation on how to use the
// returned value. This function uses `Neo.Storage.Find` syscall.
func Find(ctx Context, key interface{}) iterator.Iterator { return iterator.Iterator{} }
