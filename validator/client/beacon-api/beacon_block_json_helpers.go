package beacon_api

import (
	"strconv"

	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/apimiddleware"
	enginev1 "github.com/theQRL/qrysm/v4/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

func jsonifyTransactions(transactions [][]byte) []string {
	jsonTransactions := make([]string, len(transactions))
	for index, transaction := range transactions {
		jsonTransaction := hexutil.Encode(transaction)
		jsonTransactions[index] = jsonTransaction
	}
	return jsonTransactions
}

func jsonifyDilithiumToExecutionChanges(dilithiumToExecutionChanges []*zondpb.SignedDilithiumToExecutionChange) []*apimiddleware.SignedDilithiumToExecutionChangeJson {
	jsonDilithiumToExecutionChanges := make([]*apimiddleware.SignedDilithiumToExecutionChangeJson, len(dilithiumToExecutionChanges))
	for index, signedDilithiumToExecutionChange := range dilithiumToExecutionChanges {
		dilithiumToExecutionChangeJson := &apimiddleware.DilithiumToExecutionChangeJson{
			ValidatorIndex:      uint64ToString(signedDilithiumToExecutionChange.Message.ValidatorIndex),
			FromDilithiumPubkey: hexutil.Encode(signedDilithiumToExecutionChange.Message.FromDilithiumPubkey),
			ToExecutionAddress:  hexutil.Encode(signedDilithiumToExecutionChange.Message.ToExecutionAddress),
		}
		signedJson := &apimiddleware.SignedDilithiumToExecutionChangeJson{
			Message:   dilithiumToExecutionChangeJson,
			Signature: hexutil.Encode(signedDilithiumToExecutionChange.Signature),
		}
		jsonDilithiumToExecutionChanges[index] = signedJson
	}
	return jsonDilithiumToExecutionChanges
}

func jsonifyZond1Data(zond1Data *zondpb.Zond1Data) *apimiddleware.Zond1DataJson {
	return &apimiddleware.Zond1DataJson{
		BlockHash:    hexutil.Encode(zond1Data.BlockHash),
		DepositCount: uint64ToString(zond1Data.DepositCount),
		DepositRoot:  hexutil.Encode(zond1Data.DepositRoot),
	}
}

func jsonifyAttestations(attestations []*zondpb.Attestation) []*apimiddleware.AttestationJson {
	jsonAttestations := make([]*apimiddleware.AttestationJson, len(attestations))
	for index, attestation := range attestations {
		jsonAttestations[index] = jsonifyAttestation(attestation)
	}
	return jsonAttestations
}

func jsonifyAttesterSlashings(attesterSlashings []*zondpb.AttesterSlashing) []*apimiddleware.AttesterSlashingJson {
	jsonAttesterSlashings := make([]*apimiddleware.AttesterSlashingJson, len(attesterSlashings))
	for index, attesterSlashing := range attesterSlashings {
		jsonAttesterSlashing := &apimiddleware.AttesterSlashingJson{
			Attestation_1: jsonifyIndexedAttestation(attesterSlashing.Attestation_1),
			Attestation_2: jsonifyIndexedAttestation(attesterSlashing.Attestation_2),
		}
		jsonAttesterSlashings[index] = jsonAttesterSlashing
	}
	return jsonAttesterSlashings
}

func jsonifyDeposits(deposits []*zondpb.Deposit) []*apimiddleware.DepositJson {
	jsonDeposits := make([]*apimiddleware.DepositJson, len(deposits))
	for depositIndex, deposit := range deposits {
		proofs := make([]string, len(deposit.Proof))
		for proofIndex, proof := range deposit.Proof {
			proofs[proofIndex] = hexutil.Encode(proof)
		}

		jsonDeposit := &apimiddleware.DepositJson{
			Data: &apimiddleware.Deposit_DataJson{
				Amount:                uint64ToString(deposit.Data.Amount),
				PublicKey:             hexutil.Encode(deposit.Data.PublicKey),
				Signature:             hexutil.Encode(deposit.Data.Signature),
				WithdrawalCredentials: hexutil.Encode(deposit.Data.WithdrawalCredentials),
			},
			Proof: proofs,
		}
		jsonDeposits[depositIndex] = jsonDeposit
	}
	return jsonDeposits
}

func jsonifyProposerSlashings(proposerSlashings []*zondpb.ProposerSlashing) []*apimiddleware.ProposerSlashingJson {
	jsonProposerSlashings := make([]*apimiddleware.ProposerSlashingJson, len(proposerSlashings))
	for index, proposerSlashing := range proposerSlashings {
		jsonProposerSlashing := &apimiddleware.ProposerSlashingJson{
			Header_1: jsonifySignedBeaconBlockHeader(proposerSlashing.Header_1),
			Header_2: jsonifySignedBeaconBlockHeader(proposerSlashing.Header_2),
		}
		jsonProposerSlashings[index] = jsonProposerSlashing
	}
	return jsonProposerSlashings
}

// JsonifySignedVoluntaryExits converts an array of voluntary exit structs to a JSON hex string compatible format.
func JsonifySignedVoluntaryExits(voluntaryExits []*zondpb.SignedVoluntaryExit) []*apimiddleware.SignedVoluntaryExitJson {
	jsonSignedVoluntaryExits := make([]*apimiddleware.SignedVoluntaryExitJson, len(voluntaryExits))
	for index, signedVoluntaryExit := range voluntaryExits {
		jsonSignedVoluntaryExit := &apimiddleware.SignedVoluntaryExitJson{
			Exit: &apimiddleware.VoluntaryExitJson{
				Epoch:          uint64ToString(signedVoluntaryExit.Exit.Epoch),
				ValidatorIndex: uint64ToString(signedVoluntaryExit.Exit.ValidatorIndex),
			},
			Signature: hexutil.Encode(signedVoluntaryExit.Signature),
		}
		jsonSignedVoluntaryExits[index] = jsonSignedVoluntaryExit
	}
	return jsonSignedVoluntaryExits
}

func jsonifySignedBeaconBlockHeader(signedBeaconBlockHeader *zondpb.SignedBeaconBlockHeader) *apimiddleware.SignedBeaconBlockHeaderJson {
	return &apimiddleware.SignedBeaconBlockHeaderJson{
		Header: &apimiddleware.BeaconBlockHeaderJson{
			BodyRoot:      hexutil.Encode(signedBeaconBlockHeader.Header.BodyRoot),
			ParentRoot:    hexutil.Encode(signedBeaconBlockHeader.Header.ParentRoot),
			ProposerIndex: uint64ToString(signedBeaconBlockHeader.Header.ProposerIndex),
			Slot:          uint64ToString(signedBeaconBlockHeader.Header.Slot),
			StateRoot:     hexutil.Encode(signedBeaconBlockHeader.Header.StateRoot),
		},
		Signature: hexutil.Encode(signedBeaconBlockHeader.Signature),
	}
}

func jsonifyIndexedAttestation(indexedAttestation *zondpb.IndexedAttestation) *apimiddleware.IndexedAttestationJson {
	attestingIndices := make([]string, len(indexedAttestation.AttestingIndices))
	for index, attestingIndex := range indexedAttestation.AttestingIndices {
		attestingIndex := uint64ToString(attestingIndex)
		attestingIndices[index] = attestingIndex
	}

	return &apimiddleware.IndexedAttestationJson{
		AttestingIndices: attestingIndices,
		Data:             jsonifyAttestationData(indexedAttestation.Data),
		Signature:        hexutil.Encode(indexedAttestation.Signatures),
	}
}

func jsonifyAttestationData(attestationData *zondpb.AttestationData) *apimiddleware.AttestationDataJson {
	return &apimiddleware.AttestationDataJson{
		BeaconBlockRoot: hexutil.Encode(attestationData.BeaconBlockRoot),
		CommitteeIndex:  uint64ToString(attestationData.CommitteeIndex),
		Slot:            uint64ToString(attestationData.Slot),
		Source: &apimiddleware.CheckpointJson{
			Epoch: uint64ToString(attestationData.Source.Epoch),
			Root:  hexutil.Encode(attestationData.Source.Root),
		},
		Target: &apimiddleware.CheckpointJson{
			Epoch: uint64ToString(attestationData.Target.Epoch),
			Root:  hexutil.Encode(attestationData.Target.Root),
		},
	}
}

func jsonifyAttestation(attestation *zondpb.Attestation) *apimiddleware.AttestationJson {
	return &apimiddleware.AttestationJson{
		AggregationBits: hexutil.Encode(attestation.AggregationBits),
		Data:            jsonifyAttestationData(attestation.Data),
		Signature:       hexutil.Encode(attestation.Signatures),
	}
}

func jsonifySignedAggregateAndProof(signedAggregateAndProof *zondpb.SignedAggregateAttestationAndProof) *apimiddleware.SignedAggregateAttestationAndProofJson {
	return &apimiddleware.SignedAggregateAttestationAndProofJson{
		Message: &apimiddleware.AggregateAttestationAndProofJson{
			AggregatorIndex: uint64ToString(signedAggregateAndProof.Message.AggregatorIndex),
			Aggregate:       jsonifyAttestation(signedAggregateAndProof.Message.Aggregate),
			SelectionProof:  hexutil.Encode(signedAggregateAndProof.Message.SelectionProof),
		},
		Signature: hexutil.Encode(signedAggregateAndProof.Signature),
	}
}

func jsonifyWithdrawals(withdrawals []*enginev1.Withdrawal) []*apimiddleware.WithdrawalJson {
	jsonWithdrawals := make([]*apimiddleware.WithdrawalJson, len(withdrawals))
	for index, withdrawal := range withdrawals {
		jsonWithdrawals[index] = &apimiddleware.WithdrawalJson{
			WithdrawalIndex:  strconv.FormatUint(withdrawal.Index, 10),
			ValidatorIndex:   strconv.FormatUint(uint64(withdrawal.ValidatorIndex), 10),
			ExecutionAddress: hexutil.Encode(withdrawal.Address),
			Amount:           strconv.FormatUint(withdrawal.Amount, 10),
		}
	}
	return jsonWithdrawals
}
