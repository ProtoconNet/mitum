package process

import (
	"context"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/prprocessor"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/launch/config"
	"github.com/spikeekips/mitum/launch/pm"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/states"
	basicstate "github.com/spikeekips/mitum/states/basic"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/storage/blockdata"
	"github.com/spikeekips/mitum/util/logging"
)

const ProcessNameConsensusStates = "consensus_states"

var ProcessorConsensusStates pm.Process

func init() {
	if i, err := pm.NewProcess(
		ProcessNameConsensusStates,
		[]string{
			ProcessNameLocalNode,
			ProcessNameDatabase,
			ProcessNameBlockData,
			ProcessNameSuffrage,
			ProcessNameProposalProcessor,
		},
		ProcessConsensusStates,
	); err != nil {
		panic(err)
	} else {
		ProcessorConsensusStates = i
	}
}

func ProcessConsensusStates(ctx context.Context) (context.Context, error) {
	var policy *isaac.LocalPolicy
	if err := LoadPolicyContextValue(ctx, &policy); err != nil {
		return ctx, err
	}

	var nodepool *network.Nodepool
	if err := LoadNodepoolContextValue(ctx, &nodepool); err != nil {
		return ctx, err
	}

	var st storage.Database
	if err := LoadDatabaseContextValue(ctx, &st); err != nil {
		return ctx, err
	}

	var blockData blockdata.BlockData
	if err := LoadBlockDataContextValue(ctx, &blockData); err != nil {
		return ctx, err
	}

	var pps *prprocessor.Processors
	if err := LoadProposalProcessorContextValue(ctx, &pps); err != nil {
		return ctx, err
	}

	var suffrage base.Suffrage
	if err := LoadSuffrageContextValue(ctx, &suffrage); err != nil {
		return ctx, err
	}

	var log logging.Logger
	if err := config.LoadLogContextValue(ctx, &log); err != nil {
		return ctx, err
	}

	if cs, err := processConsensusStates(st, blockData, policy, nodepool, pps, suffrage, log); err != nil {
		return ctx, err
	} else {
		if i, ok := cs.(logging.SetLogger); ok {
			_ = i.SetLogger(log)
		}

		return context.WithValue(ctx, ContextValueConsensusStates, cs), nil
	}
}

func processConsensusStates(
	st storage.Database,
	blockData blockdata.BlockData,
	policy *isaac.LocalPolicy,
	nodepool *network.Nodepool,
	pps *prprocessor.Processors,
	suffrage base.Suffrage,
	log logging.Logger,
) (states.States, error) {
	ballotbox := isaac.NewBallotbox(
		suffrage.Nodes,
		func() base.Threshold {
			if t, err := base.NewThreshold(
				uint(len(suffrage.Nodes())),
				policy.ThresholdRatio(),
			); err != nil {
				panic(err)
			} else {
				return t
			}
		},
	)
	_ = ballotbox.SetLogger(log)

	proposalMaker := isaac.NewProposalMaker(nodepool.Local(), st, policy)

	stopped := basicstate.NewStoppedState()
	booting := basicstate.NewBootingState(st, blockData, policy, suffrage)
	joining := basicstate.NewJoiningState(nodepool.Local(), st, policy, suffrage, ballotbox)
	consensus := basicstate.NewConsensusState(st, policy, nodepool, suffrage, proposalMaker, pps)
	syncing := basicstate.NewSyncingState(st, blockData, policy, nodepool)

	return basicstate.NewStates(
		st,
		policy,
		nodepool,
		suffrage,
		ballotbox,
		stopped,
		booting,
		joining,
		consensus,
		syncing,
	)
}
