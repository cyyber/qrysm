package apimiddleware

import (
	"strings"

	"github.com/theQRL/qrysm/api/gateway/apimiddleware"
	"github.com/theQRL/qrysm/beacon-chain/rpc/zond/shared"
	zondpbv1 "github.com/theQRL/qrysm/proto/zond/v1"
)

//----------------
// Requests and responses.
//----------------

// WeakSubjectivityResponse is used to marshal/unmarshal the response for the
// /zond/v1/beacon/weak_subjectivity endpoint.
type WeakSubjectivityResponse struct {
	Data *struct {
		Checkpoint *CheckpointJson `json:"ws_checkpoint"`
		StateRoot  string          `json:"state_root" hex:"true"`
	} `json:"data"`
}

type StateRootResponseJson struct {
	Data                *StateRootResponse_StateRootJson `json:"data"`
	ExecutionOptimistic bool                             `json:"execution_optimistic"`
	Finalized           bool                             `json:"finalized"`
}

type StateRootResponse_StateRootJson struct {
	StateRoot string `json:"root" hex:"true"`
}

type SyncCommitteesResponseJson struct {
	Data                *SyncCommitteeValidatorsJson `json:"data"`
	ExecutionOptimistic bool                         `json:"execution_optimistic"`
	Finalized           bool                         `json:"finalized"`
}

type RandaoResponseJson struct {
	Data *struct {
		Randao string `json:"randao" hex:"true"`
	} `json:"data"`
	ExecutionOptimistic bool `json:"execution_optimistic"`
	Finalized           bool `json:"finalized"`
}

type BlockResponseJson struct {
	Version             string                          `json:"version" enum:"true"`
	Data                *SignedBeaconBlockContainerJson `json:"data"`
	ExecutionOptimistic bool                            `json:"execution_optimistic"`
	Finalized           bool                            `json:"finalized"`
}

type BlindedBlockResponseJson struct {
	Version             string                                 `json:"version" enum:"true"`
	Data                *SignedBlindedBeaconBlockContainerJson `json:"data"`
	ExecutionOptimistic bool                                   `json:"execution_optimistic"`
	Finalized           bool                                   `json:"finalized"`
}

type BlockRootResponseJson struct {
	Data                *BlockRootContainerJson `json:"data"`
	ExecutionOptimistic bool                    `json:"execution_optimistic"`
	Finalized           bool                    `json:"finalized"`
}

type BlockAttestationsResponseJson struct {
	Data                []*AttestationJson `json:"data"`
	ExecutionOptimistic bool               `json:"execution_optimistic"`
	Finalized           bool               `json:"finalized"`
}

type AttesterSlashingsPoolResponseJson struct {
	Data []*AttesterSlashingJson `json:"data"`
}

type ProposerSlashingsPoolResponseJson struct {
	Data []*ProposerSlashingJson `json:"data"`
}

type DilithiumToExecutionChangesPoolResponseJson struct {
	Data []*SignedDilithiumToExecutionChangeJson `json:"data"`
}

type IdentityResponseJson struct {
	Data *IdentityJson `json:"data"`
}

type PeersResponseJson struct {
	Data []*PeerJson `json:"data"`
}

type PeerResponseJson struct {
	Data *PeerJson `json:"data"`
}

type PeerCountResponseJson struct {
	Data PeerCountResponse_PeerCountJson `json:"data"`
}

type PeerCountResponse_PeerCountJson struct {
	Disconnected  string `json:"disconnected"`
	Connecting    string `json:"connecting"`
	Connected     string `json:"connected"`
	Disconnecting string `json:"disconnecting"`
}

type VersionResponseJson struct {
	Data *VersionJson `json:"data"`
}

type SyncingResponseJson struct {
	Data *shared.SyncDetails `json:"data"`
}

type BeaconStateResponseJson struct {
	Version             string                    `json:"version" enum:"true"`
	Data                *BeaconStateContainerJson `json:"data"`
	ExecutionOptimistic bool                      `json:"execution_optimistic"`
	Finalized           bool                      `json:"finalized"`
}

