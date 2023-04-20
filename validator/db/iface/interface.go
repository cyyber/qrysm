// Package iface defines an interface for the validator database.
package iface

import (
	"context"
	"io"

	"github.com/prysmaticlabs/prysm/v4/consensus-types/primitives"
	"github.com/prysmaticlabs/prysm/v4/monitoring/backup"
	ethpb "github.com/prysmaticlabs/prysm/v4/proto/prysm/v1alpha1"
	"github.com/prysmaticlabs/prysm/v4/validator/db/kv"
	dilithium2 "github.com/theQRL/go-qrllib/dilithium"
)

// Ensure the kv store implements the interface.
var _ = ValidatorDB(&kv.Store{})

// ValidatorDB defines the necessary methods for a Prysm validator DB.
type ValidatorDB interface {
	io.Closer
	backup.BackupExporter
	DatabasePath() string
	ClearDB() error
	RunUpMigrations(ctx context.Context) error
	RunDownMigrations(ctx context.Context) error
	UpdatePublicKeysBuckets(publicKeys [][dilithium2.CryptoPublicKeyBytes]byte) error

	// Genesis information related methods.
	GenesisValidatorsRoot(ctx context.Context) ([]byte, error)
	SaveGenesisValidatorsRoot(ctx context.Context, genValRoot []byte) error

	// Proposer protection related methods.
	HighestSignedProposal(ctx context.Context, publicKey [dilithium2.CryptoPublicKeyBytes]byte) (primitives.Slot, bool, error)
	LowestSignedProposal(ctx context.Context, publicKey [dilithium2.CryptoPublicKeyBytes]byte) (primitives.Slot, bool, error)
	ProposalHistoryForPubKey(ctx context.Context, publicKey [dilithium2.CryptoPublicKeyBytes]byte) ([]*kv.Proposal, error)
	ProposalHistoryForSlot(ctx context.Context, publicKey [dilithium2.CryptoPublicKeyBytes]byte, slot primitives.Slot) ([32]byte, bool, error)
	SaveProposalHistoryForSlot(ctx context.Context, pubKey [dilithium2.CryptoPublicKeyBytes]byte, slot primitives.Slot, signingRoot []byte) error
	ProposedPublicKeys(ctx context.Context) ([][dilithium2.CryptoPublicKeyBytes]byte, error)

	// Attester protection related methods.
	// Methods to store and read blacklisted public keys from EIP-3076
	// slashing protection imports.
	EIPImportBlacklistedPublicKeys(ctx context.Context) ([][dilithium2.CryptoPublicKeyBytes]byte, error)
	SaveEIPImportBlacklistedPublicKeys(ctx context.Context, publicKeys [][dilithium2.CryptoPublicKeyBytes]byte) error
	SigningRootAtTargetEpoch(ctx context.Context, publicKey [dilithium2.CryptoPublicKeyBytes]byte, target primitives.Epoch) ([32]byte, error)
	LowestSignedTargetEpoch(ctx context.Context, publicKey [dilithium2.CryptoPublicKeyBytes]byte) (primitives.Epoch, bool, error)
	LowestSignedSourceEpoch(ctx context.Context, publicKey [dilithium2.CryptoPublicKeyBytes]byte) (primitives.Epoch, bool, error)
	AttestedPublicKeys(ctx context.Context) ([][dilithium2.CryptoPublicKeyBytes]byte, error)
	CheckSlashableAttestation(
		ctx context.Context, pubKey [dilithium2.CryptoPublicKeyBytes]byte, signingRoot [32]byte, att *ethpb.IndexedAttestation,
	) (kv.SlashingKind, error)
	SaveAttestationForPubKey(
		ctx context.Context, pubKey [dilithium2.CryptoPublicKeyBytes]byte, signingRoot [32]byte, att *ethpb.IndexedAttestation,
	) error
	SaveAttestationsForPubKey(
		ctx context.Context, pubKey [dilithium2.CryptoPublicKeyBytes]byte, signingRoots [][32]byte, atts []*ethpb.IndexedAttestation,
	) error
	AttestationHistoryForPubKey(
		ctx context.Context, pubKey [dilithium2.CryptoPublicKeyBytes]byte,
	) ([]*kv.AttestationRecord, error)

	// Graffiti ordered index related methods
	SaveGraffitiOrderedIndex(ctx context.Context, index uint64) error
	GraffitiOrderedIndex(ctx context.Context, fileHash [32]byte) (uint64, error)
}
