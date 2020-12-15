package transaction

import "github.com/neophora/neo2go/pkg/util"

// Result represents the Result of a transaction.
type Result struct {
	AssetID util.Uint256
	Amount  util.Fixed8
}
