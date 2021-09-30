package blockdata

import (
	"github.com/pkg/errors"
	"github.com/spikeekips/mitum/base"
	"github.com/spikeekips/mitum/base/block"
	"github.com/spikeekips/mitum/storage"
	"github.com/spikeekips/mitum/util"
)

// Clean makes Database and BlockData to be empty. If 'remove' is true, remove
// the BlockData directory itself.
func Clean(db storage.Database, blockData BlockData, remove bool) error {
	if err := db.Clean(); err != nil {
		return err
	} else if err := blockData.Clean(remove); err != nil {
		return err
	} else {
		return nil
	}
}

func CleanByHeight(db storage.Database, blockData BlockData, height base.Height) error {
	if err := db.LocalBlockDataMapsByHeight(height, func(bd block.BlockDataMap) (bool, error) {
		if err := blockData.RemoveAll(bd.Height()); err != nil {
			if errors.Is(err, util.NotFoundError) {
				return true, nil
			}

			return false, err
		}
		return true, nil
	}); err != nil {
		return err
	}

	return db.CleanByHeight(height)
}

func CheckBlock(db storage.Database, blockData BlockData, networkID base.NetworkID) (block.Manifest, error) {
	m, err := storage.CheckBlock(db, networkID)
	if err != nil {
		return m, err
	}

	if found, err := blockData.Exists(m.Height()); err != nil {
		return m, err
	} else if !found {
		return m, util.NotFoundError.Errorf("block, %d not found in block data", m.Height())
	}

	return m, nil
}
