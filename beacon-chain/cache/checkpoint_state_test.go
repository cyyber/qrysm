package cache

import (
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/state"
	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"google.golang.org/protobuf/proto"
)

func TestCheckpointStateCache_StateByCheckpoint(t *testing.T) {
	cache := NewCheckpointStateCache()

	cp1 := &zondpb.Checkpoint{Epoch: 1, Root: bytesutil.PadTo([]byte{'A'}, 32)}
	st, err := state_native.InitializeFromProtoCapella(&zondpb.BeaconStateCapella{
		GenesisValidatorsRoot: params.BeaconConfig().ZeroHash[:],
		Slot:                  64,
	})
	require.NoError(t, err)

	s, err := cache.StateByCheckpoint(cp1)
	require.NoError(t, err)
	assert.Equal(t, state.BeaconState(nil), s, "Expected state not to exist in empty cache")

	require.NoError(t, cache.AddCheckpointState(cp1, st))

	s, err = cache.StateByCheckpoint(cp1)
	require.NoError(t, err)

	pbState1, err := state_native.ProtobufBeaconStateCapella(s.ToProtoUnsafe())
	require.NoError(t, err)
	pbstate, err := state_native.ProtobufBeaconStateCapella(st.ToProtoUnsafe())
	require.NoError(t, err)
	if !proto.Equal(pbState1, pbstate) {
		t.Error("incorrectly cached state")
	}

	cp2 := &zondpb.Checkpoint{Epoch: 2, Root: bytesutil.PadTo([]byte{'B'}, 32)}
	st2, err := state_native.InitializeFromProtoCapella(&zondpb.BeaconStateCapella{
		Slot: 128,
	})
	require.NoError(t, err)
	require.NoError(t, cache.AddCheckpointState(cp2, st2))

	s, err = cache.StateByCheckpoint(cp2)
	require.NoError(t, err)
	assert.DeepEqual(t, st2.ToProto(), s.ToProto(), "incorrectly cached state")

	s, err = cache.StateByCheckpoint(cp1)
	require.NoError(t, err)
	assert.DeepEqual(t, st.ToProto(), s.ToProto(), "incorrectly cached state")
}

func TestCheckpointStateCache_MaxSize(t *testing.T) {
	c := NewCheckpointStateCache()
	st, err := state_native.InitializeFromProtoCapella(&zondpb.BeaconStateCapella{
		Slot: 0,
	})
	require.NoError(t, err)

	for i := uint64(0); i < uint64(maxCheckpointStateSize+100); i++ {
		require.NoError(t, st.SetSlot(primitives.Slot(i)))
		require.NoError(t, c.AddCheckpointState(&zondpb.Checkpoint{Epoch: primitives.Epoch(i), Root: make([]byte, 32)}, st))
	}

	assert.Equal(t, maxCheckpointStateSize, len(c.cache.Keys()))
}
