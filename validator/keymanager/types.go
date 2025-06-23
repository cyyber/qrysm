package keymanager

import (
	"context"
	"fmt"
	"strings"

	"github.com/theQRL/qrysm/async/event"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/crypto/dilithium"
	validatorpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1/validator-client"
	zondpbservice "github.com/theQRL/qrysm/proto/zond/service"
)

// IKeymanager defines a general keymanager interface for Qrysm wallets.
type IKeymanager interface {
	PublicKeysFetcher
	Signer
	KeyChangeSubscriber
	KeyStoreExtractor
	AccountLister
	Deleter
}

// KeysFetcher for validating private and public keys.
type KeysFetcher interface {
	FetchValidatingSeeds(ctx context.Context) ([][field_params.DilithiumSeedLength]byte, error)
	PublicKeysFetcher
}

// PublicKeysFetcher for validating public keys.
type PublicKeysFetcher interface {
	FetchValidatingPublicKeys(ctx context.Context) ([][field_params.DilithiumPubkeyLength]byte, error)
}

// Signer allows signing messages using a validator private key.
type Signer interface {
	Sign(context.Context, *validatorpb.SignRequest) (dilithium.Signature, error)
}

// Importer can import new keystores into the keymanager.
type Importer interface {
	ImportKeystores(
		ctx context.Context, keystores []*Keystore, passwords []string,
	) ([]*zondpbservice.ImportedKeystoreStatus, error)
}

// Deleter can delete keystores from the keymanager.
type Deleter interface {
	DeleteKeystores(ctx context.Context, publicKeys [][]byte) ([]*zondpbservice.DeletedKeystoreStatus, error)
}

// KeyChangeSubscriber allows subscribing to changes made to the underlying keys.
type KeyChangeSubscriber interface {
	SubscribeAccountChanges(pubKeysChan chan [][field_params.DilithiumPubkeyLength]byte) event.Subscription
}

// KeyStoreExtractor allows keys to be extracted from the keymanager.
type KeyStoreExtractor interface {
	ExtractKeystores(ctx context.Context, publicKeys []dilithium.PublicKey, password string) ([]*Keystore, error)
}

// PublicKeyAdder allows adding public keys to the keymanager.
// type PublicKeyAdder interface {
// 	AddPublicKeys(ctx context.Context, publicKeys [][field_params.DilithiumPubkeyLength]byte) ([]*zondpbservice.ImportedRemoteKeysStatus, error)
// }

// PublicKeyDeleter allows deleting public keys set in keymanager.
// type PublicKeyDeleter interface {
// 	DeletePublicKeys(ctx context.Context, publicKeys [][field_params.DilithiumPubkeyLength]byte) ([]*zondpbservice.DeletedRemoteKeysStatus, error)
// }

type ListKeymanagerAccountConfig struct {
	ShowDepositData          bool
	ShowPrivateKeys          bool
	WalletAccountsDir        string
	KeymanagerConfigFileName string
}

type AccountLister interface {
	ListKeymanagerAccounts(ctx context.Context, cfg ListKeymanagerAccountConfig) error
}

// Keystore json file representation as a Go struct.
type Keystore struct {
	Crypto      map[string]interface{} `json:"crypto"`
	ID          string                 `json:"uuid"`
	Pubkey      string                 `json:"pubkey"`
	Version     uint                   `json:"version"`
	Description string                 `json:"description"`
	Path        string                 `json:"path"`
}

// Kind defines an enum for either local, derived, or remote-signing
// keystores for Qrysm wallets.
type Kind int

const (
	// Local keymanager defines an on-disk, encrypted keystore-capable store.
	Local Kind = iota
	// Derived keymanager using a hierarchical-deterministic algorithm.
	// Derived
	// Web3Signer keymanager capable of signing data using a remote signer called Web3Signer.
	// Web3Signer
)

// @NOTE(rgeraldes24):
// QRL Version - https://github.com/theQRL/go-zond-wallet-encryptor-keystore/blob/main/decrypt.go#L87
// Eth Version - https://github.com/wealdtech/go-eth2-wallet-encryptor-keystorev1/blob/master/decrypt.go#L168
// IncorrectPasswordErrMsg defines a common error string representing an EIP-2335
// keystore password was incorrect.
// const IncorrectPasswordErrMsg = "invalid checksum"
const IncorrectPasswordErrMsg = "checksum mismatch"

// String marshals a keymanager kind to a string value.
func (k Kind) String() string {
	switch k {
	// case Derived:
	// 	return "derived"
	case Local:
		// TODO(#10181) need a safe way to migrate away from using direct.
		// function is used for directory creation, dangerous to change which may result in multiple directories.
		// multiple directories will cause the isValid function to fail in wallet.go
		// and may result in using a unintended wallet.
		return "direct"
	// case Web3Signer:
	// 	return "web3signer"
	default:
		return fmt.Sprintf("%d", int(k))
	}
}

// ParseKind from a raw string, returning a keymanager kind.
func ParseKind(k string) (Kind, error) {
	switch strings.ToLower(k) {
	// case "derived":
	// 	return Derived, nil
	case "direct", "imported", "local":
		return Local, nil
	// case "web3signer":
	// 	return Web3Signer, nil
	default:
		return 0, fmt.Errorf("%s is not an allowed keymanager", k)
	}
}
