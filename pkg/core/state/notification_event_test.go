package state

import (
	"testing"

	"github.com/neophora/neo2go/pkg/internal/random"
	"github.com/neophora/neo2go/pkg/internal/testserdes"
	"github.com/neophora/neo2go/pkg/smartcontract"
	"github.com/neophora/neo2go/pkg/vm"
)

func TestEncodeDecodeNotificationEvent(t *testing.T) {
	event := &NotificationEvent{
		ScriptHash: random.Uint160(),
		Item:       vm.NewBoolItem(true),
	}

	testserdes.EncodeDecodeBinary(t, event, new(NotificationEvent))
}

func TestEncodeDecodeAppExecResult(t *testing.T) {
	appExecResult := &AppExecResult{
		TxHash:      random.Uint256(),
		Trigger:     1,
		VMState:     "Hault",
		GasConsumed: 10,
		Stack:       []smartcontract.Parameter{},
		Events:      []NotificationEvent{},
	}

	testserdes.EncodeDecodeBinary(t, appExecResult, new(AppExecResult))
}
