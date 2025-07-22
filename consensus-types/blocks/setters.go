package blocks

import (
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	zond "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

// SetSignature sets the signature of the signed beacon block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetSignature(sig []byte) {
	copy(b.signature[:], sig)
}

// SetSlot sets the respective slot of the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetSlot(slot primitives.Slot) {
	b.block.slot = slot
}

// SetProposerIndex sets the proposer index of the beacon block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetProposerIndex(proposerIndex primitives.ValidatorIndex) {
	b.block.proposerIndex = proposerIndex
}

// SetParentRoot sets the parent root of beacon block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetParentRoot(parentRoot []byte) {
	copy(b.block.parentRoot[:], parentRoot)
}

// SetStateRoot sets the state root of the underlying beacon block
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetStateRoot(root []byte) {
	copy(b.block.stateRoot[:], root)
}

// SetBlinded sets the blinded flag of the beacon block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetBlinded(blinded bool) {
	b.block.body.isBlinded = blinded
}

// SetRandaoReveal sets the randao reveal in the block body.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetRandaoReveal(r []byte) {
	copy(b.block.body.randaoReveal[:], r)
}

// SetGraffiti sets the graffiti in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetGraffiti(g []byte) {
	copy(b.block.body.graffiti[:], g)
}

// SetExecutionNodeData sets the eth1 data in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetExecutionNodeData(e *zond.ExecutionNodeData) {
	b.block.body.executionNodeData = e
}

// SetProposerSlashings sets the proposer slashings in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetProposerSlashings(p []*zond.ProposerSlashing) {
	b.block.body.proposerSlashings = p
}

// SetAttesterSlashings sets the attester slashings in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetAttesterSlashings(a []*zond.AttesterSlashing) {
	b.block.body.attesterSlashings = a
}

// SetAttestations sets the attestations in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetAttestations(a []*zond.Attestation) {
	b.block.body.attestations = a
}

// SetDeposits sets the deposits in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetDeposits(d []*zond.Deposit) {
	b.block.body.deposits = d
}

// SetVoluntaryExits sets the voluntary exits in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetVoluntaryExits(v []*zond.SignedVoluntaryExit) {
	b.block.body.voluntaryExits = v
}

// SetSyncAggregate sets the sync aggregate in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetSyncAggregate(s *zond.SyncAggregate) error {
	b.block.body.syncAggregate = s
	return nil
}

// SetExecution sets the execution payload of the block body.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetExecution(e interfaces.ExecutionData) error {
	if b.block.body.isBlinded {
		b.block.body.executionPayloadHeader = e
		return nil
	}
	b.block.body.executionPayload = e
	return nil
}

// SetDilithiumToExecutionChanges sets the Dilithium to execution changes in the block.
// This function is not thread safe, it is only used during block creation.
func (b *SignedBeaconBlock) SetDilithiumToExecutionChanges(dilithiumToExecutionChanges []*zond.SignedDilithiumToExecutionChange) error {
	b.block.body.dilithiumToExecutionChanges = dilithiumToExecutionChanges
	return nil
}
