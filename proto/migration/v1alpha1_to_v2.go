package migration

import (
	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/v4/beacon-chain/state"
	fieldparams "github.com/theQRL/qrysm/v4/config/fieldparams"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	"github.com/theQRL/qrysm/v4/encoding/ssz"
	enginev1 "github.com/theQRL/qrysm/v4/proto/engine/v1"
	zondpbalpha "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	zondpbv1 "github.com/theQRL/qrysm/v4/proto/zond/v1"
	zondpbv2 "github.com/theQRL/qrysm/v4/proto/zond/v2"
	"google.golang.org/protobuf/proto"
)

// V1Alpha1BeaconBlockAltairToV2 converts a v1alpha1 Altair beacon block to a v2 Altair block.
func V1Alpha1BeaconBlockAltairToV2(v1alpha1Block *zondpbalpha.BeaconBlockAltair) (*zondpbv2.BeaconBlockAltair, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.BeaconBlockAltair{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BeaconBlockAltairToV2Signed converts a v1alpha1 Altair signed beacon block to a v2 Altair block.
func V1Alpha1BeaconBlockAltairToV2Signed(v1alpha1Block *zondpbalpha.SignedBeaconBlockAltair) (*zondpbv2.SignedBeaconBlockAltair, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.SignedBeaconBlockAltair{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BeaconBlockBellatrixToV2 converts a v1alpha1 Bellatrix beacon block to a v2
// Bellatrix block.
func V1Alpha1BeaconBlockBellatrixToV2(v1alpha1Block *zondpbalpha.BeaconBlockBellatrix) (*zondpbv2.BeaconBlockBellatrix, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.BeaconBlockBellatrix{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BeaconBlockCapellaToV2 converts a v1alpha1 Capella beacon block to a v2
// Capella block.
func V1Alpha1BeaconBlockCapellaToV2(v1alpha1Block *zondpbalpha.BeaconBlockCapella) (*zondpbv2.BeaconBlockCapella, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.BeaconBlockCapella{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BeaconBlockDenebToV2 converts a v1alpha1 Deneb beacon block to a v2
// Deneb block.
func V1Alpha1BeaconBlockDenebToV2(v1alpha1Block *zondpbalpha.BeaconBlockDeneb) (*zondpbv2.BeaconBlockDeneb, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.BeaconBlockDeneb{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1SignedBeaconBlockDenebToV2 converts a v1alpha1 signed Deneb beacon block to a v2
// Deneb block.
func V1Alpha1SignedBeaconBlockDenebToV2(v1alpha1Block *zondpbalpha.SignedBeaconBlockDeneb) (*zondpbv2.SignedBeaconBlockDeneb, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.SignedBeaconBlockDeneb{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BlobSidecarsToV2 converts an array of v1alpha1 blinded blob sidecars to its v2 equivalent.
func V1Alpha1BlobSidecarsToV2(v1alpha1Blobs []*zondpbalpha.BlobSidecar) ([]*zondpbv2.BlobSidecar, error) {
	v2Blobs := make([]*zondpbv2.BlobSidecar, len(v1alpha1Blobs))
	for index, v1Blob := range v1alpha1Blobs {
		marshaledBlob, err := proto.Marshal(v1Blob)
		if err != nil {
			return nil, errors.Wrap(err, "could not marshal blob sidecar")
		}
		v2Blob := &zondpbv2.BlobSidecar{}
		if err := proto.Unmarshal(marshaledBlob, v2Blob); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal blob sidecar")
		}
		v2Blobs[index] = v2Blob
	}
	return v2Blobs, nil
}

// V1Alpha1BlindedBlobSidecarsToV2 converts an array of v1alpha1 blinded blob sidecars to its v2 equivalent.
func V1Alpha1BlindedBlobSidecarsToV2(v1alpha1Blobs []*zondpbalpha.BlindedBlobSidecar) ([]*zondpbv2.BlindedBlobSidecar, error) {
	v2Blobs := make([]*zondpbv2.BlindedBlobSidecar, len(v1alpha1Blobs))
	for index, v1Blob := range v1alpha1Blobs {
		marshaledBlob, err := proto.Marshal(v1Blob)
		if err != nil {
			return nil, errors.Wrap(err, "could not marshal blob sidecar")
		}
		v2Blob := &zondpbv2.BlindedBlobSidecar{}
		if err := proto.Unmarshal(marshaledBlob, v2Blob); err != nil {
			return nil, errors.Wrap(err, "could not unmarshal blob sidecar")
		}
		v2Blobs[index] = v2Blob
	}
	return v2Blobs, nil
}

// V1Alpha1SignedBlindedBlobSidecarsToV2 converts an array of v1alpha1 objects to its v2 SignedBlindedBlobSidecar equivalent.
func V1Alpha1SignedBlindedBlobSidecarsToV2(sidecars []*zondpbalpha.SignedBlindedBlobSidecar) []*zondpbv2.SignedBlindedBlobSidecar {
	result := make([]*zondpbv2.SignedBlindedBlobSidecar, len(sidecars))
	for i, sc := range sidecars {
		result[i] = &zondpbv2.SignedBlindedBlobSidecar{
			Message: &zondpbv2.BlindedBlobSidecar{
				BlockRoot:       bytesutil.SafeCopyBytes(sc.Message.BlockRoot),
				Index:           sc.Message.Index,
				Slot:            sc.Message.Slot,
				BlockParentRoot: bytesutil.SafeCopyBytes(sc.Message.BlockParentRoot),
				ProposerIndex:   sc.Message.ProposerIndex,
				BlobRoot:        bytesutil.SafeCopyBytes(sc.Message.BlobRoot),
				KzgCommitment:   bytesutil.SafeCopyBytes(sc.Message.KzgCommitment),
				KzgProof:        bytesutil.SafeCopyBytes(sc.Message.KzgProof),
			},
			Signature: bytesutil.SafeCopyBytes(sc.Signature),
		}
	}
	return result
}

// V1Alpha1SignedBlobsToV2 converts an array of v1alpha1 objects to its v2 SignedBlobSidecar equivalent.
func V1Alpha1SignedBlobsToV2(sidecars []*zondpbalpha.SignedBlobSidecar) []*zondpbv2.SignedBlobSidecar {
	result := make([]*zondpbv2.SignedBlobSidecar, len(sidecars))
	for i, sc := range sidecars {
		result[i] = V1Alpha1SignedBlobToV2(sc)
	}
	return result
}

// V1Alpha1SignedBlobToV2 converts a v1alpha1 object to its v2 SignedBlobSidecar equivalent.
func V1Alpha1SignedBlobToV2(sidecar *zondpbalpha.SignedBlobSidecar) *zondpbv2.SignedBlobSidecar {
	if sidecar == nil || sidecar.Message == nil {
		return &zondpbv2.SignedBlobSidecar{}
	}
	return &zondpbv2.SignedBlobSidecar{
		Message: &zondpbv2.BlobSidecar{
			BlockRoot:       bytesutil.SafeCopyBytes(sidecar.Message.BlockRoot),
			Index:           sidecar.Message.Index,
			Slot:            sidecar.Message.Slot,
			BlockParentRoot: bytesutil.SafeCopyBytes(sidecar.Message.BlockParentRoot),
			ProposerIndex:   sidecar.Message.ProposerIndex,
			Blob:            bytesutil.SafeCopyBytes(sidecar.Message.Blob),
			KzgCommitment:   bytesutil.SafeCopyBytes(sidecar.Message.KzgCommitment),
			KzgProof:        bytesutil.SafeCopyBytes(sidecar.Message.KzgProof),
		},
		Signature: bytesutil.SafeCopyBytes(sidecar.Signature),
	}
}

// V1Alpha1BeaconBlockDenebAndBlobsToV2 converts a v1alpha1 Deneb beacon block and blobs to a v2
// Deneb block.
func V1Alpha1BeaconBlockDenebAndBlobsToV2(v1alpha1Block *zondpbalpha.BeaconBlockAndBlobsDeneb) (*zondpbv2.BeaconBlockContentsDeneb, error) {
	v2Block, err := V1Alpha1BeaconBlockDenebToV2(v1alpha1Block.Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert block")
	}
	v2Blobs, err := V1Alpha1BlobSidecarsToV2(v1alpha1Block.Blobs)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert blobs")
	}
	return &zondpbv2.BeaconBlockContentsDeneb{Block: v2Block, BlobSidecars: v2Blobs}, nil
}

// V1Alpha1SignedBeaconBlockDenebAndBlobsToV2 converts a signed v1alpha1 Deneb beacon block and blobs to a v2
// Deneb block.
func V1Alpha1SignedBeaconBlockDenebAndBlobsToV2(v1alpha1Block *zondpbalpha.SignedBeaconBlockAndBlobsDeneb) (*zondpbv2.SignedBeaconBlockContentsDeneb, error) {
	v2Block, err := V1Alpha1SignedBeaconBlockDenebToV2(v1alpha1Block.Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert block")
	}
	v2Blobs := V1Alpha1SignedBlobsToV2(v1alpha1Block.Blobs)
	return &zondpbv2.SignedBeaconBlockContentsDeneb{
		SignedBlock:        v2Block,
		SignedBlobSidecars: v2Blobs,
	}, nil
}

// V1Alpha1BeaconBlockBlindedBellatrixToV2Blinded converts a v1alpha1 Blinded Bellatrix beacon block to a v2 Blinded Bellatrix block.
func V1Alpha1BeaconBlockBlindedBellatrixToV2Blinded(v1alpha1Block *zondpbalpha.BlindedBeaconBlockBellatrix) (*zondpbv2.BlindedBeaconBlockBellatrix, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.BlindedBeaconBlockBellatrix{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BeaconBlockBlindedCapellaToV2Blinded converts a v1alpha1 Blinded Capella beacon block to a v2 Blinded Capella block.
func V1Alpha1BeaconBlockBlindedCapellaToV2Blinded(v1alpha1Block *zondpbalpha.BlindedBeaconBlockCapella) (*zondpbv2.BlindedBeaconBlockCapella, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.BlindedBeaconBlockCapella{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BeaconBlockBlindedDenebToV2Blinded converts a v1alpha1 Blinded Deneb beacon block to a v2 Blinded Deneb block.
func V1Alpha1BeaconBlockBlindedDenebToV2Blinded(v1alpha1Block *zondpbalpha.BlindedBeaconBlockDeneb) (*zondpbv2.BlindedBeaconBlockDeneb, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.BlindedBeaconBlockDeneb{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1SignedBeaconBlockBlindedDenebToV2Blinded converts a v1alpha1 Signed Blinded Deneb beacon block to a v2 Blinded Deneb block.
func V1Alpha1SignedBeaconBlockBlindedDenebToV2Blinded(v1alpha1Block *zondpbalpha.SignedBlindedBeaconBlockDeneb) (*zondpbv2.SignedBlindedBeaconBlockDeneb, error) {
	marshaledBlk, err := proto.Marshal(v1alpha1Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not marshal block")
	}
	v2Block := &zondpbv2.SignedBlindedBeaconBlockDeneb{}
	if err := proto.Unmarshal(marshaledBlk, v2Block); err != nil {
		return nil, errors.Wrap(err, "could not unmarshal block")
	}
	return v2Block, nil
}

// V1Alpha1BlindedBlockAndBlobsDenebToV2Blinded converts a v1alpha1 Deneb blinded beacon block and blobs to v2 blinded block contents.
func V1Alpha1BlindedBlockAndBlobsDenebToV2Blinded(
	v1Alpha1BlkAndBlobs *zondpbalpha.BlindedBeaconBlockAndBlobsDeneb,
) (*zondpbv2.BlindedBeaconBlockContentsDeneb, error) {
	v2Block, err := V1Alpha1BeaconBlockBlindedDenebToV2Blinded(v1Alpha1BlkAndBlobs.Block)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert block")
	}
	v2Blobs, err := V1Alpha1BlindedBlobSidecarsToV2(v1Alpha1BlkAndBlobs.Blobs)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert blobs")
	}
	return &zondpbv2.BlindedBeaconBlockContentsDeneb{BlindedBlock: v2Block, BlindedBlobSidecars: v2Blobs}, nil
}

// V1Alpha1SignedBlindedBlockAndBlobsDenebToV2Blinded converts a v1alpha1 signed Deneb blinded beacon block and blobs to v2 blinded block contents.
func V1Alpha1SignedBlindedBlockAndBlobsDenebToV2Blinded(
	v1Alpha1BlkAndBlobs *zondpbalpha.SignedBlindedBeaconBlockAndBlobsDeneb,
) (*zondpbv2.SignedBlindedBeaconBlockContentsDeneb, error) {
	v2Block, err := V1Alpha1SignedBeaconBlockBlindedDenebToV2Blinded(v1Alpha1BlkAndBlobs.SignedBlindedBlock)
	if err != nil {
		return nil, errors.Wrap(err, "could not convert block")
	}
	v2Blobs := V1Alpha1SignedBlindedBlobSidecarsToV2(v1Alpha1BlkAndBlobs.SignedBlindedBlobSidecars)
	return &zondpbv2.SignedBlindedBeaconBlockContentsDeneb{
		SignedBlindedBlock:        v2Block,
		SignedBlindedBlobSidecars: v2Blobs,
	}, nil
}

// V1Alpha1BeaconBlockBellatrixToV2Blinded converts a v1alpha1 Bellatrix beacon block to a v2
// blinded Bellatrix block.
func V1Alpha1BeaconBlockBellatrixToV2Blinded(v1alpha1Block *zondpbalpha.BeaconBlockBellatrix) (*zondpbv2.BlindedBeaconBlockBellatrix, error) {
	sourceProposerSlashings := v1alpha1Block.Body.ProposerSlashings
	resultProposerSlashings := make([]*zondpbv1.ProposerSlashing, len(sourceProposerSlashings))
	for i, s := range sourceProposerSlashings {
		resultProposerSlashings[i] = &zondpbv1.ProposerSlashing{
			SignedHeader_1: &zondpbv1.SignedBeaconBlockHeader{
				Message: &zondpbv1.BeaconBlockHeader{
					Slot:          s.Header_1.Header.Slot,
					ProposerIndex: s.Header_1.Header.ProposerIndex,
					ParentRoot:    bytesutil.SafeCopyBytes(s.Header_1.Header.ParentRoot),
					StateRoot:     bytesutil.SafeCopyBytes(s.Header_1.Header.StateRoot),
					BodyRoot:      bytesutil.SafeCopyBytes(s.Header_1.Header.BodyRoot),
				},
				Signature: bytesutil.SafeCopyBytes(s.Header_1.Signature),
			},
			SignedHeader_2: &zondpbv1.SignedBeaconBlockHeader{
				Message: &zondpbv1.BeaconBlockHeader{
					Slot:          s.Header_2.Header.Slot,
					ProposerIndex: s.Header_2.Header.ProposerIndex,
					ParentRoot:    bytesutil.SafeCopyBytes(s.Header_2.Header.ParentRoot),
					StateRoot:     bytesutil.SafeCopyBytes(s.Header_2.Header.StateRoot),
					BodyRoot:      bytesutil.SafeCopyBytes(s.Header_2.Header.BodyRoot),
				},
				Signature: bytesutil.SafeCopyBytes(s.Header_2.Signature),
			},
		}
	}

	sourceAttesterSlashings := v1alpha1Block.Body.AttesterSlashings
	resultAttesterSlashings := make([]*zondpbv1.AttesterSlashing, len(sourceAttesterSlashings))
	for i, s := range sourceAttesterSlashings {
		att1Indices := make([]uint64, len(s.Attestation_1.AttestingIndices))
		copy(att1Indices, s.Attestation_1.AttestingIndices)
		att2Indices := make([]uint64, len(s.Attestation_2.AttestingIndices))
		copy(att2Indices, s.Attestation_2.AttestingIndices)
		resultAttesterSlashings[i] = &zondpbv1.AttesterSlashing{
			Attestation_1: &zondpbv1.IndexedAttestation{
				AttestingIndices: att1Indices,
				Data: &zondpbv1.AttestationData{
					Slot:            s.Attestation_1.Data.Slot,
					Index:           s.Attestation_1.Data.CommitteeIndex,
					BeaconBlockRoot: bytesutil.SafeCopyBytes(s.Attestation_1.Data.BeaconBlockRoot),
					Source: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_1.Data.Source.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_1.Data.Source.Root),
					},
					Target: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_1.Data.Target.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_1.Data.Target.Root),
					},
				},
				Signature: bytesutil.SafeCopyBytes(s.Attestation_1.Signature),
			},
			Attestation_2: &zondpbv1.IndexedAttestation{
				AttestingIndices: att2Indices,
				Data: &zondpbv1.AttestationData{
					Slot:            s.Attestation_2.Data.Slot,
					Index:           s.Attestation_2.Data.CommitteeIndex,
					BeaconBlockRoot: bytesutil.SafeCopyBytes(s.Attestation_2.Data.BeaconBlockRoot),
					Source: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_2.Data.Source.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_2.Data.Source.Root),
					},
					Target: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_2.Data.Target.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_2.Data.Target.Root),
					},
				},
				Signature: bytesutil.SafeCopyBytes(s.Attestation_2.Signature),
			},
		}
	}

	sourceAttestations := v1alpha1Block.Body.Attestations
	resultAttestations := make([]*zondpbv1.Attestation, len(sourceAttestations))
	for i, a := range sourceAttestations {
		resultAttestations[i] = &zondpbv1.Attestation{
			AggregationBits: bytesutil.SafeCopyBytes(a.AggregationBits),
			Data: &zondpbv1.AttestationData{
				Slot:            a.Data.Slot,
				Index:           a.Data.CommitteeIndex,
				BeaconBlockRoot: bytesutil.SafeCopyBytes(a.Data.BeaconBlockRoot),
				Source: &zondpbv1.Checkpoint{
					Epoch: a.Data.Source.Epoch,
					Root:  bytesutil.SafeCopyBytes(a.Data.Source.Root),
				},
				Target: &zondpbv1.Checkpoint{
					Epoch: a.Data.Target.Epoch,
					Root:  bytesutil.SafeCopyBytes(a.Data.Target.Root),
				},
			},
			Signature: bytesutil.SafeCopyBytes(a.Signature),
		}
	}

	sourceDeposits := v1alpha1Block.Body.Deposits
	resultDeposits := make([]*zondpbv1.Deposit, len(sourceDeposits))
	for i, d := range sourceDeposits {
		resultDeposits[i] = &zondpbv1.Deposit{
			Proof: bytesutil.SafeCopy2dBytes(d.Proof),
			Data: &zondpbv1.Deposit_Data{
				Pubkey:                bytesutil.SafeCopyBytes(d.Data.PublicKey),
				WithdrawalCredentials: bytesutil.SafeCopyBytes(d.Data.WithdrawalCredentials),
				Amount:                d.Data.Amount,
				Signature:             bytesutil.SafeCopyBytes(d.Data.Signature),
			},
		}
	}

	sourceExits := v1alpha1Block.Body.VoluntaryExits
	resultExits := make([]*zondpbv1.SignedVoluntaryExit, len(sourceExits))
	for i, e := range sourceExits {
		resultExits[i] = &zondpbv1.SignedVoluntaryExit{
			Message: &zondpbv1.VoluntaryExit{
				Epoch:          e.Exit.Epoch,
				ValidatorIndex: e.Exit.ValidatorIndex,
			},
			Signature: bytesutil.SafeCopyBytes(e.Signature),
		}
	}

	transactionsRoot, err := ssz.TransactionsRoot(v1alpha1Block.Body.ExecutionPayload.Transactions)
	if err != nil {
		return nil, errors.Wrapf(err, "could not calculate transactions root")
	}

	resultBlockBody := &zondpbv2.BlindedBeaconBlockBodyBellatrix{
		RandaoReveal: bytesutil.SafeCopyBytes(v1alpha1Block.Body.RandaoReveal),
		ZondData: &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(v1alpha1Block.Body.ZondData.DepositRoot),
			DepositCount: v1alpha1Block.Body.ZondData.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(v1alpha1Block.Body.ZondData.BlockHash),
		},
		Graffiti:          bytesutil.SafeCopyBytes(v1alpha1Block.Body.Graffiti),
		ProposerSlashings: resultProposerSlashings,
		AttesterSlashings: resultAttesterSlashings,
		Attestations:      resultAttestations,
		Deposits:          resultDeposits,
		VoluntaryExits:    resultExits,
		SyncAggregate: &zondpbv1.SyncAggregate{
			SyncCommitteeBits:      bytesutil.SafeCopyBytes(v1alpha1Block.Body.SyncAggregate.SyncCommitteeBits),
			SyncCommitteeSignature: bytesutil.SafeCopyBytes(v1alpha1Block.Body.SyncAggregate.SyncCommitteeSignature),
		},
		ExecutionPayloadHeader: &enginev1.ExecutionPayloadHeader{
			ParentHash:       bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.ParentHash),
			FeeRecipient:     bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.FeeRecipient),
			StateRoot:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.StateRoot),
			ReceiptsRoot:     bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.ReceiptsRoot),
			LogsBloom:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.LogsBloom),
			PrevRandao:       bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.PrevRandao),
			BlockNumber:      v1alpha1Block.Body.ExecutionPayload.BlockNumber,
			GasLimit:         v1alpha1Block.Body.ExecutionPayload.GasLimit,
			GasUsed:          v1alpha1Block.Body.ExecutionPayload.GasUsed,
			Timestamp:        v1alpha1Block.Body.ExecutionPayload.Timestamp,
			ExtraData:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.ExtraData),
			BaseFeePerGas:    bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.BaseFeePerGas),
			BlockHash:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.BlockHash),
			TransactionsRoot: transactionsRoot[:],
		},
	}
	v2Block := &zondpbv2.BlindedBeaconBlockBellatrix{
		Slot:          v1alpha1Block.Slot,
		ProposerIndex: v1alpha1Block.ProposerIndex,
		ParentRoot:    bytesutil.SafeCopyBytes(v1alpha1Block.ParentRoot),
		StateRoot:     bytesutil.SafeCopyBytes(v1alpha1Block.StateRoot),
		Body:          resultBlockBody,
	}
	return v2Block, nil
}

