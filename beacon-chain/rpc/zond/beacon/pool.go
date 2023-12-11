package beacon

import (
	"context"
	"time"

	"github.com/theQRL/qrysm/v4/api/grpc"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/blocks"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/feed"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/feed/operation"
	corehelpers "github.com/theQRL/qrysm/v4/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/transition"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/zond/helpers"
	"github.com/theQRL/qrysm/v4/config/features"
	"github.com/theQRL/qrysm/v4/crypto/dilithium"
	"github.com/theQRL/qrysm/v4/proto/migration"
	zondpbalpha "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	zondpbv1 "github.com/theQRL/qrysm/v4/proto/zond/v1"
	"github.com/theQRL/qrysm/v4/runtime/version"
	"github.com/theQRL/qrysm/v4/time/slots"
	"go.opencensus.io/trace"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

const broadcastDilithiumChangesRateLimit = 128

// ListPoolAttestations retrieves attestations known by the node but
// not necessarily incorporated into any block. Allows filtering by committee index or slot.
func (bs *Server) ListPoolAttestations(ctx context.Context, req *zondpbv1.AttestationsPoolRequest) (*zondpbv1.AttestationsPoolResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListPoolAttestations")
	defer span.End()

	attestations := bs.AttestationsPool.AggregatedAttestations()
	unaggAtts, err := bs.AttestationsPool.UnaggregatedAttestations()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get unaggregated attestations: %v", err)
	}
	attestations = append(attestations, unaggAtts...)
	isEmptyReq := req.Slot == nil && req.CommitteeIndex == nil
	if isEmptyReq {
		allAtts := make([]*zondpbv1.Attestation, len(attestations))
		for i, att := range attestations {
			allAtts[i] = migration.V1Alpha1ToV1Attestation(att)
		}
		return &zondpbv1.AttestationsPoolResponse{Data: allAtts}, nil
	}

	filteredAtts := make([]*zondpbv1.Attestation, 0, len(attestations))
	for _, att := range attestations {
		bothDefined := req.Slot != nil && req.CommitteeIndex != nil
		committeeIndexMatch := req.CommitteeIndex != nil && att.Data.CommitteeIndex == *req.CommitteeIndex
		slotMatch := req.Slot != nil && att.Data.Slot == *req.Slot

		if bothDefined && committeeIndexMatch && slotMatch {
			filteredAtts = append(filteredAtts, migration.V1Alpha1ToV1Attestation(att))
		} else if !bothDefined && (committeeIndexMatch || slotMatch) {
			filteredAtts = append(filteredAtts, migration.V1Alpha1ToV1Attestation(att))
		}
	}
	return &zondpbv1.AttestationsPoolResponse{Data: filteredAtts}, nil
}

// SubmitAttestations submits Attestation object to node. If attestation passes all validation
// constraints, node MUST publish attestation on appropriate subnet.
func (bs *Server) SubmitAttestations(ctx context.Context, req *zondpbv1.SubmitAttestationsRequest) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.SubmitAttestation")
	defer span.End()

	var validAttestations []*zondpbalpha.Attestation
	var attFailures []*helpers.SingleIndexedVerificationFailure
	for i, sourceAtt := range req.Data {
		att := migration.V1ToV1Alpha1Attestation(sourceAtt)
		sigErr := false
		for _, sig := range att.Signatures {
			if _, err := dilithium.SignatureFromBytes(sig); err != nil {
				attFailures = append(attFailures, &helpers.SingleIndexedVerificationFailure{
					Index:   i,
					Message: "Incorrect attestation signature: " + err.Error(),
				})
				sigErr = true
				break
			}
		}
		if sigErr {
			continue
		}

		// Broadcast the unaggregated attestation on a feed to notify other services in the beacon node
		// of a received unaggregated attestation.
		// Note we can't send for aggregated att because we don't have selection proof.
		if !corehelpers.IsAggregated(att) {
			bs.OperationNotifier.OperationFeed().Send(&feed.Event{
				Type: operation.UnaggregatedAttReceived,
				Data: &operation.UnAggregatedAttReceivedData{
					Attestation: att,
				},
			})
		}

		validAttestations = append(validAttestations, att)
	}

	broadcastFailed := false
	for _, att := range validAttestations {
		// Determine subnet to broadcast attestation to
		wantedEpoch := slots.ToEpoch(att.Data.Slot)
		vals, err := bs.HeadFetcher.HeadValidatorsIndices(ctx, wantedEpoch)
		if err != nil {
			return nil, err
		}
		subnet := corehelpers.ComputeSubnetFromCommitteeAndSlot(uint64(len(vals)), att.Data.CommitteeIndex, att.Data.Slot)

		if err := bs.Broadcaster.BroadcastAttestation(ctx, subnet, att); err != nil {
			broadcastFailed = true
		}

		if corehelpers.IsAggregated(att) {
			if err := bs.AttestationsPool.SaveAggregatedAttestation(att); err != nil {
				log.WithError(err).Error("could not save aggregated att")
			}
		} else {
			if err := bs.AttestationsPool.SaveUnaggregatedAttestation(att); err != nil {
				log.WithError(err).Error("could not save unaggregated att")
			}
		}
	}
	if broadcastFailed {
		return nil, status.Errorf(
			codes.Internal,
			"Could not publish one or more attestations. Some attestations could be published successfully.")
	}

	if len(attFailures) > 0 {
		failuresContainer := &helpers.IndexedVerificationFailure{Failures: attFailures}
		err := grpc.AppendCustomErrorHeader(ctx, failuresContainer)
		if err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"One or more attestations failed validation. Could not prepare attestation failure information: %v",
				err,
			)
		}
		return nil, status.Errorf(codes.InvalidArgument, "One or more attestations failed validation")
	}

	return &emptypb.Empty{}, nil
}

