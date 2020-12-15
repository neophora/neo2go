package dao

import (
	"testing"

	"github.com/neophora/neo2go/pkg/core/block"
	"github.com/neophora/neo2go/pkg/core/state"
	"github.com/neophora/neo2go/pkg/core/storage"
	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/internal/random"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/smartcontract"
	"github.com/neophora/neo2go/pkg/vm/opcode"
	"github.com/stretchr/testify/require"
)

func TestPutGetAndDecode(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	serializable := &TestSerializable{field: random.String(4)}
	hash := []byte{1}
	err := dao.Put(serializable, hash)
	require.NoError(t, err)

	gotAndDecoded := &TestSerializable{}
	err = dao.GetAndDecode(gotAndDecoded, hash)
	require.NoError(t, err)
}

// TestSerializable structure used in testing.
type TestSerializable struct {
	field string
}

func (t *TestSerializable) EncodeBinary(writer *io.BinWriter) {
	writer.WriteString(t.field)
}

func (t *TestSerializable) DecodeBinary(reader *io.BinReader) {
	t.field = reader.ReadString()
}

func TestGetAccountStateOrNew_New(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint160()
	createdAccount, err := dao.GetAccountStateOrNew(hash)
	require.NoError(t, err)
	require.NotNil(t, createdAccount)
}

func TestPutAndGetAccountStateOrNew(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint160()
	accountState := &state.Account{ScriptHash: hash}
	err := dao.PutAccountState(accountState)
	require.NoError(t, err)
	gotAccount, err := dao.GetAccountStateOrNew(hash)
	require.NoError(t, err)
	require.Equal(t, accountState.ScriptHash, gotAccount.ScriptHash)
}

func TestPutAndGetAssetState(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	id := random.Uint256()
	assetState := &state.Asset{ID: id, Owner: keys.PublicKey{}}
	err := dao.PutAssetState(assetState)
	require.NoError(t, err)
	gotAssetState, err := dao.GetAssetState(id)
	require.NoError(t, err)
	require.Equal(t, assetState, gotAssetState)
}

func TestPutAndGetContractState(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	contractState := &state.Contract{Script: []byte{}, ParamList: []smartcontract.ParamType{}}
	hash := contractState.ScriptHash()
	err := dao.PutContractState(contractState)
	require.NoError(t, err)
	gotContractState, err := dao.GetContractState(hash)
	require.NoError(t, err)
	require.Equal(t, contractState, gotContractState)
}

func TestDeleteContractState(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	contractState := &state.Contract{Script: []byte{}, ParamList: []smartcontract.ParamType{}}
	hash := contractState.ScriptHash()
	err := dao.PutContractState(contractState)
	require.NoError(t, err)
	err = dao.DeleteContractState(hash)
	require.NoError(t, err)
	gotContractState, err := dao.GetContractState(hash)
	require.Error(t, err)
	require.Nil(t, gotContractState)
}

func TestGetUnspentCoinState_Err(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint256()
	gotUnspentCoinState, err := dao.GetUnspentCoinState(hash)
	require.Error(t, err)
	require.Nil(t, gotUnspentCoinState)
}

func TestPutGetUnspentCoinState(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint256()
	unspentCoinState := &state.UnspentCoin{Height: 42, States: []state.OutputState{}}
	err := dao.PutUnspentCoinState(hash, unspentCoinState)
	require.NoError(t, err)
	gotUnspentCoinState, err := dao.GetUnspentCoinState(hash)
	require.NoError(t, err)
	require.Equal(t, unspentCoinState, gotUnspentCoinState)
}

func TestGetValidatorStateOrNew_New(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	publicKey := &keys.PublicKey{}
	validatorState, err := dao.GetValidatorStateOrNew(publicKey)
	require.NoError(t, err)
	require.NotNil(t, validatorState)
}

func TestPutGetValidatorState(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	publicKey := &keys.PublicKey{}
	validatorState := &state.Validator{
		PublicKey:  publicKey,
		Registered: false,
		Votes:      0,
	}
	err := dao.PutValidatorState(validatorState)
	require.NoError(t, err)
	gotValidatorState, err := dao.GetValidatorState(publicKey)
	require.NoError(t, err)
	require.Equal(t, validatorState, gotValidatorState)
}

