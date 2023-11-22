package state_native

import (
	"testing"

	"github.com/theQRL/qrysm/v4/beacon-chain/state"
	testtmpl "github.com/theQRL/qrysm/v4/beacon-chain/state/testing"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

func TestBeaconState_SlotDataRace_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateSlotDataRace(t, func() (state.BeaconState, error) {
		return InitializeFromProtoCapella(&zondpb.BeaconState{Slot: 1})
	})
}

func TestBeaconState_MatchCurrentJustifiedCheckpt_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchCurrentJustifiedCheckptNative(
		t,
		func(cp *zondpb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&zondpb.BeaconState{CurrentJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchPreviousJustifiedCheckpt_Phase0(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchPreviousJustifiedCheckptNative(
		t,
		func(cp *zondpb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&zondpb.BeaconState{PreviousJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchPreviousJustifiedCheckpt_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchPreviousJustifiedCheckptNative(
		t,
		func(cp *zondpb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&zondpb.BeaconState{PreviousJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_ValidatorByPubkey_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorByPubkey(t, func() (state.BeaconState, error) {
		return InitializeFromProtoCapella(&zondpb.BeaconState{})
	})
}