// V1Alpha1BeaconBlockCapellaToV2Blinded converts a v1alpha1 Capella beacon block to a v2
// blinded Capella block.
func V1Alpha1BeaconBlockCapellaToV2Blinded(v1alpha1Block *zondpbalpha.BeaconBlockCapella) (*zondpbv2.BlindedBeaconBlockCapella, error) {
	sourceProposerSlashings := v1alpha1Block.Body.ProposerSlashings
	resultProposerSlashings := make([]*zondpbv1.ProposerSlashing, len(sourceProposerSlashings))
	for i, s := range sourceProposerSlashings {
		resultProposerSlashings[i] = &zondpbv1.ProposerSlashing{
			SignedHeader_1: &zondpbv1.SignedBeaconBlockHeader{
				Message: &zondpbv1.BeaconBlockHeader{
					Slot:          s.Header_1.Header.Slot,
					ProposerIndex: s.Header_1.Header.ProposerIndex,
					ParentRoot:    bytesutil.SafeCopyBytes(s.Header_1.Header.ParentRoot),
					StateRoot:     bytesutil.SafeCopyBytes(s.Header_1.Header.StateRoot),
					BodyRoot:      bytesutil.SafeCopyBytes(s.Header_1.Header.BodyRoot),
				},
				Signature: bytesutil.SafeCopyBytes(s.Header_1.Signature),
			},
			SignedHeader_2: &zondpbv1.SignedBeaconBlockHeader{
				Message: &zondpbv1.BeaconBlockHeader{
					Slot:          s.Header_2.Header.Slot,
					ProposerIndex: s.Header_2.Header.ProposerIndex,
					ParentRoot:    bytesutil.SafeCopyBytes(s.Header_2.Header.ParentRoot),
					StateRoot:     bytesutil.SafeCopyBytes(s.Header_2.Header.StateRoot),
					BodyRoot:      bytesutil.SafeCopyBytes(s.Header_2.Header.BodyRoot),
				},
				Signature: bytesutil.SafeCopyBytes(s.Header_2.Signature),
			},
		}
	}

	sourceAttesterSlashings := v1alpha1Block.Body.AttesterSlashings
	resultAttesterSlashings := make([]*zondpbv1.AttesterSlashing, len(sourceAttesterSlashings))
	for i, s := range sourceAttesterSlashings {
		att1Indices := make([]uint64, len(s.Attestation_1.AttestingIndices))
		copy(att1Indices, s.Attestation_1.AttestingIndices)
		att2Indices := make([]uint64, len(s.Attestation_2.AttestingIndices))
		copy(att2Indices, s.Attestation_2.AttestingIndices)
		resultAttesterSlashings[i] = &zondpbv1.AttesterSlashing{
			Attestation_1: &zondpbv1.IndexedAttestation{
				AttestingIndices: att1Indices,
				Data: &zondpbv1.AttestationData{
					Slot:            s.Attestation_1.Data.Slot,
					Index:           s.Attestation_1.Data.CommitteeIndex,
					BeaconBlockRoot: bytesutil.SafeCopyBytes(s.Attestation_1.Data.BeaconBlockRoot),
					Source: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_1.Data.Source.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_1.Data.Source.Root),
					},
					Target: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_1.Data.Target.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_1.Data.Target.Root),
					},
				},
				Signature: bytesutil.SafeCopyBytes(s.Attestation_1.Signature),
			},
			Attestation_2: &zondpbv1.IndexedAttestation{
				AttestingIndices: att2Indices,
				Data: &zondpbv1.AttestationData{
					Slot:            s.Attestation_2.Data.Slot,
					Index:           s.Attestation_2.Data.CommitteeIndex,
					BeaconBlockRoot: bytesutil.SafeCopyBytes(s.Attestation_2.Data.BeaconBlockRoot),
					Source: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_2.Data.Source.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_2.Data.Source.Root),
					},
					Target: &zondpbv1.Checkpoint{
						Epoch: s.Attestation_2.Data.Target.Epoch,
						Root:  bytesutil.SafeCopyBytes(s.Attestation_2.Data.Target.Root),
					},
				},
				Signature: bytesutil.SafeCopyBytes(s.Attestation_2.Signature),
			},
		}
	}

	sourceAttestations := v1alpha1Block.Body.Attestations
	resultAttestations := make([]*zondpbv1.Attestation, len(sourceAttestations))
	for i, a := range sourceAttestations {
		resultAttestations[i] = &zondpbv1.Attestation{
			AggregationBits: bytesutil.SafeCopyBytes(a.AggregationBits),
			Data: &zondpbv1.AttestationData{
				Slot:            a.Data.Slot,
				Index:           a.Data.CommitteeIndex,
				BeaconBlockRoot: bytesutil.SafeCopyBytes(a.Data.BeaconBlockRoot),
				Source: &zondpbv1.Checkpoint{
					Epoch: a.Data.Source.Epoch,
					Root:  bytesutil.SafeCopyBytes(a.Data.Source.Root),
				},
				Target: &zondpbv1.Checkpoint{
					Epoch: a.Data.Target.Epoch,
					Root:  bytesutil.SafeCopyBytes(a.Data.Target.Root),
				},
			},
			Signature: bytesutil.SafeCopyBytes(a.Signature),
		}
	}

	sourceDeposits := v1alpha1Block.Body.Deposits
	resultDeposits := make([]*zondpbv1.Deposit, len(sourceDeposits))
	for i, d := range sourceDeposits {
		resultDeposits[i] = &zondpbv1.Deposit{
			Proof: bytesutil.SafeCopy2dBytes(d.Proof),
			Data: &zondpbv1.Deposit_Data{
				Pubkey:                bytesutil.SafeCopyBytes(d.Data.PublicKey),
				WithdrawalCredentials: bytesutil.SafeCopyBytes(d.Data.WithdrawalCredentials),
				Amount:                d.Data.Amount,
				Signature:             bytesutil.SafeCopyBytes(d.Data.Signature),
			},
		}
	}

	sourceExits := v1alpha1Block.Body.VoluntaryExits
	resultExits := make([]*zondpbv1.SignedVoluntaryExit, len(sourceExits))
	for i, e := range sourceExits {
		resultExits[i] = &zondpbv1.SignedVoluntaryExit{
			Message: &zondpbv1.VoluntaryExit{
				Epoch:          e.Exit.Epoch,
				ValidatorIndex: e.Exit.ValidatorIndex,
			},
			Signature: bytesutil.SafeCopyBytes(e.Signature),
		}
	}

	transactionsRoot, err := ssz.TransactionsRoot(v1alpha1Block.Body.ExecutionPayload.Transactions)
	if err != nil {
		return nil, errors.Wrapf(err, "could not calculate transactions root")
	}

	withdrawalsRoot, err := ssz.WithdrawalSliceRoot(v1alpha1Block.Body.ExecutionPayload.Withdrawals, fieldparams.MaxWithdrawalsPerPayload)
	if err != nil {
		return nil, errors.Wrapf(err, "could not calculate transactions root")
	}

	changes := make([]*zondpbv2.SignedDilithiumToExecutionChange, len(v1alpha1Block.Body.DilithiumToExecutionChanges))
	for i, change := range v1alpha1Block.Body.DilithiumToExecutionChanges {
		changes[i] = &zondpbv2.SignedDilithiumToExecutionChange{
			Message: &zondpbv2.DilithiumToExecutionChange{
				ValidatorIndex:      change.Message.ValidatorIndex,
				FromDilithiumPubkey: bytesutil.SafeCopyBytes(change.Message.FromDilithiumPubkey),
				ToExecutionAddress:  bytesutil.SafeCopyBytes(change.Message.ToExecutionAddress),
			},
			Signature: bytesutil.SafeCopyBytes(change.Signature),
		}
	}

	resultBlockBody := &zondpbv2.BlindedBeaconBlockBodyCapella{
		RandaoReveal: bytesutil.SafeCopyBytes(v1alpha1Block.Body.RandaoReveal),
		ZondData: &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(v1alpha1Block.Body.ZondData.DepositRoot),
			DepositCount: v1alpha1Block.Body.ZondData.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(v1alpha1Block.Body.ZondData.BlockHash),
		},
		Graffiti:          bytesutil.SafeCopyBytes(v1alpha1Block.Body.Graffiti),
		ProposerSlashings: resultProposerSlashings,
		AttesterSlashings: resultAttesterSlashings,
		Attestations:      resultAttestations,
		Deposits:          resultDeposits,
		VoluntaryExits:    resultExits,
		SyncAggregate: &zondpbv1.SyncAggregate{
			SyncCommitteeBits:      bytesutil.SafeCopyBytes(v1alpha1Block.Body.SyncAggregate.SyncCommitteeBits),
			SyncCommitteeSignature: bytesutil.SafeCopyBytes(v1alpha1Block.Body.SyncAggregate.SyncCommitteeSignature),
		},
		ExecutionPayloadHeader: &enginev1.ExecutionPayloadHeaderCapella{
			ParentHash:       bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.ParentHash),
			FeeRecipient:     bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.FeeRecipient),
			StateRoot:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.StateRoot),
			ReceiptsRoot:     bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.ReceiptsRoot),
			LogsBloom:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.LogsBloom),
			PrevRandao:       bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.PrevRandao),
			BlockNumber:      v1alpha1Block.Body.ExecutionPayload.BlockNumber,
			GasLimit:         v1alpha1Block.Body.ExecutionPayload.GasLimit,
			GasUsed:          v1alpha1Block.Body.ExecutionPayload.GasUsed,
			Timestamp:        v1alpha1Block.Body.ExecutionPayload.Timestamp,
			ExtraData:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.ExtraData),
			BaseFeePerGas:    bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.BaseFeePerGas),
			BlockHash:        bytesutil.SafeCopyBytes(v1alpha1Block.Body.ExecutionPayload.BlockHash),
			TransactionsRoot: transactionsRoot[:],
			WithdrawalsRoot:  withdrawalsRoot[:],
		},
		DilithiumToExecutionChanges: changes,
	}
	v2Block := &zondpbv2.BlindedBeaconBlockCapella{
		Slot:          v1alpha1Block.Slot,
		ProposerIndex: v1alpha1Block.ProposerIndex,
		ParentRoot:    bytesutil.SafeCopyBytes(v1alpha1Block.ParentRoot),
		StateRoot:     bytesutil.SafeCopyBytes(v1alpha1Block.StateRoot),
		Body:          resultBlockBody,
	}
	return v2Block, nil
}