// ListPoolAttesterSlashings retrieves attester slashings known by the node but
// not necessarily incorporated into any block.
func (bs *Server) ListPoolAttesterSlashings(ctx context.Context, _ *emptypb.Empty) (*zondpbv1.AttesterSlashingsPoolResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListPoolAttesterSlashings")
	defer span.End()

	headState, err := bs.ChainInfoFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	sourceSlashings := bs.SlashingsPool.PendingAttesterSlashings(ctx, headState, true /* return unlimited slashings */)

	slashings := make([]*zondpbv1.AttesterSlashing, len(sourceSlashings))
	for i, s := range sourceSlashings {
		slashings[i] = migration.V1Alpha1ToV1AttSlashing(s)
	}

	return &zondpbv1.AttesterSlashingsPoolResponse{
		Data: slashings,
	}, nil
}

// SubmitAttesterSlashing submits AttesterSlashing object to node's pool and
// if passes validation node MUST broadcast it to network.
func (bs *Server) SubmitAttesterSlashing(ctx context.Context, req *zondpbv1.AttesterSlashing) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.SubmitAttesterSlashing")
	defer span.End()

	headState, err := bs.ChainInfoFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	headState, err = transition.ProcessSlotsIfPossible(ctx, headState, req.Attestation_1.Data.Slot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not process slots: %v", err)
	}

	alphaSlashing := migration.V1ToV1Alpha1AttSlashing(req)
	err = blocks.VerifyAttesterSlashing(ctx, headState, alphaSlashing)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid attester slashing: %v", err)
	}

	err = bs.SlashingsPool.InsertAttesterSlashing(ctx, headState, alphaSlashing)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not insert attester slashing into pool: %v", err)
	}
	if !features.Get().DisableBroadcastSlashings {
		if err := bs.Broadcaster.Broadcast(ctx, alphaSlashing); err != nil {
			return nil, status.Errorf(codes.Internal, "Could not broadcast slashing object: %v", err)
		}
	}

	return &emptypb.Empty{}, nil
}

// ListPoolProposerSlashings retrieves proposer slashings known by the node
// but not necessarily incorporated into any block.
func (bs *Server) ListPoolProposerSlashings(ctx context.Context, _ *emptypb.Empty) (*zondpbv1.ProposerSlashingPoolResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListPoolProposerSlashings")
	defer span.End()

	headState, err := bs.ChainInfoFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	sourceSlashings := bs.SlashingsPool.PendingProposerSlashings(ctx, headState, true /* return unlimited slashings */)

	slashings := make([]*zondpbv1.ProposerSlashing, len(sourceSlashings))
	for i, s := range sourceSlashings {
		slashings[i] = migration.V1Alpha1ToV1ProposerSlashing(s)
	}

	return &zondpbv1.ProposerSlashingPoolResponse{
		Data: slashings,
	}, nil
}

