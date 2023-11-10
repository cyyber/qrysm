package test_helpers

import (
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/apimiddleware"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

func GenerateProtoPhase0BeaconBlock() *zondpb.BeaconBlock {
	return &zondpb.BeaconBlock{
		Slot:          1,
		ProposerIndex: 2,
		ParentRoot:    FillByteSlice(32, 3),
		StateRoot:     FillByteSlice(32, 4),
		Body: &zondpb.BeaconBlockBody{
			RandaoReveal: FillByteSlice(4595, 5),
			Zond1Data: &zondpb.Zond1Data{
				DepositRoot:  FillByteSlice(32, 6),
				DepositCount: 7,
				BlockHash:    FillByteSlice(32, 8),
			},
			Graffiti: FillByteSlice(32, 9),
			ProposerSlashings: []*zondpb.ProposerSlashing{
				{
					Header_1: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							Slot:          10,
							ProposerIndex: 11,
							ParentRoot:    FillByteSlice(32, 12),
							StateRoot:     FillByteSlice(32, 13),
							BodyRoot:      FillByteSlice(32, 14),
						},
						Signature: FillByteSlice(4595, 15),
					},
					Header_2: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							Slot:          16,
							ProposerIndex: 17,
							ParentRoot:    FillByteSlice(32, 18),
							StateRoot:     FillByteSlice(32, 19),
							BodyRoot:      FillByteSlice(32, 20),
						},
						Signature: FillByteSlice(4595, 21),
					},
				},
				{
					Header_1: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							Slot:          22,
							ProposerIndex: 23,
							ParentRoot:    FillByteSlice(32, 24),
							StateRoot:     FillByteSlice(32, 25),
							BodyRoot:      FillByteSlice(32, 26),
						},
						Signature: FillByteSlice(4595, 27),
					},
					Header_2: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							Slot:          28,
							ProposerIndex: 29,
							ParentRoot:    FillByteSlice(32, 30),
							StateRoot:     FillByteSlice(32, 31),
							BodyRoot:      FillByteSlice(32, 32),
						},
						Signature: FillByteSlice(4595, 33),
					},
				},
			},
			AttesterSlashings: []*zondpb.AttesterSlashing{
				{
					Attestation_1: &zondpb.IndexedAttestation{
						AttestingIndices: []uint64{34, 35},
						Data: &zondpb.AttestationData{
							Slot:            36,
							CommitteeIndex:  37,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &zondpb.Checkpoint{
								Epoch: 39,
								Root:  FillByteSlice(32, 40),
							},
							Target: &zondpb.Checkpoint{
								Epoch: 41,
								Root:  FillByteSlice(32, 42),
							},
						},
						Signatures: FillByteSlice(4595, 43),
					},
					Attestation_2: &zondpb.IndexedAttestation{
						AttestingIndices: []uint64{44, 45},
						Data: &zondpb.AttestationData{
							Slot:            46,
							CommitteeIndex:  47,
							BeaconBlockRoot: FillByteSlice(32, 48),
							Source: &zondpb.Checkpoint{
								Epoch: 49,
								Root:  FillByteSlice(32, 50),
							},
							Target: &zondpb.Checkpoint{
								Epoch: 51,
								Root:  FillByteSlice(32, 52),
							},
						},
						Signatures: FillByteSlice(4595, 53),
					},
				},
				{
					Attestation_1: &zondpb.IndexedAttestation{
						AttestingIndices: []uint64{54, 55},
						Data: &zondpb.AttestationData{
							Slot:            56,
							CommitteeIndex:  57,
							BeaconBlockRoot: FillByteSlice(32, 58),
							Source: &zondpb.Checkpoint{
								Epoch: 59,
								Root:  FillByteSlice(32, 60),
							},
							Target: &zondpb.Checkpoint{
								Epoch: 61,
								Root:  FillByteSlice(32, 62),
							},
						},
						Signatures: FillByteSlice(4595, 63),
					},
					Attestation_2: &zondpb.IndexedAttestation{
						AttestingIndices: []uint64{64, 65},
						Data: &zondpb.AttestationData{
							Slot:            66,
							CommitteeIndex:  67,
							BeaconBlockRoot: FillByteSlice(32, 68),
							Source: &zondpb.Checkpoint{
								Epoch: 69,
								Root:  FillByteSlice(32, 70),
							},
							Target: &zondpb.Checkpoint{
								Epoch: 71,
								Root:  FillByteSlice(32, 72),
							},
						},
						Signatures: FillByteSlice(4595, 73),
					},
				},
			},
			Attestations: []*zondpb.Attestation{
				{
					AggregationBits: FillByteSlice(32, 74),
					Data: &zondpb.AttestationData{
						Slot:            75,
						CommitteeIndex:  76,
						BeaconBlockRoot: FillByteSlice(32, 77),
						Source: &zondpb.Checkpoint{
							Epoch: 78,
							Root:  FillByteSlice(32, 79),
						},
						Target: &zondpb.Checkpoint{
							Epoch: 80,
							Root:  FillByteSlice(32, 81),
						},
					},
					Signatures: FillByteSlice(4595, 82),
				},
				{
					AggregationBits: FillByteSlice(4, 83),
					Data: &zondpb.AttestationData{
						Slot:            84,
						CommitteeIndex:  85,
						BeaconBlockRoot: FillByteSlice(32, 38),
						Source: &zondpb.Checkpoint{
							Epoch: 87,
							Root:  FillByteSlice(32, 88),
						},
						Target: &zondpb.Checkpoint{
							Epoch: 89,
							Root:  FillByteSlice(32, 90),
						},
					},
					Signatures: FillByteSlice(4595, 91),
				},
			},
			Deposits: []*zondpb.Deposit{
				{
					Proof: FillByteArraySlice(33, FillByteSlice(32, 92)),
					Data: &zondpb.Deposit_Data{
						PublicKey:             FillByteSlice(48, 94),
						WithdrawalCredentials: FillByteSlice(32, 95),
						Amount:                96,
						Signature:             FillByteSlice(4595, 97),
					},
				},
				{
					Proof: FillByteArraySlice(33, FillByteSlice(32, 98)),
					Data: &zondpb.Deposit_Data{
						PublicKey:             FillByteSlice(48, 100),
						WithdrawalCredentials: FillByteSlice(32, 101),
						Amount:                102,
						Signature:             FillByteSlice(4595, 103),
					},
				},
			},
			VoluntaryExits: []*zondpb.SignedVoluntaryExit{
				{
					Exit: &zondpb.VoluntaryExit{
						Epoch:          104,
						ValidatorIndex: 105,
					},
					Signature: FillByteSlice(4595, 106),
				},
				{
					Exit: &zondpb.VoluntaryExit{
						Epoch:          107,
						ValidatorIndex: 108,
					},
					Signature: FillByteSlice(4595, 109),
				},
			},
		},
	}
}

