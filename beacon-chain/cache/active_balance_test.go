//go:build !fuzz

package cache

import (
	"encoding/binary"
	"math"
	"testing"

	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
)

func TestBalanceCache_AddGetBalance(t *testing.T) {
	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := 0; i < len(blockRoots); i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(i))
		blockRoots[i] = b
	}
	raw := &zondpb.BeaconStateCapella{
		BlockRoots: blockRoots,
	}
	st, err := state_native.InitializeFromProtoCapella(raw)
	require.NoError(t, err)

	cache := NewEffectiveBalanceCache()
	_, err = cache.Get(st)
	require.ErrorContains(t, ErrNotFound.Error(), err)

	b := uint64(100)
	require.NoError(t, cache.AddTotalEffectiveBalance(st, b))
	cachedB, err := cache.Get(st)
	require.NoError(t, err)
	require.Equal(t, b, cachedB)

	require.NoError(t, st.SetSlot(1000))
	_, err = cache.Get(st)
	require.ErrorContains(t, ErrNotFound.Error(), err)

	b = uint64(200)
	require.NoError(t, cache.AddTotalEffectiveBalance(st, b))
	cachedB, err = cache.Get(st)
	require.NoError(t, err)
	require.Equal(t, b, cachedB)

	require.NoError(t, st.SetSlot(1000+params.BeaconConfig().SlotsPerHistoricalRoot))
	_, err = cache.Get(st)
	require.ErrorContains(t, ErrNotFound.Error(), err)

	b = uint64(300)
	require.NoError(t, cache.AddTotalEffectiveBalance(st, b))
	cachedB, err = cache.Get(st)
	require.NoError(t, err)
	require.Equal(t, b, cachedB)
}

func TestBalanceCache_BalanceKey(t *testing.T) {
	blockRoots := make([][]byte, params.BeaconConfig().SlotsPerHistoricalRoot)
	for i := 0; i < len(blockRoots); i++ {
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, uint64(i))
		blockRoots[i] = b
	}
	raw := &zondpb.BeaconStateCapella{
		BlockRoots: blockRoots,
	}
	st, err := state_native.InitializeFromProtoCapella(raw)
	require.NoError(t, err)
	require.NoError(t, st.SetSlot(primitives.Slot(math.MaxUint64)))

	_, err = balanceCacheKey(st)
	require.NoError(t, err)
}
