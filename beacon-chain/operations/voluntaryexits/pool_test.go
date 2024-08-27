package voluntaryexits

import (
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/beacon-chain/core/time"
	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/config/params"
	types "github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/dilithium"
	"github.com/theQRL/qrysm/crypto/dilithium/common"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/time/slots"
)

func TestPendingExits(t *testing.T) {
	t.Run("empty pool", func(t *testing.T) {
		pool := NewPool()
		changes, err := pool.PendingExits()
		require.NoError(t, err)
		assert.Equal(t, 0, len(changes))
	})
	t.Run("non-empty pool", func(t *testing.T) {
		pool := NewPool()
		pool.InsertVoluntaryExit(&zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				Epoch:          0,
				ValidatorIndex: 0,
			},
		})
		pool.InsertVoluntaryExit(&zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				Epoch:          0,
				ValidatorIndex: 1,
			},
		})
		changes, err := pool.PendingExits()
		require.NoError(t, err)
		assert.Equal(t, 2, len(changes))
	})
}

func TestExitsForInclusion(t *testing.T) {
	spb := &zondpb.BeaconStateCapella{
		Fork: &zondpb.Fork{
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
		},
	}
	stateSlot := types.Slot(uint64(params.BeaconConfig().ShardCommitteePeriod) * uint64(params.BeaconConfig().SlotsPerEpoch))
	spb.Slot = stateSlot
	numValidators := 2 * params.BeaconConfig().MaxVoluntaryExits
	validators := make([]*zondpb.Validator, numValidators)
	exits := make([]*zondpb.VoluntaryExit, numValidators)
	privKeys := make([]common.SecretKey, numValidators)

	for i := range validators {
		v := &zondpb.Validator{}
		if i == len(validators)-2 {
			// exit for this validator is invalid
			v.ExitEpoch = 0
		} else {
			v.ExitEpoch = params.BeaconConfig().FarFutureEpoch
		}
		priv, err := dilithium.RandKey()
		require.NoError(t, err)
		privKeys[i] = priv
		pubkey := priv.PublicKey().Marshal()
		v.PublicKey = pubkey

		message := &zondpb.VoluntaryExit{
			ValidatorIndex: types.ValidatorIndex(i),
		}
		// exit for future slot
		if i == len(validators)-1 {
			message.Epoch = slots.ToEpoch(stateSlot) + 1
		}

		validators[i] = v
		exits[i] = message
	}
	spb.Validators = validators
	st, err := state_native.InitializeFromProtoCapella(spb)
	require.NoError(t, err)

	signedExits := make([]*zondpb.SignedVoluntaryExit, numValidators)
	for i, message := range exits {
		signature, err := signing.ComputeDomainAndSign(st, time.CurrentEpoch(st), message, params.BeaconConfig().DomainVoluntaryExit, privKeys[i])
		require.NoError(t, err)

		signed := &zondpb.SignedVoluntaryExit{
			Exit:      message,
			Signature: signature,
		}
		signedExits[i] = signed
	}

	t.Run("empty pool", func(t *testing.T) {
		pool := NewPool()
		exits, err := pool.ExitsForInclusion(st, stateSlot)
		require.NoError(t, err)
		assert.Equal(t, 0, len(exits))
	})
	t.Run("less than MaxVoluntaryExits in pool", func(t *testing.T) {
		pool := NewPool()
		for i := uint64(0); i < params.BeaconConfig().MaxVoluntaryExits-1; i++ {
			pool.InsertVoluntaryExit(signedExits[i])
		}
		exits, err := pool.ExitsForInclusion(st, stateSlot)
		require.NoError(t, err)
		assert.Equal(t, int(params.BeaconConfig().MaxVoluntaryExits)-1, len(exits))
	})
	t.Run("MaxVoluntaryExits in pool", func(t *testing.T) {
		pool := NewPool()
		for i := uint64(0); i < params.BeaconConfig().MaxVoluntaryExits; i++ {
			pool.InsertVoluntaryExit(signedExits[i])
		}
		exits, err := pool.ExitsForInclusion(st, stateSlot)
		require.NoError(t, err)
		assert.Equal(t, int(params.BeaconConfig().MaxVoluntaryExits), len(exits))
	})
	t.Run("more than MaxVoluntaryExits in pool", func(t *testing.T) {
		pool := NewPool()
		for i := uint64(0); i < numValidators; i++ {
			pool.InsertVoluntaryExit(signedExits[i])
		}
		exits, err := pool.ExitsForInclusion(st, stateSlot)
		require.NoError(t, err)
		assert.Equal(t, int(params.BeaconConfig().MaxVoluntaryExits), len(exits))
		for _, ch := range exits {
			assert.NotEqual(t, types.ValidatorIndex(params.BeaconConfig().MaxVoluntaryExits), ch.Exit.ValidatorIndex)
		}
	})
	t.Run("exit for future epoch not returned", func(t *testing.T) {
		pool := NewPool()
		pool.InsertVoluntaryExit(signedExits[len(signedExits)-1])
		exits, err := pool.ExitsForInclusion(st, stateSlot)
		require.NoError(t, err)
		assert.Equal(t, 0, len(exits))
	})
	t.Run("invalid exit not returned", func(t *testing.T) {
		pool := NewPool()
		pool.InsertVoluntaryExit(signedExits[len(signedExits)-2])
		exits, err := pool.ExitsForInclusion(st, stateSlot)
		require.NoError(t, err)
		assert.Equal(t, 0, len(exits))
	})
}

