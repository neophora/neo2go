package state

import (
	"testing"

	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/internal/random"
	"github.com/neophora/neo2go/pkg/internal/testserdes"
	"github.com/neophora/neo2go/pkg/util"
)

func TestDecodeEncodeUnspentCoin(t *testing.T) {
	unspent := &UnspentCoin{
		Height: 100500,
		States: []OutputState{
			{
				Output: transaction.Output{
					AssetID:    random.Uint256(),
					Amount:     util.Fixed8(42),
					ScriptHash: random.Uint160(),
				},
				SpendHeight: 201000,
				State:       CoinSpent,
			},
			{
				Output: transaction.Output{
					AssetID:    random.Uint256(),
					Amount:     util.Fixed8(420),
					ScriptHash: random.Uint160(),
				},
				SpendHeight: 0,
				State:       CoinConfirmed,
			},
			{
				Output: transaction.Output{
					AssetID:    random.Uint256(),
					Amount:     util.Fixed8(4200),
					ScriptHash: random.Uint160(),
				},
				SpendHeight: 111000,
				State:       CoinSpent & CoinClaimed,
			},
		},
	}

	testserdes.EncodeDecodeBinary(t, unspent, new(UnspentCoin))
}
