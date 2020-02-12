package common

import (
	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/logging"
)

// FixedSuffrage will be used only for testing.
type FixedSuffrage struct {
	*logging.Logger
	proposer  isaac.Node
	nodes     map[isaac.Address]isaac.Node
	nodeSlice []isaac.Node
}

func NewFixedSuffrage(proposer isaac.Node, nodes []isaac.Node) *FixedSuffrage {
	ns := map[isaac.Address]isaac.Node{}
	for _, n := range nodes {
		ns[n.Address()] = n
	}

	return &FixedSuffrage{
		Logger: logging.NewLogger(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "fixed-suffrage")
		}),
		proposer:  proposer,
		nodes:     ns,
		nodeSlice: nodes,
	}
}

func (ff *FixedSuffrage) Name() string {
	return "fixed-suffrage"
}

func (ff *FixedSuffrage) Acting(height isaac.Height, round isaac.Round) isaac.ActingSuffrage {
	return isaac.NewActingSuffrage(height, round, ff.proposer, ff.nodeSlice)
}

func (ff *FixedSuffrage) IsActing(_ isaac.Height, _ isaac.Round, node isaac.Address) bool {
	_, found := ff.nodes[node]
	return found
}

func (ff *FixedSuffrage) IsProposer(_ isaac.Height, _ isaac.Round, node isaac.Address) bool {
	return ff.proposer.Address().Equal(node)
}