// SubmitProposerSlashing submits AttesterSlashing object to node's pool and if
// passes validation node MUST broadcast it to network.
func (bs *Server) SubmitProposerSlashing(ctx context.Context, req *zondpbv1.ProposerSlashing) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.SubmitProposerSlashing")
	defer span.End()

	headState, err := bs.ChainInfoFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	headState, err = transition.ProcessSlotsIfPossible(ctx, headState, req.SignedHeader_1.Message.Slot)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not process slots: %v", err)
	}

	alphaSlashing := migration.V1ToV1Alpha1ProposerSlashing(req)
	err = blocks.VerifyProposerSlashing(headState, alphaSlashing)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid proposer slashing: %v", err)
	}

	err = bs.SlashingsPool.InsertProposerSlashing(ctx, headState, alphaSlashing)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not insert proposer slashing into pool: %v", err)
	}
	if !features.Get().DisableBroadcastSlashings {
		if err := bs.Broadcaster.Broadcast(ctx, alphaSlashing); err != nil {
			return nil, status.Errorf(codes.Internal, "Could not broadcast slashing object: %v", err)
		}
	}

	return &emptypb.Empty{}, nil
}

// ListPoolVoluntaryExits retrieves voluntary exits known by the node but
// not necessarily incorporated into any block.
func (bs *Server) ListPoolVoluntaryExits(ctx context.Context, _ *emptypb.Empty) (*zondpbv1.VoluntaryExitsPoolResponse, error) {
	_, span := trace.StartSpan(ctx, "beacon.ListPoolVoluntaryExits")
	defer span.End()

	sourceExits, err := bs.VoluntaryExitsPool.PendingExits()
	if err != nil {
		return nil, status.Error(codes.Internal, "Could not get exits from the pool")
	}
	exits := make([]*zondpbv1.SignedVoluntaryExit, len(sourceExits))
	for i, s := range sourceExits {
		exits[i] = migration.V1Alpha1ToV1Exit(s)
	}

	return &zondpbv1.VoluntaryExitsPoolResponse{
		Data: exits,
	}, nil
}

