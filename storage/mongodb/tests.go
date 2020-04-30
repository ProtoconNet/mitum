// +build test mongodb

package mongodbstorage

import (
	"fmt"
	"os"

	"github.com/rs/zerolog"

	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/logging"
)

var BaseTestMongodbURI = "mongodb://localhost:27017"

var log logging.Logger // nolint

func init() {
	l := zerolog.
		New(os.Stderr).
		With().
		Timestamp().
		Caller().
		Stack().
		Logger().Level(zerolog.DebugLevel)

	log = logging.NewLogger(&l, true)
}

func TestMongodbURI() string {
	uri := "localhost"
	if s := os.Getenv("MITUM_TEST_MONGODB_URI"); len(s) > 0 {
		uri = s
	}

	return fmt.Sprintf("mongodb://%s/t_%s", uri, util.UUID().String())
}