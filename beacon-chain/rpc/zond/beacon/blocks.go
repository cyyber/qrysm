package beacon

import (
	"context"
	"strconv"
	"strings"

	"google.golang.org/grpc"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/v4/api"
	"github.com/theQRL/qrysm/v4/beacon-chain/db/filters"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/lookup"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/qrysm/v1alpha1/validator"
	rpchelpers "github.com/theQRL/qrysm/v4/beacon-chain/rpc/zond/helpers"
	"github.com/theQRL/qrysm/v4/config/params"
	consensus_types "github.com/theQRL/qrysm/v4/consensus-types"
	"github.com/theQRL/qrysm/v4/consensus-types/blocks"
	"github.com/theQRL/qrysm/v4/consensus-types/interfaces"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	"github.com/theQRL/qrysm/v4/encoding/ssz/detect"
	"github.com/theQRL/qrysm/v4/network/forks"
	"github.com/theQRL/qrysm/v4/proto/migration"
	zond "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	zondpbv1 "github.com/theQRL/qrysm/v4/proto/zond/v1"
	"github.com/theQRL/qrysm/v4/runtime/version"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

var (
	errNilBlock = errors.New("nil block")
)

// GetBlockHeader retrieves block header for given block id.
func (bs *Server) GetBlockHeader(ctx context.Context, req *zondpbv1.BlockRequest) (*zondpbv1.BlockHeaderResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.GetBlockHeader")
	defer span.End()

	blk, err := bs.Blocker.Block(ctx, req.BlockId)
	err = handleGetBlockError(blk, err)
	if err != nil {
		return nil, err
	}
	v1alpha1Header, err := blk.Header()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get block header from block: %v", err)
	}
	header := migration.V1Alpha1ToV1SignedHeader(v1alpha1Header)
	headerRoot, err := header.Message.HashTreeRoot()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not hash block header: %v", err)
	}
	blkRoot, err := blk.Block().HashTreeRoot()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not hash block: %v", err)
	}
	canonical, err := bs.ChainInfoFetcher.IsCanonical(ctx, blkRoot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not determine if block root is canonical: %v", err)
	}
	isOptimistic, err := bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, blkRoot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not check if block is optimistic: %v", err)
	}

	return &zondpbv1.BlockHeaderResponse{
		Data: &zondpbv1.BlockHeaderContainer{
			Root:      headerRoot[:],
			Canonical: canonical,
			Header: &zondpbv1.BeaconBlockHeaderContainer{
				Message:   header.Message,
				Signature: header.Signature,
			},
		},
		ExecutionOptimistic: isOptimistic,
		Finalized:           bs.FinalizationFetcher.IsFinalized(ctx, blkRoot),
	}, nil
}

// ListBlockHeaders retrieves block headers matching given query. By default it will fetch current head slot blocks.
func (bs *Server) ListBlockHeaders(ctx context.Context, req *zondpbv1.BlockHeadersRequest) (*zondpbv1.BlockHeadersResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListBlockHeaders")
	defer span.End()

	var err error
	var blks []interfaces.ReadOnlySignedBeaconBlock
	var blkRoots [][32]byte
	if len(req.ParentRoot) == 32 {
		blks, blkRoots, err = bs.BeaconDB.Blocks(ctx, filters.NewFilter().SetParentRoot(req.ParentRoot))
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not retrieve blocks: %v", err)
		}
	} else {
		slot := bs.ChainInfoFetcher.HeadSlot()
		if req.Slot != nil {
			slot = *req.Slot
		}
		blks, err = bs.BeaconDB.BlocksBySlot(ctx, slot)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not retrieve blocks for slot %d: %v", req.Slot, err)
		}
		_, blkRoots, err = bs.BeaconDB.BlockRootsBySlot(ctx, slot)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not retrieve block roots for slot %d: %v", req.Slot, err)
		}
	}
	if len(blks) == 0 {
		return nil, status.Error(codes.NotFound, "Could not find requested blocks")
	}

	isOptimistic := false
	isFinalized := true
	blkHdrs := make([]*zondpbv1.BlockHeaderContainer, len(blks))
	for i, bl := range blks {
		v1alpha1Header, err := bl.Header()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not get block header from block: %v", err)
		}
		header := migration.V1Alpha1ToV1SignedHeader(v1alpha1Header)
		headerRoot, err := header.Message.HashTreeRoot()
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not hash block header: %v", err)
		}
		canonical, err := bs.ChainInfoFetcher.IsCanonical(ctx, blkRoots[i])
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not determine if block root is canonical: %v", err)
		}
		if !isOptimistic {
			isOptimistic, err = bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, blkRoots[i])
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Could not check if block is optimistic: %v", err)
			}
		}
		if isFinalized {
			isFinalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoots[i])
		}
		blkHdrs[i] = &zondpbv1.BlockHeaderContainer{
			Root:      headerRoot[:],
			Canonical: canonical,
			Header: &zondpbv1.BeaconBlockHeaderContainer{
				Message:   header.Message,
				Signature: header.Signature,
			},
		}
	}

	return &zondpbv1.BlockHeadersResponse{Data: blkHdrs, ExecutionOptimistic: isOptimistic, Finalized: isFinalized}, nil
}