// BeaconStateAltairToProto converts a state.BeaconState object to its protobuf equivalent.
func BeaconStateAltairToProto(altairState state.BeaconState) (*zondpbv2.BeaconState, error) {
	sourceFork := altairState.Fork()
	sourceLatestBlockHeader := altairState.LatestBlockHeader()
	sourceZondData := altairState.ZondData()
	sourceZondDataVotes := altairState.ZondDataVotes()
	sourceValidators := altairState.Validators()
	sourceJustificationBits := altairState.JustificationBits()
	sourcePrevJustifiedCheckpoint := altairState.PreviousJustifiedCheckpoint()
	sourceCurrJustifiedCheckpoint := altairState.CurrentJustifiedCheckpoint()
	sourceFinalizedCheckpoint := altairState.FinalizedCheckpoint()

	resultZondDataVotes := make([]*zondpbv1.ZondData, len(sourceZondDataVotes))
	for i, vote := range sourceZondDataVotes {
		resultZondDataVotes[i] = &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(vote.DepositRoot),
			DepositCount: vote.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(vote.BlockHash),
		}
	}
	resultValidators := make([]*zondpbv1.Validator, len(sourceValidators))
	for i, validator := range sourceValidators {
		resultValidators[i] = &zondpbv1.Validator{
			Pubkey:                     bytesutil.SafeCopyBytes(validator.PublicKey),
			WithdrawalCredentials:      bytesutil.SafeCopyBytes(validator.WithdrawalCredentials),
			EffectiveBalance:           validator.EffectiveBalance,
			Slashed:                    validator.Slashed,
			ActivationEligibilityEpoch: validator.ActivationEligibilityEpoch,
			ActivationEpoch:            validator.ActivationEpoch,
			ExitEpoch:                  validator.ExitEpoch,
			WithdrawableEpoch:          validator.WithdrawableEpoch,
		}
	}

	sourcePrevEpochParticipation, err := altairState.PreviousEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get previous epoch participation")
	}
	sourceCurrEpochParticipation, err := altairState.CurrentEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current epoch participation")
	}
	sourceInactivityScores, err := altairState.InactivityScores()
	if err != nil {
		return nil, errors.Wrap(err, "could not get inactivity scores")
	}
	sourceCurrSyncCommittee, err := altairState.CurrentSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current sync committee")
	}
	sourceNextSyncCommittee, err := altairState.NextSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next sync committee")
	}

	hrs, err := altairState.HistoricalRoots()
	if err != nil {
		return nil, errors.Wrap(err, "could not get historical roots")
	}

	result := &zondpbv2.BeaconState{
		GenesisTime:           altairState.GenesisTime(),
		GenesisValidatorsRoot: bytesutil.SafeCopyBytes(altairState.GenesisValidatorsRoot()),
		Slot:                  altairState.Slot(),
		Fork: &zondpbv1.Fork{
			PreviousVersion: bytesutil.SafeCopyBytes(sourceFork.PreviousVersion),
			CurrentVersion:  bytesutil.SafeCopyBytes(sourceFork.CurrentVersion),
			Epoch:           sourceFork.Epoch,
		},
		LatestBlockHeader: &zondpbv1.BeaconBlockHeader{
			Slot:          sourceLatestBlockHeader.Slot,
			ProposerIndex: sourceLatestBlockHeader.ProposerIndex,
			ParentRoot:    bytesutil.SafeCopyBytes(sourceLatestBlockHeader.ParentRoot),
			StateRoot:     bytesutil.SafeCopyBytes(sourceLatestBlockHeader.StateRoot),
			BodyRoot:      bytesutil.SafeCopyBytes(sourceLatestBlockHeader.BodyRoot),
		},
		BlockRoots:      bytesutil.SafeCopy2dBytes(altairState.BlockRoots()),
		StateRoots:      bytesutil.SafeCopy2dBytes(altairState.StateRoots()),
		HistoricalRoots: bytesutil.SafeCopy2dBytes(hrs),
		ZondData: &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(sourceZondData.DepositRoot),
			DepositCount: sourceZondData.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(sourceZondData.BlockHash),
		},
		ZondDataVotes:              resultZondDataVotes,
		ZondDepositIndex:           altairState.ZondDepositIndex(),
		Validators:                 resultValidators,
		Balances:                   altairState.Balances(),
		RandaoMixes:                bytesutil.SafeCopy2dBytes(altairState.RandaoMixes()),
		Slashings:                  altairState.Slashings(),
		PreviousEpochParticipation: bytesutil.SafeCopyBytes(sourcePrevEpochParticipation),
		CurrentEpochParticipation:  bytesutil.SafeCopyBytes(sourceCurrEpochParticipation),
		JustificationBits:          bytesutil.SafeCopyBytes(sourceJustificationBits),
		PreviousJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourcePrevJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourcePrevJustifiedCheckpoint.Root),
		},
		CurrentJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceCurrJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceCurrJustifiedCheckpoint.Root),
		},
		FinalizedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceFinalizedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceFinalizedCheckpoint.Root),
		},
		InactivityScores: sourceInactivityScores,
		CurrentSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceCurrSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceCurrSyncCommittee.AggregatePubkey),
		},
		NextSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceNextSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceNextSyncCommittee.AggregatePubkey),
		},
	}

	return result, nil
}

