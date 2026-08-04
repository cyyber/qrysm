package util

import (
	qrlmldsa "github.com/theQRL/go-qrllib/crypto/ml_dsa_87"
	walletcommon "github.com/theQRL/go-qrllib/wallet/common"
	walletmldsa "github.com/theQRL/go-qrllib/wallet/ml_dsa_87"
	"github.com/theQRL/qrysm/crypto/ml_dsa_87"
)

func deterministicSign(key ml_dsa_87.MLDSA87Key, msg []byte) ([]byte, error) {
	var seed walletcommon.Seed
	copy(seed[:], key.Marshal())
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
	return signature[:], nil
}
