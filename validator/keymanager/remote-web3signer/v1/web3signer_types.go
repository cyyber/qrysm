// Package v1 defines mappings of types as defined by the web3signer official specification for its v1 version i.e. /api/v1/zond2
/* Web3Signer Specs are found by searching Consensys' Web3Signer API specification*/
package v1

import (
	"github.com/theQRL/go-zond/common/hexutil"
)

// AggregationSlotSignRequest is a request object for web3signer sign api.
type AggregationSlotSignRequest struct {
	Type            string           `json:"type" validate:"required"`
	ForkInfo        *ForkInfo        `json:"fork_info" validate:"required"`
	SigningRoot     hexutil.Bytes    `json:"signingRoot"`
	AggregationSlot *AggregationSlot `json:"aggregation_slot" validate:"required"`
}

// AggregateAndProofSignRequest is a request object for web3signer sign api.
type AggregateAndProofSignRequest struct {
	Type              string             `json:"type" validate:"required"`
	ForkInfo          *ForkInfo          `json:"fork_info" validate:"required"`
	SigningRoot       hexutil.Bytes      `json:"signingRoot"`
	AggregateAndProof *AggregateAndProof `json:"aggregate_and_proof" validate:"required"`
}

// AttestationSignRequest is a request object for web3signer sign api.
type AttestationSignRequest struct {
	Type        string           `json:"type" validate:"required"`
	ForkInfo    *ForkInfo        `json:"fork_info" validate:"required"`
	SigningRoot hexutil.Bytes    `json:"signingRoot"`
	Attestation *AttestationData `json:"attestation" validate:"required"`
}

// BlockSignRequest is a request object for web3signer sign api.
type BlockSignRequest struct {
	Type        string        `json:"type" validate:"required"`
	ForkInfo    *ForkInfo     `json:"fork_info" validate:"required"`
	SigningRoot hexutil.Bytes `json:"signingRoot"`
	Block       *BeaconBlock  `json:"block" validate:"required"`
}

// BlockBlindedSignRequest is a request object for web3signer sign api
type BlockBlindedSignRequest struct {
	Type        string              `json:"type" validate:"required"`
	ForkInfo    *ForkInfo           `json:"fork_info" validate:"required"`
	SigningRoot hexutil.Bytes       `json:"signingRoot"`
	BeaconBlock *BeaconBlockBlinded `json:"beacon_block" validate:"required"`
}

// DepositSignRequest Not currently supported by Qrysm.
// DepositSignRequest is a request object for web3signer sign api.

// RandaoRevealSignRequest is a request object for web3signer sign api.
type RandaoRevealSignRequest struct {
	Type         string        `json:"type" validate:"required"`
	ForkInfo     *ForkInfo     `json:"fork_info" validate:"required"`
	SigningRoot  hexutil.Bytes `json:"signingRoot"`
	RandaoReveal *RandaoReveal `json:"randao_reveal" validate:"required"`
}

// VoluntaryExitSignRequest is a request object for web3signer sign api.
type VoluntaryExitSignRequest struct {
	Type          string         `json:"type" validate:"required"`
	ForkInfo      *ForkInfo      `json:"fork_info"`
	SigningRoot   hexutil.Bytes  `json:"signingRoot" validate:"required"`
	VoluntaryExit *VoluntaryExit `json:"voluntary_exit" validate:"required"`
}

// SyncCommitteeMessageSignRequest is a request object for web3signer sign api.
type SyncCommitteeMessageSignRequest struct {
	Type                 string                `json:"type" validate:"required"`
	ForkInfo             *ForkInfo             `json:"fork_info" validate:"required"`
	SigningRoot          hexutil.Bytes         `json:"signingRoot"`
	SyncCommitteeMessage *SyncCommitteeMessage `json:"sync_committee_message" validate:"required"`
}

// SyncCommitteeSelectionProofSignRequest is a request object for web3signer sign api.
type SyncCommitteeSelectionProofSignRequest struct {
	Type                        string                       `json:"type" validate:"required"`
	ForkInfo                    *ForkInfo                    `json:"fork_info" validate:"required"`
	SigningRoot                 hexutil.Bytes                `json:"signingRoot"`
	SyncAggregatorSelectionData *SyncAggregatorSelectionData `json:"sync_aggregator_selection_data" validate:"required"`
}

// SyncCommitteeContributionAndProofSignRequest is a request object for web3signer sign api.
type SyncCommitteeContributionAndProofSignRequest struct {
	Type                 string                `json:"type" validate:"required"`
	ForkInfo             *ForkInfo             `json:"fork_info" validate:"required"`
	SigningRoot          hexutil.Bytes         `json:"signingRoot"`
	ContributionAndProof *ContributionAndProof `json:"contribution_and_proof" validate:"required"`
}