func TestInsertExit(t *testing.T) {
	t.Run("empty pool", func(t *testing.T) {
		pool := NewPool()
		exit := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			},
		}
		pool.InsertVoluntaryExit(exit)
		require.Equal(t, 1, pool.pending.Len())
		require.Equal(t, 1, len(pool.m))
		n, ok := pool.m[0]
		require.Equal(t, true, ok)
		v, err := n.Value()
		require.NoError(t, err)
		assert.DeepEqual(t, exit, v)
	})
	t.Run("item in pool", func(t *testing.T) {
		pool := NewPool()
		old := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			},
		}
		exit := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(1),
			},
		}
		pool.InsertVoluntaryExit(old)
		pool.InsertVoluntaryExit(exit)
		require.Equal(t, 2, pool.pending.Len())
		require.Equal(t, 2, len(pool.m))
		n, ok := pool.m[0]
		require.Equal(t, true, ok)
		v, err := n.Value()
		require.NoError(t, err)
		assert.DeepEqual(t, old, v)
		n, ok = pool.m[1]
		require.Equal(t, true, ok)
		v, err = n.Value()
		require.NoError(t, err)
		assert.DeepEqual(t, exit, v)
	})
	t.Run("validator index already exists", func(t *testing.T) {
		pool := NewPool()
		old := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			},
			Signature: []byte("old"),
		}
		exit := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			},
			Signature: []byte("exit"),
		}
		pool.InsertVoluntaryExit(old)
		pool.InsertVoluntaryExit(exit)
		assert.Equal(t, 1, pool.pending.Len())
		require.Equal(t, 1, len(pool.m))
		n, ok := pool.m[0]
		require.Equal(t, true, ok)
		v, err := n.Value()
		require.NoError(t, err)
		assert.DeepEqual(t, old, v)
	})
}

func TestMarkIncluded(t *testing.T) {
	t.Run("one element in pool", func(t *testing.T) {
		pool := NewPool()
		exit := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			}}
		pool.InsertVoluntaryExit(exit)
		pool.MarkIncluded(exit)
		assert.Equal(t, 0, pool.pending.Len())
		_, ok := pool.m[0]
		assert.Equal(t, false, ok)
	})
	t.Run("first of multiple elements", func(t *testing.T) {
		pool := NewPool()
		first := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			}}
		second := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(1),
			}}
		third := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(2),
			}}
		pool.InsertVoluntaryExit(first)
		pool.InsertVoluntaryExit(second)
		pool.InsertVoluntaryExit(third)
		pool.MarkIncluded(first)
		require.Equal(t, 2, pool.pending.Len())
		_, ok := pool.m[0]
		assert.Equal(t, false, ok)
	})
	t.Run("last of multiple elements", func(t *testing.T) {
		pool := NewPool()
		first := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			}}
		second := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(1),
			}}
		third := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(2),
			}}
		pool.InsertVoluntaryExit(first)
		pool.InsertVoluntaryExit(second)
		pool.InsertVoluntaryExit(third)
		pool.MarkIncluded(third)
		require.Equal(t, 2, pool.pending.Len())
		_, ok := pool.m[2]
		assert.Equal(t, false, ok)
	})
	t.Run("in the middle of multiple elements", func(t *testing.T) {
		pool := NewPool()
		first := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			}}
		second := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(1),
			}}
		third := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(2),
			}}
		pool.InsertVoluntaryExit(first)
		pool.InsertVoluntaryExit(second)
		pool.InsertVoluntaryExit(third)
		pool.MarkIncluded(second)
		require.Equal(t, 2, pool.pending.Len())
		_, ok := pool.m[1]
		assert.Equal(t, false, ok)
	})
	t.Run("not in pool", func(t *testing.T) {
		pool := NewPool()
		first := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(0),
			}}
		second := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(1),
			}}
		exit := &zondpb.SignedVoluntaryExit{
			Exit: &zondpb.VoluntaryExit{
				ValidatorIndex: types.ValidatorIndex(2),
			}}
		pool.InsertVoluntaryExit(first)
		pool.InsertVoluntaryExit(second)
		pool.MarkIncluded(exit)
		require.Equal(t, 2, pool.pending.Len())
		_, ok := pool.m[0]
		require.Equal(t, true, ok)
		assert.NotNil(t, pool.m[0])
		_, ok = pool.m[1]
		require.Equal(t, true, ok)
		assert.NotNil(t, pool.m[1])
	})
}
