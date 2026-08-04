package ml_dsa_87t

import (
	"fmt"

	qrlmldsa "github.com/theQRL/go-qrllib/crypto/ml_dsa_87"
	walletcommon "github.com/theQRL/go-qrllib/wallet/common"
	walletmldsa "github.com/theQRL/go-qrllib/wallet/ml_dsa_87"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/crypto/ml_dsa_87/common"
	"github.com/theQRL/qrysm/crypto/rand"
)

type mlDSA87Key struct {
	w *walletmldsa.Wallet
}

func RandKey() (common.SecretKey, error) {
	var seed [field_params.MLDSA87SeedLength]uint8
	_, err := rand.NewGenerator().Read(seed[:])
	if err != nil {
		return nil, err
	}
	w, err := walletmldsa.NewWalletFromSeed(seed)
	if err != nil {
		return nil, err
	}
	return &mlDSA87Key{w: w}, nil
}

func SecretKeyFromSeed(seed []byte) (common.SecretKey, error) {
	if len(seed) != field_params.MLDSA87SeedLength {
		return nil, fmt.Errorf("secret key must be %d bytes", field_params.MLDSA87SeedLength)
	}
	var sizedSeed [field_params.MLDSA87SeedLength]uint8
	copy(sizedSeed[:], seed)

	w, err := walletmldsa.NewWalletFromSeed(sizedSeed)
	if err != nil {
		return nil, err
	}
	return &mlDSA87Key{w: w}, nil
}

// PublicKey obtains the public key corresponding to the ML-DSA-87 secret key.
func (m *mlDSA87Key) PublicKey() common.PublicKey {
	p := m.w.GetPK()
	return &PublicKey{p: &p}
}

func (m *mlDSA87Key) Sign(msg []byte) common.Signature {
	signature, err := m.w.Sign(msg)
	if err != nil {
		return nil
	}
	return &Signature{s: &signature}
}

func (m *mlDSA87Key) SignDeterministic(msg []byte) (common.Signature, error) {
	seed := m.w.GetSeed()
	signer, err := qrlmldsa.NewMLDSA87FromSeed(seed.HashSHA256())
	if err != nil {
		return nil, err
	}
	descriptor, err := walletmldsa.NewMLDSA87Descriptor()
	if err != nil {
		return nil, err
	}
	signature, err := signer.SignDeterministic(walletcommon.SigningContext(descriptor.ToDescriptor()), msg)
	if err != nil {
		return nil, err
	}
	return &Signature{s: &signature}, nil
}

// Marshal a secret key into a LittleEndian byte slice.
func (m *mlDSA87Key) Marshal() []byte {
	keyBytes := m.w.GetSeed()
	return keyBytes[:]
}
