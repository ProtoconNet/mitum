package isaac

import (
	"time"

	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/hint"
	"github.com/spikeekips/mitum/isvalid"
	"github.com/spikeekips/mitum/operation"
	"github.com/spikeekips/mitum/valuehash"
)

var VoteproofV0Hint hint.Hint = hint.MustHint(VoteproofType, "0.0.1")

type VoteproofV0 struct {
	height     Height
	round      Round
	threshold  Threshold
	result     VoteproofResultType
	closed     bool
	stage      Stage
	majority   operation.Fact
	facts      map[valuehash.Hash]operation.Fact // key: Fact.Hash(), value: Fact
	ballots    map[Address]valuehash.Hash        // key: node Address, value: ballot hash
	votes      map[Address]VoteproofNodeFact     // key: node Address, value: VoteproofNodeFact
	finishedAt time.Time
}

func (vp VoteproofV0) Hint() hint.Hint {
	return VoteproofV0Hint
}

func (vp VoteproofV0) IsFinished() bool {
	return vp.result != VoteproofNotYet
}

func (vp VoteproofV0) FinishedAt() time.Time {
	return vp.finishedAt
}

func (vp VoteproofV0) IsClosed() bool {
	return vp.closed
}

func (vp VoteproofV0) Height() Height {
	return vp.height
}

func (vp VoteproofV0) Round() Round {
	return vp.round
}

func (vp VoteproofV0) Stage() Stage {
	return vp.stage
}

func (vp VoteproofV0) Result() VoteproofResultType {
	return vp.result
}

func (vp VoteproofV0) Majority() operation.Fact {
	return vp.majority
}

func (vp VoteproofV0) Ballots() map[Address]valuehash.Hash {
	return vp.ballots
}

func (vp VoteproofV0) Bytes() []byte {
	// TODO returns proper bytes
	return nil
}

func (vp VoteproofV0) IsValid(b []byte) error {
	if err := vp.isValidFields(b); err != nil {
		return err
	}

	if err := vp.isValidFacts(b); err != nil {
		return err
	}

	// check majority
	if len(vp.votes) < int(vp.threshold.Threshold) {
		if vp.result != VoteproofNotYet {
			return xerrors.Errorf("result should be not-yet: %s", vp.result)
		}

		return nil
	}

	return vp.isValidCheckMajority()
}

func (vp VoteproofV0) isValidCheckMajority() error {
	counts := map[valuehash.Hash]uint{}
	for _, f := range vp.votes { // nolint
		counts[f.fact]++
	}

	set := make([]uint, len(counts))
	byCount := map[uint]valuehash.Hash{}

	var index int
	for h, c := range counts {
		set[index] = c
		index++
		byCount[c] = h
	}

	var fact operation.Fact
	var factHash valuehash.Hash
	var result VoteproofResultType
	switch index := FindMajority(vp.threshold.Total, vp.threshold.Threshold, set...); index {
	case -1:
		result = VoteproofNotYet
	case -2:
		result = VoteproofDraw
	default:
		result = VoteproofMajority
		factHash = byCount[set[index]]
		fact = vp.facts[factHash]
	}

	if vp.result != result {
		return xerrors.Errorf("result mismatch; vp.result=%s != result=%s", vp.result, result)
	}

	if fact == nil {
		if vp.majority != nil {
			return xerrors.Errorf("result should be nil, but not")
		}
	} else {
		mhash := vp.majority.Hash()
		if !mhash.Equal(factHash) {
			return xerrors.Errorf("fact hash mismatch; vp.majority=%s != fact=%s", mhash, factHash)
		}
	}

	return nil
}

func (vp VoteproofV0) isValidFields(b []byte) error {
	if err := isvalid.Check([]isvalid.IsValider{
		vp.height,
		vp.stage,
		vp.threshold,
		vp.result,
	}, b, false); err != nil {
		return err
	}
	if vp.finishedAt.IsZero() {
		return isvalid.InvalidError.Wrapf("empty finishedAt")
	}

	if vp.result != VoteproofMajority && vp.result != VoteproofDraw {
		return isvalid.InvalidError.Wrapf("invalid result; result=%v", vp.result)
	}

	if vp.majority == nil {
		if vp.result != VoteproofDraw {
			return isvalid.InvalidError.Wrapf("empty majority, but result is not draw; result=%v", vp.result)
		}
	} else if err := vp.majority.IsValid(b); err != nil {
		return err
	}

	if len(vp.facts) < 1 {
		return isvalid.InvalidError.Wrapf("empty facts")
	}

	if len(vp.ballots) < 1 {
		return isvalid.InvalidError.Wrapf("empty ballots")
	}

	if len(vp.votes) < 1 {
		return isvalid.InvalidError.Wrapf("empty votes")
	}

	if len(vp.ballots) != len(vp.votes) {
		return isvalid.InvalidError.Wrapf("vote count does not match: ballots=%d votes=%d", len(vp.ballots), len(vp.votes))
	}

	for k := range vp.ballots {
		if _, found := vp.votes[k]; !found {
			return xerrors.Errorf("unknown node found: %v", k)
		}
	}

	return nil
}

func (vp VoteproofV0) isValidFacts(b []byte) error {
	factHashes := map[valuehash.Hash]bool{}
	for node, f := range vp.votes { // nolint
		if err := node.IsValid(b); err != nil {
			return err
		}

		if _, found := vp.facts[f.fact]; !found {
			return xerrors.Errorf("missing fact found in facts: %s", f.fact.String())
		}
		factHashes[f.fact] = true
	}

	if len(factHashes) != len(vp.facts) {
		return xerrors.Errorf("unknown facts found in facts: %d", len(vp.facts)-len(factHashes))
	}

	for k, v := range vp.facts {
		if err := isvalid.Check([]isvalid.IsValider{k, v}, b, false); err != nil {
			return err
		}
		if h := v.Hash(); !h.Equal(k) {
			return xerrors.Errorf(
				"factHash and Fact.Hash() does not match: factHash=%v != Fact.Hash()=%v",
				k.String(), h.String(),
			)
		}
	}

	for k, v := range vp.ballots {
		if err := isvalid.Check([]isvalid.IsValider{k, v}, b, false); err != nil {
			return err
		}
	}

	return nil
}
