package validator

import (
	"bytes"
	"context"
	"time"

	"github.com/theQRL/qrysm/beacon-chain/cache"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	coreTime "github.com/theQRL/qrysm/beacon-chain/core/time"
	"github.com/theQRL/qrysm/beacon-chain/core/transition"
	"github.com/theQRL/qrysm/beacon-chain/rpc/core"
	beaconState "github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	qrysmTime "github.com/theQRL/qrysm/time"
	"github.com/theQRL/qrysm/time/slots"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// GetDuties returns the duties assigned to a list of validators specified
// in the request object.
func (vs *Server) GetDuties(ctx context.Context, req *qrysmpb.DutiesRequest) (*qrysmpb.DutiesResponse, error) {
	if vs.SyncChecker.Syncing() {
		return nil, status.Error(codes.Unavailable, "Syncing to latest head, not ready to respond")
	}
	return vs.duties(ctx, req)
}

// Compute the validator duties from the head state's corresponding epoch
// for validators public key / indices requested.
func (vs *Server) duties(ctx context.Context, req *qrysmpb.DutiesRequest) (*qrysmpb.DutiesResponse, error) {
	currentEpoch := slots.ToEpoch(vs.TimeFetcher.CurrentSlot())
	if req.Epoch > currentEpoch+1 {
		return nil, status.Errorf(codes.Unavailable, "Request epoch %d can not be greater than next epoch %d", req.Epoch, currentEpoch+1)
	}

	s, err := vs.HeadFetcher.HeadState(ctx)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not get head state: %v", err)
	}

	// Advance state with empty transitions up to the requested epoch start slot.
	epochStartSlot, err := slots.EpochStart(req.Epoch)
	if err != nil {
		return nil, err
	}
	if s.Slot() < epochStartSlot {
		headRoot, err := vs.HeadFetcher.HeadRoot(ctx)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not retrieve head root: %v", err)
		}
		s, err = transition.ProcessSlotsUsingNextSlotCache(ctx, s, headRoot, epochStartSlot)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "Could not process slots up to %d: %v", epochStartSlot, err)
		}
	}
	committeeAssignments, proposerIndexToSlots, err := helpers.CommitteeAssignments(ctx, s, req.Epoch)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not compute committee assignments: %v", err)
	}
	// Query the next epoch assignments for committee subnet subscriptions.
	nextCommitteeAssignments, nextProposerIndexToSlots, err := helpers.CommitteeAssignments(ctx, s, req.Epoch+1)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "Could not compute next committee assignments: %v", err)
	}

	validatorAssignments := make([]*qrysmpb.DutiesResponse_Duty, 0, len(req.PublicKeys))
	nextValidatorAssignments := make([]*qrysmpb.DutiesResponse_Duty, 0, len(req.PublicKeys))

	for _, pubKey := range req.PublicKeys {
		if ctx.Err() != nil {
			return nil, status.Errorf(codes.Aborted, "Could not continue fetching assignments: %v", ctx.Err())
		}
		assignment := &qrysmpb.DutiesResponse_Duty{
			PublicKey: pubKey,
		}
		nextAssignment := &qrysmpb.DutiesResponse_Duty{
			PublicKey: pubKey,
		}
		idx, ok := s.ValidatorIndexByPubkey(bytesutil.ToBytes2592(pubKey))
		if ok {
			s := assignmentStatus(s, idx)

			assignment.ValidatorIndex = idx
			assignment.Status = s
			assignment.ProposerSlots = proposerIndexToSlots[idx]

			// The next epoch has no lookup for proposer indexes.
			nextAssignment.ValidatorIndex = idx
			nextAssignment.Status = s

			ca, ok := committeeAssignments[idx]
			if ok {
				assignment.Committee = ca.Committee
				assignment.AttesterSlot = ca.AttesterSlot
				assignment.CommitteeIndex = ca.CommitteeIndex
			}
			// Save the next epoch assignments.
			ca, ok = nextCommitteeAssignments[idx]
			if ok {
				nextAssignment.Committee = ca.Committee
				nextAssignment.AttesterSlot = ca.AttesterSlot
				nextAssignment.CommitteeIndex = ca.CommitteeIndex
			}
			// Cache proposer assignment for the current epoch.
			for _, slot := range proposerIndexToSlots[idx] {
				// Head root is empty because it can't be known until slot - 1. Same with payload id.
				vs.ProposerSlotIndexCache.SetProposerAndPayloadIDs(slot, idx, [8]byte{} /* payloadID */, [32]byte{} /* head root */)
			}
			// Cache proposer assignment for the next epoch.
			for _, slot := range nextProposerIndexToSlots[idx] {
				vs.ProposerSlotIndexCache.SetProposerAndPayloadIDs(slot, idx, [8]byte{} /* payloadID */, [32]byte{} /* head root */)
			}
		} else {
			// If the validator isn't in the beacon state, try finding their deposit to determine their status.
			// We don't need the lastActiveValidatorFn because we don't use the response in this.
			vStatus, _ := vs.validatorStatus(ctx, s, pubKey, nil)
			assignment.Status = vStatus.Status
		}

		// Are the validators in current or next epoch sync committee.
		if ok {
			assignment.IsSyncCommittee, err = helpers.IsCurrentPeriodSyncCommittee(s, idx)
			if err != nil {
				return nil, status.Errorf(codes.Internal, "Could not determine current epoch sync committee: %v", err)
			}
			if assignment.IsSyncCommittee {
				if err := registerSyncSubnetCurrentPeriod(s, req.Epoch, pubKey, assignment.Status); err != nil {
					return nil, err
				}
			}
			nextAssignment.IsSyncCommittee = assignment.IsSyncCommittee

			// Next epoch sync committee duty is assigned with next period sync committee only during
			// sync period epoch boundary (ie. EPOCHS_PER_SYNC_COMMITTEE_PERIOD - 1). Else wise
			// next epoch sync committee duty is the same as current epoch.
			nextSlotToEpoch := slots.ToEpoch(s.Slot() + 1)
			currentEpoch := coreTime.CurrentEpoch(s)
			if slots.SyncCommitteePeriod(nextSlotToEpoch) == slots.SyncCommitteePeriod(currentEpoch)+1 {
				nextAssignment.IsSyncCommittee, err = helpers.IsNextPeriodSyncCommittee(s, idx)
				if err != nil {
					return nil, status.Errorf(codes.Internal, "Could not determine next epoch sync committee: %v", err)
				}
				if nextAssignment.IsSyncCommittee {
					if err := registerSyncSubnetNextPeriod(s, req.Epoch, pubKey, nextAssignment.Status); err != nil {
						return nil, err
					}
				}
			}
		}

		validatorAssignments = append(validatorAssignments, assignment)
		nextValidatorAssignments = append(nextValidatorAssignments, nextAssignment)
		// Assign relevant validator to subnet.
		core.AssignValidatorToSubnetProto(pubKey, assignment.Status)
		core.AssignValidatorToSubnetProto(pubKey, nextAssignment.Status)
	}
	// Prune payload ID cache for any slots before request slot.
	vs.ProposerSlotIndexCache.PrunePayloadIDs(epochStartSlot)

	return &qrysmpb.DutiesResponse{
		CurrentEpochDuties: validatorAssignments,
		NextEpochDuties:    nextValidatorAssignments,
	}, nil
}

