package state_native

import (
	"context"
	"testing"

	"github.com/theQRL/qrysm/v4/beacon-chain/state/state-native/types"
	"github.com/theQRL/qrysm/v4/config/params"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
)

func TestBeaconState_RotateAttestations(t *testing.T) {
	st, err := InitializeFromProtoPhase0(&zondpb.BeaconState{
		Slot:                      1,
		CurrentEpochAttestations:  []*zondpb.PendingAttestation{{Data: &zondpb.AttestationData{Slot: 456}}},
		PreviousEpochAttestations: []*zondpb.PendingAttestation{{Data: &zondpb.AttestationData{Slot: 123}}},
	})
	require.NoError(t, err)

	require.NoError(t, st.RotateAttestations())
	s, ok := st.(*BeaconState)
	require.Equal(t, true, ok)
	require.Equal(t, 0, len(s.currentEpochAttestationsVal()))
	require.Equal(t, primitives.Slot(456), s.previousEpochAttestationsVal()[0].Data.Slot)
}

func TestAppendBeyondIndicesLimit(t *testing.T) {
	zeroHash := params.BeaconConfig().ZeroHash
	mockblockRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := 0; i < len(mockblockRoots); i++ {
		mockblockRoots[i] = zeroHash[:]
	}

	mockstateRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := 0; i < len(mockstateRoots); i++ {
		mockstateRoots[i] = zeroHash[:]
	}
	mockrandaoMixes := make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector)
	for i := 0; i < len(mockrandaoMixes); i++ {
		mockrandaoMixes[i] = zeroHash[:]
	}
	st, err := InitializeFromProtoPhase0(&zondpb.BeaconState{
		Slot:                      1,
		CurrentEpochAttestations:  []*zondpb.PendingAttestation{{Data: &zondpb.AttestationData{Slot: 456}}},
		PreviousEpochAttestations: []*zondpb.PendingAttestation{{Data: &zondpb.AttestationData{Slot: 123}}},
		Validators:                []*zondpb.Validator{},
		Eth1Data:                  &zondpb.Eth1Data{},
		BlockRoots:                mockblockRoots,
		StateRoots:                mockstateRoots,
		RandaoMixes:               mockrandaoMixes,
	})
	require.NoError(t, err)
	_, err = st.HashTreeRoot(context.Background())
	require.NoError(t, err)
	s, ok := st.(*BeaconState)
	require.Equal(t, true, ok)
	for i := types.FieldIndex(0); i < types.FieldIndex(params.BeaconConfig().BeaconStateFieldCount); i++ {
		s.dirtyFields[i] = true
	}
	_, err = st.HashTreeRoot(context.Background())
	require.NoError(t, err)
	for i := 0; i < 10; i++ {
		assert.NoError(t, st.AppendValidator(&zondpb.Validator{}))
	}
	assert.Equal(t, false, s.rebuildTrie[types.Validators])
	assert.NotEqual(t, len(s.dirtyIndices[types.Validators]), 0)

	for i := 0; i < indicesLimit; i++ {
		assert.NoError(t, st.AppendValidator(&zondpb.Validator{}))
	}
	assert.Equal(t, true, s.rebuildTrie[types.Validators])
	assert.Equal(t, len(s.dirtyIndices[types.Validators]), 0)
}
