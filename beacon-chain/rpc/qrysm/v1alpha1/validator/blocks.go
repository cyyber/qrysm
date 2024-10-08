package validator

import (
	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/async/event"
	"github.com/theQRL/qrysm/beacon-chain/core/blocks"
	"github.com/theQRL/qrysm/beacon-chain/core/feed"
	blockfeed "github.com/theQRL/qrysm/beacon-chain/core/feed/block"
	statefeed "github.com/theQRL/qrysm/beacon-chain/core/feed/state"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/runtime/version"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StreamBlocksAltair to clients every single time a block is received by the beacon node.
func (vs *Server) StreamBlocksAltair(req *zondpb.StreamBlocksRequest, stream zondpb.BeaconNodeValidator_StreamBlocksAltairServer) error {
	blocksChannel := make(chan *feed.Event, 1)
	var blockSub event.Subscription
	if req.VerifiedOnly {
		blockSub = vs.StateNotifier.StateFeed().Subscribe(blocksChannel)
	} else {
		blockSub = vs.BlockNotifier.BlockFeed().Subscribe(blocksChannel)
	}
	defer blockSub.Unsubscribe()

	for {
		select {
		case blockEvent := <-blocksChannel:
			if req.VerifiedOnly {
				if err := sendVerifiedBlocks(stream, blockEvent); err != nil {
					return err
				}
			} else {
				if err := vs.sendBlocks(stream, blockEvent); err != nil {
					return err
				}
			}
		case <-blockSub.Err():
			return status.Error(codes.Aborted, "Subscriber closed, exiting goroutine")
		case <-vs.Ctx.Done():
			return status.Error(codes.Canceled, "Context canceled")
		case <-stream.Context().Done():
			return status.Error(codes.Canceled, "Context canceled")
		}
	}
}

func sendVerifiedBlocks(stream zondpb.BeaconNodeValidator_StreamBlocksAltairServer, blockEvent *feed.Event) error {
	if blockEvent.Type != statefeed.BlockProcessed {
		return nil
	}
	data, ok := blockEvent.Data.(*statefeed.BlockProcessedData)
	if !ok || data == nil {
		return nil
	}
	b := &zondpb.StreamBlocksResponse{}
	switch data.SignedBlock.Version() {
	case version.Capella:
		pb, err := data.SignedBlock.Proto()
		if err != nil {
			return errors.Wrap(err, "could not get protobuf block")
		}
		phBlk, ok := pb.(*zondpb.SignedBeaconBlockCapella)
		if !ok {
			log.Warn("Mismatch between version and block type, was expecting SignedBeaconBlockCapella")
			return nil
		}
		b.Block = &zondpb.StreamBlocksResponse_CapellaBlock{CapellaBlock: phBlk}
	}

	if err := stream.Send(b); err != nil {
		return status.Errorf(codes.Unavailable, "Could not send over stream: %v", err)
	}

	return nil
}

func (vs *Server) sendBlocks(stream zondpb.BeaconNodeValidator_StreamBlocksAltairServer, blockEvent *feed.Event) error {
	if blockEvent.Type != blockfeed.ReceivedBlock {
		return nil
	}

	data, ok := blockEvent.Data.(*blockfeed.ReceivedBlockData)
	if !ok || data == nil {
		// Got bad data over the stream.
		return nil
	}
	if data.SignedBlock == nil {
		// One nil block shouldn't stop the stream.
		return nil
	}
	log := log.WithField("blockSlot", data.SignedBlock.Block().Slot())
	headState, err := vs.HeadFetcher.HeadStateReadOnly(vs.Ctx)
	if err != nil {
		log.WithError(err).Error("Could not get head state")
		return nil
	}
	signed := data.SignedBlock
	sig := signed.Signature()
	if err := blocks.VerifyBlockSignature(headState, signed.Block().ProposerIndex(), sig[:], signed.Block().HashTreeRoot); err != nil {
		log.WithError(err).Error("Could not verify block signature")
		return nil
	}
	b := &zondpb.StreamBlocksResponse{}
	pb, err := data.SignedBlock.Proto()
	if err != nil {
		return errors.Wrap(err, "could not get protobuf block")
	}
	switch p := pb.(type) {
	case *zondpb.SignedBeaconBlockCapella:
		b.Block = &zondpb.StreamBlocksResponse_CapellaBlock{CapellaBlock: p}
	default:
		log.Errorf("Unknown block type %T", p)
	}
	if err := stream.Send(b); err != nil {
		return status.Errorf(codes.Unavailable, "Could not send over stream: %v", err)
	}

	return nil
}