type ForkChoiceHeadsResponseJson struct {
	Data []*ForkChoiceHeadJson `json:"data"`
}

type ForkScheduleResponseJson struct {
	Data []*ForkJson `json:"data"`
}

type DepositContractResponseJson struct {
	Data *DepositContractJson `json:"data"`
}

type SpecResponseJson struct {
	Data interface{} `json:"data"`
}

type ProduceBlockResponseJson struct {
	Version string                    `json:"version" enum:"true"`
	Data    *BeaconBlockContainerJson `json:"data"`
}

type ProduceBlindedBlockResponseJson struct {
	Version string                           `json:"version" enum:"true"`
	Data    *BlindedBeaconBlockContainerJson `json:"data"`
}

type AggregateAttestationResponseJson struct {
	Data *AttestationJson `json:"data"`
}

type BeaconCommitteeSubscribeJson struct {
	ValidatorIndex   string `json:"validator_index"`
	CommitteeIndex   string `json:"committee_index"`
	CommitteesAtSlot string `json:"committees_at_slot"`
	Slot             string `json:"slot"`
	IsAggregator     bool   `json:"is_aggregator"`
}

type ProduceSyncCommitteeContributionResponseJson struct {
	Data *SyncCommitteeContributionJson `json:"data"`
}

type ForkChoiceNodeResponseJson struct {
	Slot               string                       `json:"slot"`
	BlockRoot          string                       `json:"block_root" hex:"true"`
	ParentRoot         string                       `json:"parent_root" hex:"true"`
	JustifiedEpoch     string                       `json:"justified_epoch"`
	FinalizedEpoch     string                       `json:"finalized_epoch"`
	Weight             string                       `json:"weight"`
	Validity           string                       `json:"validity" enum:"true"`
	ExecutionBlockHash string                       `json:"execution_block_hash" hex:"true"`
	ExtraData          *ForkChoiceNodeExtraDataJson `json:"extra_data"`
}

type ForkChoiceNodeExtraDataJson struct {
	UnrealizedJustifiedEpoch string `json:"unrealized_justified_epoch"`
	UnrealizedFinalizedEpoch string `json:"unrealized_finalized_epoch"`
	Balance                  string `json:"balance"`
	ExecutionOptimistic      bool   `json:"execution_optimistic"`
	TimeStamp                string `json:"timestamp"`
}

type ForkChoiceResponseJson struct {
	JustifiedCheckpoint *CheckpointJson                  `json:"justified_checkpoint"`
	FinalizedCheckpoint *CheckpointJson                  `json:"finalized_checkpoint"`
	ForkChoiceNodes     []*ForkChoiceNodeResponseJson    `json:"fork_choice_nodes"`
	ExtraData           *ForkChoiceResponseExtraDataJson `json:"extra_data"`
}

type ForkChoiceResponseExtraDataJson struct {
	BestJustifiedCheckpoint       *CheckpointJson `json:"best_justified_checkpoint"`
	UnrealizedJustifiedCheckpoint *CheckpointJson `json:"unrealized_justified_checkpoint"`
	UnrealizedFinalizedCheckpoint *CheckpointJson `json:"unrealized_finalized_checkpoint"`
	ProposerBoostRoot             string          `json:"proposer_boost_root" hex:"true"`
	PreviousProposerBoostRoot     string          `json:"previous_proposer_boost_root" hex:"true"`
	HeadRoot                      string          `json:"head_root" hex:"true"`
}

//----------------
// Reusable types.
//----------------

type CheckpointJson struct {
	Epoch string `json:"epoch"`
	Root  string `json:"root" hex:"true"`
}

type BlockRootContainerJson struct {
	Root string `json:"root" hex:"true"`
}

type SignedBeaconBlockJson struct {
	Message   *BeaconBlockJson `json:"message"`
	Signature string           `json:"signature" hex:"true"`
}

