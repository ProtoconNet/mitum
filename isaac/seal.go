package isaac

import (
	"context"

	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/operation"
	"github.com/spikeekips/mitum/network"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
	"github.com/spikeekips/mitum/util/valuehash"
)

var KnownSealError = util.NewError("seal is known")

type SealsExtracter struct {
	*logging.Logging
	local    base.Address
	proposer base.Address
	database storage.Database
	nodepool *network.Nodepool
	seals    []valuehash.Hash
	founds   map[string]struct{}
}

func NewSealsExtracter(
	local base.Address,
	proposer base.Address,
	db storage.Database,
	nodepool *network.Nodepool,
	seals []valuehash.Hash,
) *SealsExtracter {
	return &SealsExtracter{
		Logging: logging.NewLogging(func(c zerolog.Context) zerolog.Context {
			return c.Str("module", "seals-extracter").
				Stringer("proposer", proposer).
				Int("seals", len(seals))
		}),
		local:    local,
		database: db,
		nodepool: nodepool,
		proposer: proposer,
		seals:    seals,
		founds:   map[string]struct{}{},
	}
}

func (se *SealsExtracter) Extract(ctx context.Context) ([]operation.Operation, error) {
	se.Log().Debug().Msg("trying to extract seals")

	var notFounds []valuehash.Hash
	var opsCount int
	opsBySeals := map[string][]operation.Operation{}

	i, f, err := se.extractFromStorage(ctx, opsBySeals)
	if err != nil {
		return nil, err
	}
	opsCount += i
	notFounds = f

	if len(notFounds) > 0 {
		i, err := se.extractFromChannel(ctx, notFounds, opsBySeals)
		if err != nil {
			return nil, err
		}
		opsCount += i
	}

	se.Log().Debug().Int("operations", opsCount).Msg("extracted seals and it's operations")

	ops := make([]operation.Operation, opsCount)

	var offset int
	for i := range se.seals {
		h := se.seals[i]
		if l, found := opsBySeals[h.String()]; !found {
			continue
		} else if len(l) > 0 {
			copy(ops[offset:], l)
			offset += len(l)
		}
	}

	return ops, nil
}

func (se *SealsExtracter) extractFromStorage(
	ctx context.Context,
	opsBySeals map[string][]operation.Operation,
) (int, []valuehash.Hash, error) {
	var notFounds []valuehash.Hash
	var count int
	for i := range se.seals {
		h := se.seals[i]
		switch ops, found, err := se.fromStorage(ctx, h); {
		case err != nil:
			if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
				return count, nil, err
			} else if errors.Is(err, util.NotFoundError) {
				notFounds = append(notFounds, h)

				continue
			}

			return count, nil, err
		case !found:
			notFounds = append(notFounds, h)

			continue
		default:
			opsBySeals[h.String()] = se.filterFounds(ops)
			count += len(opsBySeals[h.String()])
		}
	}

	se.Log().Debug().Int("operations", count).Msg("extracted from storage")

	return count, notFounds, nil
}

func (se *SealsExtracter) extractFromChannel(
	ctx context.Context,
	seals []valuehash.Hash,
	opsBySeals map[string][]operation.Operation,
) (int, error) {
	finished := make(chan error)

	var m map[string][]operation.Operation
	go func() {
		i, err := se.fromChannel(seals)
		if err == nil {
			m = i
		}

		finished <- err
	}()

	var count int
	select {
	case <-ctx.Done():
		return count, ctx.Err()
	case err := <-finished:
		if err != nil {
			return count, err
		}
	}

	for k := range m {
		opsBySeals[k] = se.filterFounds(m[k])
		count += len(opsBySeals[k])
	}

	se.Log().Debug().Int("operations", count).Msg("extracted from remote")

	return count, nil
}

func (se *SealsExtracter) filterFounds(ops []operation.Operation) []operation.Operation {
	if len(ops) < 1 {
		return nil
	}

	var nops []operation.Operation
	for i := range ops {
		h := ops[i].Hash().String()
		if _, found := se.founds[h]; !found {
			nops = append(nops, ops[i])

			se.founds[h] = struct{}{}
		}
	}

	return nops
}

func (*SealsExtracter) filterDuplicated(ops []operation.Operation) []operation.Operation {
	if len(ops) < 1 {
		return nil
	}

	founds := map[string]struct{}{}
	var nops []operation.Operation
	for i := range ops {
		op := ops[i]
		fk := op.Hash().String()
		if _, found := founds[fk]; !found {
			nops = append(nops, op)
			founds[fk] = struct{}{}
		}
	}

	return nops
}

func (se *SealsExtracter) fromStorage(
	ctx context.Context,
	h valuehash.Hash, /* seal.Hash() */
) ([]operation.Operation, bool, error) {
	var ops []operation.Operation
	var found bool
	f := func(h valuehash.Hash) error {
		if sl, found0, err := se.database.Seal(h); err != nil {
			return err
		} else if !found0 {
			return nil
		} else if os, ok := sl.(operation.Seal); !ok {
			return errors.Errorf("not operation.Seal: %T", sl)
		} else {
			ops = se.filterDuplicated(os.Operations())
			found = true

			return nil
		}
	}

	finished := make(chan error)
	go func() {
		finished <- f(h)
	}()

	select {
	case <-ctx.Done():
		return nil, false, ctx.Err()
	case err := <-finished:
		if err != nil {
			return nil, false, err
		}
	}

	return ops, found, nil
}

func (se *SealsExtracter) fromChannel(notFounds []valuehash.Hash) (map[string][]operation.Operation, error) {
	if se.local.Equal(se.proposer) {
		return nil, errors.Errorf("proposer is local, but it does not have seals. Hmmm")
	}

	_, proposerch, found := se.nodepool.Node(se.proposer)
	if !found {
		return nil, errors.Errorf("proposer is not in nodes: %v", se.proposer)
	} else if proposerch == nil {
		return nil, errors.Errorf("proposer is dead: %v", se.proposer)
	}

	received, err := proposerch.Seals(context.TODO(), notFounds)
	if err != nil {
		return nil, err
	}

	if err := se.database.NewSeals(received); err != nil {
		if !errors.Is(err, util.DuplicatedError) {
			return nil, err
		}
	}

	bySeals := map[string][]operation.Operation{}
	for i := range received {
		sl := received[i]
		os, ok := sl.(operation.Seal)
		if !ok {
			return nil, errors.Errorf("not operation.Seal: %T", sl)
		}
		bySeals[sl.Hash().String()] = os.Operations()
	}

	return bySeals, nil
}
