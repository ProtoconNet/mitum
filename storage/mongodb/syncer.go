package mongodbstorage

import (
	"context"
	"fmt"
	"sync"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/xerrors"

	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/block"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
)

type SyncerSession struct {
	sync.RWMutex
	*logging.Logging
	main             *Database
	manifestDatabase *Database
	session          *Database
	heightFrom       base.Height
	heightTo         base.Height
}

func NewSyncerSession(main *Database) (*SyncerSession, error) {
	var manifestDatabase, session *Database

	if s, err := newTempDatabase(main, "manifest"); err != nil {
		return nil, err
	} else if err := s.CreateIndex(ColNameManifest, manifestIndexModels, IndexPrefix); err != nil {
		return nil, err
	} else {
		manifestDatabase = s
	}
	if s, err := newTempDatabase(main, "block"); err != nil {
		return nil, err
	} else {
		session = s
	}

	return &SyncerSession{
		Logging: logging.NewLogging(func(c logging.Context) logging.Emitter {
			return c.Str("module", "mongodb-syncer-database")
		}),
		main:             main,
		manifestDatabase: manifestDatabase,
		session:          session,
		heightFrom:       base.NilHeight,
		heightTo:         base.NilHeight,
	}, nil
}

func (st *SyncerSession) Manifest(height base.Height) (block.Manifest, bool, error) {
	return st.manifestDatabase.ManifestByHeight(height)
}

func (st *SyncerSession) Manifests(heights []base.Height) ([]block.Manifest, error) {
	var bs []block.Manifest
	for i := range heights {
		if b, found, err := st.manifestDatabase.ManifestByHeight(heights[i]); !found {
			return nil, storage.NotFoundError.Errorf("manifest not found")
		} else if err != nil {
			return nil, err
		} else {
			bs = append(bs, b)
		}
	}

	return bs, nil
}

func (st *SyncerSession) SetManifests(manifests []block.Manifest) error {
	st.Lock()
	defer st.Unlock()

	var lastManifest block.Manifest
	for _, m := range manifests {
		if lastManifest == nil {
			lastManifest = m
		} else if m.Height() > lastManifest.Height() {
			lastManifest = m
		}
	}

	var models []mongo.WriteModel
	for i := range manifests {
		m := manifests[i]
		if doc, err := NewManifestDoc(m, st.session.Encoder()); err != nil {
			return err
		} else {
			models = append(models, mongo.NewInsertOneModel().SetDocument(doc))
		}

		if h := m.Height(); st.heightFrom <= base.NilHeight || h < st.heightFrom {
			st.heightFrom = h
		}

		if h := m.Height(); h > st.heightTo {
			st.heightTo = h
		}
	}

	if err := st.manifestDatabase.Client().Bulk(context.Background(), ColNameManifest, models, true); err != nil {
		return err
	}

	st.Log().VerboseFunc(func(e *logging.Event) logging.Emitter {
		var heights []base.Height
		for i := range manifests {
			heights = append(heights, manifests[i].Height())
		}

		return e.Interface("heights", heights)
	}).
		Hinted("from_height", st.heightFrom).
		Hinted("to_height", st.heightTo).
		Int("manifests", len(manifests)).
		Msg("set manifests")

	return st.manifestDatabase.setLastManifest(lastManifest, false, false)
}

func (st *SyncerSession) HasBlock(height base.Height) (bool, error) {
	return st.session.client.Exists(ColNameManifest, util.NewBSONFilter("height", height).D())
}

func (st *SyncerSession) SetBlocks(blocks []block.Block, maps []block.BlockDataMap) error {
	if len(blocks) != len(maps) {
		return xerrors.Errorf("blocks and maps has different size, %d != %d", len(blocks), len(maps))
	} else {
		for i := range blocks {
			if err := block.CompareManifestWithMap(blocks[i], maps[i]); err != nil {
				return err
			}
		}
	}

	st.Log().VerboseFunc(func(e *logging.Event) logging.Emitter {
		var heights []base.Height
		for i := range blocks {
			heights = append(heights, blocks[i].Height())
		}

		return e.Interface("heights", heights)
	}).
		Int("blocks", len(blocks)).
		Msg("set blocks")

	var lastBlock block.Block
	for i := range blocks {
		blk := blocks[i]
		m := maps[i]

		if err := st.setBlock(blk, m); err != nil {
			return err
		}

		if lastBlock == nil {
			lastBlock = blk
		} else if blk.Height() > lastBlock.Height() {
			lastBlock = blk
		}
	}

	return st.session.setLastBlock(lastBlock, true, false)
}

func (st *SyncerSession) setBlock(blk block.Block, m block.BlockDataMap) error {
	var bs storage.DatabaseSession
	if st, err := st.session.NewSession(blk); err != nil {
		return err
	} else {
		bs = st
	}

	defer func() {
		_ = bs.Close()
	}()

	if err := bs.SetBlock(context.Background(), blk); err != nil {
		return err
	} else if err := bs.Commit(context.Background(), m); err != nil {
		return err
	}

	return nil
}

func (st *SyncerSession) Commit() error {
	l := st.Log().WithLogger(func(ctx logging.Context) logging.Emitter {
		return ctx.Hinted("from_height", st.heightFrom).
			Hinted("to_height", st.heightTo)
	})

	l.Debug().Msg("trying to commit blocks to main database")

	var last block.Manifest
	if m, found, err := st.session.LastManifest(); err != nil || !found {
		return xerrors.Errorf("failed to get last manifest fromm storage: %w", err)
	} else {
		last = m
	}

	for _, col := range []string{
		ColNameManifest,
		ColNameSeal,
		ColNameOperation,
		ColNameOperationSeal,
		ColNameProposal,
		ColNameState,
		ColNameVoteproof,
		ColNameBlockDataMap,
	} {
		if err := moveWithinCol(st.session, col, st.main, col, bson.D{}); err != nil {
			l.Error().Err(err).Str("collection", col).Msg("failed to move collection")

			return err
		}
		l.Debug().Str("collection", col).Msg("moved collection")
	}

	return st.main.setLastBlock(last, false, false)
}

func (st *SyncerSession) Close() error {
	// NOTE drop tmp database
	if err := st.manifestDatabase.client.DropDatabase(); err != nil {
		return err
	}

	if err := st.session.client.DropDatabase(); err != nil {
		return err
	}

	return nil
}

func newTempDatabase(main *Database, prefix string) (*Database, error) {
	// NOTE create new mongodb database with prefix
	var tmpClient *Client
	if c, err := main.client.New(fmt.Sprintf("sync-%s_%s", prefix, util.UUID().String())); err != nil {
		return nil, err
	} else {
		tmpClient = c
	}

	return NewDatabase(tmpClient, main.Encoders(), main.Encoder(), main.Cache())
}

func moveWithinCol(from *Database, fromCol string, to *Database, toCol string, filter bson.D) error {
	var limit int = 100
	var models []mongo.WriteModel
	err := from.Client().Find(context.Background(), fromCol, filter, func(cursor *mongo.Cursor) (bool, error) {
		if len(models) == limit {
			if err := to.Client().Bulk(context.Background(), toCol, models, false); err != nil {
				return false, err
			} else {
				models = nil
			}
		}

		raw := util.CopyBytes(cursor.Current)
		models = append(models, mongo.NewInsertOneModel().SetDocument(bson.Raw(raw)))

		return true, nil
	})
	if err != nil {
		return err
	}

	if len(models) > 0 {
		if err := to.Client().Bulk(context.Background(), toCol, models, false); err != nil {
			return err
		}
	}

	return nil
}