type BeaconBlockJson struct {
	Slot          string               `json:"slot"`
	ProposerIndex string               `json:"proposer_index"`
	ParentRoot    string               `json:"parent_root" hex:"true"`
	StateRoot     string               `json:"state_root" hex:"true"`
	Body          *BeaconBlockBodyJson `json:"body"`
}

type BeaconBlockBodyJson struct {
	RandaoReveal      string                     `json:"randao_reveal" hex:"true"`
	Eth1Data          *Eth1DataJson              `json:"eth1_data"`
	Graffiti          string                     `json:"graffiti" hex:"true"`
	ProposerSlashings []*ProposerSlashingJson    `json:"proposer_slashings"`
	AttesterSlashings []*AttesterSlashingJson    `json:"attester_slashings"`
	Attestations      []*AttestationJson         `json:"attestations"`
	Deposits          []*DepositJson             `json:"deposits"`
	VoluntaryExits    []*SignedVoluntaryExitJson `json:"voluntary_exits"`
}

type SignedBeaconBlockContainerJson struct {
	CapellaBlock *BeaconBlockCapellaJson `json:"capella_block"`
	Signature    string                  `json:"signature" hex:"true"`
}

type SignedBlindedBeaconBlockContainerJson struct {
	CapellaBlock *BlindedBeaconBlockCapellaJson `json:"capella_block"`
	Signature    string                         `json:"signature" hex:"true"`
}

type SignedBlindedBeaconBlockContentsContainerJson struct {
	CapellaBlock *SignedBlindedBeaconBlockCapellaJson `json:"capella_block"`
}

type BeaconBlockContainerJson struct {
	CapellaBlock *BeaconBlockCapellaJson `json:"capella_block"`
}

type BlindedBeaconBlockContainerJson struct {
	CapellaBlock *BlindedBeaconBlockCapellaJson `json:"capella_block"`
}

type SignedBeaconBlockCapellaJson struct {
	Message   *BeaconBlockCapellaJson `json:"message"`
	Signature string                  `json:"signature" hex:"true"`
}

type SignedBlindedBeaconBlockCapellaJson struct {
	Message   *BlindedBeaconBlockCapellaJson `json:"message"`
	Signature string                         `json:"signature" hex:"true"`
}

type BeaconBlockCapellaJson struct {
	Slot          string                      `json:"slot"`
	ProposerIndex string                      `json:"proposer_index"`
	ParentRoot    string                      `json:"parent_root" hex:"true"`
	StateRoot     string                      `json:"state_root" hex:"true"`
	Body          *BeaconBlockBodyCapellaJson `json:"body"`
}

type BlindedBeaconBlockCapellaJson struct {
	Slot          string                             `json:"slot"`
	ProposerIndex string                             `json:"proposer_index"`
	ParentRoot    string                             `json:"parent_root" hex:"true"`
	StateRoot     string                             `json:"state_root" hex:"true"`
	Body          *BlindedBeaconBlockBodyCapellaJson `json:"body"`
}

type BeaconBlockBodyCapellaJson struct {
	RandaoReveal                string                                  `json:"randao_reveal" hex:"true"`
	Eth1Data                    *Eth1DataJson                           `json:"eth1_data"`
	Graffiti                    string                                  `json:"graffiti" hex:"true"`
	ProposerSlashings           []*ProposerSlashingJson                 `json:"proposer_slashings"`
	AttesterSlashings           []*AttesterSlashingJson                 `json:"attester_slashings"`
	Attestations                []*AttestationJson                      `json:"attestations"`
	Deposits                    []*DepositJson                          `json:"deposits"`
	VoluntaryExits              []*SignedVoluntaryExitJson              `json:"voluntary_exits"`
	SyncAggregate               *SyncAggregateJson                      `json:"sync_aggregate"`
	ExecutionPayload            *ExecutionPayloadCapellaJson            `json:"execution_payload"`
	DilithiumToExecutionChanges []*SignedDilithiumToExecutionChangeJson `json:"dilithium_to_execution_changes"`
}

