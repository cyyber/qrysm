package stategen

import (
	"context"
	"testing"

	testDB "github.com/theQRL/qrysm/beacon-chain/db/testing"
	doublylinkedtree "github.com/theQRL/qrysm/beacon-chain/forkchoice/doubly-linked-tree"
	"github.com/theQRL/qrysm/config/params"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
)

func TestResume(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)

	service := New(beaconDB, doublylinkedtree.New())
	b := util.NewBeaconBlockZond()
	util.SaveBlock(t, ctx, service.beaconDB, b)
	root, err := b.Block.HashTreeRoot()
	require.NoError(t, err)
	beaconState, _ := util.DeterministicGenesisStateZond(t, 32)
	require.NoError(t, beaconState.SetSlot(params.BeaconConfig().SlotsPerEpoch))
	require.NoError(t, service.beaconDB.SaveState(ctx, beaconState, root))
	require.NoError(t, service.beaconDB.SaveGenesisBlockRoot(ctx, root))
	require.NoError(t, service.beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Root: root[:]}))

	resumeState, err := service.Resume(ctx, beaconState)
	require.NoError(t, err)
	require.DeepSSZEqual(t, beaconState.ToProtoUnsafe(), resumeState.ToProtoUnsafe())
	assert.Equal(t, params.BeaconConfig().SlotsPerEpoch, service.finalizedInfo.slot, "Did not get watned slot")
	assert.Equal(t, service.finalizedInfo.root, root, "Did not get wanted root")
	assert.NotNil(t, service.finalizedState(), "Wanted a non nil finalized state")
}

// The root comparison and the state copy must happen atomically: a mismatched
// root returns nil (so latestAncestor falls through to the other lookup paths)
// rather than the finalized state of a different root. Backport of upstream
// PR #16881.
func TestFinalizedStateIfRoot(t *testing.T) {
	beaconDB := testDB.SetupDB(t)
	service := New(beaconDB, doublylinkedtree.New())

	// Nothing cached yet: any root yields nil instead of a nil-state panic.
	assert.Equal(t, nil, service.finalizedStateIfRoot([32]byte{'a'}))

	beaconState, _ := util.DeterministicGenesisStateZond(t, 32)
	fRoot := [32]byte{'f'}
	service.SaveFinalizedState(0, fRoot, beaconState)

	got := service.finalizedStateIfRoot(fRoot)
	require.NotNil(t, got, "Wanted the finalized state for the matching root")
	require.DeepSSZEqual(t, beaconState.ToProtoUnsafe(), got.ToProtoUnsafe())
	assert.Equal(t, nil, service.finalizedStateIfRoot([32]byte{'o'}), "Wanted nil for a non-finalized root")

	// After the finalized info advances, the old root no longer matches.
	newState, _ := util.DeterministicGenesisStateZond(t, 32)
	require.NoError(t, newState.SetSlot(params.BeaconConfig().SlotsPerEpoch))
	newRoot := [32]byte{'n'}
	service.SaveFinalizedState(params.BeaconConfig().SlotsPerEpoch, newRoot, newState)
	assert.Equal(t, nil, service.finalizedStateIfRoot(fRoot), "Wanted nil for the stale finalized root")
	require.NotNil(t, service.finalizedStateIfRoot(newRoot))
}
