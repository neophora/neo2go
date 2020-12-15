package runtimecontract

import (
	"github.com/neophora/neo2go/pkg/interop/runtime"
	"github.com/neophora/neo2go/pkg/interop/util"
)

// Check if the invoker of the contract is the specified owner
var owner = util.FromAddress("Aej1fe4mUgou48Zzup5j8sPrE3973cJ5oz")

// Main is something to be ran from outside.
func Main(operation string, args []interface{}) bool {
	trigger := runtime.GetTrigger()

	// Log owner upon Verification trigger
	if trigger == runtime.Verification() {
		if runtime.CheckWitness(owner) {
			runtime.Log("Verified Owner")
		}
		return true
	}

	// Discerns between log and notify for this test
	if trigger == runtime.Application() {
		return handleOperation(operation, args)
	}

	return false
}

func handleOperation(operation string, args []interface{}) bool {
	if operation == "log" {
		message := args[0].(string)
		runtime.Log(message)
		return true
	}

	if operation == "notify" {
		runtime.Notify(args[0])
		return true
	}

	return false
}