// SubmitBlock instructs the beacon node to broadcast a newly signed beacon block to the beacon network, to be
// included in the beacon chain. The beacon node is not required to validate the signed ReadOnlyBeaconBlock, and a successful
// response (20X) only indicates that the broadcast has been successful. The beacon node is expected to integrate the
// new block into its state, and therefore validate the block internally, however blocks which fail the validation are
// still broadcast but a different status code is returned (202).
func (bs *Server) SubmitBlock(ctx context.Context, req *zondpbv1.SignedBeaconBlockContainer) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.SubmitBlock")
	defer span.End()

	if err := rpchelpers.ValidateSyncGRPC(ctx, bs.SyncChecker, bs.HeadFetcher, bs.TimeFetcher, bs.OptimisticModeFetcher); err != nil {
		// We simply return the error because it's already a gRPC error.
		return nil, err
	}

	switch blkContainer := req.Message.(type) {
	case *zondpbv1.SignedBeaconBlockContainer_CapellaBlock:
		if err := bs.submitCapellaBlock(ctx, blkContainer.CapellaBlock, req.Signature); err != nil {
			return nil, err
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "Unsupported block container type %T", blkContainer)
	}

	return &emptypb.Empty{}, nil
}

// SubmitBlockSSZ instructs the beacon node to broadcast a newly signed beacon block to the beacon network, to be
// included in the beacon chain. The beacon node is not required to validate the signed ReadOnlyBeaconBlock, and a successful
// response (20X) only indicates that the broadcast has been successful. The beacon node is expected to integrate the
// new block into its state, and therefore validate the block internally, however blocks which fail the validation are
// still broadcast but a different status code is returned (202).
//
// The provided block must be SSZ-serialized.
func (bs *Server) SubmitBlockSSZ(ctx context.Context, req *zondpbv1.SSZContainer) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.SubmitBlockSSZ")
	defer span.End()

	if err := rpchelpers.ValidateSyncGRPC(ctx, bs.SyncChecker, bs.HeadFetcher, bs.TimeFetcher, bs.OptimisticModeFetcher); err != nil {
		// We simply return the error because it's already a gRPC error.
		return nil, err
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not read "+api.VersionHeader+" header")
	}
	ver := md.Get(api.VersionHeader)
	if len(ver) == 0 {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not read "+api.VersionHeader+" header")
	}
	schedule := forks.NewOrderedSchedule(params.BeaconConfig())
	forkVer, err := schedule.VersionForName(ver[0])
	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not determine fork version: %v", err)
	}
	unmarshaler, err := detect.FromForkVersion(forkVer)
	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not create unmarshaler: %v", err)
	}
	block, err := unmarshaler.UnmarshalBeaconBlock(req.Data)
	if err != nil {
		return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not unmarshal request data into block: %v", err)
	}

	switch forkVer {
	case bytesutil.ToBytes4(params.BeaconConfig().CapellaForkVersion):
		if block.IsBlinded() {
			return nil, status.Error(codes.InvalidArgument, "Submitted block is blinded")
		}
		b, err := block.PbCapellaBlock()
		if err != nil {
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not get proto block: %v", err)
		}
		_, err = bs.V1Alpha1ValidatorServer.ProposeBeaconBlock(ctx, &zond.GenericSignedBeaconBlock{
			Block: &zond.GenericSignedBeaconBlock_Capella{
				Capella: b,
			},
		})
		if err != nil {
			if strings.Contains(err.Error(), validator.CouldNotDecodeBlock) {
				return &emptypb.Empty{}, status.Error(codes.InvalidArgument, err.Error())
			}
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not propose block: %v", err)
		}
		return &emptypb.Empty{}, nil
	case bytesutil.ToBytes4(params.BeaconConfig().BellatrixForkVersion):
		if block.IsBlinded() {
			return nil, status.Error(codes.InvalidArgument, "Submitted block is blinded")
		}
		b, err := block.PbBellatrixBlock()
		if err != nil {
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not get proto block: %v", err)
		}
		_, err = bs.V1Alpha1ValidatorServer.ProposeBeaconBlock(ctx, &zond.GenericSignedBeaconBlock{
			Block: &zond.GenericSignedBeaconBlock_Bellatrix{
				Bellatrix: b,
			},
		})
		if err != nil {
			if strings.Contains(err.Error(), validator.CouldNotDecodeBlock) {
				return &emptypb.Empty{}, status.Error(codes.InvalidArgument, err.Error())
			}
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not propose block: %v", err)
		}
		return &emptypb.Empty{}, nil
	case bytesutil.ToBytes4(params.BeaconConfig().AltairForkVersion):
		b, err := block.PbAltairBlock()
		if err != nil {
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not get proto block: %v", err)
		}
		_, err = bs.V1Alpha1ValidatorServer.ProposeBeaconBlock(ctx, &zond.GenericSignedBeaconBlock{
			Block: &zond.GenericSignedBeaconBlock_Altair{
				Altair: b,
			},
		})
		if err != nil {
			if strings.Contains(err.Error(), validator.CouldNotDecodeBlock) {
				return &emptypb.Empty{}, status.Error(codes.InvalidArgument, err.Error())
			}
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not propose block: %v", err)
		}
		return &emptypb.Empty{}, nil
	case bytesutil.ToBytes4(params.BeaconConfig().GenesisForkVersion):
		b, err := block.PbPhase0Block()
		if err != nil {
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not get proto block: %v", err)
		}
		_, err = bs.V1Alpha1ValidatorServer.ProposeBeaconBlock(ctx, &zond.GenericSignedBeaconBlock{
			Block: &zond.GenericSignedBeaconBlock_Phase0{
				Phase0: b,
			},
		})
		if err != nil {
			if strings.Contains(err.Error(), validator.CouldNotDecodeBlock) {
				return &emptypb.Empty{}, status.Error(codes.InvalidArgument, err.Error())
			}
			return &emptypb.Empty{}, status.Errorf(codes.Internal, "Could not propose block: %v", err)
		}
		return &emptypb.Empty{}, nil
	default:
		return &emptypb.Empty{}, status.Errorf(codes.InvalidArgument, "Unsupported fork %s", string(forkVer[:]))
	}
}

