package smartcontract

import (
	"testing"

	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/vm/opcode"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateMultiSigRedeemScript(t *testing.T) {
	val1, _ := keys.NewPublicKeyFromString("03b209fd4f53a7170ea4444e0cb0a6bb6a53c2bd016926989cf85f9b0fba17a70c")
	val2, _ := keys.NewPublicKeyFromString("02df48f60e8f3e01c48ff40b9b7f1310d7a8b2a193188befe1c2e3df740e895093")
	val3, _ := keys.NewPublicKeyFromString("03b8d9d5771d8f513aa0869b9cc8d50986403b78c6da36890638c3d46a5adce04a")

	validators := []*keys.PublicKey{val1, val2, val3}

	out, err := CreateMultiSigRedeemScript(3, validators)
	require.NoError(t, err)

	br := io.NewBinReaderFromBuf(out)
	assert.Equal(t, opcode.PUSH3, opcode.Opcode(br.ReadB()))

	for i := 0; i < len(validators); i++ {
		bb := br.ReadVarBytes()
		require.NoError(t, br.Err)
		assert.Equal(t, validators[i].Bytes(), bb)
	}

	assert.Equal(t, opcode.PUSH3, opcode.Opcode(br.ReadB()))
	assert.Equal(t, opcode.CHECKMULTISIG, opcode.Opcode(br.ReadB()))
}
