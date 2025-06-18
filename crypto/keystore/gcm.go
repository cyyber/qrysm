package keystore

import (
	"crypto/aes"
	"crypto/cipher"
	"fmt"
)

const GCMNonceSize = 12

// encryptGCM encrypts plaintext using AES-GCM with the given key and nonce. The ciphertext is
// appended to dest, which must not overlap with plaintext.
func encryptGCM(dest, key, nonce, plaintext, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("can't create block cipher: %v", err))
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, GCMNonceSize)
	if err != nil {
		panic(fmt.Errorf("can't create GCM: %v", err))
	}
	return aesgcm.Seal(dest, nonce, plaintext, additionalData), nil
}

// decryptGCM decrypts cyphertext using AES-GCM with the given key and nonce.
func decryptGCM(key, nonce, cyphertext, additionalData []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("can't create block cipher: %v", err)
	}
	if len(nonce) != GCMNonceSize {
		return nil, fmt.Errorf("invalid GCM nonce size: %d", len(nonce))
	}
	aesgcm, err := cipher.NewGCMWithNonceSize(block, GCMNonceSize)
	if err != nil {
		return nil, fmt.Errorf("can't create GCM: %v", err)
	}
	pt := make([]byte, 0, len(cyphertext))
	return aesgcm.Open(pt, nonce, cyphertext, additionalData)
}