// GetBlock retrieves block details for given block ID.
func (bs *Server) GetBlock(ctx context.Context, req *zondpbv1.BlockRequest) (*zondpbv1.BlockResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.GetBlock")
	defer span.End()

	blk, err := bs.Blocker.Block(ctx, req.BlockId)
	err = handleGetBlockError(blk, err)
	if err != nil {
		return nil, err
	}
	blkRoot, err := blk.Block().HashTreeRoot()
	if err != nil {
		return nil, errors.Wrapf(err, "could not get block root")
	}

	result, err := getBlockPhase0(blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}
	if err := grpc.SetHeader(ctx, metadata.Pairs(api.VersionHeader, version.String(blk.Version()))); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not set "+api.VersionHeader+" header: %v", err)
	}
	result, err = getBlockAltair(blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}
	result, err = bs.getBlockBellatrix(ctx, blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}
	result, err = bs.getBlockCapella(ctx, blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}
	return nil, status.Errorf(codes.Internal, "Unknown block type %T", blk)
}

// GetBlockSSZ returns the SSZ-serialized version of the beacon block for given block ID.
func (bs *Server) GetBlockSSZ(ctx context.Context, req *zondpbv1.BlockRequest) (*zondpbv1.SSZContainer, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.GetBlockSSZ")
	defer span.End()

	blk, err := bs.Blocker.Block(ctx, req.BlockId)
	err = handleGetBlockError(blk, err)
	if err != nil {
		return nil, err
	}
	blkRoot, err := blk.Block().HashTreeRoot()
	if err != nil {
		return nil, errors.Wrapf(err, "could not get block root")
	}

	result, err := getSSZBlockPhase0(blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}
	result, err = getSSZBlockAltair(blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}
	result, err = bs.getSSZBlockBellatrix(ctx, blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}
	result, err = bs.getSSZBlockCapella(ctx, blk)
	if result != nil {
		result.Finalized = bs.FinalizationFetcher.IsFinalized(ctx, blkRoot)
		return result, nil
	}
	// ErrUnsupportedField means that we have another block type
	if !errors.Is(err, consensus_types.ErrUnsupportedField) {
		return nil, status.Errorf(codes.Internal, "Could not get signed beacon block: %v", err)
	}

	return nil, status.Errorf(codes.Internal, "Unknown block type %T", blk)
}