// BeaconStateBellatrixToProto converts a state.BeaconState object to its protobuf equivalent.
func BeaconStateBellatrixToProto(st state.BeaconState) (*zondpbv2.BeaconStateBellatrix, error) {
	sourceFork := st.Fork()
	sourceLatestBlockHeader := st.LatestBlockHeader()
	sourceZondData := st.ZondData()
	sourceZondDataVotes := st.ZondDataVotes()
	sourceValidators := st.Validators()
	sourceJustificationBits := st.JustificationBits()
	sourcePrevJustifiedCheckpoint := st.PreviousJustifiedCheckpoint()
	sourceCurrJustifiedCheckpoint := st.CurrentJustifiedCheckpoint()
	sourceFinalizedCheckpoint := st.FinalizedCheckpoint()

	resultZondDataVotes := make([]*zondpbv1.ZondData, len(sourceZondDataVotes))
	for i, vote := range sourceZondDataVotes {
		resultZondDataVotes[i] = &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(vote.DepositRoot),
			DepositCount: vote.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(vote.BlockHash),
		}
	}
	resultValidators := make([]*zondpbv1.Validator, len(sourceValidators))
	for i, validator := range sourceValidators {
		resultValidators[i] = &zondpbv1.Validator{
			Pubkey:                     bytesutil.SafeCopyBytes(validator.PublicKey),
			WithdrawalCredentials:      bytesutil.SafeCopyBytes(validator.WithdrawalCredentials),
			EffectiveBalance:           validator.EffectiveBalance,
			Slashed:                    validator.Slashed,
			ActivationEligibilityEpoch: validator.ActivationEligibilityEpoch,
			ActivationEpoch:            validator.ActivationEpoch,
			ExitEpoch:                  validator.ExitEpoch,
			WithdrawableEpoch:          validator.WithdrawableEpoch,
		}
	}

	sourcePrevEpochParticipation, err := st.PreviousEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get previous epoch participation")
	}
	sourceCurrEpochParticipation, err := st.CurrentEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current epoch participation")
	}
	sourceInactivityScores, err := st.InactivityScores()
	if err != nil {
		return nil, errors.Wrap(err, "could not get inactivity scores")
	}
	sourceCurrSyncCommittee, err := st.CurrentSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current sync committee")
	}
	sourceNextSyncCommittee, err := st.NextSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next sync committee")
	}
	executionPayloadHeaderInterface, err := st.LatestExecutionPayloadHeader()
	if err != nil {
		return nil, errors.Wrap(err, "could not get latest execution payload header")
	}
	sourceLatestExecutionPayloadHeader, ok := executionPayloadHeaderInterface.Proto().(*enginev1.ExecutionPayloadHeader)
	if !ok {
		return nil, errors.New("execution payload header has incorrect type")
	}

	hRoots, err := st.HistoricalRoots()
	if err != nil {
		return nil, errors.Wrap(err, "could not get historical roots")
	}

	result := &zondpbv2.BeaconStateBellatrix{
		GenesisTime:           st.GenesisTime(),
		GenesisValidatorsRoot: bytesutil.SafeCopyBytes(st.GenesisValidatorsRoot()),
		Slot:                  st.Slot(),
		Fork: &zondpbv1.Fork{
			PreviousVersion: bytesutil.SafeCopyBytes(sourceFork.PreviousVersion),
			CurrentVersion:  bytesutil.SafeCopyBytes(sourceFork.CurrentVersion),
			Epoch:           sourceFork.Epoch,
		},
		LatestBlockHeader: &zondpbv1.BeaconBlockHeader{
			Slot:          sourceLatestBlockHeader.Slot,
			ProposerIndex: sourceLatestBlockHeader.ProposerIndex,
			ParentRoot:    bytesutil.SafeCopyBytes(sourceLatestBlockHeader.ParentRoot),
			StateRoot:     bytesutil.SafeCopyBytes(sourceLatestBlockHeader.StateRoot),
			BodyRoot:      bytesutil.SafeCopyBytes(sourceLatestBlockHeader.BodyRoot),
		},
		BlockRoots:      bytesutil.SafeCopy2dBytes(st.BlockRoots()),
		StateRoots:      bytesutil.SafeCopy2dBytes(st.StateRoots()),
		HistoricalRoots: bytesutil.SafeCopy2dBytes(hRoots),
		ZondData: &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(sourceZondData.DepositRoot),
			DepositCount: sourceZondData.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(sourceZondData.BlockHash),
		},
		ZondDataVotes:              resultZondDataVotes,
		ZondDepositIndex:           st.ZondDepositIndex(),
		Validators:                 resultValidators,
		Balances:                   st.Balances(),
		RandaoMixes:                bytesutil.SafeCopy2dBytes(st.RandaoMixes()),
		Slashings:                  st.Slashings(),
		PreviousEpochParticipation: bytesutil.SafeCopyBytes(sourcePrevEpochParticipation),
		CurrentEpochParticipation:  bytesutil.SafeCopyBytes(sourceCurrEpochParticipation),
		JustificationBits:          bytesutil.SafeCopyBytes(sourceJustificationBits),
		PreviousJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourcePrevJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourcePrevJustifiedCheckpoint.Root),
		},
		CurrentJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceCurrJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceCurrJustifiedCheckpoint.Root),
		},
		FinalizedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceFinalizedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceFinalizedCheckpoint.Root),
		},
		InactivityScores: sourceInactivityScores,
		CurrentSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceCurrSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceCurrSyncCommittee.AggregatePubkey),
		},
		NextSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceNextSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceNextSyncCommittee.AggregatePubkey),
		},
		LatestExecutionPayloadHeader: &enginev1.ExecutionPayloadHeader{
			ParentHash:       bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ParentHash),
			FeeRecipient:     bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.FeeRecipient),
			StateRoot:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.StateRoot),
			ReceiptsRoot:     bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ReceiptsRoot),
			LogsBloom:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.LogsBloom),
			PrevRandao:       bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.PrevRandao),
			BlockNumber:      sourceLatestExecutionPayloadHeader.BlockNumber,
			GasLimit:         sourceLatestExecutionPayloadHeader.GasLimit,
			GasUsed:          sourceLatestExecutionPayloadHeader.GasUsed,
			Timestamp:        sourceLatestExecutionPayloadHeader.Timestamp,
			ExtraData:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ExtraData),
			BaseFeePerGas:    bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.BaseFeePerGas),
			BlockHash:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.BlockHash),
			TransactionsRoot: bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.TransactionsRoot),
		},
	}

	return result, nil
}