type BlindedBeaconBlockBodyCapellaJson struct {
	RandaoReveal                string                                  `json:"randao_reveal" hex:"true"`
	Eth1Data                    *Eth1DataJson                           `json:"eth1_data"`
	Graffiti                    string                                  `json:"graffiti" hex:"true"`
	ProposerSlashings           []*ProposerSlashingJson                 `json:"proposer_slashings"`
	AttesterSlashings           []*AttesterSlashingJson                 `json:"attester_slashings"`
	Attestations                []*AttestationJson                      `json:"attestations"`
	Deposits                    []*DepositJson                          `json:"deposits"`
	VoluntaryExits              []*SignedVoluntaryExitJson              `json:"voluntary_exits"`
	SyncAggregate               *SyncAggregateJson                      `json:"sync_aggregate"`
	ExecutionPayloadHeader      *ExecutionPayloadHeaderCapellaJson      `json:"execution_payload_header"`
	DilithiumToExecutionChanges []*SignedDilithiumToExecutionChangeJson `json:"dilithium_to_execution_changes"`
}

type ExecutionPayloadCapellaJson struct {
	ParentHash    string            `json:"parent_hash" hex:"true"`
	FeeRecipient  string            `json:"fee_recipient" hex:"true"`
	StateRoot     string            `json:"state_root" hex:"true"`
	ReceiptsRoot  string            `json:"receipts_root" hex:"true"`
	LogsBloom     string            `json:"logs_bloom" hex:"true"`
	PrevRandao    string            `json:"prev_randao" hex:"true"`
	BlockNumber   string            `json:"block_number"`
	GasLimit      string            `json:"gas_limit"`
	GasUsed       string            `json:"gas_used"`
	TimeStamp     string            `json:"timestamp"`
	ExtraData     string            `json:"extra_data" hex:"true"`
	BaseFeePerGas string            `json:"base_fee_per_gas" uint256:"true"`
	BlockHash     string            `json:"block_hash" hex:"true"`
	Transactions  []string          `json:"transactions" hex:"true"`
	Withdrawals   []*WithdrawalJson `json:"withdrawals"`
}

type ExecutionPayloadHeaderCapellaJson struct {
	ParentHash       string `json:"parent_hash" hex:"true"`
	FeeRecipient     string `json:"fee_recipient" hex:"true"`
	StateRoot        string `json:"state_root" hex:"true"`
	ReceiptsRoot     string `json:"receipts_root" hex:"true"`
	LogsBloom        string `json:"logs_bloom" hex:"true"`
	PrevRandao       string `json:"prev_randao" hex:"true"`
	BlockNumber      string `json:"block_number"`
	GasLimit         string `json:"gas_limit"`
	GasUsed          string `json:"gas_used"`
	TimeStamp        string `json:"timestamp"`
	ExtraData        string `json:"extra_data" hex:"true"`
	BaseFeePerGas    string `json:"base_fee_per_gas" uint256:"true"`
	BlockHash        string `json:"block_hash" hex:"true"`
	TransactionsRoot string `json:"transactions_root" hex:"true"`
	WithdrawalsRoot  string `json:"withdrawals_root" hex:"true"`
}

type SyncAggregateJson struct {
	SyncCommitteeBits       string   `json:"sync_committee_bits" hex:"true"`
	SyncCommitteeSignatures []string `json:"sync_committee_signatures" hex:"true"`
}

type SignedBeaconBlockHeaderJson struct {
	Header    *BeaconBlockHeaderJson `json:"message"`
	Signature string                 `json:"signature" hex:"true"`
}

type BeaconBlockHeaderJson struct {
	Slot          string `json:"slot"`
	ProposerIndex string `json:"proposer_index"`
	ParentRoot    string `json:"parent_root" hex:"true"`
	StateRoot     string `json:"state_root" hex:"true"`
	BodyRoot      string `json:"body_root" hex:"true"`
}