// GetBlockRoot retrieves hashTreeRoot of ReadOnlyBeaconBlock/BeaconBlockHeader.
func (bs *Server) GetBlockRoot(ctx context.Context, req *zondpbv1.BlockRequest) (*zondpbv1.BlockRootResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.GetBlockRoot")
	defer span.End()

	var root []byte
	var err error
	switch string(req.BlockId) {
	case "head":
		root, err = bs.ChainInfoFetcher.HeadRoot(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not retrieve head block: %v", err)
		}
		if root == nil {
			return nil, status.Errorf(codes.NotFound, "No head root was found")
		}
	case "finalized":
		finalized := bs.ChainInfoFetcher.FinalizedCheckpt()
		root = finalized.Root
	case "genesis":
		blk, err := bs.BeaconDB.GenesisBlock(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not retrieve blocks for genesis slot: %v", err)
		}
		if err := blocks.BeaconBlockIsNil(blk); err != nil {
			return nil, status.Errorf(codes.NotFound, "Could not find genesis block: %v", err)
		}
		blkRoot, err := blk.Block().HashTreeRoot()
		if err != nil {
			return nil, status.Error(codes.Internal, "Could not hash genesis block")
		}
		root = blkRoot[:]
	default:
		if len(req.BlockId) == 32 {
			blk, err := bs.BeaconDB.Block(ctx, bytesutil.ToBytes32(req.BlockId))
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Could not retrieve block for block root %#x: %v", req.BlockId, err)
			}
			if err := blocks.BeaconBlockIsNil(blk); err != nil {
				return nil, status.Errorf(codes.NotFound, "Could not find block: %v", err)
			}
			root = req.BlockId
		} else {
			slot, err := strconv.ParseUint(string(req.BlockId), 10, 64)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "Could not parse block ID: %v", err)
			}
			hasRoots, roots, err := bs.BeaconDB.BlockRootsBySlot(ctx, primitives.Slot(slot))
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Could not retrieve blocks for slot %d: %v", slot, err)
			}

			if !hasRoots {
				return nil, status.Error(codes.NotFound, "Could not find any blocks with given slot")
			}
			root = roots[0][:]
			if len(roots) == 1 {
				break
			}
			for _, blockRoot := range roots {
				canonical, err := bs.ChainInfoFetcher.IsCanonical(ctx, blockRoot)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "Could not determine if block root is canonical: %v", err)
				}
				if canonical {
					root = blockRoot[:]
					break
				}
			}
		}
	}

	b32Root := bytesutil.ToBytes32(root)
	isOptimistic, err := bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, b32Root)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not check if block is optimistic: %v", err)
	}

	return &zondpbv1.BlockRootResponse{
		Data: &zondpbv1.BlockRootContainer{
			Root: root,
		},
		ExecutionOptimistic: isOptimistic,
		Finalized:           bs.FinalizationFetcher.IsFinalized(ctx, b32Root),
	}, nil
}

