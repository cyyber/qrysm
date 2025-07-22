package state_native

import (
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/state"
	testtmpl "github.com/theQRL/qrysm/beacon-chain/state/testing"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func TestBeaconState_SlotDataRace_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateSlotDataRace(t, func() (state.BeaconState, error) {
		return InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{Slot: 1})
	})
}

func TestBeaconState_MatchCurrentJustifiedCheckpt_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchCurrentJustifiedCheckptNative(
		t,
		func(cp *qrysmpb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{CurrentJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_MatchPreviousJustifiedCheckpt_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateMatchPreviousJustifiedCheckptNative(
		t,
		func(cp *qrysmpb.Checkpoint) (state.BeaconState, error) {
			return InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{PreviousJustifiedCheckpoint: cp})
		},
	)
}

func TestBeaconState_ValidatorByPubkey_Capella(t *testing.T) {
	testtmpl.VerifyBeaconStateValidatorByPubkey(t, func() (state.BeaconState, error) {
		return InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{})
	})
}