type Eth1DataJson struct {
	DepositRoot  string `json:"deposit_root" hex:"true"`
	DepositCount string `json:"deposit_count"`
	BlockHash    string `json:"block_hash" hex:"true"`
}

type ProposerSlashingJson struct {
	Header_1 *SignedBeaconBlockHeaderJson `json:"signed_header_1"`
	Header_2 *SignedBeaconBlockHeaderJson `json:"signed_header_2"`
}

type AttesterSlashingJson struct {
	Attestation_1 *IndexedAttestationJson `json:"attestation_1"`
	Attestation_2 *IndexedAttestationJson `json:"attestation_2"`
}

type IndexedAttestationJson struct {
	AttestingIndices []string             `json:"attesting_indices"`
	Data             *AttestationDataJson `json:"data"`
	Signatures       []string             `json:"signatures" hex:"true"`
}

type AttestationJson struct {
	AggregationBits string               `json:"aggregation_bits" hex:"true"`
	Data            *AttestationDataJson `json:"data"`
	Signatures      []string             `json:"signatures" hex:"true"`
}

type AttestationDataJson struct {
	Slot            string          `json:"slot"`
	CommitteeIndex  string          `json:"index"`
	BeaconBlockRoot string          `json:"beacon_block_root" hex:"true"`
	Source          *CheckpointJson `json:"source"`
	Target          *CheckpointJson `json:"target"`
}

type SignedDilithiumToExecutionChangeJson struct {
	Message   *DilithiumToExecutionChangeJson `json:"message"`
	Signature string                          `json:"signature" hex:"true"`
}

type DilithiumToExecutionChangeJson struct {
	ValidatorIndex      string `json:"validator_index"`
	FromDilithiumPubkey string `json:"from_dilithium_pubkey" hex:"true"`
	ToExecutionAddress  string `json:"to_execution_address" hex:"true"`
}

type SubmitDilithiumToExecutionChangesRequest struct {
	Changes []*SignedDilithiumToExecutionChangeJson `json:"changes"`
}

type DepositJson struct {
	Proof []string          `json:"proof" hex:"true"`
	Data  *Deposit_DataJson `json:"data"`
}

type Deposit_DataJson struct {
	PublicKey             string `json:"pubkey" hex:"true"`
	WithdrawalCredentials string `json:"withdrawal_credentials" hex:"true"`
	Amount                string `json:"amount"`
	Signature             string `json:"signature" hex:"true"`
}

type SignedVoluntaryExitJson struct {
	Exit      *VoluntaryExitJson `json:"message"`
	Signature string             `json:"signature" hex:"true"`
}

type VoluntaryExitJson struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validator_index"`
}

type IdentityJson struct {
	PeerId             string        `json:"peer_id"`
	Qnr                string        `json:"qnr"`
	P2PAddresses       []string      `json:"p2p_addresses"`
	DiscoveryAddresses []string      `json:"discovery_addresses"`
	Metadata           *MetadataJson `json:"metadata"`
}

type MetadataJson struct {
	SeqNumber string `json:"seq_number"`
	Attnets   string `json:"attnets" hex:"true"`
}

type PeerJson struct {
	PeerId    string `json:"peer_id"`
	Qnr       string `json:"qnr"`
	Address   string `json:"last_seen_p2p_address"`
	State     string `json:"state" enum:"true"`
	Direction string `json:"direction" enum:"true"`
}

type VersionJson struct {
	Version string `json:"version" enum:"true"`
}

type WithdrawalJson struct {
	WithdrawalIndex  string `json:"index"`
	ValidatorIndex   string `json:"validator_index"`
	ExecutionAddress string `json:"address" hex:"true"`
	Amount           string `json:"amount"`
}