// ListBlockAttestations retrieves attestation included in requested block.
func (bs *Server) ListBlockAttestations(ctx context.Context, req *zondpbv1.BlockRequest) (*zondpbv1.BlockAttestationsResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListBlockAttestations")
	defer span.End()

	blk, err := bs.Blocker.Block(ctx, req.BlockId)
	err = handleGetBlockError(blk, err)
	if err != nil {
		return nil, err
	}

	v1Alpha1Attestations := blk.Block().Body().Attestations()
	v1Attestations := make([]*zondpbv1.Attestation, 0, len(v1Alpha1Attestations))
	for _, att := range v1Alpha1Attestations {
		migratedAtt := migration.V1Alpha1ToV1Attestation(att)
		v1Attestations = append(v1Attestations, migratedAtt)
	}
	root, err := blk.Block().HashTreeRoot()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get block root: %v", err)
	}
	isOptimistic, err := bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, root)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not check if block is optimistic: %v", err)
	}
	return &zondpbv1.BlockAttestationsResponse{
		Data:                v1Attestations,
		ExecutionOptimistic: isOptimistic,
		Finalized:           bs.FinalizationFetcher.IsFinalized(ctx, root),
	}, nil
}

func handleGetBlockError(blk interfaces.ReadOnlySignedBeaconBlock, err error) error {
	if invalidBlockIdErr, ok := err.(*lookup.BlockIdParseError); ok {
		return status.Errorf(codes.InvalidArgument, "Invalid block ID: %v", invalidBlockIdErr)
	}
	if err != nil {
		return status.Errorf(codes.Internal, "Could not get block from block ID: %v", err)
	}
	if err := blocks.BeaconBlockIsNil(blk); err != nil {
		return status.Errorf(codes.NotFound, "Could not find requested block: %v", err)
	}
	return nil
}

func (bs *Server) getBlockCapella(ctx context.Context, blk interfaces.ReadOnlySignedBeaconBlock) (*zondpbv1.BlockResponse, error) {
	capellaBlk, err := blk.PbCapellaBlock()
	if err != nil {
		// ErrUnsupportedField means that we have another block type
		if errors.Is(err, consensus_types.ErrUnsupportedField) {
			if blindedCapellaBlk, err := blk.PbBlindedCapellaBlock(); err == nil {
				if blindedCapellaBlk == nil {
					return nil, errNilBlock
				}
				signedFullBlock, err := bs.ExecutionPayloadReconstructor.ReconstructFullBlock(ctx, blk)
				if err != nil {
					return nil, errors.Wrapf(err, "could not reconstruct full execution payload to create signed beacon block")
				}
				capellaBlk, err = signedFullBlock.PbCapellaBlock()
				if err != nil {
					return nil, errors.Wrapf(err, "could not get signed beacon block")
				}
				v2Blk, err := migration.V1Alpha1ToV1Block(capellaBlk.Block)
				if err != nil {
					return nil, errors.Wrapf(err, "could not convert beacon block")
				}
				root, err := blk.Block().HashTreeRoot()
				if err != nil {
					return nil, errors.Wrapf(err, "could not get block root")
				}
				isOptimistic, err := bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, root)
				if err != nil {
					return nil, errors.Wrapf(err, "could not check if block is optimistic")
				}
				sig := blk.Signature()
				return &zondpbv1.BlockResponse{
					Version: zondpbv1.Version_CAPELLA,
					Data: &zondpbv1.SignedBeaconBlockContainer{
						Message:   &zondpbv1.SignedBeaconBlockContainer_CapellaBlock{CapellaBlock: v2Blk},
						Signature: sig[:],
					},
					ExecutionOptimistic: isOptimistic,
				}, nil
			}
			return nil, err
		}
		return nil, err
	}

	if capellaBlk == nil {
		return nil, errNilBlock
	}
	v2Blk, err := migration.V1Alpha1ToV1Block(capellaBlk.Block)
	if err != nil {
		return nil, errors.Wrapf(err, "could not convert beacon block")
	}
	root, err := blk.Block().HashTreeRoot()
	if err != nil {
		return nil, errors.Wrapf(err, "could not get block root")
	}
	isOptimistic, err := bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, root)
	if err != nil {
		return nil, errors.Wrapf(err, "could not check if block is optimistic")
	}
	sig := blk.Signature()
	return &zondpbv1.BlockResponse{
		Version: zondpbv1.Version_CAPELLA,
		Data: &zondpbv1.SignedBeaconBlockContainer{
			Message:   &zondpbv1.SignedBeaconBlockContainer_CapellaBlock{CapellaBlock: v2Blk},
			Signature: sig[:],
		},
		ExecutionOptimistic: isOptimistic,
	}, nil
}

