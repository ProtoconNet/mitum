package isaac

import (
	"testing"

	"github.com/stretchr/testify/suite"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/ballot"
)

type testConsensusStates struct {
	baseTestStateHandler
}

type dummySyncingStateHandler struct {
	*StateSyncingHandler
}

func (ss *dummySyncingStateHandler) Activate(_ StateChangeContext) error {
	return nil
}

func (ss *dummySyncingStateHandler) NewVoteproof(_ base.Voteproof) error {
	return nil
}

func (t *testConsensusStates) TestINITVoteproofHigherHeight() {
	thr, _ := base.NewThreshold(2, 67)
	_ = t.localstate.Policy().SetThreshold(thr)
	_ = t.remoteState.Policy().SetThreshold(thr)

	cs, err := NewStateSyncingHandler(t.localstate, nil)
	t.NoError(err)

	css := NewConsensusStates(t.localstate, nil, nil, nil, nil, nil, &dummySyncingStateHandler{cs}, nil)
	t.NotNil(css)

	manifest := t.lastManifest(t.localstate.Storage())
	initFact := ballot.NewINITBallotV0(
		t.localstate.Node().Address(),
		manifest.Height()+3,
		base.Round(2), // round is not important to go
		manifest.Hash(),
		nil,
	).Fact()

	vp, err := t.newVoteproof(base.StageINIT, initFact, t.localstate, t.remoteState)
	t.NoError(err)

	t.NoError(css.newVoteproof(vp))

	t.NotNil(css.ActiveHandler())
	t.Equal(base.StateSyncing, css.ActiveHandler().State())
}

func TestConsensusStates(t *testing.T) {
	suite.Run(t, new(testConsensusStates))
}
