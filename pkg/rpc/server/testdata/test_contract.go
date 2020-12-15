package testdata

import (
	"github.com/neophora/neo2go/pkg/interop/contract"
	"github.com/neophora/neo2go/pkg/interop/engine"
	"github.com/neophora/neo2go/pkg/interop/runtime"
	"github.com/neophora/neo2go/pkg/interop/storage"
)

const (
	totalSupply = 1000000
	decimals    = 2
)

func Main(operation string, args []interface{}) interface{} {
	runtime.Notify("contract call", operation, args)
	switch operation {
	case "Put":
		ctx := storage.GetContext()
		storage.Put(ctx, args[0].([]byte), args[1].([]byte))
		return true
	case "totalSupply":
		return totalSupply
	case "decimals":
		return decimals
	case "name":
		return "Rubl"
	case "symbol":
		return "RUB"
	case "balanceOf":
		ctx := storage.GetContext()
		addr := args[0].([]byte)
		if len(addr) != 20 {
			runtime.Log("invalid address")
			return false
		}
		amount := storage.Get(ctx, addr).(int)
		runtime.Notify("balanceOf", addr, amount)
		return amount
	case "transfer":
		ctx := storage.GetContext()
		from := args[0].([]byte)
		if len(from) != 20 {
			runtime.Log("invalid 'from' address")
			return false
		}
		to := args[1].([]byte)
		if len(to) != 20 {
			runtime.Log("invalid 'to' address")
			return false
		}
		amount := args[2].(int)
		if amount < 0 {
			runtime.Log("invalid amount")
			return false
		}

		fromBalance := storage.Get(ctx, from).(int)
		if fromBalance < amount {
			runtime.Log("insufficient funds")
			return false
		}
		fromBalance -= amount
		storage.Put(ctx, from, fromBalance)

		toBalance := storage.Get(ctx, to).(int)
		toBalance += amount
		storage.Put(ctx, to, toBalance)

		runtime.Notify("transfer", from, to, amount)

		return true
	case "init":
		ctx := storage.GetContext()
		h := engine.GetExecutingScriptHash()
		amount := totalSupply
		storage.Put(ctx, h, amount)
		runtime.Notify("transfer", []byte{}, h, amount)
		return true
	case "migrate":
		script := args[0].([]byte)
		params := []byte{0x07, 0x10} // string + array
		description := args[1].(string)
		email := args[2].(string)
		author := args[3].(string)
		version := args[4].(string)
		name := args[5].(string)
		contract.Migrate(script, params, 0x05, 0x01, name, version, author, email, description)
		return true
	default:
		panic("invalid operation")
	}
}
