package beacon

import (
	"context"
	"sort"
	"strconv"
	"strings"

	"github.com/theQRL/qrysm/api/pagination"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/beacon-chain/db/filters"
	"github.com/theQRL/qrysm/cmd"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/proto/qrysm/v1alpha1/attestation"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// sortableAttestations implements the Sort interface to sort attestations
// by slot as the canonical sorting attribute.
type sortableAttestations []*zondpb.Attestation

// Len is the number of elements in the collection.
func (s sortableAttestations) Len() int { return len(s) }

// Swap swaps the elements with indexes i and j.
func (s sortableAttestations) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// Less reports whether the element with index i must sort before the element with index j.
func (s sortableAttestations) Less(i, j int) bool {
	return s[i].Data.Slot < s[j].Data.Slot
}

func mapAttestationsByTargetRoot(atts []*zondpb.Attestation) map[[32]byte][]*zondpb.Attestation {
	attsMap := make(map[[32]byte][]*zondpb.Attestation, len(atts))
	if len(atts) == 0 {
		return attsMap
	}
	for _, att := range atts {
		attsMap[bytesutil.ToBytes32(att.Data.Target.Root)] = append(attsMap[bytesutil.ToBytes32(att.Data.Target.Root)], att)
	}
	return attsMap
}

// ListAttestations retrieves attestations by block root, slot, or epoch.
// Attestations are sorted by data slot by default.
//
// The server may return an empty list when no attestations match the given
// filter criteria. This RPC should not return NOT_FOUND. Only one filter
// criteria should be used.
func (bs *Server) ListAttestations(
	ctx context.Context, req *zondpb.ListAttestationsRequest,
) (*zondpb.ListAttestationsResponse, error) {
	if int(req.PageSize) > cmd.Get().MaxRPCPageSize {
		return nil, status.Errorf(codes.InvalidArgument, "Requested page size %d can not be greater than max size %d",
			req.PageSize, cmd.Get().MaxRPCPageSize)
	}
	var blocks []interfaces.ReadOnlySignedBeaconBlock
	var err error
	switch q := req.QueryFilter.(type) {
	case *zondpb.ListAttestationsRequest_GenesisEpoch:
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(0).SetEndEpoch(0))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	case *zondpb.ListAttestationsRequest_Epoch:
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(q.Epoch).SetEndEpoch(q.Epoch))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "Must specify a filter criteria for fetching attestations")
	}
	atts := make([]*zondpb.Attestation, 0, params.BeaconConfig().MaxAttestations*uint64(len(blocks)))
	for _, blk := range blocks {
		atts = append(atts, blk.Block().Body().Attestations()...)
	}
	// We sort attestations according to the Sortable interface.
	sort.Sort(sortableAttestations(atts))
	numAttestations := len(atts)

	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if numAttestations == 0 {
		return &zondpb.ListAttestationsResponse{
			Attestations:  make([]*zondpb.Attestation, 0),
			TotalSize:     int32(0),
			NextPageToken: strconv.Itoa(0),
		}, nil
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), numAttestations)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}
	return &zondpb.ListAttestationsResponse{
		Attestations:  atts[start:end],
		TotalSize:     int32(numAttestations),
		NextPageToken: nextPageToken,
	}, nil
}

