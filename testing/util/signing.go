package util

import (
	fssz "github.com/prysmaticlabs/fastssz"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/ml_dsa_87"
)

func deterministicSign(key ml_dsa_87.MLDSA87Key, msg []byte) ([]byte, error) {
	signature, err := deterministicSignature(key, msg)
	if err != nil {
		return nil, err
	}
	return signature.Marshal(), nil
}

func deterministicSignature(key ml_dsa_87.MLDSA87Key, msg []byte) (ml_dsa_87.Signature, error) {
	return ml_dsa_87.SignDeterministic(key, msg)
}

func computeDomainAndSignDeterministic(
	st state.ReadOnlyBeaconState,
	epoch primitives.Epoch,
	object fssz.HashRoot,
	domain [4]byte,
	key ml_dsa_87.MLDSA87Key,
) ([]byte, error) {
	signatureDomain, err := signing.Domain(st.Fork(), epoch, domain, st.GenesisValidatorsRoot())
	if err != nil {
		return nil, err
	}
	signingRoot, err := signing.ComputeSigningRoot(object, signatureDomain)
	if err != nil {
		return nil, err
	}
	return deterministicSign(key, signingRoot[:])
}
