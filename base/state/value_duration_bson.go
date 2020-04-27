package state

import (
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base/valuehash"
	bsonencoder "github.com/spikeekips/mitum/util/encoder/bson"
)

func (dv DurationValue) MarshalBSON() ([]byte, error) {
	return bsonencoder.Marshal(bsonencoder.MergeBSONM(
		bsonencoder.NewHintedDoc(dv.Hint()),
		bson.M{
			"hash":  dv.Hash(),
			"value": dv.v.Nanoseconds(),
		},
	))
}

type DurationValueUnpackerBSON struct {
	H bson.Raw `bson:"hash"`
	V int64    `bson:"value"`
}

func (dv *DurationValue) UnpackBSON(b []byte, enc *bsonencoder.Encoder) error {
	var uv DurationValueUnpackerBSON
	if err := enc.Unmarshal(b, &uv); err != nil {
		return err
	}

	var err error
	var h valuehash.Hash
	if h, err = valuehash.Decode(enc, uv.H); err != nil {
		return err
	}

	dv.h = h
	dv.v = time.Duration(uv.V)

	return nil
}
