package ecdsa

import (
	"crypto/ecdsa"
	"crypto/rand"
	"math/big"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/libp2p/go-libp2p/core/crypto"
	gcrypto "github.com/theQRL/go-zond/crypto"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
)

func TestConvertToInterfacePubkey(t *testing.T) {
	privKey, err := gcrypto.GenerateKey()
	require.NoError(t, err)

	pubkey, ok := privKey.Public().(*ecdsa.PublicKey)
	require.NotEqual(t, false, ok)

	altPubkey, err := ConvertToInterfacePubkey(pubkey)
	require.NoError(t, err)

	nKey := *(altPubkey.(*crypto.Secp256k1PublicKey))
	rawKey := btcec.PublicKey(nKey).SerializeUncompressed()
	origRawKey := gcrypto.FromECDSAPub(pubkey)
	assert.DeepEqual(t, origRawKey, rawKey)
}

func TestConvertToInterfacePrivkey_HandlesShorterKeys(t *testing.T) {
	priv, _, err := crypto.GenerateSecp256k1Key(rand.Reader)
	assert.NoError(t, err)
	rawBytes, err := priv.Raw()
	assert.NoError(t, err)
	// Zero-out most significant byte so that the big int normalizes
	// it by removing it.
	rawBytes[0] = 0
	privKey := new(ecdsa.PrivateKey)
	k := new(big.Int).SetBytes(rawBytes)
	privKey.D = k
	privKey.Curve = gcrypto.S256()
	privKey.X, privKey.Y = gcrypto.S256().ScalarBaseMult(rawBytes)
	_, err = ConvertToInterfacePrivkey(privKey)
	assert.NoError(t, err)
}