// BeaconStateCapellaToProto converts a state.BeaconState object to its protobuf equivalent.
func BeaconStateCapellaToProto(st state.BeaconState) (*zondpbv2.BeaconStateCapella, error) {
	sourceFork := st.Fork()
	sourceLatestBlockHeader := st.LatestBlockHeader()
	sourceZondData := st.ZondData()
	sourceZondDataVotes := st.ZondDataVotes()
	sourceValidators := st.Validators()
	sourceJustificationBits := st.JustificationBits()
	sourcePrevJustifiedCheckpoint := st.PreviousJustifiedCheckpoint()
	sourceCurrJustifiedCheckpoint := st.CurrentJustifiedCheckpoint()
	sourceFinalizedCheckpoint := st.FinalizedCheckpoint()

	resultZondDataVotes := make([]*zondpbv1.ZondData, len(sourceZondDataVotes))
	for i, vote := range sourceZondDataVotes {
		resultZondDataVotes[i] = &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(vote.DepositRoot),
			DepositCount: vote.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(vote.BlockHash),
		}
	}
	resultValidators := make([]*zondpbv1.Validator, len(sourceValidators))
	for i, validator := range sourceValidators {
		resultValidators[i] = &zondpbv1.Validator{
			Pubkey:                     bytesutil.SafeCopyBytes(validator.PublicKey),
			WithdrawalCredentials:      bytesutil.SafeCopyBytes(validator.WithdrawalCredentials),
			EffectiveBalance:           validator.EffectiveBalance,
			Slashed:                    validator.Slashed,
			ActivationEligibilityEpoch: validator.ActivationEligibilityEpoch,
			ActivationEpoch:            validator.ActivationEpoch,
			ExitEpoch:                  validator.ExitEpoch,
			WithdrawableEpoch:          validator.WithdrawableEpoch,
		}
	}

	sourcePrevEpochParticipation, err := st.PreviousEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get previous epoch participation")
	}
	sourceCurrEpochParticipation, err := st.CurrentEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current epoch participation")
	}
	sourceInactivityScores, err := st.InactivityScores()
	if err != nil {
		return nil, errors.Wrap(err, "could not get inactivity scores")
	}
	sourceCurrSyncCommittee, err := st.CurrentSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current sync committee")
	}
	sourceNextSyncCommittee, err := st.NextSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next sync committee")
	}
	executionPayloadHeaderInterface, err := st.LatestExecutionPayloadHeader()
	if err != nil {
		return nil, errors.Wrap(err, "could not get latest execution payload header")
	}
	sourceLatestExecutionPayloadHeader, ok := executionPayloadHeaderInterface.Proto().(*enginev1.ExecutionPayloadHeaderCapella)
	if !ok {
		return nil, errors.New("execution payload header has incorrect type")
	}
	sourceNextWithdrawalIndex, err := st.NextWithdrawalIndex()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next withdrawal index")
	}
	sourceNextWithdrawalValIndex, err := st.NextWithdrawalValidatorIndex()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next withdrawal validator index")
	}
	summaries, err := st.HistoricalSummaries()
	if err != nil {
		return nil, errors.Wrap(err, "could not get historical summaries")
	}
	sourceHistoricalSummaries := make([]*zondpbv2.HistoricalSummary, len(summaries))
	for i, summary := range summaries {
		sourceHistoricalSummaries[i] = &zondpbv2.HistoricalSummary{
			BlockSummaryRoot: summary.BlockSummaryRoot,
			StateSummaryRoot: summary.StateSummaryRoot,
		}
	}
	hRoots, err := st.HistoricalRoots()
	if err != nil {
		return nil, errors.Wrap(err, "could not get historical roots")
	}

	result := &zondpbv2.BeaconStateCapella{
		GenesisTime:           st.GenesisTime(),
		GenesisValidatorsRoot: bytesutil.SafeCopyBytes(st.GenesisValidatorsRoot()),
		Slot:                  st.Slot(),
		Fork: &zondpbv1.Fork{
			PreviousVersion: bytesutil.SafeCopyBytes(sourceFork.PreviousVersion),
			CurrentVersion:  bytesutil.SafeCopyBytes(sourceFork.CurrentVersion),
			Epoch:           sourceFork.Epoch,
		},
		LatestBlockHeader: &zondpbv1.BeaconBlockHeader{
			Slot:          sourceLatestBlockHeader.Slot,
			ProposerIndex: sourceLatestBlockHeader.ProposerIndex,
			ParentRoot:    bytesutil.SafeCopyBytes(sourceLatestBlockHeader.ParentRoot),
			StateRoot:     bytesutil.SafeCopyBytes(sourceLatestBlockHeader.StateRoot),
			BodyRoot:      bytesutil.SafeCopyBytes(sourceLatestBlockHeader.BodyRoot),
		},
		BlockRoots: bytesutil.SafeCopy2dBytes(st.BlockRoots()),
		StateRoots: bytesutil.SafeCopy2dBytes(st.StateRoots()),
		ZondData: &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(sourceZondData.DepositRoot),
			DepositCount: sourceZondData.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(sourceZondData.BlockHash),
		},
		ZondDataVotes:              resultZondDataVotes,
		ZondDepositIndex:           st.ZondDepositIndex(),
		Validators:                 resultValidators,
		Balances:                   st.Balances(),
		RandaoMixes:                bytesutil.SafeCopy2dBytes(st.RandaoMixes()),
		Slashings:                  st.Slashings(),
		PreviousEpochParticipation: bytesutil.SafeCopyBytes(sourcePrevEpochParticipation),
		CurrentEpochParticipation:  bytesutil.SafeCopyBytes(sourceCurrEpochParticipation),
		JustificationBits:          bytesutil.SafeCopyBytes(sourceJustificationBits),
		PreviousJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourcePrevJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourcePrevJustifiedCheckpoint.Root),
		},
		CurrentJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceCurrJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceCurrJustifiedCheckpoint.Root),
		},
		FinalizedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceFinalizedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceFinalizedCheckpoint.Root),
		},
		InactivityScores: sourceInactivityScores,
		CurrentSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceCurrSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceCurrSyncCommittee.AggregatePubkey),
		},
		NextSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceNextSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceNextSyncCommittee.AggregatePubkey),
		},
		LatestExecutionPayloadHeader: &enginev1.ExecutionPayloadHeaderCapella{
			ParentHash:       bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ParentHash),
			FeeRecipient:     bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.FeeRecipient),
			StateRoot:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.StateRoot),
			ReceiptsRoot:     bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ReceiptsRoot),
			LogsBloom:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.LogsBloom),
			PrevRandao:       bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.PrevRandao),
			BlockNumber:      sourceLatestExecutionPayloadHeader.BlockNumber,
			GasLimit:         sourceLatestExecutionPayloadHeader.GasLimit,
			GasUsed:          sourceLatestExecutionPayloadHeader.GasUsed,
			Timestamp:        sourceLatestExecutionPayloadHeader.Timestamp,
			ExtraData:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ExtraData),
			BaseFeePerGas:    bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.BaseFeePerGas),
			BlockHash:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.BlockHash),
			TransactionsRoot: bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.TransactionsRoot),
			WithdrawalsRoot:  bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.WithdrawalsRoot),
		},
		NextWithdrawalIndex:          sourceNextWithdrawalIndex,
		NextWithdrawalValidatorIndex: sourceNextWithdrawalValIndex,
		HistoricalSummaries:          sourceHistoricalSummaries,
		HistoricalRoots:              hRoots,
	}

	return result, nil
}