// ValidatorRegistrationSignRequest a request object for web3signer sign api.
type ValidatorRegistrationSignRequest struct {
	Type                  string                 `json:"type" validate:"required"`
	SigningRoot           hexutil.Bytes          `json:"signingRoot"`
	ValidatorRegistration *ValidatorRegistration `json:"validator_registration" validate:"required"`
}

////////////////////////////////////////////////////////////////////////////////
// sub properties of Sign Requests /////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////

// ForkInfo a sub property object of the Sign request
type ForkInfo struct {
	Fork                  *Fork         `json:"fork"`
	GenesisValidatorsRoot hexutil.Bytes `json:"genesis_validators_root"`
}

// Fork a sub property of ForkInfo.
type Fork struct {
	PreviousVersion hexutil.Bytes `json:"previous_version"`
	CurrentVersion  hexutil.Bytes `json:"current_version"`
	Epoch           string        `json:"epoch"` /*uint64*/
}

// AggregationSlot a sub property of AggregationSlotSignRequest.
type AggregationSlot struct {
	Slot string `json:"slot"`
}

// AggregateAndProof a sub property of AggregateAndProofSignRequest.
type AggregateAndProof struct {
	AggregatorIndex string        `json:"aggregator_index"` /* uint64 */
	Aggregate       *Attestation  `json:"aggregate"`
	SelectionProof  hexutil.Bytes `json:"selection_proof"` /* 4595 bytes */
}

// Attestation a sub property of AggregateAndProofSignRequest.
type Attestation struct {
	ParticipationBits hexutil.Bytes    `json:"aggregation_bits"` /*hex bitlist*/
	Data              *AttestationData `json:"data"`
	Signatures        []hexutil.Bytes  `json:"signatures"`
}

// AttestationData a sub property of Attestation.
type AttestationData struct {
	Slot            string        `json:"slot"`  /* uint64 */
	Index           string        `json:"index"` /* uint64 */ // Qrysm uses CommitteeIndex but web3signer uses index.
	BeaconBlockRoot hexutil.Bytes `json:"beacon_block_root"`
	Source          *Checkpoint   `json:"source"`
	Target          *Checkpoint   `json:"target"`
}

// Checkpoint a sub property of AttestationData.
type Checkpoint struct {
	Epoch string `json:"epoch"`
	Root  string `json:"root"`
}

// BeaconBlock a sub property of BeaconBlockBlock.
type BeaconBlock struct {
	Slot          string           `json:"slot"`           /* uint64 */
	ProposerIndex string           `json:"proposer_index"` /* uint64 */
	ParentRoot    hexutil.Bytes    `json:"parent_root"`
	StateRoot     hexutil.Bytes    `json:"state_root"`
	Body          *BeaconBlockBody `json:"body"`
}

// BeaconBlockBody a sub property of BeaconBlock.
type BeaconBlockBody struct {
	RandaoReveal      hexutil.Bytes          `json:"randao_reveal"`
	Zond1Data         *Zond1Data             `json:"zond1_data"`
	Graffiti          hexutil.Bytes          `json:"graffiti"` // 32 bytes
	ProposerSlashings []*ProposerSlashing    `json:"proposer_slashings"`
	AttesterSlashings []*AttesterSlashing    `json:"attester_slashings"`
	Attestations      []*Attestation         `json:"attestations"`
	Deposits          []*Deposit             `json:"deposits"`
	VoluntaryExits    []*SignedVoluntaryExit `json:"voluntary_exits"`
}

// Zond1Data a sub property of BeaconBlockBody.
type Zond1Data struct {
	DepositRoot  hexutil.Bytes `json:"deposit_root"`
	DepositCount string        `json:"deposit_count"` /* uint64 */
	BlockHash    hexutil.Bytes `json:"block_hash"`
}

// ProposerSlashing a sub property of BeaconBlockBody.
type ProposerSlashing struct {
	// Qrysm uses Header_1 but web3signer uses signed_header_1.
	Signedheader1 *SignedBeaconBlockHeader `json:"signed_header_1"`
	// Qrysm uses Header_2 but web3signer uses signed_header_2.
	Signedheader2 *SignedBeaconBlockHeader `json:"signed_header_2"`
}

// SignedBeaconBlockHeader is a sub property of ProposerSlashing.
type SignedBeaconBlockHeader struct {
	Message   *BeaconBlockHeader `json:"message"`
	Signature hexutil.Bytes      `json:"signature"`
}

// BeaconBlockHeader is a sub property of SignedBeaconBlockHeader.
type BeaconBlockHeader struct {
	Slot          string        `json:"slot"`           /* uint64 */
	ProposerIndex string        `json:"proposer_index"` /* uint64 */
	ParentRoot    hexutil.Bytes `json:"parent_root"`    /* Hash32 */
	StateRoot     hexutil.Bytes `json:"state_root"`     /* Hash32 */
	BodyRoot      hexutil.Bytes `json:"body_root"`      /* Hash32 */
}

