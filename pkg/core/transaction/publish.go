package transaction

import (
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/smartcontract"
	"github.com/neophora/neo2go/pkg/util"
)

// PublishTX represents a publish transaction.
// NOTE: This is deprecated and should no longer be used.
type PublishTX struct {
	Script      []byte
	ParamList   []smartcontract.ParamType
	ReturnType  smartcontract.ParamType
	NeedStorage bool
	Name        string
	CodeVersion string
	Author      string
	Email       string
	Description string
	Version     uint8 // Version of the parent struct Transaction. Used in reading NeedStorage flag.
}

// DecodeBinary implements Serializable interface.
func (tx *PublishTX) DecodeBinary(br *io.BinReader) {
	tx.Script = br.ReadVarBytes()

	lenParams := br.ReadVarUint()
	tx.ParamList = make([]smartcontract.ParamType, lenParams)
	for i := 0; i < int(lenParams); i++ {
		tx.ParamList[i] = smartcontract.ParamType(br.ReadB())
	}

	tx.ReturnType = smartcontract.ParamType(br.ReadB())

	if tx.Version >= 1 {
		tx.NeedStorage = br.ReadBool()
	} else {
		tx.NeedStorage = false
	}

	tx.Name = br.ReadString()
	tx.CodeVersion = br.ReadString()
	tx.Author = br.ReadString()
	tx.Email = br.ReadString()
	tx.Description = br.ReadString()
}

// EncodeBinary implements Serializable interface.
func (tx *PublishTX) EncodeBinary(bw *io.BinWriter) {
	bw.WriteVarBytes(tx.Script)
	bw.WriteVarUint(uint64(len(tx.ParamList)))
	for _, param := range tx.ParamList {
		bw.WriteB(byte(param))
	}
	bw.WriteB(byte(tx.ReturnType))
	if tx.Version >= 1 {
		bw.WriteBool(tx.NeedStorage)
	}
	bw.WriteString(tx.Name)
	bw.WriteString(tx.CodeVersion)
	bw.WriteString(tx.Author)
	bw.WriteString(tx.Email)
	bw.WriteString(tx.Description)
}

// publishedContract is a JSON wrapper for PublishTransaction
type publishedContract struct {
	Code        publishedCode `json:"code"`
	NeedStorage bool          `json:"needstorage,omitempty"`
	Name        string        `json:"name,omitempty"`
	CodeVersion string        `json:"version,omitempty"`
	Author      string        `json:"author,omitempty"`
	Email       string        `json:"email,omitempty"`
	Description string        `json:"description,omitempty"`
}

// publishedCode is a JSON wrapper for PublishTransaction Code
type publishedCode struct {
	Hash       util.Uint160              `json:"hash,omitempty"`
	Script     string                    `json:"script,omitempty"`
	ParamList  []smartcontract.ParamType `json:"parameters,omitempty"`
	ReturnType smartcontract.ParamType   `json:"returntype,omitempty"`
}
