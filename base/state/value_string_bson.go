package state

import (
	"go.mongodb.org/mongo-driver/bson"

	"github.com/spikeekips/mitum/base/valuehash"
	bsonencoder "github.com/spikeekips/mitum/util/encoder/bson"
)

func (sv StringValue) MarshalBSON() ([]byte, error) {
	return bsonencoder.Marshal(bsonencoder.MergeBSONM(
		bsonencoder.NewHintedDoc(sv.Hint()),
		bson.M{
			"hash":  sv.Hash(),
			"value": sv.v,
		},
	))
}

type StringValueUnpackerBSON struct {
	H bson.Raw `bson:"hash"`
	V string   `bson:"value"`
}

func (sv *StringValue) UnpackBSON(b []byte, enc *bsonencoder.Encoder) error {
	var uv StringValueUnpackerBSON
	if err := enc.Unmarshal(b, &uv); err != nil {
		return err
	}

	var err error
	var h valuehash.Hash
	if h, err = valuehash.Decode(enc, uv.H); err != nil {
		return err
	}

	sv.h = h
	sv.v = uv.V

	return nil
}