func GenerateJsonPhase0BeaconBlock() *apimiddleware.BeaconBlockJson {
	return &apimiddleware.BeaconBlockJson{
		Slot:          "1",
		ProposerIndex: "2",
		ParentRoot:    FillEncodedByteSlice(32, 3),
		StateRoot:     FillEncodedByteSlice(32, 4),
		Body: &apimiddleware.BeaconBlockBodyJson{
			RandaoReveal: FillEncodedByteSlice(4595, 5),
			Zond1Data: &apimiddleware.Zond1DataJson{
				DepositRoot:  FillEncodedByteSlice(32, 6),
				DepositCount: "7",
				BlockHash:    FillEncodedByteSlice(32, 8),
			},
			Graffiti: FillEncodedByteSlice(32, 9),
			ProposerSlashings: []*apimiddleware.ProposerSlashingJson{
				{
					Header_1: &apimiddleware.SignedBeaconBlockHeaderJson{
						Header: &apimiddleware.BeaconBlockHeaderJson{
							Slot:          "10",
							ProposerIndex: "11",
							ParentRoot:    FillEncodedByteSlice(32, 12),
							StateRoot:     FillEncodedByteSlice(32, 13),
							BodyRoot:      FillEncodedByteSlice(32, 14),
						},
						Signature: FillEncodedByteSlice(4595, 15),
					},
					Header_2: &apimiddleware.SignedBeaconBlockHeaderJson{
						Header: &apimiddleware.BeaconBlockHeaderJson{
							Slot:          "16",
							ProposerIndex: "17",
							ParentRoot:    FillEncodedByteSlice(32, 18),
							StateRoot:     FillEncodedByteSlice(32, 19),
							BodyRoot:      FillEncodedByteSlice(32, 20),
						},
						Signature: FillEncodedByteSlice(4595, 21),
					},
				},
				{
					Header_1: &apimiddleware.SignedBeaconBlockHeaderJson{
						Header: &apimiddleware.BeaconBlockHeaderJson{
							Slot:          "22",
							ProposerIndex: "23",
							ParentRoot:    FillEncodedByteSlice(32, 24),
							StateRoot:     FillEncodedByteSlice(32, 25),
							BodyRoot:      FillEncodedByteSlice(32, 26),
						},
						Signature: FillEncodedByteSlice(4595, 27),
					},
					Header_2: &apimiddleware.SignedBeaconBlockHeaderJson{
						Header: &apimiddleware.BeaconBlockHeaderJson{
							Slot:          "28",
							ProposerIndex: "29",
							ParentRoot:    FillEncodedByteSlice(32, 30),
							StateRoot:     FillEncodedByteSlice(32, 31),
							BodyRoot:      FillEncodedByteSlice(32, 32),
						},
						Signature: FillEncodedByteSlice(4595, 33),
					},
				},
			},
			AttesterSlashings: []*apimiddleware.AttesterSlashingJson{
				{
					Attestation_1: &apimiddleware.IndexedAttestationJson{
						AttestingIndices: []string{"34", "35"},
						Data: &apimiddleware.AttestationDataJson{
							Slot:            "36",
							CommitteeIndex:  "37",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &apimiddleware.CheckpointJson{
								Epoch: "39",
								Root:  FillEncodedByteSlice(32, 40),
							},
							Target: &apimiddleware.CheckpointJson{
								Epoch: "41",
								Root:  FillEncodedByteSlice(32, 42),
							},
						},
						Signature: FillEncodedByteSlice(4595, 43),
					},
					Attestation_2: &apimiddleware.IndexedAttestationJson{
						AttestingIndices: []string{"44", "45"},
						Data: &apimiddleware.AttestationDataJson{
							Slot:            "46",
							CommitteeIndex:  "47",
							BeaconBlockRoot: FillEncodedByteSlice(32, 48),
							Source: &apimiddleware.CheckpointJson{
								Epoch: "49",
								Root:  FillEncodedByteSlice(32, 50),
							},
							Target: &apimiddleware.CheckpointJson{
								Epoch: "51",
								Root:  FillEncodedByteSlice(32, 52),
							},
						},
						Signature: FillEncodedByteSlice(4595, 53),
					},
				},
				{
					Attestation_1: &apimiddleware.IndexedAttestationJson{
						AttestingIndices: []string{"54", "55"},
						Data: &apimiddleware.AttestationDataJson{
							Slot:            "56",
							CommitteeIndex:  "57",
							BeaconBlockRoot: FillEncodedByteSlice(32, 58),
							Source: &apimiddleware.CheckpointJson{
								Epoch: "59",
								Root:  FillEncodedByteSlice(32, 60),
							},
							Target: &apimiddleware.CheckpointJson{
								Epoch: "61",
								Root:  FillEncodedByteSlice(32, 62),
							},
						},
						Signature: FillEncodedByteSlice(4595, 63),
					},
					Attestation_2: &apimiddleware.IndexedAttestationJson{
						AttestingIndices: []string{"64", "65"},
						Data: &apimiddleware.AttestationDataJson{
							Slot:            "66",
							CommitteeIndex:  "67",
							BeaconBlockRoot: FillEncodedByteSlice(32, 68),
							Source: &apimiddleware.CheckpointJson{
								Epoch: "69",
								Root:  FillEncodedByteSlice(32, 70),
							},
							Target: &apimiddleware.CheckpointJson{
								Epoch: "71",
								Root:  FillEncodedByteSlice(32, 72),
							},
						},
						Signature: FillEncodedByteSlice(4595, 73),
					},
				},
			},
			Attestations: []*apimiddleware.AttestationJson{
				{
					AggregationBits: FillEncodedByteSlice(32, 74),
					Data: &apimiddleware.AttestationDataJson{
						Slot:            "75",
						CommitteeIndex:  "76",
						BeaconBlockRoot: FillEncodedByteSlice(32, 77),
						Source: &apimiddleware.CheckpointJson{
							Epoch: "78",
							Root:  FillEncodedByteSlice(32, 79),
						},
						Target: &apimiddleware.CheckpointJson{
							Epoch: "80",
							Root:  FillEncodedByteSlice(32, 81),
						},
					},
					Signature: FillEncodedByteSlice(4595, 82),
				},
				{
					AggregationBits: FillEncodedByteSlice(4, 83),
					Data: &apimiddleware.AttestationDataJson{
						Slot:            "84",
						CommitteeIndex:  "85",
						BeaconBlockRoot: FillEncodedByteSlice(32, 38),
						Source: &apimiddleware.CheckpointJson{
							Epoch: "87",
							Root:  FillEncodedByteSlice(32, 88),
						},
						Target: &apimiddleware.CheckpointJson{
							Epoch: "89",
							Root:  FillEncodedByteSlice(32, 90),
						},
					},
					Signature: FillEncodedByteSlice(4595, 91),
				},
			},
			Deposits: []*apimiddleware.DepositJson{
				{
					Proof: FillEncodedByteArraySlice(33, FillEncodedByteSlice(32, 92)),
					Data: &apimiddleware.Deposit_DataJson{
						PublicKey:             FillEncodedByteSlice(48, 94),
						WithdrawalCredentials: FillEncodedByteSlice(32, 95),
						Amount:                "96",
						Signature:             FillEncodedByteSlice(4595, 97),
					},
				},
				{
					Proof: FillEncodedByteArraySlice(33, FillEncodedByteSlice(32, 98)),
					Data: &apimiddleware.Deposit_DataJson{
						PublicKey:             FillEncodedByteSlice(48, 100),
						WithdrawalCredentials: FillEncodedByteSlice(32, 101),
						Amount:                "102",
						Signature:             FillEncodedByteSlice(4595, 103),
					},
				},
			},
			VoluntaryExits: []*apimiddleware.SignedVoluntaryExitJson{
				{
					Exit: &apimiddleware.VoluntaryExitJson{
						Epoch:          "104",
						ValidatorIndex: "105",
					},
					Signature: FillEncodedByteSlice(4595, 106),
				},
				{
					Exit: &apimiddleware.VoluntaryExitJson{
						Epoch:          "107",
						ValidatorIndex: "108",
					},
					Signature: FillEncodedByteSlice(4595, 109),
				},
			},
		},
	}
}
