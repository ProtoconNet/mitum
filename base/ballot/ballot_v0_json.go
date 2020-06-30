package ballot

import (
	"encoding/json"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/key"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/spikeekips/mitum/util/hint"
	"github.com/spikeekips/mitum/util/localtime"
	"github.com/spikeekips/mitum/util/valuehash"
)

type BaseBallotV0PackerJSON struct {
	jsonenc.HintedHead
	H   valuehash.Hash     `json:"hash"`
	SN  key.Publickey      `json:"signer"`
	SG  key.Signature      `json:"signature"`
	SA  localtime.JSONTime `json:"signed_at"`
	HT  base.Height        `json:"height"`
	RD  base.Round         `json:"round"`
	N   base.Address       `json:"node"`
	BH  valuehash.Hash     `json:"body_hash"`
	FH  valuehash.Hash     `json:"fact_hash"`
	FSG key.Signature      `json:"fact_signature"`
}

func PackBaseBallotV0JSON(ballot Ballot) (BaseBallotV0PackerJSON, error) {
	return BaseBallotV0PackerJSON{
		HintedHead: jsonenc.NewHintedHead(ballot.Hint()),
		H:          ballot.Hash(),
		SN:         ballot.Signer(),
		SG:         ballot.Signature(),
		SA:         localtime.NewJSONTime(ballot.SignedAt()),
		HT:         ballot.Height(),
		RD:         ballot.Round(),
		N:          ballot.Node(),
		BH:         ballot.BodyHash(),
		FH:         ballot.FactHash(),
		FSG:        ballot.FactSignature(),
	}, nil
}

type BaseBallotV0UnpackerJSON struct {
	jsonenc.HintedHead
	H   valuehash.Bytes    `json:"hash"`
	SN  json.RawMessage    `json:"signer"`
	SG  key.Signature      `json:"signature"`
	SA  localtime.JSONTime `json:"signed_at"`
	HT  base.Height        `json:"height"`
	RD  base.Round         `json:"round"`
	N   json.RawMessage    `json:"node"`
	BH  valuehash.Bytes    `json:"body_hash"`
	FH  valuehash.Bytes    `json:"fact_hash"`
	FSG key.Signature      `json:"fact_signature"`
}

func UnpackBaseBallotV0JSON(nib BaseBallotV0UnpackerJSON, enc *jsonenc.Encoder) (
	BaseBallotV0,
	BaseBallotFactV0,
	error,
) {
	var err error

	// signer
	var signer key.Publickey
	if signer, err = key.DecodePublickey(enc, nib.SN); err != nil {
		return BaseBallotV0{}, BaseBallotFactV0{}, err
	}

	var node base.Address
	if node, err = base.DecodeAddress(enc, nib.N); err != nil {
		return BaseBallotV0{}, BaseBallotFactV0{}, err
	}

	var h, bh, fh valuehash.Hash
	if !nib.H.Empty() {
		h = nib.H
	}
	if !nib.BH.Empty() {
		bh = nib.BH
	}
	if !nib.FH.Empty() {
		fh = nib.FH
	}

	return BaseBallotV0{
			h:             h,
			bodyHash:      bh,
			signer:        signer,
			signature:     nib.SG,
			signedAt:      nib.SA.Time,
			node:          node,
			factHash:      fh,
			factSignature: nib.FSG,
		},
		BaseBallotFactV0{
			height: nib.HT,
			round:  nib.RD,
		}, nil
}

type BaseBallotFactV0PackerJSON struct {
	jsonenc.HintedHead
	HT base.Height `json:"height"`
	RD base.Round  `json:"round"`
}

func NewBaseBallotFactV0PackerJSON(bbf BaseBallotFactV0, ht hint.Hint) BaseBallotFactV0PackerJSON {
	return BaseBallotFactV0PackerJSON{
		HintedHead: jsonenc.NewHintedHead(ht),
		HT:         bbf.height,
		RD:         bbf.round,
	}
}