// BeaconStateDenebToProto converts a state.BeaconState object to its protobuf equivalent.
func BeaconStateDenebToProto(st state.BeaconState) (*zondpbv2.BeaconStateDeneb, error) {
	sourceFork := st.Fork()
	sourceLatestBlockHeader := st.LatestBlockHeader()
	sourceZondData := st.ZondData()
	sourceZondDataVotes := st.ZondDataVotes()
	sourceValidators := st.Validators()
	sourceJustificationBits := st.JustificationBits()
	sourcePrevJustifiedCheckpoint := st.PreviousJustifiedCheckpoint()
	sourceCurrJustifiedCheckpoint := st.CurrentJustifiedCheckpoint()
	sourceFinalizedCheckpoint := st.FinalizedCheckpoint()

	resultZondDataVotes := make([]*zondpbv1.ZondData, len(sourceZondDataVotes))
	for i, vote := range sourceZondDataVotes {
		resultZondDataVotes[i] = &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(vote.DepositRoot),
			DepositCount: vote.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(vote.BlockHash),
		}
	}
	resultValidators := make([]*zondpbv1.Validator, len(sourceValidators))
	for i, validator := range sourceValidators {
		resultValidators[i] = &zondpbv1.Validator{
			Pubkey:                     bytesutil.SafeCopyBytes(validator.PublicKey),
			WithdrawalCredentials:      bytesutil.SafeCopyBytes(validator.WithdrawalCredentials),
			EffectiveBalance:           validator.EffectiveBalance,
			Slashed:                    validator.Slashed,
			ActivationEligibilityEpoch: validator.ActivationEligibilityEpoch,
			ActivationEpoch:            validator.ActivationEpoch,
			ExitEpoch:                  validator.ExitEpoch,
			WithdrawableEpoch:          validator.WithdrawableEpoch,
		}
	}

	sourcePrevEpochParticipation, err := st.PreviousEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get previous epoch participation")
	}
	sourceCurrEpochParticipation, err := st.CurrentEpochParticipation()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current epoch participation")
	}
	sourceInactivityScores, err := st.InactivityScores()
	if err != nil {
		return nil, errors.Wrap(err, "could not get inactivity scores")
	}
	sourceCurrSyncCommittee, err := st.CurrentSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get current sync committee")
	}
	sourceNextSyncCommittee, err := st.NextSyncCommittee()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next sync committee")
	}
	executionPayloadHeaderInterface, err := st.LatestExecutionPayloadHeader()
	if err != nil {
		return nil, errors.Wrap(err, "could not get latest execution payload header")
	}
	sourceLatestExecutionPayloadHeader, ok := executionPayloadHeaderInterface.Proto().(*enginev1.ExecutionPayloadHeaderDeneb)
	if !ok {
		return nil, errors.New("execution payload header has incorrect type")
	}
	sourceNextWithdrawalIndex, err := st.NextWithdrawalIndex()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next withdrawal index")
	}
	sourceNextWithdrawalValIndex, err := st.NextWithdrawalValidatorIndex()
	if err != nil {
		return nil, errors.Wrap(err, "could not get next withdrawal validator index")
	}
	summaries, err := st.HistoricalSummaries()
	if err != nil {
		return nil, errors.Wrap(err, "could not get historical summaries")
	}
	sourceHistoricalSummaries := make([]*zondpbv2.HistoricalSummary, len(summaries))
	for i, summary := range summaries {
		sourceHistoricalSummaries[i] = &zondpbv2.HistoricalSummary{
			BlockSummaryRoot: summary.BlockSummaryRoot,
			StateSummaryRoot: summary.StateSummaryRoot,
		}
	}

	hr, err := st.HistoricalRoots()
	if err != nil {
		return nil, errors.Wrap(err, "could not get historical roots")
	}

	result := &zondpbv2.BeaconStateDeneb{
		GenesisTime:           st.GenesisTime(),
		GenesisValidatorsRoot: bytesutil.SafeCopyBytes(st.GenesisValidatorsRoot()),
		Slot:                  st.Slot(),
		Fork: &zondpbv1.Fork{
			PreviousVersion: bytesutil.SafeCopyBytes(sourceFork.PreviousVersion),
			CurrentVersion:  bytesutil.SafeCopyBytes(sourceFork.CurrentVersion),
			Epoch:           sourceFork.Epoch,
		},
		LatestBlockHeader: &zondpbv1.BeaconBlockHeader{
			Slot:          sourceLatestBlockHeader.Slot,
			ProposerIndex: sourceLatestBlockHeader.ProposerIndex,
			ParentRoot:    bytesutil.SafeCopyBytes(sourceLatestBlockHeader.ParentRoot),
			StateRoot:     bytesutil.SafeCopyBytes(sourceLatestBlockHeader.StateRoot),
			BodyRoot:      bytesutil.SafeCopyBytes(sourceLatestBlockHeader.BodyRoot),
		},
		BlockRoots:      bytesutil.SafeCopy2dBytes(st.BlockRoots()),
		StateRoots:      bytesutil.SafeCopy2dBytes(st.StateRoots()),
		HistoricalRoots: bytesutil.SafeCopy2dBytes(hr),
		ZondData: &zondpbv1.ZondData{
			DepositRoot:  bytesutil.SafeCopyBytes(sourceZondData.DepositRoot),
			DepositCount: sourceZondData.DepositCount,
			BlockHash:    bytesutil.SafeCopyBytes(sourceZondData.BlockHash),
		},
		ZondDataVotes:              resultZondDataVotes,
		ZondDepositIndex:           st.ZondDepositIndex(),
		Validators:                 resultValidators,
		Balances:                   st.Balances(),
		RandaoMixes:                bytesutil.SafeCopy2dBytes(st.RandaoMixes()),
		Slashings:                  st.Slashings(),
		PreviousEpochParticipation: bytesutil.SafeCopyBytes(sourcePrevEpochParticipation),
		CurrentEpochParticipation:  bytesutil.SafeCopyBytes(sourceCurrEpochParticipation),
		JustificationBits:          bytesutil.SafeCopyBytes(sourceJustificationBits),
		PreviousJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourcePrevJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourcePrevJustifiedCheckpoint.Root),
		},
		CurrentJustifiedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceCurrJustifiedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceCurrJustifiedCheckpoint.Root),
		},
		FinalizedCheckpoint: &zondpbv1.Checkpoint{
			Epoch: sourceFinalizedCheckpoint.Epoch,
			Root:  bytesutil.SafeCopyBytes(sourceFinalizedCheckpoint.Root),
		},
		InactivityScores: sourceInactivityScores,
		CurrentSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceCurrSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceCurrSyncCommittee.AggregatePubkey),
		},
		NextSyncCommittee: &zondpbv2.SyncCommittee{
			Pubkeys:         bytesutil.SafeCopy2dBytes(sourceNextSyncCommittee.Pubkeys),
			AggregatePubkey: bytesutil.SafeCopyBytes(sourceNextSyncCommittee.AggregatePubkey),
		},
		LatestExecutionPayloadHeader: &enginev1.ExecutionPayloadHeaderDeneb{
			ParentHash:       bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ParentHash),
			FeeRecipient:     bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.FeeRecipient),
			StateRoot:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.StateRoot),
			ReceiptsRoot:     bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ReceiptsRoot),
			LogsBloom:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.LogsBloom),
			PrevRandao:       bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.PrevRandao),
			BlockNumber:      sourceLatestExecutionPayloadHeader.BlockNumber,
			GasLimit:         sourceLatestExecutionPayloadHeader.GasLimit,
			GasUsed:          sourceLatestExecutionPayloadHeader.GasUsed,
			Timestamp:        sourceLatestExecutionPayloadHeader.Timestamp,
			ExtraData:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.ExtraData),
			BaseFeePerGas:    bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.BaseFeePerGas),
			BlobGasUsed:      sourceLatestExecutionPayloadHeader.BlobGasUsed,
			ExcessBlobGas:    sourceLatestExecutionPayloadHeader.ExcessBlobGas,
			BlockHash:        bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.BlockHash),
			TransactionsRoot: bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.TransactionsRoot),
			WithdrawalsRoot:  bytesutil.SafeCopyBytes(sourceLatestExecutionPayloadHeader.WithdrawalsRoot),
		},
		NextWithdrawalIndex:          sourceNextWithdrawalIndex,
		NextWithdrawalValidatorIndex: sourceNextWithdrawalValIndex,
		HistoricalSummaries:          sourceHistoricalSummaries,
	}

	return result, nil
}