type BeaconStateCapellaJson struct {
	GenesisTime                  string                             `json:"genesis_time"`
	GenesisValidatorsRoot        string                             `json:"genesis_validators_root" hex:"true"`
	Slot                         string                             `json:"slot"`
	Fork                         *ForkJson                          `json:"fork"`
	LatestBlockHeader            *BeaconBlockHeaderJson             `json:"latest_block_header"`
	BlockRoots                   []string                           `json:"block_roots" hex:"true"`
	StateRoots                   []string                           `json:"state_roots" hex:"true"`
	HistoricalRoots              []string                           `json:"historical_roots" hex:"true"`
	Eth1Data                     *Eth1DataJson                      `json:"eth1_data"`
	Eth1DataVotes                []*Eth1DataJson                    `json:"eth1_data_votes"`
	Eth1DepositIndex             string                             `json:"eth1_deposit_index"`
	Validators                   []*ValidatorJson                   `json:"validators"`
	Balances                     []string                           `json:"balances"`
	RandaoMixes                  []string                           `json:"randao_mixes" hex:"true"`
	Slashings                    []string                           `json:"slashings"`
	PreviousEpochParticipation   EpochParticipation                 `json:"previous_epoch_participation"`
	CurrentEpochParticipation    EpochParticipation                 `json:"current_epoch_participation"`
	JustificationBits            string                             `json:"justification_bits" hex:"true"`
	PreviousJustifiedCheckpoint  *CheckpointJson                    `json:"previous_justified_checkpoint"`
	CurrentJustifiedCheckpoint   *CheckpointJson                    `json:"current_justified_checkpoint"`
	FinalizedCheckpoint          *CheckpointJson                    `json:"finalized_checkpoint"`
	InactivityScores             []string                           `json:"inactivity_scores"`
	CurrentSyncCommittee         *SyncCommitteeJson                 `json:"current_sync_committee"`
	NextSyncCommittee            *SyncCommitteeJson                 `json:"next_sync_committee"`
	LatestExecutionPayloadHeader *ExecutionPayloadHeaderCapellaJson `json:"latest_execution_payload_header"`
	NextWithdrawalIndex          string                             `json:"next_withdrawal_index"`
	NextWithdrawalValidatorIndex string                             `json:"next_withdrawal_validator_index"`
	HistoricalSummaries          []*HistoricalSummaryJson           `json:"historical_summaries"`
}

type BeaconStateContainerJson struct {
	CapellaState *BeaconStateCapellaJson `json:"capella_state"`
}

type ForkJson struct {
	PreviousVersion string `json:"previous_version" hex:"true"`
	CurrentVersion  string `json:"current_version" hex:"true"`
	Epoch           string `json:"epoch"`
}

type ValidatorJson struct {
	PublicKey                  string `json:"pubkey" hex:"true"`
	WithdrawalCredentials      string `json:"withdrawal_credentials" hex:"true"`
	EffectiveBalance           string `json:"effective_balance"`
	Slashed                    bool   `json:"slashed"`
	ActivationEligibilityEpoch string `json:"activation_eligibility_epoch"`
	ActivationEpoch            string `json:"activation_epoch"`
	ExitEpoch                  string `json:"exit_epoch"`
	WithdrawableEpoch          string `json:"withdrawable_epoch"`
}

type SyncCommitteeJson struct {
	Pubkeys []string `json:"pubkeys" hex:"true"`
}

type SyncCommitteeValidatorsJson struct {
	Validators          []string   `json:"validators"`
	ValidatorAggregates [][]string `json:"validator_aggregates"`
}

type PendingAttestationJson struct {
	AggregationBits string               `json:"aggregation_bits" hex:"true"`
	Data            *AttestationDataJson `json:"data"`
	InclusionDelay  string               `json:"inclusion_delay"`
	ProposerIndex   string               `json:"proposer_index"`
}

