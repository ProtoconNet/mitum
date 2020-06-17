package isaac

import (
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/ballot"
	"github.com/spikeekips/mitum/base/block"
	"github.com/spikeekips/mitum/util/errors"
	"github.com/spikeekips/mitum/util/logging"
)

var (
	IgnoreVoteproofError = errors.NewError("Voteproof should be ignored")
	stateToBeChangeError = errors.NewError("State needs to be changed")
)

type StateToBeChangeError struct {
	*errors.NError
	ToState   base.State
	Voteproof base.Voteproof
	Ballot    ballot.Ballot
	Err       error
}

func NewStateToBeChangeError(
	toState base.State,
	voteproof base.Voteproof,
	blt ballot.Ballot,
	err error,
) *StateToBeChangeError {
	return &StateToBeChangeError{
		NError:    stateToBeChangeError,
		ToState:   toState,
		Voteproof: voteproof,
		Ballot:    blt,
		Err:       err,
	}
}

type VoteProofChecker struct {
	*logging.Logging
	voteproof  base.Voteproof
	suffrage   base.Suffrage
	localstate *Localstate
}

// NOTE VoteProofChecker should check the signer of VoteproofNodeFact is valid
// Ballot.Signer(), but it takes a little bit time to gather the Ballots from
// the other node, so this will be ignored at this time for performance reason.

func NewVoteProofChecker(voteproof base.Voteproof, localstate *Localstate, suffrage base.Suffrage) *VoteProofChecker {
	return &VoteProofChecker{
		Logging: logging.NewLogging(func(c logging.Context) logging.Emitter {
			return c.Str("module", "voteproof-checker")
		}),
		voteproof:  voteproof,
		suffrage:   suffrage,
		localstate: localstate,
	}
}

func (vc *VoteProofChecker) CheckIsValid() (bool, error) {
	networkID := vc.localstate.Policy().NetworkID()
	if err := vc.voteproof.IsValid(networkID); err != nil {
		return false, err
	}

	return true, nil
}

func (vc *VoteProofChecker) CheckNodeIsInSuffrage() (bool, error) {
	for n := range vc.voteproof.Ballots() {
		if !vc.suffrage.IsInside(n) {
			vc.Log().Debug().Str("node", n.String()).Msg("voteproof has the vote from unknown node")
			return false, nil
		}
	}

	return true, nil
}

// CheckThreshold checks Threshold only for new incoming Voteproof.
func (vc *VoteProofChecker) CheckThreshold() (bool, error) {
	tr := vc.localstate.Policy().ThresholdRatio()
	if tr != vc.voteproof.ThresholdRatio() {
		vc.Log().Debug().
			Interface("threshold_ratio", vc.voteproof.ThresholdRatio()).
			Interface("expected", tr).
			Msg("voteproof has different threshold ratio")
		return false, nil
	}

	return true, nil
}

type VoteproofConsensusStateChecker struct {
	*logging.Logging
	lastManifest      block.Manifest
	lastINITVoteproof base.Voteproof
	voteproof         base.Voteproof
	css               *ConsensusStates
}

func NewVoteproofConsensusStateChecker(
	lastManifest block.Manifest,
	lastINITVoteproof base.Voteproof,
	voteproof base.Voteproof,
	css *ConsensusStates,
) *VoteproofConsensusStateChecker {
	return &VoteproofConsensusStateChecker{
		Logging: logging.NewLogging(func(c logging.Context) logging.Emitter {
			return c.Str("module", "voteproof-validation-checker")
		}),
		lastManifest:      lastManifest,
		lastINITVoteproof: lastINITVoteproof,
		voteproof:         voteproof,
		css:               css,
	}
}

func (vpc *VoteproofConsensusStateChecker) CheckHeight() (bool, error) {
	l := loggerWithVoteproof(vpc.voteproof, vpc.Log())

	var height base.Height
	if vpc.lastManifest == nil {
		height = base.NilHeight
	} else {
		height = vpc.lastManifest.Height()
	}

	d := vpc.voteproof.Height() - (height + 1)

	if d > 0 {
		l.Debug().
			Hinted("local_block_height", height).
			Msg("Voteproof has higher height from local block")

		return false, NewStateToBeChangeError(
			base.StateSyncing, vpc.voteproof, nil,
			xerrors.Errorf("Voteproof has higher height from local block"),
		)
	}

	if d < 0 {
		l.Debug().
			Hinted("local_block_height", height).
			Msg("Voteproof has lower height from local block; ignore it")

		return false, IgnoreVoteproofError
	}

	return true, nil
}

func (vpc *VoteproofConsensusStateChecker) CheckINITVoteproof() (bool, error) {
	if vpc.voteproof.Stage() != base.StageINIT {
		return true, nil
	}

	l := loggerWithVoteproof(vpc.voteproof, vpc.Log())

	if err := checkBlockWithINITVoteproof(vpc.lastManifest, vpc.voteproof); err != nil {
		l.Error().Err(err).Msg("werid init voteproof found")

		return false, NewStateToBeChangeError(base.StateSyncing, vpc.voteproof, nil, err)
	}

	return true, nil
}

func (vpc *VoteproofConsensusStateChecker) CheckACCEPTVoteproof() (bool, error) {
	if vpc.voteproof.Stage() != base.StageACCEPT {
		return true, nil
	}

	if vpc.lastINITVoteproof.Round() != vpc.voteproof.Round() {
		// BLOCK valid voteproof should be passed without error
		return false, xerrors.Errorf("Voteproof has different round from last init voteproof: voteproof=%d last=%d",
			vpc.voteproof.Round(), vpc.lastINITVoteproof.Round(),
		)
	}

	return true, nil
}