func (bs *Server) getSSZBlockCapella(ctx context.Context, blk interfaces.ReadOnlySignedBeaconBlock) (*zondpbv1.SSZContainer, error) {
	capellaBlk, err := blk.PbCapellaBlock()
	if err != nil {
		// ErrUnsupportedField means that we have another block type
		if errors.Is(err, consensus_types.ErrUnsupportedField) {
			if blindedCapellaBlk, err := blk.PbBlindedCapellaBlock(); err == nil {
				if blindedCapellaBlk == nil {
					return nil, errNilBlock
				}
				signedFullBlock, err := bs.ExecutionPayloadReconstructor.ReconstructFullBlock(ctx, blk)
				if err != nil {
					return nil, errors.Wrapf(err, "could not reconstruct full execution payload to create signed beacon block")
				}
				capellaBlk, err = signedFullBlock.PbCapellaBlock()
				if err != nil {
					return nil, errors.Wrapf(err, "could not get signed beacon block")
				}
				v2Blk, err := migration.V1Alpha1ToV1Block(capellaBlk.Block)
				if err != nil {
					return nil, errors.Wrapf(err, "could not convert signed beacon block")
				}
				root, err := blk.Block().HashTreeRoot()
				if err != nil {
					return nil, errors.Wrapf(err, "could not get block root")
				}
				isOptimistic, err := bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, root)
				if err != nil {
					return nil, errors.Wrapf(err, "could not check if block is optimistic")
				}
				sig := blk.Signature()
				data := &zondpbv1.SignedBeaconBlock{
					Message:   v2Blk,
					Signature: sig[:],
				}
				sszData, err := data.MarshalSSZ()
				if err != nil {
					return nil, errors.Wrapf(err, "could not marshal block into SSZ")
				}
				return &zondpbv1.SSZContainer{
					Version:             zondpbv1.Version_CAPELLA,
					ExecutionOptimistic: isOptimistic,
					Data:                sszData,
				}, nil
			}
			return nil, err
		}
		return nil, err
	}

	if capellaBlk == nil {
		return nil, errNilBlock
	}
	v2Blk, err := migration.V1Alpha1ToV1Block(capellaBlk.Block)
	if err != nil {
		return nil, errors.Wrapf(err, "could not convert signed beacon block")
	}
	root, err := blk.Block().HashTreeRoot()
	if err != nil {
		return nil, errors.Wrapf(err, "could not get block root")
	}
	isOptimistic, err := bs.OptimisticModeFetcher.IsOptimisticForRoot(ctx, root)
	if err != nil {
		return nil, errors.Wrapf(err, "could not check if block is optimistic")
	}
	sig := blk.Signature()
	data := &zondpbv1.SignedBeaconBlock{
		Message:   v2Blk,
		Signature: sig[:],
	}
	sszData, err := data.MarshalSSZ()
	if err != nil {
		return nil, errors.Wrapf(err, "could not marshal block into SSZ")
	}
	return &zondpbv1.SSZContainer{Version: zondpbv1.Version_CAPELLA, ExecutionOptimistic: isOptimistic, Data: sszData}, nil
}

func (bs *Server) submitCapellaBlock(ctx context.Context, capellaBlk *zondpbv1.BeaconBlock, sig []byte) error {
	v1alpha1Blk, err := migration.V1ToV1Alpha1SignedBlock(&zondpbv1.SignedBeaconBlock{Message: capellaBlk, Signature: sig})
	if err != nil {
		return status.Errorf(codes.InvalidArgument, "Could not convert block to v1 block")
	}
	_, err = bs.V1Alpha1ValidatorServer.ProposeBeaconBlock(ctx, &zond.GenericSignedBeaconBlock{
		Block: &zond.GenericSignedBeaconBlock_Capella{
			Capella: v1alpha1Blk,
		},
	})
	if err != nil {
		if strings.Contains(err.Error(), validator.CouldNotDecodeBlock) {
			return status.Error(codes.InvalidArgument, err.Error())
		}
		return status.Errorf(codes.Internal, "Could not propose block: %v", err)
	}
	return nil
}