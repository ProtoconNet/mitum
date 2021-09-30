//go:build test
// +build test

package network

import (
	"testing"

	"github.com/spikeekips/mitum/util/encoder"
	bsonenc "github.com/spikeekips/mitum/util/encoder/bson"
	jsonenc "github.com/spikeekips/mitum/util/encoder/json"
	"github.com/stretchr/testify/suite"
)

type testHTTPConnInfo struct {
	suite.Suite
}

func (t *testHTTPConnInfo) TestEqual() {
	t.Run("localhost", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://localhost", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.True(conn.Equal(b))
	})

	t.Run("insecure not same", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://localhost", false)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.False(conn.Equal(b))
	})

	t.Run("ignore root path", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost/", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://localhost", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.True(conn.Equal(b))
	})

	t.Run("fragment not ignored", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost#showme", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://localhost", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.False(conn.Equal(b))
	})

	t.Run("end path slash", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost/showme/", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://localhost/showme", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.False(conn.Equal(b))
	})

	t.Run("omit wellknown port", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost/showme/", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://localhost:443/showme/", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.True(conn.Equal(b))
	})

	t.Run("complex query", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost/?a=a&b=b", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://localhost?b=b&a=a", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.True(conn.Equal(b))
	})

	t.Run("localhost: no surplus data", func() {
		conn, err := NewHTTPConnInfoFromString("https://localhost", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		b, err := NewHTTPConnInfoFromString("https://127.0.0.1", true)
		t.NoError(err)
		t.NoError(conn.IsValid(nil))

		t.True(conn.Equal(b))
	})
}

func TestHTTPConnInfo(t *testing.T) {
	suite.Run(t, new(testHTTPConnInfo))
}

type testNilConnInfoEncode struct {
	suite.Suite

	enc encoder.Encoder
}

func (t *testNilConnInfoEncode) SetupSuite() {
	t.enc.Add(NilConnInfo{})
}

func (t *testNilConnInfoEncode) TestMarshal() {
	conn := NewNilConnInfo("showme")
	t.NoError(conn.IsValid(nil))

	b, err := t.enc.Marshal(conn)
	t.NoError(err)
	t.NotNil(b)

	hinter, err := DecodeConnInfo(b, t.enc)
	t.NoError(err)

	_, ok := hinter.(NilConnInfo)
	t.True(ok)

	t.False(conn.Equal(nil))
	t.True(conn.Equal(hinter))
}

func TestNilConnInfoEncodeJSON(t *testing.T) {
	b := new(testNilConnInfoEncode)
	b.enc = jsonenc.NewEncoder()

	suite.Run(t, b)
}

func TestNilConnInfoEncodeBSON(t *testing.T) {
	b := new(testNilConnInfoEncode)
	b.enc = bsonenc.NewEncoder()

	suite.Run(t, b)
}

type testHTTPConnInfoEncode struct {
	suite.Suite

	enc encoder.Encoder
}

func (t *testHTTPConnInfoEncode) SetupSuite() {
	t.enc.Add(HTTPConnInfo{})
}

func (t *testHTTPConnInfoEncode) TestMarshal() {
	conn, err := NewHTTPConnInfoFromString("https://a.b.c:1234/showme/findme#killme", true)
	t.NoError(err)
	t.NoError(conn.IsValid(nil))

	b, err := t.enc.Marshal(conn)
	t.NoError(err)
	t.NotNil(b)

	hinter, err := DecodeConnInfo(b, t.enc)
	t.NoError(err)

	_, ok := hinter.(HTTPConnInfo)
	t.True(ok)

	t.False(conn.Equal(nil))
	t.True(conn.Equal(hinter))
}

func TestHTTPConnInfoEncodeJSON(t *testing.T) {
	b := new(testHTTPConnInfoEncode)
	b.enc = jsonenc.NewEncoder()

	suite.Run(t, b)
}

func TestHTTPConnInfoEncodeBSON(t *testing.T) {
	b := new(testHTTPConnInfoEncode)
	b.enc = bsonenc.NewEncoder()

	suite.Run(t, b)
}
