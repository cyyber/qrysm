// Package operation contains types for block operation-specific events fired during the runtime of a beacon node.
package operation

import (
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

const (
	// UnaggregatedAttReceived is sent after an unaggregated attestation object has been received
	// from the outside world. (eg. in RPC or sync)
	UnaggregatedAttReceived = iota + 1

	// AggregatedAttReceived is sent after an aggregated attestation object has been received
	// from the outside world. (eg. in sync)
	AggregatedAttReceived

	// ExitReceived is sent after an voluntary exit object has been received from the outside world (eg in RPC or sync)
	ExitReceived

	// SyncCommitteeContributionReceived is sent after a sync committee contribution object has been received.
	SyncCommitteeContributionReceived

	// MLDSA87ToExecutionChangeReceived is sent after a ML-DSA-87 to execution change object has been received from gossip or rpc.
	MLDSA87ToExecutionChangeReceived
)

// UnAggregatedAttReceivedData is the data sent with UnaggregatedAttReceived events.
type UnAggregatedAttReceivedData struct {
	// Attestation is the unaggregated attestation object.
	Attestation *qrysmpb.Attestation
}

// AggregatedAttReceivedData is the data sent with AggregatedAttReceived events.
type AggregatedAttReceivedData struct {
	// Attestation is the aggregated attestation object.
	Attestation *qrysmpb.AggregateAttestationAndProof
}

// ExitReceivedData is the data sent with ExitReceived events.
type ExitReceivedData struct {
	// Exit is the voluntary exit object.
	Exit *qrysmpb.SignedVoluntaryExit
}

// SyncCommitteeContributionReceivedData is the data sent with SyncCommitteeContributionReceived objects.
type SyncCommitteeContributionReceivedData struct {
	// Contribution is the sync committee contribution object.
	Contribution *qrysmpb.SignedContributionAndProof
}

// MLDSA87ToExecutionChangeReceivedData is the data sent with MLDSA87ToExecutionChangeReceived events.
type MLDSA87ToExecutionChangeReceivedData struct {
	Change *qrysmpb.SignedMLDSA87ToExecutionChange
}
