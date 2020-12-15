package enginecontract

import (
	"github.com/neophora/neo2go/pkg/interop/engine"
	"github.com/neophora/neo2go/pkg/interop/runtime"
)

// Main is that famous Main() function, you know.
func Main() bool {
	tx := engine.GetScriptContainer()
	runtime.Notify(tx)

	callingScriptHash := engine.GetCallingScriptHash()
	runtime.Notify(callingScriptHash)

	execScriptHash := engine.GetExecutingScriptHash()
	runtime.Notify(execScriptHash)

	entryScriptHash := engine.GetEntryScriptHash()
	runtime.Notify(entryScriptHash)

	return true
}
