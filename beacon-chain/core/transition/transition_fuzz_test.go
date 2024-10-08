package transition

import (
	"context"
	"testing"

	fuzz "github.com/google/gofuzz"
	"github.com/theQRL/qrysm/beacon-chain/core/time"
	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
)

func TestFuzzExecuteStateTransition_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := context.Background()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	sb := &zondpb.SignedBeaconBlockCapella{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 1000; i++ {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(sb)
		if sb.Block == nil || sb.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(sb)
		require.NoError(t, err)
		s, err := ExecuteStateTransition(ctx, state, wsb)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v and signed block: %v", s, err, state, sb)
		}
	}
}

func TestFuzzCalculateStateRoot_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := context.Background()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	sb := &zondpb.SignedBeaconBlockCapella{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 1000; i++ {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(sb)
		if sb.Block == nil || sb.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(sb)
		require.NoError(t, err)
		stateRoot, err := CalculateStateRoot(ctx, state, wsb)
		if err != nil && stateRoot != [32]byte{} {
			t.Fatalf("state root should be empty on err. found: %v on error: %v for signed block: %v", stateRoot, err, sb)
		}
	}
}

func TestFuzzProcessSlot_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := context.Background()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 1000; i++ {
		fuzzer.Fuzz(state)
		s, err := ProcessSlot(ctx, state)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v", s, err, state)
		}
	}
}

func TestFuzzProcessSlots_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := context.Background()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	slot := primitives.Slot(0)
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 1000; i++ {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(&slot)
		s, err := ProcessSlots(ctx, state, slot)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for state: %v", s, err, state)
		}
	}
}

func TestFuzzprocessOperationsNoVerify_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := context.Background()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	bb := &zondpb.SignedBeaconBlockCapella{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 1000; i++ {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(bb)
		if bb.Block == nil || bb.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(bb)
		require.NoError(t, err)
		s, err := ProcessOperationsNoVerifyAttsSigs(ctx, state, wsb)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for block body: %v", s, err, bb)
		}
	}
}

func TestFuzzverifyOperationLengths_10000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	bb := &zondpb.SignedBeaconBlockCapella{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 10000; i++ {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(bb)
		if bb.Block == nil || bb.Block.Body == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(bb)
		require.NoError(t, err)
		_, err = VerifyOperationLengths(context.Background(), state, wsb)
		_ = err
	}
}

func TestFuzzCanProcessEpoch_10000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 10000; i++ {
		fuzzer.Fuzz(state)
		time.CanProcessEpoch(state)
	}
}

func TestFuzzProcessBlockForStateRoot_1000(t *testing.T) {
	SkipSlotCache.Disable()
	defer SkipSlotCache.Enable()
	ctx := context.Background()
	state, err := state_native.InitializeFromProtoUnsafeCapella(&zondpb.BeaconStateCapella{})
	require.NoError(t, err)
	sb := &zondpb.SignedBeaconBlockCapella{}
	fuzzer := fuzz.NewWithSeed(0)
	fuzzer.NilChance(0.1)
	for i := 0; i < 1000; i++ {
		fuzzer.Fuzz(state)
		fuzzer.Fuzz(sb)
		if sb.Block == nil || sb.Block.Body == nil || sb.Block.Body.Eth1Data == nil {
			continue
		}
		wsb, err := blocks.NewSignedBeaconBlock(sb)
		require.NoError(t, err)
		s, err := ProcessBlockForStateRoot(ctx, state, wsb)
		if err != nil && s != nil {
			t.Fatalf("state should be nil on err. found: %v on error: %v for signed block: %v", s, err, sb)
		}
	}
}
