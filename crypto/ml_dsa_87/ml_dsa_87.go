package ml_dsa_87

import (
	"errors"

	"github.com/theQRL/qrysm/crypto/ml_dsa_87/common"
	"github.com/theQRL/qrysm/crypto/ml_dsa_87/ml_dsa_87t"
)

type deterministicSigner interface {
	SignDeterministic([]byte) (common.Signature, error)
}

// SecretKeyFromSeed creates a ML-DSA-87 private key from a seed.
func SecretKeyFromSeed(seed []byte) (MLDSA87Key, error) {
	return ml_dsa_87t.SecretKeyFromSeed(seed)
}

// PublicKeyFromBytes creates a ML-DSA-87 public key from a byte slice.
func PublicKeyFromBytes(pubKey []byte) (PublicKey, error) {
	return ml_dsa_87t.PublicKeyFromBytes(pubKey)
}

// SignatureFromBytes creates a ML-DSA-87 signature from a byte slice.
func SignatureFromBytes(sig []byte) (Signature, error) {
	return ml_dsa_87t.SignatureFromBytes(sig)
}

// VerifySignature verifies a single signature. For performance reason, always use VerifyMultipleSignatures if possible.
func VerifySignature(sig []byte, msg [32]byte, pubKey common.PublicKey) (bool, error) {
	return ml_dsa_87t.VerifySignature(sig, msg, pubKey)
}

// VerifyMultipleSignatures verifies multiple signatures for distinct messages securely.
func VerifyMultipleSignatures(sigs [][][]byte, msgs [][32]byte, pubKeys [][]common.PublicKey) (bool, error) {
	return ml_dsa_87t.VerifyMultipleSignatures(sigs, msgs, pubKeys)
}

// SignDeterministic signs a message without per-signature randomness.
func SignDeterministic(key MLDSA87Key, msg []byte) (Signature, error) {
	signer, ok := key.(deterministicSigner)
	if !ok {
		return nil, errors.New("ML-DSA-87 key does not support deterministic signing")
	}
	return signer.SignDeterministic(msg)
}

// RandKey creates a new private key using a random input.
func RandKey() (common.SecretKey, error) {
	return ml_dsa_87t.RandKey()
}