// AttesterSlashing a sub property of BeaconBlockBody.
type AttesterSlashing struct {
	Attestation1 *IndexedAttestation `json:"attestation_1"`
	Attestation2 *IndexedAttestation `json:"attestation_2"`
}

// IndexedAttestation a sub property of AttesterSlashing.
type IndexedAttestation struct {
	AttestingIndices []string         `json:"attesting_indices"` /* uint64[] */
	Data             *AttestationData `json:"data"`
	Signatures       []hexutil.Bytes  `json:"signatures"`
}

// Deposit a sub property of DepositSignRequest.
type Deposit struct {
	Proof []string     `json:"proof"`
	Data  *DepositData `json:"data"`
}

// DepositData a sub property of Deposit.
// DepositData Qrysm uses Deposit_data instead of DepositData which is inconsistent naming
type DepositData struct {
	PublicKey             hexutil.Bytes `json:"pubkey"`
	WithdrawalCredentials hexutil.Bytes `json:"withdrawal_credentials"`
	Amount                string        `json:"amount"` /* uint64 */
	Signature             hexutil.Bytes `json:"signature"`
}

// SignedVoluntaryExit is a sub property of BeaconBlockBody.
type SignedVoluntaryExit struct {
	// Qrysm uses Exit instead of Message
	Message   *VoluntaryExit `json:"message"`
	Signature hexutil.Bytes  `json:"signature"`
}

// VoluntaryExit a sub property of SignedVoluntaryExit.
type VoluntaryExit struct {
	Epoch          string `json:"epoch"`           /* uint64 */
	ValidatorIndex string `json:"validator_index"` /* uint64 */
}

// BeaconBlockBlinded a field of BlockBlindedSignRequest.
type BeaconBlockBlinded struct {
	Version     string             `json:"version" enum:"true"`
	BlockHeader *BeaconBlockHeader `json:"block_header"`
}

// BeaconBlockBlock a sub property of BlockSignRequest.
type BeaconBlockBlock struct {
	Version       string           `json:"version" enum:"true"`
	Slot          string           `json:"slot"`           /* uint64 */
	ProposerIndex string           `json:"proposer_index"` /* uint64 */
	ParentRoot    hexutil.Bytes    `json:"parent_root"`
	StateRoot     hexutil.Bytes    `json:"state_root"`
	Body          *BeaconBlockBody `json:"body"`
}

// RandaoReveal a sub property of RandaoRevealSignRequest.
type RandaoReveal struct {
	Epoch string `json:"epoch"` /* uint64 */
}

// SyncCommitteeMessage a sub property of SyncCommitteeSignRequest.
type SyncCommitteeMessage struct {
	BeaconBlockRoot hexutil.Bytes `json:"beacon_block_root"` /* Hash32 */
	Slot            string        `json:"slot"`              /* uint64 */
	// Qrysm uses BlockRoot instead of BeaconBlockRoot and has the following extra properties : ValidatorIndex, Signature
}

// SyncAggregatorSelectionData a sub property of SyncAggregatorSelectionSignRequest.
type SyncAggregatorSelectionData struct {
	Slot              string `json:"slot"`               /* uint64 */
	SubcommitteeIndex string `json:"subcommittee_index"` /* uint64 */
}

// ContributionAndProof a sub property of AggregatorSelectionSignRequest.
type ContributionAndProof struct {
	AggregatorIndex string                     `json:"aggregator_index"` /* uint64 */
	SelectionProof  hexutil.Bytes              `json:"selection_proof"`  /* 4595 byte hexadecimal */
	Contribution    *SyncCommitteeContribution `json:"contribution"`
}

// SyncCommitteeContribution a sub property of AggregatorSelectionSignRequest.
type SyncCommitteeContribution struct {
	Slot              string          `json:"slot"`               /* uint64 */
	BeaconBlockRoot   hexutil.Bytes   `json:"beacon_block_root"`  /* Hash32 */ // Qrysm uses BlockRoot instead of BeaconBlockRoot
	SubcommitteeIndex string          `json:"subcommittee_index"` /* uint64 */
	ParticipationBits hexutil.Bytes   `json:"aggregation_bits"`   /* SSZ hexadecimal string */
	Signatures        []hexutil.Bytes `json:"signatures"`
}

// ValidatorRegistration a sub property of ValidatorRegistrationSignRequest
type ValidatorRegistration struct {
	FeeRecipient hexutil.Bytes `json:"fee_recipient" validate:"required"` /* 42 hexadecimal string */
	GasLimit     string        `json:"gas_limit" validate:"required"`     /* uint64 */
	Timestamp    string        `json:"timestamp" validate:"required"`     /* uint64 */
	Pubkey       hexutil.Bytes `json:"pubkey"  validate:"required"`       /* dilithium hexadecimal string */
}

////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////
