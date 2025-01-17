package basicstates

import (
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/node"
	"github.com/spikeekips/mitum/isaac"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/storage/blockdata"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
)

type BootingState struct {
	*logging.Logging
	*BaseState
	local     node.Local
	database  storage.Database
	blockdata blockdata.Blockdata
	policy    *isaac.LocalPolicy
	suffrage  base.Suffrage
}

func NewBootingState(
	local node.Local,
	db storage.Database,
	bd blockdata.Blockdata,
	policy *isaac.LocalPolicy,
	suffrage base.Suffrage,
) *BootingState {
	return &BootingState{
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "basic-booting-state")
		}),
		BaseState: NewBaseState(base.StateBooting),
		local:     local,
		database:  db,
		blockdata: bd,
		policy:    policy,
		suffrage:  suffrage,
	}
}

func (st *BootingState) Enter(sctx StateSwitchContext) (func() error, error) {
	callback := EmptySwitchFunc
	if i, err := st.BaseState.Enter(sctx); err != nil {
		return nil, err
	} else if i != nil {
		callback = i
	}

	st.resetBallotbox()

	if _, err := storage.CheckBlock(st.database, st.policy.NetworkID()); err != nil {
		st.Log().Error().Err(err).Msg("something wrong to check blocks")

		if !errors.Is(err, util.NotFoundError) {
			return nil, err
		}

		st.Log().Debug().Msg("empty blocks found; cleaning up")
		// NOTE empty block
		if err := blockdata.Clean(st.database, st.blockdata, false); err != nil {
			return nil, err
		}

		return st.enterSyncing(callback)
	}

	if st.suffrage.IsInside(st.local.Address()) {
		return func() error {
			if err := callback(); err != nil {
				return err
			}

			st.Log().Debug().Msg("block checked; moves to joining")

			return st.NewStateSwitchContext(base.StateJoining)
		}, nil
	}
	return func() error {
		if err := callback(); err != nil {
			return err
		}

		st.Log().Debug().Msg("block checked; none-suffrage node moves to syncing")

		return st.NewStateSwitchContext(base.StateSyncing)
	}, nil
}

func (st *BootingState) enterSyncing(callback func() error) (func() error, error) {
	if len(st.syncableChannels()) < 1 {
		st.Log().Debug().Msg("empty blocks; no channels for syncing; can not sync")

		return nil, errors.Errorf("empty blocks, but no channels for syncing; can not sync")
	}

	st.Log().Debug().Msg("empty blocks; will sync")

	return func() error {
		if err := callback(); err != nil {
			return err
		}

		return st.NewStateSwitchContext(base.StateSyncing)
	}, nil
}
