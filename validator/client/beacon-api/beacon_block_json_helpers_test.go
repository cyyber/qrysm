package beacon_api

import (
	"testing"

	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/beacon-chain/rpc/apimiddleware"
	enginev1 "github.com/theQRL/qrysm/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
)

func TestBeaconBlockJsonHelpers_JsonifyTransactions(t *testing.T) {
	input := [][]byte{{1}, {2}, {3}, {4}}

	expectedResult := []string{
		hexutil.Encode([]byte{1}),
		hexutil.Encode([]byte{2}),
		hexutil.Encode([]byte{3}),
		hexutil.Encode([]byte{4}),
	}

	result := jsonifyTransactions(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyDilithiumToExecutionChanges(t *testing.T) {
	input := []*zondpb.SignedDilithiumToExecutionChange{
		{
			Message: &zondpb.DilithiumToExecutionChange{
				ValidatorIndex:      1,
				FromDilithiumPubkey: []byte{2},
				ToExecutionAddress:  []byte{3},
			},
			Signature: []byte{7},
		},
		{
			Message: &zondpb.DilithiumToExecutionChange{
				ValidatorIndex:      4,
				FromDilithiumPubkey: []byte{5},
				ToExecutionAddress:  []byte{6},
			},
			Signature: []byte{8},
		},
	}

	expectedResult := []*apimiddleware.SignedDilithiumToExecutionChangeJson{
		{
			Message: &apimiddleware.DilithiumToExecutionChangeJson{
				ValidatorIndex:      "1",
				FromDilithiumPubkey: hexutil.Encode([]byte{2}),
				ToExecutionAddress:  hexutil.Encode([]byte{3}),
			},
			Signature: hexutil.Encode([]byte{7}),
		},
		{
			Message: &apimiddleware.DilithiumToExecutionChangeJson{
				ValidatorIndex:      "4",
				FromDilithiumPubkey: hexutil.Encode([]byte{5}),
				ToExecutionAddress:  hexutil.Encode([]byte{6}),
			},
			Signature: hexutil.Encode([]byte{8}),
		},
	}

	result := jsonifyDilithiumToExecutionChanges(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyEth1Data(t *testing.T) {
	input := &zondpb.Eth1Data{
		DepositRoot:  []byte{1},
		DepositCount: 2,
		BlockHash:    []byte{3},
	}

	expectedResult := &apimiddleware.Eth1DataJson{
		DepositRoot:  hexutil.Encode([]byte{1}),
		DepositCount: "2",
		BlockHash:    hexutil.Encode([]byte{3}),
	}

	result := jsonifyEth1Data(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyAttestations(t *testing.T) {
	input := []*zondpb.Attestation{
		{
			AggregationBits: []byte{1},
			Data: &zondpb.AttestationData{
				Slot:            2,
				CommitteeIndex:  3,
				BeaconBlockRoot: []byte{4},
				Source: &zondpb.Checkpoint{
					Epoch: 5,
					Root:  []byte{6},
				},
				Target: &zondpb.Checkpoint{
					Epoch: 7,
					Root:  []byte{8},
				},
			},
			Signatures: [][]byte{{9}},
		},
		{
			AggregationBits: []byte{10},
			Data: &zondpb.AttestationData{
				Slot:            11,
				CommitteeIndex:  12,
				BeaconBlockRoot: []byte{13},
				Source: &zondpb.Checkpoint{
					Epoch: 14,
					Root:  []byte{15},
				},
				Target: &zondpb.Checkpoint{
					Epoch: 16,
					Root:  []byte{17},
				},
			},
			Signatures: [][]byte{{18}},
		},
	}

	expectedResult := []*apimiddleware.AttestationJson{
		{
			AggregationBits: hexutil.Encode([]byte{1}),
			Data: &apimiddleware.AttestationDataJson{
				Slot:            "2",
				CommitteeIndex:  "3",
				BeaconBlockRoot: hexutil.Encode([]byte{4}),
				Source: &apimiddleware.CheckpointJson{
					Epoch: "5",
					Root:  hexutil.Encode([]byte{6}),
				},
				Target: &apimiddleware.CheckpointJson{
					Epoch: "7",
					Root:  hexutil.Encode([]byte{8}),
				},
			},
			Signatures: []string{hexutil.Encode([]byte{9})},
		},
		{
			AggregationBits: hexutil.Encode([]byte{10}),
			Data: &apimiddleware.AttestationDataJson{
				Slot:            "11",
				CommitteeIndex:  "12",
				BeaconBlockRoot: hexutil.Encode([]byte{13}),
				Source: &apimiddleware.CheckpointJson{
					Epoch: "14",
					Root:  hexutil.Encode([]byte{15}),
				},
				Target: &apimiddleware.CheckpointJson{
					Epoch: "16",
					Root:  hexutil.Encode([]byte{17}),
				},
			},
			Signatures: []string{hexutil.Encode([]byte{18})},
		},
	}

	result := jsonifyAttestations(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyAttesterSlashings(t *testing.T) {
	input := []*zondpb.AttesterSlashing{
		{
			Attestation_1: &zondpb.IndexedAttestation{
				AttestingIndices: []uint64{1, 2},
				Data: &zondpb.AttestationData{
					Slot:            3,
					CommitteeIndex:  4,
					BeaconBlockRoot: []byte{5},
					Source: &zondpb.Checkpoint{
						Epoch: 6,
						Root:  []byte{7},
					},
					Target: &zondpb.Checkpoint{
						Epoch: 8,
						Root:  []byte{9},
					},
				},
				Signatures: [][]byte{{10}},
			},
			Attestation_2: &zondpb.IndexedAttestation{
				AttestingIndices: []uint64{11, 12},
				Data: &zondpb.AttestationData{
					Slot:            13,
					CommitteeIndex:  14,
					BeaconBlockRoot: []byte{15},
					Source: &zondpb.Checkpoint{
						Epoch: 16,
						Root:  []byte{17},
					},
					Target: &zondpb.Checkpoint{
						Epoch: 18,
						Root:  []byte{19},
					},
				},
				Signatures: [][]byte{{20}},
			},
		},
		{
			Attestation_1: &zondpb.IndexedAttestation{
				AttestingIndices: []uint64{21, 22},
				Data: &zondpb.AttestationData{
					Slot:            23,
					CommitteeIndex:  24,
					BeaconBlockRoot: []byte{25},
					Source: &zondpb.Checkpoint{
						Epoch: 26,
						Root:  []byte{27},
					},
					Target: &zondpb.Checkpoint{
						Epoch: 28,
						Root:  []byte{29},
					},
				},
				Signatures: [][]byte{{30}},
			},
			Attestation_2: &zondpb.IndexedAttestation{
				AttestingIndices: []uint64{31, 32},
				Data: &zondpb.AttestationData{
					Slot:            33,
					CommitteeIndex:  34,
					BeaconBlockRoot: []byte{35},
					Source: &zondpb.Checkpoint{
						Epoch: 36,
						Root:  []byte{37},
					},
					Target: &zondpb.Checkpoint{
						Epoch: 38,
						Root:  []byte{39},
					},
				},
				Signatures: [][]byte{{40}},
			},
		},
	}

	expectedResult := []*apimiddleware.AttesterSlashingJson{
		{
			Attestation_1: &apimiddleware.IndexedAttestationJson{
				AttestingIndices: []string{"1", "2"},
				Data: &apimiddleware.AttestationDataJson{
					Slot:            "3",
					CommitteeIndex:  "4",
					BeaconBlockRoot: hexutil.Encode([]byte{5}),
					Source: &apimiddleware.CheckpointJson{
						Epoch: "6",
						Root:  hexutil.Encode([]byte{7}),
					},
					Target: &apimiddleware.CheckpointJson{
						Epoch: "8",
						Root:  hexutil.Encode([]byte{9}),
					},
				},
				Signatures: []string{hexutil.Encode([]byte{10})},
			},
			Attestation_2: &apimiddleware.IndexedAttestationJson{
				AttestingIndices: []string{"11", "12"},
				Data: &apimiddleware.AttestationDataJson{
					Slot:            "13",
					CommitteeIndex:  "14",
					BeaconBlockRoot: hexutil.Encode([]byte{15}),
					Source: &apimiddleware.CheckpointJson{
						Epoch: "16",
						Root:  hexutil.Encode([]byte{17}),
					},
					Target: &apimiddleware.CheckpointJson{
						Epoch: "18",
						Root:  hexutil.Encode([]byte{19}),
					},
				},
				Signatures: []string{hexutil.Encode([]byte{20})},
			},
		},
		{
			Attestation_1: &apimiddleware.IndexedAttestationJson{
				AttestingIndices: []string{"21", "22"},
				Data: &apimiddleware.AttestationDataJson{
					Slot:            "23",
					CommitteeIndex:  "24",
					BeaconBlockRoot: hexutil.Encode([]byte{25}),
					Source: &apimiddleware.CheckpointJson{
						Epoch: "26",
						Root:  hexutil.Encode([]byte{27}),
					},
					Target: &apimiddleware.CheckpointJson{
						Epoch: "28",
						Root:  hexutil.Encode([]byte{29}),
					},
				},
				Signatures: []string{hexutil.Encode([]byte{30})},
			},
			Attestation_2: &apimiddleware.IndexedAttestationJson{
				AttestingIndices: []string{"31", "32"},
				Data: &apimiddleware.AttestationDataJson{
					Slot:            "33",
					CommitteeIndex:  "34",
					BeaconBlockRoot: hexutil.Encode([]byte{35}),
					Source: &apimiddleware.CheckpointJson{
						Epoch: "36",
						Root:  hexutil.Encode([]byte{37}),
					},
					Target: &apimiddleware.CheckpointJson{
						Epoch: "38",
						Root:  hexutil.Encode([]byte{39}),
					},
				},
				Signatures: []string{hexutil.Encode([]byte{40})},
			},
		},
	}

	result := jsonifyAttesterSlashings(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyDeposits(t *testing.T) {
	input := []*zondpb.Deposit{
		{
			Proof: [][]byte{{1}, {2}},
			Data: &zondpb.Deposit_Data{
				PublicKey:             []byte{3},
				WithdrawalCredentials: []byte{4},
				Amount:                5,
				Signature:             []byte{6},
			},
		},
		{
			Proof: [][]byte{
				{7},
				{8},
			},
			Data: &zondpb.Deposit_Data{
				PublicKey:             []byte{9},
				WithdrawalCredentials: []byte{10},
				Amount:                11,
				Signature:             []byte{12},
			},
		},
	}

	expectedResult := []*apimiddleware.DepositJson{
		{
			Proof: []string{
				hexutil.Encode([]byte{1}),
				hexutil.Encode([]byte{2}),
			},
			Data: &apimiddleware.Deposit_DataJson{
				PublicKey:             hexutil.Encode([]byte{3}),
				WithdrawalCredentials: hexutil.Encode([]byte{4}),
				Amount:                "5",
				Signature:             hexutil.Encode([]byte{6}),
			},
		},
		{
			Proof: []string{
				hexutil.Encode([]byte{7}),
				hexutil.Encode([]byte{8}),
			},
			Data: &apimiddleware.Deposit_DataJson{
				PublicKey:             hexutil.Encode([]byte{9}),
				WithdrawalCredentials: hexutil.Encode([]byte{10}),
				Amount:                "11",
				Signature:             hexutil.Encode([]byte{12}),
			},
		},
	}

	result := jsonifyDeposits(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyProposerSlashings(t *testing.T) {
	input := []*zondpb.ProposerSlashing{
		{
			Header_1: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					Slot:          1,
					ProposerIndex: 2,
					ParentRoot:    []byte{3},
					StateRoot:     []byte{4},
					BodyRoot:      []byte{5},
				},
				Signature: []byte{6},
			},
			Header_2: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					Slot:          7,
					ProposerIndex: 8,
					ParentRoot:    []byte{9},
					StateRoot:     []byte{10},
					BodyRoot:      []byte{11},
				},
				Signature: []byte{12},
			},
		},
		{
			Header_1: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					Slot:          13,
					ProposerIndex: 14,
					ParentRoot:    []byte{15},
					StateRoot:     []byte{16},
					BodyRoot:      []byte{17},
				},
				Signature: []byte{18},
			},
			Header_2: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					Slot:          19,
					ProposerIndex: 20,
					ParentRoot:    []byte{21},
					StateRoot:     []byte{22},
					BodyRoot:      []byte{23},
				},
				Signature: []byte{24},
			},
		},
	}

	expectedResult := []*apimiddleware.ProposerSlashingJson{
		{
			Header_1: &apimiddleware.SignedBeaconBlockHeaderJson{
				Header: &apimiddleware.BeaconBlockHeaderJson{
					Slot:          "1",
					ProposerIndex: "2",
					ParentRoot:    hexutil.Encode([]byte{3}),
					StateRoot:     hexutil.Encode([]byte{4}),
					BodyRoot:      hexutil.Encode([]byte{5}),
				},
				Signature: hexutil.Encode([]byte{6}),
			},
			Header_2: &apimiddleware.SignedBeaconBlockHeaderJson{
				Header: &apimiddleware.BeaconBlockHeaderJson{
					Slot:          "7",
					ProposerIndex: "8",
					ParentRoot:    hexutil.Encode([]byte{9}),
					StateRoot:     hexutil.Encode([]byte{10}),
					BodyRoot:      hexutil.Encode([]byte{11}),
				},
				Signature: hexutil.Encode([]byte{12}),
			},
		},
		{
			Header_1: &apimiddleware.SignedBeaconBlockHeaderJson{
				Header: &apimiddleware.BeaconBlockHeaderJson{
					Slot:          "13",
					ProposerIndex: "14",
					ParentRoot:    hexutil.Encode([]byte{15}),
					StateRoot:     hexutil.Encode([]byte{16}),
					BodyRoot:      hexutil.Encode([]byte{17}),
				},
				Signature: hexutil.Encode([]byte{18}),
			},
			Header_2: &apimiddleware.SignedBeaconBlockHeaderJson{
				Header: &apimiddleware.BeaconBlockHeaderJson{
					Slot:          "19",
					ProposerIndex: "20",
					ParentRoot:    hexutil.Encode([]byte{21}),
					StateRoot:     hexutil.Encode([]byte{22}),
					BodyRoot:      hexutil.Encode([]byte{23}),
				},
				Signature: hexutil.Encode([]byte{24}),
			},
		},
	}

	result := jsonifyProposerSlashings(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifySignedVoluntaryExits(t *testing.T) {
	input := []*zondpb.SignedVoluntaryExit{
		{
			Exit: &zondpb.VoluntaryExit{
				Epoch:          1,
				ValidatorIndex: 2,
			},
			Signature: []byte{3},
		},
		{
			Exit: &zondpb.VoluntaryExit{
				Epoch:          4,
				ValidatorIndex: 5,
			},
			Signature: []byte{6},
		},
	}

	expectedResult := []*apimiddleware.SignedVoluntaryExitJson{
		{
			Exit: &apimiddleware.VoluntaryExitJson{
				Epoch:          "1",
				ValidatorIndex: "2",
			},
			Signature: hexutil.Encode([]byte{3}),
		},
		{
			Exit: &apimiddleware.VoluntaryExitJson{
				Epoch:          "4",
				ValidatorIndex: "5",
			},
			Signature: hexutil.Encode([]byte{6}),
		},
	}

	result := JsonifySignedVoluntaryExits(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifySignedBeaconBlockHeader(t *testing.T) {
	input := &zondpb.SignedBeaconBlockHeader{
		Header: &zondpb.BeaconBlockHeader{
			Slot:          1,
			ProposerIndex: 2,
			ParentRoot:    []byte{3},
			StateRoot:     []byte{4},
			BodyRoot:      []byte{5},
		},
		Signature: []byte{6},
	}

	expectedResult := &apimiddleware.SignedBeaconBlockHeaderJson{
		Header: &apimiddleware.BeaconBlockHeaderJson{
			Slot:          "1",
			ProposerIndex: "2",
			ParentRoot:    hexutil.Encode([]byte{3}),
			StateRoot:     hexutil.Encode([]byte{4}),
			BodyRoot:      hexutil.Encode([]byte{5}),
		},
		Signature: hexutil.Encode([]byte{6}),
	}

	result := jsonifySignedBeaconBlockHeader(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyIndexedAttestation(t *testing.T) {
	input := &zondpb.IndexedAttestation{
		AttestingIndices: []uint64{1, 2},
		Data: &zondpb.AttestationData{
			Slot:            3,
			CommitteeIndex:  4,
			BeaconBlockRoot: []byte{5},
			Source: &zondpb.Checkpoint{
				Epoch: 6,
				Root:  []byte{7},
			},
			Target: &zondpb.Checkpoint{
				Epoch: 8,
				Root:  []byte{9},
			},
		},
		Signatures: [][]byte{{10}},
	}

	expectedResult := &apimiddleware.IndexedAttestationJson{
		AttestingIndices: []string{"1", "2"},
		Data: &apimiddleware.AttestationDataJson{
			Slot:            "3",
			CommitteeIndex:  "4",
			BeaconBlockRoot: hexutil.Encode([]byte{5}),
			Source: &apimiddleware.CheckpointJson{
				Epoch: "6",
				Root:  hexutil.Encode([]byte{7}),
			},
			Target: &apimiddleware.CheckpointJson{
				Epoch: "8",
				Root:  hexutil.Encode([]byte{9}),
			},
		},
		Signatures: []string{hexutil.Encode([]byte{10})},
	}

	result := jsonifyIndexedAttestation(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyAttestationData(t *testing.T) {
	input := &zondpb.AttestationData{
		Slot:            1,
		CommitteeIndex:  2,
		BeaconBlockRoot: []byte{3},
		Source: &zondpb.Checkpoint{
			Epoch: 4,
			Root:  []byte{5},
		},
		Target: &zondpb.Checkpoint{
			Epoch: 6,
			Root:  []byte{7},
		},
	}

	expectedResult := &apimiddleware.AttestationDataJson{
		Slot:            "1",
		CommitteeIndex:  "2",
		BeaconBlockRoot: hexutil.Encode([]byte{3}),
		Source: &apimiddleware.CheckpointJson{
			Epoch: "4",
			Root:  hexutil.Encode([]byte{5}),
		},
		Target: &apimiddleware.CheckpointJson{
			Epoch: "6",
			Root:  hexutil.Encode([]byte{7}),
		},
	}

	result := jsonifyAttestationData(input)
	assert.DeepEqual(t, expectedResult, result)
}

func TestBeaconBlockJsonHelpers_JsonifyWithdrawals(t *testing.T) {
	input := []*enginev1.Withdrawal{
		{
			Index:          1,
			ValidatorIndex: 2,
			Address:        []byte{3},
			Amount:         4,
		},
		{
			Index:          5,
			ValidatorIndex: 6,
			Address:        []byte{7},
			Amount:         8,
		},
	}

	expectedResult := []*apimiddleware.WithdrawalJson{
		{
			WithdrawalIndex:  "1",
			ValidatorIndex:   "2",
			ExecutionAddress: hexutil.Encode([]byte{3}),
			Amount:           "4",
		},
		{
			WithdrawalIndex:  "5",
			ValidatorIndex:   "6",
			ExecutionAddress: hexutil.Encode([]byte{7}),
			Amount:           "8",
		},
	}

	result := jsonifyWithdrawals(input)
	assert.DeepEqual(t, expectedResult, result)
}
