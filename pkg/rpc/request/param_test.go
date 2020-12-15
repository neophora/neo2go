package request

import (
	"encoding/hex"
	"encoding/json"
	"testing"

	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/encoding/address"
	"github.com/neophora/neo2go/pkg/smartcontract"
	"github.com/neophora/neo2go/pkg/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParam_UnmarshalJSON(t *testing.T) {
	msg := `["str1", 123, null, ["str2", 3], [{"type": "String", "value": "jajaja"}],
                 {"type": "MinerTransaction"},
                 {"contract": "f84d6a337fbc3d3a201d41da99e86b479e7a2554"},
                 {"state": "HALT"}]`
	contr, err := util.Uint160DecodeStringLE("f84d6a337fbc3d3a201d41da99e86b479e7a2554")
	require.NoError(t, err)
	expected := Params{
		{
			Type:  StringT,
			Value: "str1",
		},
		{
			Type:  NumberT,
			Value: 123,
		},
		{
			Type: defaultT,
		},
		{
			Type: ArrayT,
			Value: []Param{
				{
					Type:  StringT,
					Value: "str2",
				},
				{
					Type:  NumberT,
					Value: 3,
				},
			},
		},
		{
			Type: ArrayT,
			Value: []Param{
				{
					Type: FuncParamT,
					Value: FuncParam{
						Type: smartcontract.StringType,
						Value: Param{
							Type:  StringT,
							Value: "jajaja",
						},
					},
				},
			},
		},
		{
			Type:  TxFilterT,
			Value: TxFilter{Type: transaction.MinerType},
		},
		{
			Type:  NotificationFilterT,
			Value: NotificationFilter{Contract: contr},
		},
		{
			Type:  ExecutionFilterT,
			Value: ExecutionFilter{State: "HALT"},
		},
	}

	var ps Params
	require.NoError(t, json.Unmarshal([]byte(msg), &ps))
	require.Equal(t, expected, ps)

	msg = `[{"2": 3}]`
	require.Error(t, json.Unmarshal([]byte(msg), &ps))
}

func TestParamGetString(t *testing.T) {
	p := Param{StringT, "jajaja"}
	str, err := p.GetString()
	assert.Equal(t, "jajaja", str)
	require.Nil(t, err)

	p = Param{StringT, int(100500)}
	_, err = p.GetString()
	require.NotNil(t, err)
}

func TestParamGetInt(t *testing.T) {
	p := Param{NumberT, int(100500)}
	i, err := p.GetInt()
	assert.Equal(t, 100500, i)
	require.Nil(t, err)

	p = Param{NumberT, "jajaja"}
	_, err = p.GetInt()
	require.NotNil(t, err)
}

func TestParamGetArray(t *testing.T) {
	p := Param{ArrayT, []Param{{NumberT, 42}}}
	a, err := p.GetArray()
	assert.Equal(t, []Param{{NumberT, 42}}, a)
	require.Nil(t, err)

	p = Param{ArrayT, 42}
	_, err = p.GetArray()
	require.NotNil(t, err)
}

func TestParamGetUint256(t *testing.T) {
	gas := "602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7"
	u256, _ := util.Uint256DecodeStringLE(gas)
	p := Param{StringT, gas}
	u, err := p.GetUint256()
	assert.Equal(t, u256, u)
	require.Nil(t, err)

	p = Param{StringT, "0x" + gas}
	u, err = p.GetUint256()
	require.NoError(t, err)
	assert.Equal(t, u256, u)

	p = Param{StringT, 42}
	_, err = p.GetUint256()
	require.NotNil(t, err)

	p = Param{StringT, "qq2c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7"}
	_, err = p.GetUint256()
	require.NotNil(t, err)
}

func TestParamGetUint160FromHex(t *testing.T) {
	in := "50befd26fdf6e4d957c11e078b24ebce6291456f"
	u160, _ := util.Uint160DecodeStringLE(in)
	p := Param{StringT, in}
	u, err := p.GetUint160FromHex()
	assert.Equal(t, u160, u)
	require.Nil(t, err)

	p = Param{StringT, 42}
	_, err = p.GetUint160FromHex()
	require.NotNil(t, err)

	p = Param{StringT, "wwbefd26fdf6e4d957c11e078b24ebce6291456f"}
	_, err = p.GetUint160FromHex()
	require.NotNil(t, err)
}

func TestParamGetUint160FromAddress(t *testing.T) {
	in := "AK2nJJpJr6o664CWJKi1QRXjqeic2zRp8y"
	u160, _ := address.StringToUint160(in)
	p := Param{StringT, in}
	u, err := p.GetUint160FromAddress()
	assert.Equal(t, u160, u)
	require.Nil(t, err)

	p = Param{StringT, 42}
	_, err = p.GetUint160FromAddress()
	require.NotNil(t, err)

	p = Param{StringT, "QK2nJJpJr6o664CWJKi1QRXjqeic2zRp8y"}
	_, err = p.GetUint160FromAddress()
	require.NotNil(t, err)
}

func TestParam_GetUint160FromAddressOrHex(t *testing.T) {
	in := "AK2nJJpJr6o664CWJKi1QRXjqeic2zRp8y"
	inHex, _ := address.StringToUint160(in)

	t.Run("Address", func(t *testing.T) {
		p := Param{StringT, in}
		u, err := p.GetUint160FromAddressOrHex()
		require.NoError(t, err)
		require.Equal(t, inHex, u)
	})

	t.Run("Hex", func(t *testing.T) {
		p := Param{StringT, inHex.StringLE()}
		u, err := p.GetUint160FromAddressOrHex()
		require.NoError(t, err)
		require.Equal(t, inHex, u)
	})
}

func TestParam_GetArrayUint160FromHex(t *testing.T) {
	in1 := util.Uint160{1, 2, 3}
	in2 := util.Uint160{4, 5, 6}
	p := Param{Type: ArrayT, Value: []Param{
		{
			Type:  StringT,
			Value: in1.StringLE(),
		},
		{
			Type:  StringT,
			Value: in2.StringLE(),
		},
	}}
	arr, err := p.GetArrayUint160FromHex()
	require.NoError(t, err)
	require.Equal(t, []util.Uint160{in1, in2}, arr)
}

func TestParamGetFuncParam(t *testing.T) {
	fp := FuncParam{
		Type: smartcontract.StringType,
		Value: Param{
			Type:  StringT,
			Value: "jajaja",
		},
	}
	p := Param{
		Type:  FuncParamT,
		Value: fp,
	}
	newfp, err := p.GetFuncParam()
	assert.Equal(t, fp, newfp)
	require.Nil(t, err)

	p = Param{FuncParamT, 42}
	_, err = p.GetFuncParam()
	require.NotNil(t, err)
}

func TestParamGetBytesHex(t *testing.T) {
	in := "602c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7"
	inb, _ := hex.DecodeString(in)
	p := Param{StringT, in}
	bh, err := p.GetBytesHex()
	assert.Equal(t, inb, bh)
	require.Nil(t, err)

	p = Param{StringT, 42}
	_, err = p.GetBytesHex()
	require.NotNil(t, err)

	p = Param{StringT, "qq2c79718b16e442de58778e148d0b1084e3b2dffd5de6b7b16cee7969282de7"}
	_, err = p.GetBytesHex()
	require.NotNil(t, err)
}
