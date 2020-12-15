package address

import (
	"errors"

	"github.com/neophora/neo2go/pkg/encoding/base58"
	"github.com/neophora/neo2go/pkg/util"
)

// Prefix is the byte used to prepend to addresses when encoding them, it can
// be changed and defaults to 23 (0x17), the standard NEO prefix.
var Prefix = byte(0x17)

// Uint160ToString returns the "NEO address" from the given Uint160.
func Uint160ToString(u util.Uint160) string {
	// Dont forget to prepend the Address version 0x17 (23) A
	b := append([]byte{Prefix}, u.BytesBE()...)
	return base58.CheckEncode(b)
}

// StringToUint160 attempts to decode the given NEO address string
// into an Uint160.
func StringToUint160(s string) (u util.Uint160, err error) {
	b, err := base58.CheckDecode(s)
	if err != nil {
		return u, err
	}
	if b[0] != Prefix {
		return u, errors.New("wrong address prefix")
	}
	return util.Uint160DecodeBytesBE(b[1:21])
}
