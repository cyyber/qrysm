package keyderivation

import (
	"errors"
	"fmt"

	"github.com/theQRL/go-qrllib/dilithium"
	"github.com/theQRL/qrysm/cmd/staking-deposit-cli/misc"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"golang.org/x/crypto/sha3"
)

// SeedAndPathToSeed TODO: (cyyber) algorithm needs to be reviewed in future
func SeedAndPathToSeed(strSeed, path string) (string, error) {
	binSeed := misc.DecodeHex(strSeed)
	if len(binSeed) != field_params.DilithiumSeedLength {
		return "", fmt.Errorf("invalid seed size %d", len(binSeed))
	}

	var seed [field_params.DilithiumSeedLength]uint8
	copy(seed[:], binSeed)

	h := sha3.NewShake256()
	if _, err := h.Write(seed[:]); err != nil {
		return "", fmt.Errorf("shake256 hash write failed %v", err)
	}
	if _, err := h.Write([]byte(path)); err != nil {
		return "", fmt.Errorf("shake256 hash write failed %v", err)
	}

	var newSeed [field_params.DilithiumSeedLength]uint8
	_, err := h.Read(newSeed[:])
	if err != nil {
		return "", err
	}

	// Try generating Dilithium from seed to ensure seed validity
	_, err = dilithium.NewDilithiumFromSeed(newSeed)
	if err != nil {
		return "", errors.New("could not generate dilithium from mnemonic")
	}

	return misc.EncodeHex(newSeed[:]), nil
}