// ListIndexedAttestations retrieves indexed attestations by block root.
// IndexedAttestationsForEpoch are sorted by data slot by default. Start-end epoch
// filter is used to retrieve blocks with.
//
// The server may return an empty list when no attestations match the given
// filter criteria. This RPC should not return NOT_FOUND.
func (bs *Server) ListIndexedAttestations(
	ctx context.Context, req *zondpb.ListIndexedAttestationsRequest,
) (*zondpb.ListIndexedAttestationsResponse, error) {
	var blocks []interfaces.ReadOnlySignedBeaconBlock
	var err error
	switch q := req.QueryFilter.(type) {
	case *zondpb.ListIndexedAttestationsRequest_GenesisEpoch:
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(0).SetEndEpoch(0))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	case *zondpb.ListIndexedAttestationsRequest_Epoch:
		blocks, _, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetStartEpoch(q.Epoch).SetEndEpoch(q.Epoch))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not fetch attestations: %v", err)
		}
	default:
		return nil, status.Error(codes.InvalidArgument, "Must specify a filter criteria for fetching attestations")
	}

	attsArray := make([]*zondpb.Attestation, 0, params.BeaconConfig().MaxAttestations*uint64(len(blocks)))
	for _, b := range blocks {
		attsArray = append(attsArray, b.Block().Body().Attestations()...)
	}
	// We sort attestations according to the Sortable interface.
	sort.Sort(sortableAttestations(attsArray))
	numAttestations := len(attsArray)

	// If there are no attestations, we simply return a response specifying this.
	// Otherwise, attempting to paginate 0 attestations below would result in an error.
	if numAttestations == 0 {
		return &zondpb.ListIndexedAttestationsResponse{
			IndexedAttestations: make([]*zondpb.IndexedAttestation, 0),
			TotalSize:           int32(0),
			NextPageToken:       strconv.Itoa(0),
		}, nil
	}
	// We use the retrieved committees for the b root to convert all attestations
	// into indexed form effectively.
	mappedAttestations := mapAttestationsByTargetRoot(attsArray)
	indexedAtts := make([]*zondpb.IndexedAttestation, 0, numAttestations)
	for targetRoot, atts := range mappedAttestations {
		attState, err := bs.StateGen.StateByRoot(ctx, targetRoot)
		if err != nil && strings.Contains(err.Error(), "unknown state summary") {
			// We shouldn't stop the request if we encounter an attestation we don't have the state for.
			log.Debugf("Could not get state for attestation target root %#x", targetRoot)
			continue
		} else if err != nil {
			return nil, status.Errorf(
				codes.Internal,
				"Could not retrieve state for attestation target root %#x: %v",
				targetRoot,
				err,
			)
		}
		for i := 0; i < len(atts); i++ {
			att := atts[i]
			committee, err := helpers.BeaconCommitteeFromState(ctx, attState, att.Data.Slot, att.Data.CommitteeIndex)
			if err != nil {
				return nil, status.Errorf(
					codes.Internal,
					"Could not retrieve committee from state %v",
					err,
				)
			}
			idxAtt, err := attestation.ConvertToIndexed(ctx, att, committee)
			if err != nil {
				return nil, err
			}
			indexedAtts = append(indexedAtts, idxAtt)
		}
	}

	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), len(indexedAtts))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}
	return &zondpb.ListIndexedAttestationsResponse{
		IndexedAttestations: indexedAtts[start:end],
		TotalSize:           int32(len(indexedAtts)),
		NextPageToken:       nextPageToken,
	}, nil
}

// AttestationPool retrieves pending attestations.
//
// The server returns a list of attestations that have been seen but not
// yet processed. Pool attestations eventually expire as the slot
// advances, so an attestation missing from this request does not imply
// that it was included in a block. The attestation may have expired.
// Refer to the ethereum consensus specification for more details on how
// attestations are processed and when they are no longer valid.
// https://github.com/ethereum/consensus-specs/blob/dev/specs/core/0_beacon-chain.md#attestations
func (bs *Server) AttestationPool(
	_ context.Context, req *zondpb.AttestationPoolRequest,
) (*zondpb.AttestationPoolResponse, error) {
	if int(req.PageSize) > cmd.Get().MaxRPCPageSize {
		return nil, status.Errorf(
			codes.InvalidArgument,
			"Requested page size %d can not be greater than max size %d",
			req.PageSize,
			cmd.Get().MaxRPCPageSize,
		)
	}
	atts := bs.AttestationsPool.AggregatedAttestations()
	numAtts := len(atts)
	if numAtts == 0 {
		return &zondpb.AttestationPoolResponse{
			Attestations:  make([]*zondpb.Attestation, 0),
			TotalSize:     int32(0),
			NextPageToken: strconv.Itoa(0),
		}, nil
	}
	start, end, nextPageToken, err := pagination.StartAndEndPage(req.PageToken, int(req.PageSize), numAtts)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not paginate attestations: %v", err)
	}
	return &zondpb.AttestationPoolResponse{
		Attestations:  atts[start:end],
		TotalSize:     int32(numAtts),
		NextPageToken: nextPageToken,
	}, nil
}
