package misc

import (
	"encoding/hex"
	"fmt"

	"github.com/theQRL/go-qrllib/common"
	"github.com/theQRL/go-qrllib/dilithium"
)

func StrSeedToBinSeed(strSeed string) [common.SeedSize]uint8 {
	var seed [common.SeedSize]uint8

	unSizedSeed := DecodeHex(strSeed)

	copy(seed[:], unSizedSeed)
	return seed
}

func DecodeHex(hexString string) []byte {
	if hexString[:2] != "0x" {
		panic(fmt.Errorf("invalid hex string prefix %s", hexString[:2]))
	}
	hexBytes, err := hex.DecodeString(hexString[2:])
	if err != nil {
		panic(fmt.Errorf("failed to decode string %s | reason %v",
			hexString, err))
	}
	return hexBytes
}

func EncodeHex(hexBytes []byte) string {
	return fmt.Sprintf("0x%x", hexBytes)
}

func ToSizedDilithiumSignature(sig []byte) [dilithium.CryptoBytes]byte {
	if len(sig) != dilithium.CryptoBytes {
		panic(fmt.Errorf("cannot convert sig to sized dilithium sig, invalid sig length %d", len(sig)))
	}
	var sizedSig [dilithium.CryptoBytes]byte
	copy(sizedSig[:], sig)
	return sizedSig
}

func ToSizedDilithiumPublicKey(pk []byte) [dilithium.CryptoPublicKeyBytes]byte {
	if len(pk) != dilithium.CryptoPublicKeyBytes {
		panic(fmt.Errorf("cannot convert pk to sized dilithium pk, invalid pk length %d", len(pk)))
	}
	var sizedPK [dilithium.CryptoPublicKeyBytes]byte
	copy(sizedPK[:], pk)
	return sizedPK
}
