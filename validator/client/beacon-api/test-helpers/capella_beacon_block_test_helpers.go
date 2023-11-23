package test_helpers

import (
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/apimiddleware"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	enginev1 "github.com/theQRL/qrysm/v4/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

func GenerateProtoCapellaBeaconBlock() *zondpb.BeaconBlock {
	return &zondpb.BeaconBlock{
		Slot:          1,
		ProposerIndex: 2,
		ParentRoot:    FillByteSlice(32, 3),
		StateRoot:     FillByteSlice(32, 4),
		Body: &zondpb.BeaconBlockBody{
			RandaoReveal: FillByteSlice(96, 5),
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
						Signature: FillByteSlice(96, 15),
					},
					Header_2: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							Slot:          16,
							ProposerIndex: 17,
							ParentRoot:    FillByteSlice(32, 18),
							StateRoot:     FillByteSlice(32, 19),
							BodyRoot:      FillByteSlice(32, 20),
						},
						Signature: FillByteSlice(96, 21),
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
						Signature: FillByteSlice(96, 27),
					},
					Header_2: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							Slot:          28,
							ProposerIndex: 29,
							ParentRoot:    FillByteSlice(32, 30),
							StateRoot:     FillByteSlice(32, 31),
							BodyRoot:      FillByteSlice(32, 32),
						},
						Signature: FillByteSlice(96, 33),
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
						Signatures: [][]byte{FillByteSlice(96, 43)},
					},
					Attestation_2: &zondpb.IndexedAttestation{
						AttestingIndices: []uint64{44, 45},
						Data: &zondpb.AttestationData{
							Slot:            46,
							CommitteeIndex:  47,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &zondpb.Checkpoint{
								Epoch: 49,
								Root:  FillByteSlice(32, 50),
							},
							Target: &zondpb.Checkpoint{
								Epoch: 51,
								Root:  FillByteSlice(32, 52),
							},
						},
						Signatures: [][]byte{FillByteSlice(96, 53)},
					},
				},
				{
					Attestation_1: &zondpb.IndexedAttestation{
						AttestingIndices: []uint64{54, 55},
						Data: &zondpb.AttestationData{
							Slot:            56,
							CommitteeIndex:  57,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &zondpb.Checkpoint{
								Epoch: 59,
								Root:  FillByteSlice(32, 60),
							},
							Target: &zondpb.Checkpoint{
								Epoch: 61,
								Root:  FillByteSlice(32, 62),
							},
						},
						Signatures: [][]byte{FillByteSlice(96, 63)},
					},
					Attestation_2: &zondpb.IndexedAttestation{
						AttestingIndices: []uint64{64, 65},
						Data: &zondpb.AttestationData{
							Slot:            66,
							CommitteeIndex:  67,
							BeaconBlockRoot: FillByteSlice(32, 38),
							Source: &zondpb.Checkpoint{
								Epoch: 69,
								Root:  FillByteSlice(32, 70),
							},
							Target: &zondpb.Checkpoint{
								Epoch: 71,
								Root:  FillByteSlice(32, 72),
							},
						},
						Signatures: [][]byte{FillByteSlice(96, 73)},
					},
				},
			},
			Attestations: []*zondpb.Attestation{
				{
					ParticipationBits: FillByteSlice(4, 74),
					Data: &zondpb.AttestationData{
						Slot:            75,
						CommitteeIndex:  76,
						BeaconBlockRoot: FillByteSlice(32, 38),
						Source: &zondpb.Checkpoint{
							Epoch: 78,
							Root:  FillByteSlice(32, 79),
						},
						Target: &zondpb.Checkpoint{
							Epoch: 80,
							Root:  FillByteSlice(32, 81),
						},
					},
					Signatures: [][]byte{FillByteSlice(96, 82)},
				},
				{
					ParticipationBits: FillByteSlice(4, 83),
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
					Signatures: [][]byte{FillByteSlice(96, 91)},
				},
			},
			Deposits: []*zondpb.Deposit{
				{
					Proof: FillByteArraySlice(33, FillByteSlice(32, 92)),
					Data: &zondpb.Deposit_Data{
						PublicKey:             FillByteSlice(48, 94),
						WithdrawalCredentials: FillByteSlice(32, 95),
						Amount:                96,
						Signature:             FillByteSlice(96, 97),
					},
				},
				{
					Proof: FillByteArraySlice(33, FillByteSlice(32, 98)),
					Data: &zondpb.Deposit_Data{
						PublicKey:             FillByteSlice(48, 100),
						WithdrawalCredentials: FillByteSlice(32, 101),
						Amount:                102,
						Signature:             FillByteSlice(96, 103),
					},
				},
			},
			VoluntaryExits: []*zondpb.SignedVoluntaryExit{
				{
					Exit: &zondpb.VoluntaryExit{
						Epoch:          104,
						ValidatorIndex: 105,
					},
					Signature: FillByteSlice(96, 106),
				},
				{
					Exit: &zondpb.VoluntaryExit{
						Epoch:          107,
						ValidatorIndex: 108,
					},
					Signature: FillByteSlice(96, 109),
				},
			},
			SyncAggregate: &zondpb.SyncAggregate{
				SyncCommitteeBits:       FillByteSlice(64, 110),
				SyncCommitteeSignatures: [][]byte{FillByteSlice(96, 111)},
			},
			ExecutionPayload: &enginev1.ExecutionPayload{
				ParentHash:    FillByteSlice(32, 112),
				FeeRecipient:  FillByteSlice(20, 113),
				StateRoot:     FillByteSlice(32, 114),
				ReceiptsRoot:  FillByteSlice(32, 115),
				LogsBloom:     FillByteSlice(256, 116),
				PrevRandao:    FillByteSlice(32, 117),
				BlockNumber:   118,
				GasLimit:      119,
				GasUsed:       120,
				Timestamp:     121,
				ExtraData:     FillByteSlice(32, 122),
				BaseFeePerGas: FillByteSlice(32, 123),
				BlockHash:     FillByteSlice(32, 124),
				Transactions: [][]byte{
					FillByteSlice(32, 125),
					FillByteSlice(32, 126),
				},
				Withdrawals: []*enginev1.Withdrawal{
					{
						Index:          127,
						ValidatorIndex: 128,
						Address:        FillByteSlice(20, 129),
						Amount:         130,
					},
					{
						Index:          131,
						ValidatorIndex: 132,
						Address:        FillByteSlice(20, 133),
						Amount:         134,
					},
				},
			},
			DilithiumToExecutionChanges: []*zondpb.SignedDilithiumToExecutionChange{
				{
					Message: &zondpb.DilithiumToExecutionChange{
						ValidatorIndex:      135,
						FromDilithiumPubkey: FillByteSlice(48, 136),
						ToExecutionAddress:  FillByteSlice(20, 137),
					},
					Signature: FillByteSlice(96, 138),
				},
				{
					Message: &zondpb.DilithiumToExecutionChange{
						ValidatorIndex:      139,
						FromDilithiumPubkey: FillByteSlice(48, 140),
						ToExecutionAddress:  FillByteSlice(20, 141),
					},
					Signature: FillByteSlice(96, 142),
				},
			},
		},
	}
}