// AssignValidatorToSubnet checks the status and pubkey of a particular validator
// to discern whether persistent subnets need to be registered for them.
func (vs *Server) AssignValidatorToSubnet(_ context.Context, req *qrysmpb.AssignValidatorToSubnetRequest) (*emptypb.Empty, error) {
	core.AssignValidatorToSubnetProto(req.PublicKey, req.Status)
	return &emptypb.Empty{}, nil
}

func registerSyncSubnetCurrentPeriod(s beaconState.BeaconState, epoch primitives.Epoch, pubKey []byte, status qrysmpb.ValidatorStatus) error {
	committee, err := s.CurrentSyncCommittee()
	if err != nil {
		return err
	}
	syncCommPeriod := slots.SyncCommitteePeriod(epoch)
	registerSyncSubnet(epoch, syncCommPeriod, pubKey, committee, status)
	return nil
}

func registerSyncSubnetNextPeriod(s beaconState.BeaconState, epoch primitives.Epoch, pubKey []byte, status qrysmpb.ValidatorStatus) error {
	committee, err := s.NextSyncCommittee()
	if err != nil {
		return err
	}
	syncCommPeriod := slots.SyncCommitteePeriod(epoch)
	registerSyncSubnet(epoch, syncCommPeriod+1, pubKey, committee, status)
	return nil
}

// registerSyncSubnet checks the status and pubkey of a particular validator
// to discern whether persistent subnets need to be registered for them.
func registerSyncSubnet(currEpoch primitives.Epoch, syncPeriod uint64, pubkey []byte,
	syncCommittee *qrysmpb.SyncCommittee, status qrysmpb.ValidatorStatus) {
	if status != qrysmpb.ValidatorStatus_ACTIVE && status != qrysmpb.ValidatorStatus_EXITING {
		return
	}
	startEpoch := primitives.Epoch(syncPeriod * uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod))
	currPeriod := slots.SyncCommitteePeriod(currEpoch)
	endEpoch := startEpoch + params.BeaconConfig().EpochsPerSyncCommitteePeriod
	_, _, ok, expTime := cache.SyncSubnetIDs.GetSyncCommitteeSubnets(pubkey, startEpoch)
	if ok && expTime.After(qrysmTime.Now()) {
		return
	}
	firstValidEpoch, err := startEpoch.SafeSub(params.BeaconConfig().SyncCommitteeSubnetCount)
	if err != nil {
		firstValidEpoch = 0
	}
	// If we are processing for a future period, we only
	// add to the relevant subscription once we are at the valid
	// bound.
	if syncPeriod != currPeriod && currEpoch < firstValidEpoch {
		return
	}
	subs := subnetsFromCommittee(pubkey, syncCommittee)
	// Handle overflow in the event current epoch is less
	// than end epoch. This is an impossible condition, so
	// it is a defensive check.
	epochsToWatch, err := endEpoch.SafeSub(uint64(currEpoch))
	if err != nil {
		epochsToWatch = 0
	}
	epochDuration := time.Duration(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot))
	totalDuration := epochDuration * time.Duration(epochsToWatch) * time.Second
	cache.SyncSubnetIDs.AddSyncCommitteeSubnets(pubkey, startEpoch, subs, totalDuration)
}

// subnetsFromCommittee retrieves the relevant subnets for the chosen validator.
func subnetsFromCommittee(pubkey []byte, comm *qrysmpb.SyncCommittee) []uint64 {
	positions := make([]uint64, 0)
	for i, pkey := range comm.Pubkeys {
		if bytes.Equal(pubkey, pkey) {
			positions = append(positions, uint64(i)/(params.BeaconConfig().SyncCommitteeSize/params.BeaconConfig().SyncCommitteeSubnetCount))
		}
	}
	return positions
}
