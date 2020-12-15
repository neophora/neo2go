package smartcontract

import (
	"fmt"
	"sort"

	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/vm/emit"
	"github.com/neophora/neo2go/pkg/vm/opcode"
)

// CreateMultiSigRedeemScript creates a script runnable by the VM.
func CreateMultiSigRedeemScript(m int, publicKeys keys.PublicKeys) ([]byte, error) {
	if m < 1 {
		return nil, fmt.Errorf("param m cannot be smaller or equal to 1 got %d", m)
	}
	if m > len(publicKeys) {
		return nil, fmt.Errorf("length of the signatures (%d) is higher then the number of public keys", m)
	}
	if m > 1024 {
		return nil, fmt.Errorf("public key count %d exceeds maximum of length 1024", len(publicKeys))
	}

	buf := io.NewBufBinWriter()
	emit.Int(buf.BinWriter, int64(m))
	sort.Sort(publicKeys)
	for _, pubKey := range publicKeys {
		emit.Bytes(buf.BinWriter, pubKey.Bytes())
	}
	emit.Int(buf.BinWriter, int64(len(publicKeys)))
	emit.Opcode(buf.BinWriter, opcode.CHECKMULTISIG)

	return buf.Bytes(), nil
}
