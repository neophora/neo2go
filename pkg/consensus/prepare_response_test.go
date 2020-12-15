package consensus

import (
	"testing"

	"github.com/neophora/neo2go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestPrepareResponse_Setters(t *testing.T) {
	var p prepareResponse

	p.SetPreparationHash(util.Uint256{1, 2, 3})
	require.Equal(t, util.Uint256{1, 2, 3}, p.PreparationHash())
}
