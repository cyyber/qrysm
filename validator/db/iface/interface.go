// Package iface defines an interface for the validator database.
package iface

import (
	"context"
	"io"

	field_params "github.com/theQRL/qrysm/config/fieldparams"
	validatorServiceConfig "github.com/theQRL/qrysm/config/validator/service"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/monitoring/backup"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/validator/db/kv"
)

// Ensure the kv store implements the interface.
var _ = ValidatorDB(&kv.Store{})

// ValidatorDB defines the necessary methods for a Qrysm validator DB.
type ValidatorDB interface {
	io.Closer
	backup.BackupExporter
	DatabasePath() string
	ClearDB() error
	RunUpMigrations(ctx context.Context) error
	RunDownMigrations(ctx context.Context) error
	UpdatePublicKeysBuckets(publicKeys [][field_params.MLDSA87PubkeyLength]byte) error

	// Genesis information related methods.
	GenesisValidatorsRoot(ctx context.Context) ([]byte, error)
	SaveGenesisValidatorsRoot(ctx context.Context, genValRoot []byte) error

	// Proposer protection related methods.
	HighestSignedProposal(ctx context.Context, publicKey [field_params.MLDSA87PubkeyLength]byte) (primitives.Slot, bool, error)
	LowestSignedProposal(ctx context.Context, publicKey [field_params.MLDSA87PubkeyLength]byte) (primitives.Slot, bool, error)
	ProposalHistoryForPubKey(ctx context.Context, publicKey [field_params.MLDSA87PubkeyLength]byte) ([]*kv.Proposal, error)
	ProposalHistoryForSlot(ctx context.Context, publicKey [field_params.MLDSA87PubkeyLength]byte, slot primitives.Slot) ([32]byte, bool, error)
	SaveProposalHistoryForSlot(ctx context.Context, pubKey [field_params.MLDSA87PubkeyLength]byte, slot primitives.Slot, signingRoot []byte) error
	ProposedPublicKeys(ctx context.Context) ([][field_params.MLDSA87PubkeyLength]byte, error)

	// Attester protection related methods.
	// Methods to store and read blacklisted public keys from EIP-3076
	// slashing protection imports.
	EIPImportBlacklistedPublicKeys(ctx context.Context) ([][field_params.MLDSA87PubkeyLength]byte, error)
	SaveEIPImportBlacklistedPublicKeys(ctx context.Context, publicKeys [][field_params.MLDSA87PubkeyLength]byte) error
	SigningRootAtTargetEpoch(ctx context.Context, publicKey [field_params.MLDSA87PubkeyLength]byte, target primitives.Epoch) ([32]byte, error)
	LowestSignedTargetEpoch(ctx context.Context, publicKey [field_params.MLDSA87PubkeyLength]byte) (primitives.Epoch, bool, error)
	LowestSignedSourceEpoch(ctx context.Context, publicKey [field_params.MLDSA87PubkeyLength]byte) (primitives.Epoch, bool, error)
	AttestedPublicKeys(ctx context.Context) ([][field_params.MLDSA87PubkeyLength]byte, error)
	CheckSlashableAttestation(
		ctx context.Context, pubKey [field_params.MLDSA87PubkeyLength]byte, signingRoot [32]byte, att *qrysmpb.IndexedAttestation,
	) (kv.SlashingKind, error)
	SaveAttestationForPubKey(
		ctx context.Context, pubKey [field_params.MLDSA87PubkeyLength]byte, signingRoot [32]byte, att *qrysmpb.IndexedAttestation,
	) error
	SaveAttestationsForPubKey(
		ctx context.Context, pubKey [field_params.MLDSA87PubkeyLength]byte, signingRoots [][32]byte, atts []*qrysmpb.IndexedAttestation,
	) error
	AttestationHistoryForPubKey(
		ctx context.Context, pubKey [field_params.MLDSA87PubkeyLength]byte,
	) ([]*kv.AttestationRecord, error)

	// Graffiti ordered index related methods
	SaveGraffitiOrderedIndex(ctx context.Context, index uint64) error
	GraffitiOrderedIndex(ctx context.Context, fileHash [32]byte) (uint64, error)

	// ProposerSettings related methods
	ProposerSettings(context.Context) (*validatorServiceConfig.ProposerSettings, error)
	ProposerSettingsExists(ctx context.Context) (bool, error)
	UpdateProposerSettingsDefault(context.Context, *validatorServiceConfig.ProposerOption) error
	UpdateProposerSettingsForPubkey(context.Context, [field_params.MLDSA87PubkeyLength]byte, *validatorServiceConfig.ProposerOption) error
	SaveProposerSettings(ctx context.Context, settings *validatorServiceConfig.ProposerSettings) error
}
