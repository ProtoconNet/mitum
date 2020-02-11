package isaac

import (
	"encoding/json"

	"github.com/spikeekips/mitum/encoder"
	"github.com/spikeekips/mitum/errors"
	"github.com/spikeekips/mitum/key"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/valuehash"
	"golang.org/x/xerrors"
)

type VoteProofV0PackJSON struct {
	encoder.JSONPackHintedHead
	HT Height              `json:"height"`
	RD Round               `json:"round"`
	TH Threshold           `json:"threshold"`
	RS VoteProofResultType `json:"result"`
	ST Stage               `json:"stage"`
	MJ Fact                `json:"majority"`
	FS [][2]interface{}    `json:"facts"`
	BS [][2]interface{}    `json:"ballots"`
	VS [][2]interface{}    `json:"votes"`
}

func (vp VoteProofV0) MarshalJSON() ([]byte, error) {
	var facts [][2]interface{} // nolint
	for h, f := range vp.facts {
		facts = append(facts, [2]interface{}{h, f})
	}

	var ballots [][2]interface{} // nolint
	for a, h := range vp.ballots {
		ballots = append(ballots, [2]interface{}{a, h})
	}

	var votes [][2]interface{} // nolint
	for a := range vp.votes {
		votes = append(votes, [2]interface{}{a, vp.votes[a]})
	}

	return util.JSONMarshal(VoteProofV0PackJSON{
		JSONPackHintedHead: encoder.NewJSONPackHintedHead(vp.Hint()),
		HT:                 vp.height,
		RD:                 vp.round,
		TH:                 vp.threshold,
		RS:                 vp.result,
		ST:                 vp.stage,
		MJ:                 vp.majority,
		FS:                 facts,
		BS:                 ballots,
		VS:                 votes,
	})
}

type VoteProofV0UnpackJSON struct {
	HT Height               `json:"height"`
	RD Round                `json:"round"`
	TH Threshold            `json:"threshold"`
	RS VoteProofResultType  `json:"result"`
	ST Stage                `json:"stage"`
	MJ json.RawMessage      `json:"majority"`
	FS [][2]json.RawMessage `json:"facts"`
	BS [][2]json.RawMessage `json:"ballots"`
	VS [][2]json.RawMessage `json:"votes"`
}

func (vp *VoteProofV0) UnpackJSON(b []byte, enc *encoder.JSONEncoder) error { // nolint
	var vpp VoteProofV0UnpackJSON
	if err := enc.Unmarshal(b, &vpp); err != nil {
		return err
	}

	var err error
	var majority Fact
	if majority, err = decodeFact(enc, vpp.MJ); err != nil {
		return err
	}

	facts := map[valuehash.Hash]Fact{}
	for i := range vpp.FS {
		l := vpp.FS[i]
		if len(l) != 2 {
			return xerrors.Errorf("invalid raw of facts; not [2]json.RawMessage")
		}

		var factHash valuehash.Hash
		if factHash, err = decodeHash(enc, l[0]); err != nil {
			return err
		}

		var fact Fact
		if fact, err = decodeFact(enc, l[1]); err != nil {
			return err
		}

		facts[factHash] = fact
	}

	ballots := map[Address]valuehash.Hash{}
	for i := range vpp.BS {
		l := vpp.BS[i]
		if len(l) != 2 {
			return xerrors.Errorf("invalid raw of ballots; not [2]json.RawMessage")
		}

		var address Address
		if address, err = decodeAddress(enc, l[0]); err != nil {
			return err
		}

		var ballot valuehash.Hash
		if ballot, err = decodeHash(enc, l[1]); err != nil {
			return err
		}

		ballots[address] = ballot
	}

	votes := map[Address]VoteProofNodeFact{}
	for i := range vpp.VS {
		l := vpp.VS[i]
		if len(l) != 2 {
			return xerrors.Errorf("invalid raw of votes; not [2]json.RawMessage")
		}

		var address Address
		if address, err = decodeAddress(enc, l[0]); err != nil {
			return err
		}

		var nodeFact VoteProofNodeFact
		if err = enc.Decode(l[1], &nodeFact); err != nil {
			return err
		}

		votes[address] = nodeFact
	}

	vp.height = vpp.HT
	vp.round = vpp.RD
	vp.threshold = vpp.TH
	vp.result = vpp.RS
	vp.stage = vpp.ST
	vp.majority = majority
	vp.facts = facts
	vp.ballots = ballots
	vp.votes = votes

	return nil
}

type VoteProofNodeFactPackJSON struct {
	FC valuehash.Hash `json:"fact"`
	FS key.Signature  `json:"fact_signature"`
	SG key.Publickey  `json:"signer"`
}

func (vf VoteProofNodeFact) MarshalJSON() ([]byte, error) {
	return util.JSONMarshal(VoteProofNodeFactPackJSON{
		FC: vf.fact,
		FS: vf.factSignature,
		SG: vf.signer,
	})
}

type VoteProofNodeFactUnpackJSON struct {
	FC json.RawMessage `json:"fact"`
	FS key.Signature   `json:"fact_signature"`
	SG json.RawMessage `json:"signer"`
}

func (vf *VoteProofNodeFact) UnpackJSON(b []byte, enc *encoder.JSONEncoder) error {
	var vpp VoteProofNodeFactUnpackJSON
	if err := enc.Unmarshal(b, &vpp); err != nil {
		return err
	}

	var fact valuehash.Hash
	if h, err := decodeHash(enc, vpp.FC); err != nil {
		return err
	} else {
		fact = h
	}

	var signer key.Publickey
	if h, err := decodePublickey(enc, vpp.SG); err != nil {
		return err
	} else {
		signer = h
	}

	vf.fact = fact
	vf.factSignature = vpp.FS
	vf.signer = signer

	return nil
}

func decodeFact(enc *encoder.JSONEncoder, b []byte) (Fact, error) {
	if i, err := enc.DecodeByHint(b); err != nil {
		return nil, err
	} else if v, ok := i.(Fact); !ok {
		return nil, errors.InvalidTypeError.Wrapf("not Fact; type=%T", i)
	} else {
		return v, nil
	}
}
