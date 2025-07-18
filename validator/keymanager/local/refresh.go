package local

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
	"github.com/pkg/errors"
	keystorev1 "github.com/theQRL/go-zond-wallet-encryptor-keystore"
	"github.com/theQRL/qrysm/async"
	"github.com/theQRL/qrysm/config/features"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/crypto/dilithium"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	"github.com/theQRL/qrysm/io/file"
	"github.com/theQRL/qrysm/validator/keymanager"
)

// Listen for changes to the all-accounts.keystore.json file in our wallet
// to load in new keys we observe into our keymanager. This uses the fsnotify
// library to listen for file-system changes and debounces these events to
// ensure we can handle thousands of events fired in a short time-span.
func (km *Keymanager) listenForAccountChanges(ctx context.Context) {
	debounceFileChangesInterval := features.Get().KeystoreImportDebounceInterval
	accountsFilePath := filepath.Join(km.wallet.AccountsDir(), AccountsPath, AccountsKeystoreFileName)
	if !file.FileExists(accountsFilePath) {
		return
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).Error("Could not initialize file watcher")
		return
	}
	defer func() {
		if err := watcher.Close(); err != nil {
			log.WithError(err).Error("Could not close file watcher")
		}
	}()
	if err := watcher.Add(accountsFilePath); err != nil {
		log.WithError(err).Errorf("Could not add file %s to file watcher", accountsFilePath)
		return
	}
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	fileChangesChan := make(chan interface{}, 100)
	defer close(fileChangesChan)

	// We debounce events sent over the file changes channel by an interval
	// to ensure we are not overwhelmed by a ton of events fired over the channel in
	// a short span of time.
	go async.Debounce(ctx, debounceFileChangesInterval, fileChangesChan, func(event interface{}) {
		ev, ok := event.(fsnotify.Event)
		if !ok {
			log.Errorf("Type %T is not a valid file system event", event)
			return
		}
		fileBytes, err := os.ReadFile(ev.Name)
		if err != nil {
			log.WithError(err).Errorf("Could not read file at path: %s", ev.Name)
			return
		}
		if fileBytes == nil {
			log.WithError(err).Errorf("Loaded in an empty file: %s", ev.Name)
			return
		}
		accountsKeystore := &AccountsKeystoreRepresentation{}
		if err := json.Unmarshal(fileBytes, accountsKeystore); err != nil {
			log.WithError(
				err,
			).Errorf("Could not read valid, EIP-2335 keystore json file at path: %s", ev.Name)
			return
		}
		if err := km.reloadAccountsFromKeystore(accountsKeystore); err != nil {
			log.WithError(
				err,
			).Error("Could not replace the accounts store from keystore file")
		}
	})
	for {
		select {
		case event := <-watcher.Events:
			// If a file was modified, we attempt to read that file
			// and parse it into our accounts store.
			fileChangesChan <- event
		case err := <-watcher.Errors:
			log.WithError(err).Errorf("Could not watch for file changes for: %s", accountsFilePath)
		case <-ctx.Done():
			return
		}
	}
}

// Replaces the accounts store struct in the local keymanager with
// the contents of a keystore file by decrypting it with the accounts password.
func (km *Keymanager) reloadAccountsFromKeystore(keystore *AccountsKeystoreRepresentation) error {
	decryptor := keystorev1.New()
	encodedAccounts, err := decryptor.Decrypt(keystore.Crypto, km.wallet.Password())
	if err != nil {
		return errors.Wrap(err, "could not decrypt keystore file")
	}
	newAccountsStore := &accountStore{}
	if err := json.Unmarshal(encodedAccounts, newAccountsStore); err != nil {
		return err
	}
	if len(newAccountsStore.PublicKeys) != len(newAccountsStore.Seeds) {
		return errors.New("number of public and private keys in keystore do not match")
	}
	pubKeys := make([][field_params.DilithiumPubkeyLength]byte, len(newAccountsStore.PublicKeys))
	for i := 0; i < len(newAccountsStore.Seeds); i++ {
		privKey, err := dilithium.SecretKeyFromSeed(newAccountsStore.Seeds[i])
		if err != nil {
			return errors.Wrap(err, "could not initialize private key")
		}
		pubKeyBytes := privKey.PublicKey().Marshal()
		pubKeys[i] = bytesutil.ToBytes2592(pubKeyBytes)
	}
	km.accountsStore = newAccountsStore
	if err := km.initializeKeysCachesFromKeystore(); err != nil {
		return err
	}
	log.Info(keymanager.KeysReloaded)
	km.accountsChangedFeed.Send(pubKeys)
	return nil
}
