package result

import (
	"github.com/neophora/neo2go/pkg/smartcontract"
)

// Invoke represents code invocation result and is used by several RPC calls
// that invoke functions, scripts and generic bytecode.
type Invoke struct {
	State       string                    `json:"state"`
	GasConsumed string                    `json:"gas_consumed"`
	Script      string                    `json:"script"`
	Stack       []smartcontract.Parameter `json:"stack"`
}
