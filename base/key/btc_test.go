package key

import (
	"testing"

	"github.com/spikeekips/mitum/util"
	"github.com/stretchr/testify/suite"
	"golang.org/x/xerrors"
)

type testBTCKeypair struct {
	suite.Suite
}

func (t *testBTCKeypair) TestNew() {
	kp, err := NewBTCPrivatekey()
	t.NoError(err)

	t.Implements((*Privatekey)(nil), kp)
}

func (t *testBTCKeypair) TestKeypairIsValid() {
	kp, _ := NewBTCPrivatekey()
	t.NoError(kp.IsValid(nil))

	// empty Keypair
	empty := BTCPrivatekey{}
	t.True(xerrors.Is(empty.IsValid(nil), InvalidKeyError))
}

func (t *testBTCKeypair) TestKeypairExportKeys() {
	priv := "L1bQZCcDZKy342x8xjK9Hk935Nttm2jkApVVS2mn4Nqyxvu7nyGC"
	kp, _ := NewBTCPrivatekeyFromString(priv)

	t.Equal("27phogA4gmbMGfg321EHfx5eABkL7KAYuDPRGFoyQtAUb", kp.Publickey().String())
	t.Equal(priv, kp.String())
}

func (t *testBTCKeypair) TestPublickey() {
	priv := "L1bQZCcDZKy342x8xjK9Hk935Nttm2jkApVVS2mn4Nqyxvu7nyGC"
	kp, _ := NewBTCPrivatekeyFromString(priv)

	t.Equal("27phogA4gmbMGfg321EHfx5eABkL7KAYuDPRGFoyQtAUb", kp.Publickey().String())

	t.NoError(kp.IsValid(nil))

	pk, err := NewBTCPublickeyFromString(kp.Publickey().String())
	t.NoError(err)

	t.True(kp.Publickey().Equal(pk))
}

func (t *testBTCKeypair) TestPublickeyEqual() {
	kp, _ := NewBTCPrivatekey()

	t.True(kp.Publickey().Equal(kp.Publickey()))

	nkp, _ := NewBTCPrivatekey()
	t.False(kp.Publickey().Equal(nkp.Publickey()))
}

func (t *testBTCKeypair) TestPrivatekey() {
	priv := "L1bQZCcDZKy342x8xjK9Hk935Nttm2jkApVVS2mn4Nqyxvu7nyGC"
	kp, _ := NewBTCPrivatekeyFromString(priv)

	t.Equal(priv, kp.String())

	t.NoError(kp.IsValid(nil))

	pk, err := NewBTCPrivatekeyFromString(kp.String())
	t.NoError(err)

	t.True(kp.Equal(pk))
}

func (t *testBTCKeypair) TestPrivatekeyEqual() {
	kp, _ := NewBTCPrivatekey()

	t.True(kp.Equal(kp))

	nkp, _ := NewBTCPrivatekey()
	t.False(kp.Equal(nkp))
}

func (t *testBTCKeypair) TestSign() {
	kp, _ := NewBTCPrivatekey()

	input := []byte("makeme")

	// sign
	sig, err := kp.Sign(input)
	t.NoError(err)
	t.NotNil(sig)

	// verify
	err = kp.Publickey().Verify(input, sig)
	t.NoError(err)
}

func (t *testBTCKeypair) TestSignInvalidInput() {
	kp, _ := NewBTCPrivatekey()

	b := []byte(util.UUID().String())

	input := b
	input = append(input, []byte("findme000")...)

	sig, err := kp.Sign(input)
	t.NoError(err)
	t.NotNil(sig)

	{
		err = kp.Publickey().Verify(input, sig)
		t.NoError(err)
	}

	{
		newInput := b
		newInput = append(newInput, []byte("showme")...)

		err = kp.Publickey().Verify(newInput, sig)
		t.True(xerrors.Is(err, SignatureVerificationFailedError))
	}
}

func TestBTCKeypair(t *testing.T) {
	suite.Run(t, new(testBTCKeypair))
}
