package util

import (
	"errors"

	fssz "github.com/prysmaticlabs/fastssz"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/ml_dsa_87"
)

func deterministicSign(key ml_dsa_87.MLDSA87Key, msg []byte) ([]byte, error) {
	sig, err := deterministicSignature(key, msg)
	if err != nil {
		return nil, err
	}
	return sig.Marshal(), nil
}

func deterministicSignature(key ml_dsa_87.MLDSA87Key, msg []byte) (ml_dsa_87.Signature, error) {
	sig := ml_dsa_87.SignDeterministic(key, msg)
	if sig == nil {
		return nil, errors.New("could not sign message deterministically")
	}
	return sig, nil
}

func computeDomainAndSignDeterministic(
	st state.ReadOnlyBeaconState,
	epoch primitives.Epoch,
	obj fssz.HashRoot,
	domain [4]byte,
	key ml_dsa_87.MLDSA87Key,
) ([]byte, error) {
	d, err := signing.Domain(st.Fork(), epoch, domain, st.GenesisValidatorsRoot())
	if err != nil {
		return nil, err
	}
	signingRoot, err := signing.ComputeSigningRoot(obj, d)
	if err != nil {
		return nil, err
	}
	return deterministicSign(key, signingRoot[:])
}
