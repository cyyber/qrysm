package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"
	keystorev1 "github.com/theQRL/go-zond-wallet-encryptor-keystore"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/crypto/dilithium"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/validator/keymanager"
	"github.com/urfave/cli/v2"
)

const password = "secretPassw0rd$1999"

type cliConfig struct {
	keystoresPath string
	password      string
	seed          string
	outputPath    string
}

func setupCliContext(
	tb testing.TB,
	conf *cliConfig,
) *cli.Context {
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String(keystoresFlag.Name, conf.keystoresPath, "")
	set.String(passwordFlag.Name, conf.password, "")
	set.String(seedFlag.Name, conf.seed, "")
	set.String(outputPathFlag.Name, conf.outputPath, "")
	assert.NoError(tb, set.Set(keystoresFlag.Name, conf.keystoresPath))
	assert.NoError(tb, set.Set(passwordFlag.Name, conf.password))
	assert.NoError(tb, set.Set(seedFlag.Name, conf.seed))
	assert.NoError(tb, set.Set(outputPathFlag.Name, conf.outputPath))
	return cli.NewContext(&app, set, nil)
}

func createRandomKeystore(t testing.TB, password string) (*keymanager.Keystore, dilithium.DilithiumKey) {
	encryptor := keystorev1.New()
	id, err := uuid.NewRandom()
	require.NoError(t, err)
	validatingKey, err := dilithium.RandKey()
	require.NoError(t, err)
	pubKey := validatingKey.PublicKey().Marshal()
	cryptoFields, err := encryptor.Encrypt(validatingKey.Marshal(), password)
	require.NoError(t, err)
	return &keymanager.Keystore{
		Crypto:  cryptoFields,
		Pubkey:  fmt.Sprintf("%x", pubKey),
		ID:      id.String(),
		Version: encryptor.Version(),
		// Name:    encryptor.Name(),
	}, validatingKey
}

func setupRandomDir(t testing.TB) string {
	randDir := t.TempDir()
	require.NoError(t, os.MkdirAll(randDir, os.ModePerm))
	return randDir
}

func TestDecrypt(t *testing.T) {
	keystoresDir := setupRandomDir(t)
	keystore, privKey := createRandomKeystore(t, password)
	// We write a random keystore to a keystores directory.
	encodedKeystore, err := json.MarshalIndent(keystore, "", "\t")
	require.NoError(t, err)
	keystoreFilePath := filepath.Join(keystoresDir, "keystore.json")
	require.NoError(t, os.WriteFile(
		keystoreFilePath, encodedKeystore, params.BeaconIoConfig().ReadWritePermissions),
	)

	cliCtx := setupCliContext(t, &cliConfig{
		keystoresPath: keystoreFilePath,
		password:      password,
	})

	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// We attempt to decrypt the keystore file we just wrote to disk.
	require.NoError(t, decrypt(cliCtx))

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	// We capture output from stdout.
	os.Stdout = rescueStdout
	stringOutput := string(out)

	// We capture the results of stdout to check the public key and private keys
	// were both printed to stdout.
	assert.Equal(t, true, strings.Contains(stringOutput, keystore.Pubkey))
	assert.Equal(t, true, strings.Contains(stringOutput, fmt.Sprintf("%#x", privKey.Marshal())))
}

func TestEncrypt(t *testing.T) {
	keystoresDir := setupRandomDir(t)
	keystoreFilePath := filepath.Join(keystoresDir, "keystore.json")
	privKey, err := dilithium.RandKey()
	require.NoError(t, err)

	cliCtx := setupCliContext(t, &cliConfig{
		outputPath: keystoreFilePath,
		password:   password,
		seed:       fmt.Sprintf("%#x", privKey.Marshal()),
	})

	rescueStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	// We attempt to encrypt the secret key and save it to the output path.
	require.NoError(t, encrypt(cliCtx))

	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)

	// We capture output from stdout.
	os.Stdout = rescueStdout
	stringOutput := string(out)

	// We capture the results of stdout to check the public key was printed to stdout.
	res := strings.Contains(stringOutput, fmt.Sprintf("%x", privKey.PublicKey().Marshal()))
	assert.Equal(t, true, res)
}