// SubmitVoluntaryExit submits SignedVoluntaryExit object to node's pool
// and if passes validation node MUST broadcast it to network.
func (bs *Server) SubmitVoluntaryExit(ctx context.Context, req *zondpbv1.SignedVoluntaryExit) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.SubmitVoluntaryExit")
	defer span.End()

	headState, err := bs.ChainInfoFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	s, err := slots.EpochStart(req.Message.Epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get epoch from message: %v", err)
	}
	headState, err = transition.ProcessSlotsIfPossible(ctx, headState, s)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not process slots: %v", err)
	}

	validator, err := headState.ValidatorAtIndexReadOnly(req.Message.ValidatorIndex)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get exiting validator: %v", err)
	}
	alphaExit := migration.V1ToV1Alpha1Exit(req)
	err = blocks.VerifyExitAndSignature(validator, headState.Slot(), headState.Fork(), alphaExit, headState.GenesisValidatorsRoot())
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "Invalid voluntary exit: %v", err)
	}

	bs.VoluntaryExitsPool.InsertVoluntaryExit(alphaExit)
	if err := bs.Broadcaster.Broadcast(ctx, alphaExit); err != nil {
		return nil, status.Errorf(codes.Internal, "Could not broadcast voluntary exit object: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// SubmitSignedDilithiumToExecutionChanges submits said object to the node's pool
// if it passes validation the node must broadcast it to the network.
func (bs *Server) SubmitSignedDilithiumToExecutionChanges(ctx context.Context, req *zondpbv1.SubmitDilithiumToExecutionChangesRequest) (*emptypb.Empty, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.SubmitSignedDilithiumToExecutionChanges")
	defer span.End()
	st, err := bs.ChainInfoFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}
	var failures []*helpers.SingleIndexedVerificationFailure
	var toBroadcast []*zondpbalpha.SignedDilithiumToExecutionChange

	for i, change := range req.GetChanges() {
		alphaChange := migration.V1ToV1Alpha1SignedDilithiumToExecutionChange(change)
		_, err = blocks.ValidateDilithiumToExecutionChange(st, alphaChange)
		if err != nil {
			failures = append(failures, &helpers.SingleIndexedVerificationFailure{
				Index:   i,
				Message: "Could not validate SignedDilithiumToExecutionChange: " + err.Error(),
			})
			continue
		}
		if err := blocks.VerifyDilithiumChangeSignature(st, change); err != nil {
			failures = append(failures, &helpers.SingleIndexedVerificationFailure{
				Index:   i,
				Message: "Could not validate signature: " + err.Error(),
			})
			continue
		}
		bs.OperationNotifier.OperationFeed().Send(&feed.Event{
			Type: operation.DilithiumToExecutionChangeReceived,
			Data: &operation.DilithiumToExecutionChangeReceivedData{
				Change: alphaChange,
			},
		})
		bs.DilithiumChangesPool.InsertDilithiumToExecChange(alphaChange)
		if st.Version() >= version.Capella {
			toBroadcast = append(toBroadcast, alphaChange)
		}
	}
	go bs.broadcastDilithiumChanges(ctx, toBroadcast)
	if len(failures) > 0 {
		failuresContainer := &helpers.IndexedVerificationFailure{Failures: failures}
		err := grpc.AppendCustomErrorHeader(ctx, failuresContainer)
		if err != nil {
			return nil, status.Errorf(
				codes.InvalidArgument,
				"One or more DilithiumToExecutionChange failed validation. Could not prepare DilithiumToExecutionChange failure information: %v",
				err,
			)
		}
		return nil, status.Errorf(codes.InvalidArgument, "One or more DilithiumToExecutionChange failed validation")
	}
	return &emptypb.Empty{}, nil
}

// broadcastDilithiumBatch broadcasts the first `broadcastDilithiumChangesRateLimit` messages from the slice pointed to by ptr.
// It validates the messages again because they could have been invalidated by being included in blocks since the last validation.
// It removes the messages from the slice and modifies it in place.
func (bs *Server) broadcastDilithiumBatch(ctx context.Context, ptr *[]*zondpbalpha.SignedDilithiumToExecutionChange) {
	limit := broadcastDilithiumChangesRateLimit
	if len(*ptr) < broadcastDilithiumChangesRateLimit {
		limit = len(*ptr)
	}
	st, err := bs.ChainInfoFetcher.HeadStateReadOnly(ctx)
	if err != nil {
		log.WithError(err).Error("could not get head state")
		return
	}
	for _, ch := range (*ptr)[:limit] {
		if ch != nil {
			_, err := blocks.ValidateDilithiumToExecutionChange(st, ch)
			if err != nil {
				log.WithError(err).Error("could not validate Dilithium to execution change")
				continue
			}
			if err := bs.Broadcaster.Broadcast(ctx, ch); err != nil {
				log.WithError(err).Error("could not broadcast Dilithium to execution changes.")
			}
		}
	}
	*ptr = (*ptr)[limit:]
}

func (bs *Server) broadcastDilithiumChanges(ctx context.Context, changes []*zondpbalpha.SignedDilithiumToExecutionChange) {
	bs.broadcastDilithiumBatch(ctx, &changes)
	if len(changes) == 0 {
		return
	}

	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			bs.broadcastDilithiumBatch(ctx, &changes)
			if len(changes) == 0 {
				return
			}
		}
	}
}

// ListDilithiumToExecutionChanges retrieves Dilithium to execution changes known by the node but not necessarily incorporated into any block
func (bs *Server) ListDilithiumToExecutionChanges(ctx context.Context, _ *emptypb.Empty) (*zondpbv1.DilithiumToExecutionChangesPoolResponse, error) {
	ctx, span := trace.StartSpan(ctx, "beacon.ListDilithiumToExecutionChanges")
	defer span.End()

	sourceChanges, err := bs.DilithiumChangesPool.PendingDilithiumToExecChanges()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get Dilithium to execution changes: %v", err)
	}

	changes := make([]*zondpbv1.SignedDilithiumToExecutionChange, len(sourceChanges))
	for i, ch := range sourceChanges {
		changes[i] = migration.V1Alpha1ToV1SignedDilithiumToExecChange(ch)
	}

	return &zondpbv1.DilithiumToExecutionChangesPoolResponse{
		Data: changes,
	}, nil
}
