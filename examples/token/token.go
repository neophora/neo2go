package tokencontract

import (
	"github.com/neophora/neo2go/examples/token/nep5"

	"github.com/neophora/neo2go/pkg/interop/storage"
	"github.com/neophora/neo2go/pkg/interop/util"
)

const (
	decimals   = 8
	multiplier = 100000000
)

var owner = util.FromAddress("AK2nJJpJr6o664CWJKi1QRXjqeic2zRp8y")

// CreateToken initializes the Token Interface for the Smart Contract to operate with
func CreateToken() nep5.Token {
	return nep5.Token{
		Name:           "Awesome NEO Token",
		Symbol:         "ANT",
		Decimals:       decimals,
		Owner:          owner,
		TotalSupply:    11000000 * multiplier,
		CirculationKey: "TokenCirculation",
	}
}

// Main function = contract entry
func Main(operation string, args []interface{}) interface{} {
	token := CreateToken()

	if operation == "name" {
		return token.Name
	}
	if operation == "symbol" {
		return token.Symbol
	}
	if operation == "decimals" {
		return token.Decimals
	}

	// The following operations need ctx
	ctx := storage.GetContext()

	if operation == "totalSupply" {
		return token.GetSupply(ctx)
	}
	if operation == "balanceOf" {
		hodler := args[0].([]byte)
		return token.BalanceOf(ctx, hodler)
	}
	if operation == "transfer" && CheckArgs(args, 3) {
		from := args[0].([]byte)
		to := args[1].([]byte)
		amount := args[2].(int)
		return token.Transfer(ctx, from, to, amount)
	}
	if operation == "mint" && CheckArgs(args, 1) {
		addr := args[0].([]byte)
		return token.Mint(ctx, addr)
	}

	return true
}

// CheckArgs checks args array against a length indicator
func CheckArgs(args []interface{}, length int) bool {
	if len(args) == length {
		return true
	}

	return false
}