func GenerateJsonCapellaBeaconBlock() *apimiddleware.BeaconBlockJson {
	return &apimiddleware.BeaconBlockJson{
		Slot:          "1",
		ProposerIndex: "2",
		ParentRoot:    FillEncodedByteSlice(32, 3),
		StateRoot:     FillEncodedByteSlice(32, 4),
		Body: &apimiddleware.BeaconBlockBodyJson{
			RandaoReveal: FillEncodedByteSlice(96, 5),
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
						Signature: FillEncodedByteSlice(96, 15),
					},
					Header_2: &apimiddleware.SignedBeaconBlockHeaderJson{
						Header: &apimiddleware.BeaconBlockHeaderJson{
							Slot:          "16",
							ProposerIndex: "17",
							ParentRoot:    FillEncodedByteSlice(32, 18),
							StateRoot:     FillEncodedByteSlice(32, 19),
							BodyRoot:      FillEncodedByteSlice(32, 20),
						},
						Signature: FillEncodedByteSlice(96, 21),
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
						Signature: FillEncodedByteSlice(96, 27),
					},
					Header_2: &apimiddleware.SignedBeaconBlockHeaderJson{
						Header: &apimiddleware.BeaconBlockHeaderJson{
							Slot:          "28",
							ProposerIndex: "29",
							ParentRoot:    FillEncodedByteSlice(32, 30),
							StateRoot:     FillEncodedByteSlice(32, 31),
							BodyRoot:      FillEncodedByteSlice(32, 32),
						},
						Signature: FillEncodedByteSlice(96, 33),
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
						Signatures: []string{FillEncodedByteSlice(4595, 43)},
					},
					Attestation_2: &apimiddleware.IndexedAttestationJson{
						AttestingIndices: []string{"44", "45"},
						Data: &apimiddleware.AttestationDataJson{
							Slot:            "46",
							CommitteeIndex:  "47",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &apimiddleware.CheckpointJson{
								Epoch: "49",
								Root:  FillEncodedByteSlice(32, 50),
							},
							Target: &apimiddleware.CheckpointJson{
								Epoch: "51",
								Root:  FillEncodedByteSlice(32, 52),
							},
						},
						Signatures: []string{FillEncodedByteSlice(96, 53)},
					},
				},
				{
					Attestation_1: &apimiddleware.IndexedAttestationJson{
						AttestingIndices: []string{"54", "55"},
						Data: &apimiddleware.AttestationDataJson{
							Slot:            "56",
							CommitteeIndex:  "57",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &apimiddleware.CheckpointJson{
								Epoch: "59",
								Root:  FillEncodedByteSlice(32, 60),
							},
							Target: &apimiddleware.CheckpointJson{
								Epoch: "61",
								Root:  FillEncodedByteSlice(32, 62),
							},
						},
						Signatures: []string{FillEncodedByteSlice(96, 63)},
					},
					Attestation_2: &apimiddleware.IndexedAttestationJson{
						AttestingIndices: []string{"64", "65"},
						Data: &apimiddleware.AttestationDataJson{
							Slot:            "66",
							CommitteeIndex:  "67",
							BeaconBlockRoot: FillEncodedByteSlice(32, 38),
							Source: &apimiddleware.CheckpointJson{
								Epoch: "69",
								Root:  FillEncodedByteSlice(32, 70),
							},
							Target: &apimiddleware.CheckpointJson{
								Epoch: "71",
								Root:  FillEncodedByteSlice(32, 72),
							},
						},
						Signatures: []string{FillEncodedByteSlice(96, 73)},
					},
				},
			},
			Attestations: []*apimiddleware.AttestationJson{
				{
					ParticipationBits: FillEncodedByteSlice(4, 74),
					Data: &apimiddleware.AttestationDataJson{
						Slot:            "75",
						CommitteeIndex:  "76",
						BeaconBlockRoot: FillEncodedByteSlice(32, 38),
						Source: &apimiddleware.CheckpointJson{
							Epoch: "78",
							Root:  FillEncodedByteSlice(32, 79),
						},
						Target: &apimiddleware.CheckpointJson{
							Epoch: "80",
							Root:  FillEncodedByteSlice(32, 81),
						},
					},
					Signatures: []string{FillEncodedByteSlice(96, 82)},
				},
				{
					ParticipationBits: FillEncodedByteSlice(4, 83),
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
					Signatures: []string{FillEncodedByteSlice(96, 91)},
				},
			},
			Deposits: []*apimiddleware.DepositJson{
				{
					Proof: FillEncodedByteArraySlice(33, FillEncodedByteSlice(32, 92)),
					Data: &apimiddleware.Deposit_DataJson{
						PublicKey:             FillEncodedByteSlice(48, 94),
						WithdrawalCredentials: FillEncodedByteSlice(32, 95),
						Amount:                "96",
						Signature:             FillEncodedByteSlice(96, 97),
					},
				},
				{
					Proof: FillEncodedByteArraySlice(33, FillEncodedByteSlice(32, 98)),
					Data: &apimiddleware.Deposit_DataJson{
						PublicKey:             FillEncodedByteSlice(48, 100),
						WithdrawalCredentials: FillEncodedByteSlice(32, 101),
						Amount:                "102",
						Signature:             FillEncodedByteSlice(96, 103),
					},
				},
			},
			VoluntaryExits: []*apimiddleware.SignedVoluntaryExitJson{
				{
					Exit: &apimiddleware.VoluntaryExitJson{
						Epoch:          "104",
						ValidatorIndex: "105",
					},
					Signature: FillEncodedByteSlice(96, 106),
				},
				{
					Exit: &apimiddleware.VoluntaryExitJson{
						Epoch:          "107",
						ValidatorIndex: "108",
					},
					Signature: FillEncodedByteSlice(96, 109),
				},
			},
			SyncAggregate: &apimiddleware.SyncAggregateJson{
				SyncCommitteeBits:           FillEncodedByteSlice(64, 110),
				SyncCommitteeSignatures:     []string{FillEncodedByteSlice(96, 111)},
				SignaturesIdxToCommitteeIdx: []string{},
			},
			ExecutionPayload: &apimiddleware.ExecutionPayloadJson{
				ParentHash:    FillEncodedByteSlice(32, 112),
				FeeRecipient:  FillEncodedByteSlice(20, 113),
				StateRoot:     FillEncodedByteSlice(32, 114),
				ReceiptsRoot:  FillEncodedByteSlice(32, 115),
				LogsBloom:     FillEncodedByteSlice(256, 116),
				PrevRandao:    FillEncodedByteSlice(32, 117),
				BlockNumber:   "118",
				GasLimit:      "119",
				GasUsed:       "120",
				TimeStamp:     "121",
				ExtraData:     FillEncodedByteSlice(32, 122),
				BaseFeePerGas: bytesutil.LittleEndianBytesToBigInt(FillByteSlice(32, 123)).String(),
				BlockHash:     FillEncodedByteSlice(32, 124),
				Transactions: []string{
					FillEncodedByteSlice(32, 125),
					FillEncodedByteSlice(32, 126),
				},
				Withdrawals: []*apimiddleware.WithdrawalJson{
					{
						WithdrawalIndex:  "127",
						ValidatorIndex:   "128",
						ExecutionAddress: FillEncodedByteSlice(20, 129),
						Amount:           "130",
					},
					{
						WithdrawalIndex:  "131",
						ValidatorIndex:   "132",
						ExecutionAddress: FillEncodedByteSlice(20, 133),
						Amount:           "134",
					},
				},
			},
			DilithiumToExecutionChanges: []*apimiddleware.SignedDilithiumToExecutionChangeJson{
				{
					Message: &apimiddleware.DilithiumToExecutionChangeJson{
						ValidatorIndex:      "135",
						FromDilithiumPubkey: FillEncodedByteSlice(48, 136),
						ToExecutionAddress:  FillEncodedByteSlice(20, 137),
					},
					Signature: FillEncodedByteSlice(96, 138),
				},
				{
					Message: &apimiddleware.DilithiumToExecutionChangeJson{
						ValidatorIndex:      "139",
						FromDilithiumPubkey: FillEncodedByteSlice(48, 140),
						ToExecutionAddress:  FillEncodedByteSlice(20, 141),
					},
					Signature: FillEncodedByteSlice(96, 142),
				},
			},
		},
	}
}
