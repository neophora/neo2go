package client_test

import (
	"context"
	"fmt"
	"os"

	"github.com/neophora/neo2go/pkg/rpc/client"
)

func Example() {
	endpoint := "http://seed5.bridgeprotocol.io:10332"
	opts := client.Options{}

	c, err := client.New(context.TODO(), endpoint, opts)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := c.Ping(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	resp, err := c.GetAccountState("ATySFJAbLW7QHsZGHScLhxq6EyNBxx3eFP")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	fmt.Println(resp.ScriptHash)
	fmt.Println(resp.Balances)
}