func TestDeleteValidatorState(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	publicKey := &keys.PublicKey{}
	validatorState := &state.Validator{
		PublicKey:  publicKey,
		Registered: false,
		Votes:      0,
	}
	err := dao.PutValidatorState(validatorState)
	require.NoError(t, err)
	err = dao.DeleteValidatorState(validatorState)
	require.NoError(t, err)
	gotValidatorState, err := dao.GetValidatorState(publicKey)
	require.Error(t, err)
	require.Nil(t, gotValidatorState)
}

func TestGetValidators(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	publicKey := &keys.PublicKey{}
	validatorState := &state.Validator{
		PublicKey:  publicKey,
		Registered: false,
		Votes:      0,
	}
	err := dao.PutValidatorState(validatorState)
	require.NoError(t, err)
	validators := dao.GetValidators()
	require.Equal(t, validatorState, validators[0])
	require.Len(t, validators, 1)
}

func TestPutGetAppExecResult(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint256()
	appExecResult := &state.AppExecResult{
		TxHash: hash,
		Events: []state.NotificationEvent{},
		Stack:  []smartcontract.Parameter{},
	}
	err := dao.PutAppExecResult(appExecResult)
	require.NoError(t, err)
	gotAppExecResult, err := dao.GetAppExecResult(hash)
	require.NoError(t, err)
	require.Equal(t, appExecResult, gotAppExecResult)
}

func TestPutGetStorageItem(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint160()
	key := []byte{0}
	storageItem := &state.StorageItem{Value: []uint8{}}
	err := dao.PutStorageItem(hash, key, storageItem)
	require.NoError(t, err)
	gotStorageItem := dao.GetStorageItem(hash, key)
	require.Equal(t, storageItem, gotStorageItem)
}

func TestDeleteStorageItem(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint160()
	key := []byte{0}
	storageItem := &state.StorageItem{Value: []uint8{}}
	err := dao.PutStorageItem(hash, key, storageItem)
	require.NoError(t, err)
	err = dao.DeleteStorageItem(hash, key)
	require.NoError(t, err)
	gotStorageItem := dao.GetStorageItem(hash, key)
	require.Nil(t, gotStorageItem)
}

func TestGetBlock_NotExists(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	hash := random.Uint256()
	block, _, err := dao.GetBlock(hash)
	require.Error(t, err)
	require.Nil(t, block)
}

func TestPutGetBlock(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	b := &block.Block{
		Base: block.Base{
			Script: transaction.Witness{
				VerificationScript: []byte{byte(opcode.PUSH1)},
				InvocationScript:   []byte{byte(opcode.NOP)},
			},
		},
	}
	hash := b.Hash()
	err := dao.StoreAsBlock(b, 42)
	require.NoError(t, err)
	gotBlock, sysfee, err := dao.GetBlock(hash)
	require.NoError(t, err)
	require.NotNil(t, gotBlock)
	require.EqualValues(t, 42, sysfee)
}

func TestGetVersion_NoVersion(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	version, err := dao.GetVersion()
	require.Error(t, err)
	require.Equal(t, "", version)
}

func TestGetVersion(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	err := dao.PutVersion("testVersion")
	require.NoError(t, err)
	version, err := dao.GetVersion()
	require.NoError(t, err)
	require.NotNil(t, version)
}

func TestGetCurrentHeaderHeight_NoHeader(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	height, err := dao.GetCurrentBlockHeight()
	require.Error(t, err)
	require.Equal(t, uint32(0), height)
}

func TestGetCurrentHeaderHeight_Store(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	b := &block.Block{
		Base: block.Base{
			Script: transaction.Witness{
				VerificationScript: []byte{byte(opcode.PUSH1)},
				InvocationScript:   []byte{byte(opcode.NOP)},
			},
		},
	}
	err := dao.StoreAsCurrentBlock(b)
	require.NoError(t, err)
	height, err := dao.GetCurrentBlockHeight()
	require.NoError(t, err)
	require.Equal(t, uint32(0), height)
}

func TestStoreAsTransaction(t *testing.T) {
	dao := NewSimple(storage.NewMemoryStore())
	tx := &transaction.Transaction{Type: transaction.IssueType, Data: &transaction.IssueTX{}}
	hash := tx.Hash()
	err := dao.StoreAsTransaction(tx, 0)
	require.NoError(t, err)
	hasTransaction := dao.HasTransaction(hash)
	require.True(t, hasTransaction)
}
