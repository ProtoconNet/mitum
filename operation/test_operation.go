// +build test

package operation

import (
	"encoding/json"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/encoder"
	"github.com/spikeekips/mitum/hint"
	"github.com/spikeekips/mitum/key"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/valuehash"
)

var (
	MaxKeyKVOperation   int = 100
	MaxValueKVOperation int = 100
)

type KVOperationFact struct {
	signer key.Publickey
	token  []byte
	Key    string
	Value  []byte
}

func (kvof KVOperationFact) IsValid(b []byte) error {
	if kvof.signer == nil {
		return xerrors.Errorf("fact has empty signer")
	}
	if err := kvof.Hint().IsValid(b); err != nil {
		return err
	}

	if l := len(kvof.Key); l < 1 {
		return xerrors.Errorf("empty Key of KVOperation")
	} else if l > MaxKeyKVOperation {
		return xerrors.Errorf("Key of KVOperation over limit; %d > %d", l, MaxKeyKVOperation)
	}

	if kvof.Value != nil {
		if l := len(kvof.Value); l > MaxValueKVOperation {
			return xerrors.Errorf("Value of KVOperation over limit; %d > %d", l, MaxValueKVOperation)
		}
	}

	return nil
}

func (kvof KVOperationFact) Hint() hint.Hint {
	return hint.MustHint(hint.Type{0xff, 0xf9}, "0.0.1")
}

func (kvof KVOperationFact) Hash() valuehash.Hash {
	return valuehash.NewSHA256(kvof.Bytes())
}

func (kvof KVOperationFact) Bytes() []byte {
	return util.ConcatSlice([][]byte{
		[]byte(kvof.signer.String()),
		kvof.token,
		[]byte(kvof.Key),
		kvof.Value,
	})
}

func (kvof KVOperationFact) Signer() key.Publickey {
	return kvof.signer
}

func (kvof KVOperationFact) Token() []byte {
	return kvof.token
}

type KVOperation struct {
	KVOperationFact
	h             valuehash.Hash
	factHash      valuehash.Hash
	factSignature key.Signature
}

func NewKVOperation(
	signer key.Privatekey,
	token []byte,
	k string,
	v []byte,
	b []byte,
) (KVOperation, error) {
	if signer == nil {
		return KVOperation{}, xerrors.Errorf("empty privatekey")
	}

	fact := KVOperationFact{
		signer: signer.Publickey(),
		token:  token,
		Key:    k,
		Value:  v,
	}
	factHash := fact.Hash()
	var factSignature key.Signature
	if fs, err := signer.Sign(util.ConcatSlice([][]byte{factHash.Bytes(), b})); err != nil {
		return KVOperation{}, err
	} else {
		factSignature = fs
	}

	kvo := KVOperation{
		KVOperationFact: fact,
		factHash:        factHash,
		factSignature:   factSignature,
	}

	if h, err := kvo.GenerateHash(b); err != nil {
		return KVOperation{}, err
	} else {
		kvo.h = h
	}

	return kvo, nil
}

func (kvo KVOperation) IsValid(b []byte) error {
	if err := IsValidOperation(kvo, b); err != nil {
		return err
	}

	return nil
}

func (kvo KVOperation) Hint() hint.Hint {
	return hint.MustHint(hint.Type{0xff, 0xfa}, "0.0.1")
}

func (kvo KVOperation) Fact() Fact {
	return kvo.KVOperationFact
}

func (kvo KVOperation) Hash() valuehash.Hash {
	return kvo.h
}

func (kvo KVOperation) GenerateHash(b []byte) (valuehash.Hash, error) {
	e := util.ConcatSlice([][]byte{
		kvo.factHash.Bytes(),
		kvo.factSignature.Bytes(),
		b,
	})

	return valuehash.NewSHA256(e), nil
}

func (kvo KVOperation) FactHash() valuehash.Hash {
	return kvo.factHash
}

func (kvo KVOperation) FactSignature() key.Signature {
	return kvo.factSignature
}

func (kvo KVOperation) MarshalJSON() ([]byte, error) {
	return util.JSONMarshal(struct {
		encoder.JSONPackHintedHead
		SG key.Publickey  `json:"signer"`
		TK []byte         `json:"token"`
		K  string         `json:"key"`
		V  []byte         `json:"value"`
		H  valuehash.Hash `json:"hash"`
		FH valuehash.Hash `json:"fact_hash"`
		FS key.Signature  `json:"fact_signature"`
	}{
		JSONPackHintedHead: encoder.NewJSONPackHintedHead(kvo.Hint()),
		SG:                 kvo.signer,
		TK:                 kvo.token,
		K:                  kvo.Key,
		V:                  kvo.Value,
		H:                  kvo.h,
		FH:                 kvo.factHash,
		FS:                 kvo.factSignature,
	})
}

func (kvo *KVOperation) UnpackJSON(b []byte, enc *encoder.JSONEncoder) error {
	var ukvo struct {
		SG json.RawMessage `json:"signer"`
		TK []byte          `json:"token"`
		K  string          `json:"key"`
		V  []byte          `json:"value"`
		H  json.RawMessage `json:"hash"`
		FH json.RawMessage `json:"fact_hash"`
		FS key.Signature   `json:"fact_signature"`
	}

	if err := enc.Unmarshal(b, &ukvo); err != nil {
		return err
	}

	var err error

	var signer key.Publickey
	if signer, err = key.DecodePublickey(enc, ukvo.SG); err != nil {
		return err
	}

	var h, factHash valuehash.Hash
	if h, err = valuehash.Decode(enc, ukvo.H); err != nil {
		return err
	}
	if factHash, err = valuehash.Decode(enc, ukvo.FH); err != nil {
		return err
	}

	kvo.KVOperationFact = KVOperationFact{
		signer: signer,
		token:  ukvo.TK,
		Key:    ukvo.K,
		Value:  ukvo.V,
	}

	kvo.h = h
	kvo.factHash = factHash
	kvo.factSignature = ukvo.FS

	return nil
}