type ForkChoiceHeadJson struct {
	Root                string `json:"root" hex:"true"`
	Slot                string `json:"slot"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
}

type DepositContractJson struct {
	ChainId string `json:"chain_id"`
	Address string `json:"address"`
}

type SignedAggregateAttestationAndProofJson struct {
	Message   *AggregateAttestationAndProofJson `json:"message"`
	Signature string                            `json:"signature" hex:"true"`
}

type AggregateAttestationAndProofJson struct {
	AggregatorIndex string           `json:"aggregator_index"`
	Aggregate       *AttestationJson `json:"aggregate"`
	SelectionProof  string           `json:"selection_proof" hex:"true"`
}

type SignedContributionAndProofJson struct {
	Message   *ContributionAndProofJson `json:"message"`
	Signature string                    `json:"signature" hex:"true"`
}

type ContributionAndProofJson struct {
	AggregatorIndex string                         `json:"aggregator_index"`
	Contribution    *SyncCommitteeContributionJson `json:"contribution"`
	SelectionProof  string                         `json:"selection_proof" hex:"true"`
}

type SyncCommitteeContributionJson struct {
	Slot              string   `json:"slot"`
	BeaconBlockRoot   string   `json:"beacon_block_root" hex:"true"`
	SubcommitteeIndex string   `json:"subcommittee_index"`
	AggregationBits   string   `json:"aggregation_bits" hex:"true"`
	Signatures        []string `json:"signatures" hex:"true"`
}

type ForkChoiceNodeJson struct {
	Slot                     string `json:"slot"`
	BlockRoot                string `json:"block_root" hex:"true"`
	ParentRoot               string `json:"parent_root" hex:"true"`
	JustifiedEpoch           string `json:"justified_epoch"`
	FinalizedEpoch           string `json:"finalized_epoch"`
	UnrealizedJustifiedEpoch string `json:"unrealized_justified_epoch"`
	UnrealizedFinalizedEpoch string `json:"unrealized_finalized_epoch"`
	Balance                  string `json:"balance"`
	Weight                   string `json:"weight"`
	ExecutionOptimistic      bool   `json:"execution_optimistic"`
	ExecutionBlockHash       string `json:"execution_block_hash" hex:"true"`
	TimeStamp                string `json:"timestamp"`
	Validity                 string `json:"validity" enum:"true"`
}

type ForkChoiceDumpJson struct {
	JustifiedCheckpoint           *CheckpointJson       `json:"justified_checkpoint"`
	FinalizedCheckpoint           *CheckpointJson       `json:"finalized_checkpoint"`
	BestJustifiedCheckpoint       *CheckpointJson       `json:"best_justified_checkpoint"`
	UnrealizedJustifiedCheckpoint *CheckpointJson       `json:"unrealized_justified_checkpoint"`
	UnrealizedFinalizedCheckpoint *CheckpointJson       `json:"unrealized_finalized_checkpoint"`
	ProposerBoostRoot             string                `json:"proposer_boost_root" hex:"true"`
	PreviousProposerBoostRoot     string                `json:"previous_proposer_boost_root" hex:"true"`
	HeadRoot                      string                `json:"head_root" hex:"true"`
	ForkChoiceNodes               []*ForkChoiceNodeJson `json:"fork_choice_nodes"`
}

type HistoricalSummaryJson struct {
	BlockSummaryRoot string `json:"block_summary_root" hex:"true"`
	StateSummaryRoot string `json:"state_summary_root" hex:"true"`
}

//----------------
// SSZ
// ---------------

type SszRequestJson struct {
	Data string `json:"data"`
}

// SszResponse is a common abstraction over all SSZ responses.
type SszResponse interface {
	SSZVersion() string
	SSZOptimistic() bool
	SSZData() string
	SSZFinalized() bool
}

type SszResponseJson struct {
	Data string `json:"data"`
}

func (ssz *SszResponseJson) SSZData() string {
	return ssz.Data
}

func (*SszResponseJson) SSZVersion() string {
	return strings.ToLower(zondpbv1.Version_PHASE0.String())
}

func (*SszResponseJson) SSZOptimistic() bool {
	return false
}

func (*SszResponseJson) SSZFinalized() bool {
	return true
}

type VersionedSSZResponseJson struct {
	Version             string `json:"version" enum:"true"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
	Finalized           bool   `json:"finalized"`
	Data                string `json:"data"`
}

