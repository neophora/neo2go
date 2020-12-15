package context

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/neophora/neo2go/pkg/core/transaction"
	"github.com/neophora/neo2go/pkg/crypto/keys"
	"github.com/neophora/neo2go/pkg/io"
	"github.com/neophora/neo2go/pkg/smartcontract"
	"github.com/neophora/neo2go/pkg/util"
	"github.com/neophora/neo2go/pkg/vm"
	"github.com/neophora/neo2go/pkg/vm/emit"
	"github.com/neophora/neo2go/pkg/wallet"
)

// ParameterContext represents smartcontract parameter's context.
type ParameterContext struct {
	// Type is a type of a verifiable item.
	Type string
	// Verifiable is an object which can be (de-)serialized.
	Verifiable io.Serializable
	// Items is a map from script hashes to context items.
	Items map[util.Uint160]*Item
}

type paramContext struct {
	Type  string                     `json:"type"`
	Hex   string                     `json:"hex"`
	Items map[string]json.RawMessage `json:"items"`
}

type sigWithIndex struct {
	index int
	sig   []byte
}

// NewParameterContext returns ParameterContext with the specified type and item to sign.
func NewParameterContext(typ string, verif io.Serializable) *ParameterContext {
	return &ParameterContext{
		Type:       typ,
		Verifiable: verif,
		Items:      make(map[util.Uint160]*Item),
	}
}

// GetWitness returns invocation and verification scripts for the specified contract.
func (c *ParameterContext) GetWitness(ctr *wallet.Contract) (*transaction.Witness, error) {
	item := c.getItemForContract(ctr)
	bw := io.NewBufBinWriter()
	for i := range item.Parameters {
		if item.Parameters[i].Type != smartcontract.SignatureType {
			return nil, errors.New("only signature parameters are supported")
		} else if item.Parameters[i].Value == nil {
			return nil, errors.New("nil parameter")
		}
		emit.Bytes(bw.BinWriter, item.Parameters[i].Value.([]byte))
	}
	return &transaction.Witness{
		InvocationScript:   bw.Bytes(),
		VerificationScript: ctr.Script,
	}, nil
}

// AddSignature adds a signature for the specified contract and public key.
func (c *ParameterContext) AddSignature(ctr *wallet.Contract, pub *keys.PublicKey, sig []byte) error {
	item := c.getItemForContract(ctr)
	if pubs, ok := vm.ParseMultiSigContract(ctr.Script); ok {
		if item.GetSignature(pub) != nil {
			return errors.New("signature is already added")
		}
		pubBytes := pub.Bytes()
		var contained bool
		for i := range pubs {
			if bytes.Equal(pubBytes, pubs[i]) {
				contained = true
				break
			}
		}
		if !contained {
			return errors.New("public key is not present in script")
		}
		item.AddSignature(pub, sig)
		if len(item.Signatures) == len(ctr.Parameters) {
			indexMap := map[string]int{}
			for i := range pubs {
				indexMap[hex.EncodeToString(pubs[i])] = i
			}
			sigs := make([]sigWithIndex, 0, len(item.Signatures))
			for pub, sig := range item.Signatures {
				sigs = append(sigs, sigWithIndex{index: indexMap[pub], sig: sig})
			}
			sort.Slice(sigs, func(i, j int) bool {
				return sigs[i].index < sigs[j].index
			})
			for i := range sigs {
				item.Parameters[i] = smartcontract.Parameter{
					Type:  smartcontract.SignatureType,
					Value: sigs[i].sig,
				}
			}
		}
		return nil
	}

	index := -1
	for i := range ctr.Parameters {
		if ctr.Parameters[i].Type == smartcontract.SignatureType {
			if index >= 0 {
				return errors.New("multiple signature parameters in non-multisig contract")
			}
			index = i
		}
	}
	if index == -1 {
		return errors.New("missing signature parameter")
	}
	item.Parameters[index].Value = sig
	return nil
}

func (c *ParameterContext) getItemForContract(ctr *wallet.Contract) *Item {
	h := ctr.ScriptHash()
	if item, ok := c.Items[h]; ok {
		return item
	}
	params := make([]smartcontract.Parameter, len(ctr.Parameters))
	for i := range params {
		params[i].Type = ctr.Parameters[i].Type
	}
	item := &Item{
		Script:     h,
		Parameters: params,
		Signatures: make(map[string][]byte),
	}
	c.Items[h] = item
	return item
}

// MarshalJSON implements json.Marshaler interface.
func (c ParameterContext) MarshalJSON() ([]byte, error) {
	bw := io.NewBufBinWriter()
	c.Verifiable.EncodeBinary(bw.BinWriter)
	if bw.Err != nil {
		return nil, bw.Err
	}
	items := make(map[string]json.RawMessage, len(c.Items))
	for u := range c.Items {
		data, err := json.Marshal(c.Items[u])
		if err != nil {
			return nil, err
		}
		items["0x"+u.StringBE()] = data
	}
	pc := &paramContext{
		Type:  c.Type,
		Hex:   hex.EncodeToString(bw.Bytes()),
		Items: items,
	}
	return json.Marshal(pc)
}

// UnmarshalJSON implements json.Unmarshaler interface.
func (c *ParameterContext) UnmarshalJSON(data []byte) error {
	pc := new(paramContext)
	if err := json.Unmarshal(data, pc); err != nil {
		return err
	}
	data, err := hex.DecodeString(pc.Hex)
	if err != nil {
		return err
	}

	var verif io.Serializable
	switch pc.Type {
	case "Neo.Core.ContractTransaction":
		verif = new(transaction.Transaction)
	default:
		return fmt.Errorf("unsupported type: %s", c.Type)
	}
	br := io.NewBinReaderFromBuf(data)
	verif.DecodeBinary(br)
	if br.Err != nil {
		return br.Err
	}
	items := make(map[util.Uint160]*Item, len(pc.Items))
	for h := range pc.Items {
		u, err := util.Uint160DecodeStringBE(strings.TrimPrefix(h, "0x"))
		if err != nil {
			return err
		}
		item := new(Item)
		if err := json.Unmarshal(pc.Items[h], item); err != nil {
			return err
		}
		items[u] = item
	}
	c.Type = pc.Type
	c.Verifiable = verif
	c.Items = items
	return nil
}
