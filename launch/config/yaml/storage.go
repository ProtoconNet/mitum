package yamlconfig

import (
	"context"

	"github.com/spikeekips/mitum/launch/config"
)

type BlockData struct {
	Path *string
}

type Database struct {
	URI   *string `yaml:",omitempty"`
	Cache *string `yaml:",omitempty"`
}

func (no Database) Set(ctx context.Context) (context.Context, error) {
	var l config.LocalNode
	if err := config.LoadConfigContextValue(ctx, &l); err != nil {
		return ctx, err
	}
	conf := l.Storage().Database()

	if no.URI != nil {
		if err := conf.SetURI(*no.URI); err != nil {
			return ctx, err
		}
	}
	if no.Cache != nil {
		if err := conf.SetCache(*no.Cache); err != nil {
			return ctx, err
		}
	}

	return ctx, nil
}

type Storage struct {
	Database  *Database  `yaml:"database,omitempty"`
	BlockData *BlockData `yaml:"blockdata,omitempty"`
}

func (no Storage) Set(ctx context.Context) (context.Context, error) {
	var l config.LocalNode
	if err := config.LoadConfigContextValue(ctx, &l); err != nil {
		return ctx, err
	}
	conf := l.Storage()

	if no.Database != nil {
		i, err := no.Database.Set(ctx)
		if err != nil {
			return ctx, err
		}
		ctx = i
	}

	if no.BlockData != nil {
		if no.BlockData.Path != nil {
			if err := conf.BlockData().SetPath(*no.BlockData.Path); err != nil {
				return ctx, err
			}
		}
	}

	return ctx, nil
}