func (ssz *VersionedSSZResponseJson) SSZData() string {
	return ssz.Data
}

func (ssz *VersionedSSZResponseJson) SSZVersion() string {
	return ssz.Version
}

func (ssz *VersionedSSZResponseJson) SSZOptimistic() bool {
	return ssz.ExecutionOptimistic
}

func (ssz *VersionedSSZResponseJson) SSZFinalized() bool {
	return ssz.Finalized
}

// ---------------
// Events.
// ---------------

type EventHeadJson struct {
	Slot                      string `json:"slot"`
	Block                     string `json:"block" hex:"true"`
	State                     string `json:"state" hex:"true"`
	EpochTransition           bool   `json:"epoch_transition"`
	ExecutionOptimistic       bool   `json:"execution_optimistic"`
	PreviousDutyDependentRoot string `json:"previous_duty_dependent_root" hex:"true"`
	CurrentDutyDependentRoot  string `json:"current_duty_dependent_root" hex:"true"`
}

type ReceivedBlockDataJson struct {
	Slot                string `json:"slot"`
	Block               string `json:"block" hex:"true"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
}

type AggregatedAttReceivedDataJson struct {
	Aggregate *AttestationJson `json:"aggregate"`
}

type UnaggregatedAttReceivedDataJson struct {
	AggregationBits string               `json:"aggregation_bits" hex:"true"`
	Data            *AttestationDataJson `json:"data"`
	Signatures      []string             `json:"signatures" hex:"true"`
}

type EventFinalizedCheckpointJson struct {
	Block               string `json:"block" hex:"true"`
	State               string `json:"state" hex:"true"`
	Epoch               string `json:"epoch"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
}

type EventChainReorgJson struct {
	Slot                string `json:"slot"`
	Depth               string `json:"depth"`
	OldHeadBlock        string `json:"old_head_block" hex:"true"`
	NewHeadBlock        string `json:"old_head_state" hex:"true"`
	OldHeadState        string `json:"new_head_block" hex:"true"`
	NewHeadState        string `json:"new_head_state" hex:"true"`
	Epoch               string `json:"epoch"`
	ExecutionOptimistic bool   `json:"execution_optimistic"`
}

type EventPayloadAttributeStreamV2Json struct {
	Version string                       `json:"version"`
	Data    *EventPayloadAttributeV2Json `json:"data"`
}

type EventPayloadAttributeV2Json struct {
	ProposerIndex     string                   `json:"proposer_index"`
	ProposalSlot      string                   `json:"proposal_slot"`
	ParentBlockNumber string                   `json:"parent_block_number"`
	ParentBlockRoot   string                   `json:"parent_block_root" hex:"true"`
	ParentBlockHash   string                   `json:"parent_block_hash" hex:"true"`
	PayloadAttributes *PayloadAttributesV2Json `json:"payload_attributes"`
}

type PayloadAttributesV2Json struct {
	Timestamp             string            `json:"timestamp"`
	Random                string            `json:"prev_randao" hex:"true"`
	SuggestedFeeRecipient string            `json:"suggested_fee_recipient" hex:"true"`
	Withdrawals           []*WithdrawalJson `json:"withdrawals"`
}

// ---------------
// Error handling.
// ---------------

// IndexedVerificationFailureErrorJson is a JSON representation of the error returned when verifying an indexed object.
type IndexedVerificationFailureErrorJson struct {
	apimiddleware.DefaultErrorJson
	Failures []*SingleIndexedVerificationFailureJson `json:"failures"`
}

// SingleIndexedVerificationFailureJson is a JSON representation of a an issue when verifying a single indexed object e.g. an item in an array.
type SingleIndexedVerificationFailureJson struct {
	Index   int    `json:"index"`
	Message string `json:"message"`
}

type EventErrorJson struct {
	StatusCode int    `json:"status_code"`
	Message    string `json:"message"`
}
