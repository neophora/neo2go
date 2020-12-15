package result

import (
	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/util"
)

// Validator used for the representation of
// state.Validator on the RPC Server.
type Validator struct {
	PublicKey keys.PublicKey `json:"publickey"`
	Votes     util.Fixed8    `json:"votes"`
	Active    bool           `json:"active"`
}