// V1Alpha1SignedContributionAndProofToV2 converts a v1alpha1 SignedContributionAndProof object to its v2 equivalent.
func V1Alpha1SignedContributionAndProofToV2(alphaContribution *zondpbalpha.SignedContributionAndProof) *zondpbv2.SignedContributionAndProof {
	result := &zondpbv2.SignedContributionAndProof{
		Message: &zondpbv2.ContributionAndProof{
			AggregatorIndex: alphaContribution.Message.AggregatorIndex,
			Contribution: &zondpbv2.SyncCommitteeContribution{
				Slot:              alphaContribution.Message.Contribution.Slot,
				BeaconBlockRoot:   alphaContribution.Message.Contribution.BlockRoot,
				SubcommitteeIndex: alphaContribution.Message.Contribution.SubcommitteeIndex,
				AggregationBits:   alphaContribution.Message.Contribution.AggregationBits,
				Signature:         alphaContribution.Message.Contribution.Signature,
			},
			SelectionProof: alphaContribution.Message.SelectionProof,
		},
		Signature: alphaContribution.Signature,
	}
	return result
}

// V2SignedDilithiumToExecutionChangeToV1Alpha1 converts a V2 SignedDilithiumToExecutionChange to its v1alpha1 equivalent.
func V2SignedDilithiumToExecutionChangeToV1Alpha1(change *zondpbv2.SignedDilithiumToExecutionChange) *zondpbalpha.SignedDilithiumToExecutionChange {
	return &zondpbalpha.SignedDilithiumToExecutionChange{
		Message: &zondpbalpha.DilithiumToExecutionChange{
			ValidatorIndex:      change.Message.ValidatorIndex,
			FromDilithiumPubkey: bytesutil.SafeCopyBytes(change.Message.FromDilithiumPubkey),
			ToExecutionAddress:  bytesutil.SafeCopyBytes(change.Message.ToExecutionAddress),
		},
		Signature: bytesutil.SafeCopyBytes(change.Signature),
	}
}

// V1Alpha1SignedDilithiumToExecChangeToV2 converts a v1alpha1 SignedDilithiumToExecutionChange object to its v2 equivalent.
func V1Alpha1SignedDilithiumToExecChangeToV2(alphaChange *zondpbalpha.SignedDilithiumToExecutionChange) *zondpbv2.SignedDilithiumToExecutionChange {
	result := &zondpbv2.SignedDilithiumToExecutionChange{
		Message: &zondpbv2.DilithiumToExecutionChange{
			ValidatorIndex:      alphaChange.Message.ValidatorIndex,
			FromDilithiumPubkey: bytesutil.SafeCopyBytes(alphaChange.Message.FromDilithiumPubkey),
			ToExecutionAddress:  bytesutil.SafeCopyBytes(alphaChange.Message.ToExecutionAddress),
		},
		Signature: bytesutil.SafeCopyBytes(alphaChange.Signature),
	}
	return result
}
