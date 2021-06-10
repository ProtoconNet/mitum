package basicstates

import (
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/ballot"
)

type EmptyState struct{}

func NewEmptyState() *EmptyState {
	return &EmptyState{}
}

func (*EmptyState) Enter(StateSwitchContext) (func() error, error) {
	return nil, nil
}

func (*EmptyState) Exit(StateSwitchContext) (func() error, error) {
	return nil, nil
}

func (*EmptyState) ProcessProposal(ballot.Proposal) error {
	return nil
}

func (*EmptyState) ProcessVoteproof(base.Voteproof) error {
	return nil
}

func (*EmptyState) SetStates(*States) State {
	return nil
}
