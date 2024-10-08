package beacon

import (
	"context"

	"github.com/theQRL/qrysm/config/features"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/container/slice"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// SubmitProposerSlashing receives a proposer slashing object via
// RPC and injects it into the beacon node's operations pool.
// Submission into this pool does not guarantee inclusion into a beacon block.
func (bs *Server) SubmitProposerSlashing(
	ctx context.Context,
	req *zondpb.ProposerSlashing,
) (*zondpb.SubmitSlashingResponse, error) {
	beaconState, err := bs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	if err := bs.SlashingsPool.InsertProposerSlashing(ctx, beaconState, req); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not insert proposer slashing into pool: %v", err)
	}
	if !features.Get().DisableBroadcastSlashings {
		if err := bs.Broadcaster.Broadcast(ctx, req); err != nil {
			return nil, status.Errorf(codes.Internal, "Could not broadcast slashing object: %v", err)
		}
	}

	return &zondpb.SubmitSlashingResponse{
		SlashedIndices: []primitives.ValidatorIndex{req.Header_1.Header.ProposerIndex},
	}, nil
}

// SubmitAttesterSlashing receives an attester slashing object via
// RPC and injects it into the beacon node's operations pool.
// Submission into this pool does not guarantee inclusion into a beacon block.
func (bs *Server) SubmitAttesterSlashing(
	ctx context.Context,
	req *zondpb.AttesterSlashing,
) (*zondpb.SubmitSlashingResponse, error) {
	beaconState, err := bs.HeadFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not retrieve head state: %v", err)
	}
	if err := bs.SlashingsPool.InsertAttesterSlashing(ctx, beaconState, req); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not insert attester slashing into pool: %v", err)
	}
	if !features.Get().DisableBroadcastSlashings {
		if err := bs.Broadcaster.Broadcast(ctx, req); err != nil {
			return nil, status.Errorf(codes.Internal, "Could not broadcast slashing object: %v", err)
		}
	}
	indices := slice.IntersectionUint64(req.Attestation_1.AttestingIndices, req.Attestation_2.AttestingIndices)
	slashedIndices := make([]primitives.ValidatorIndex, len(indices))
	for i, index := range indices {
		slashedIndices[i] = primitives.ValidatorIndex(index)
	}
	return &zondpb.SubmitSlashingResponse{
		SlashedIndices: slashedIndices,
	}, nil
}
