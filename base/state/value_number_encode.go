package state

import (
	"reflect"

	"github.com/spikeekips/mitum/base/valuehash"
	"github.com/spikeekips/mitum/util"
	"github.com/spikeekips/mitum/util/encoder"
	"golang.org/x/xerrors"
)

func (nv *NumberValue) unpack(enc encoder.Encoder, bHash, bValue []byte, t reflect.Kind) error {
	var h valuehash.Hash
	if i, err := valuehash.Decode(enc, bHash); err != nil {
		return err
	} else {
		h = i
	}

	var v interface{}
	switch t {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := util.BytesToInt64(bValue); err != nil {
			return err
		} else {
			switch t {
			case reflect.Int:
				v = int(i)
			case reflect.Int8:
				v = int8(i)
			case reflect.Int16:
				v = int16(i)
			case reflect.Int32:
				v = int32(i)
			case reflect.Int64:
				v = i
			}
		}
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if i, err := util.BytesToUint64(bValue); err != nil {
			return err
		} else {
			switch t {
			case reflect.Uint:
				v = uint(i)
			case reflect.Uint8:
				v = uint8(i)
			case reflect.Uint16:
				v = uint16(i)
			case reflect.Uint32:
				v = uint32(i)
			case reflect.Uint64:
				v = i
			}
		}
	case reflect.Float64:
		v = util.BytesToFloat64(bValue)
	default:
		return xerrors.Errorf("unsupported type for NumberValue: %v", t)
	}

	nv.h = h
	nv.v = v
	nv.b = bValue

	return nil
}
