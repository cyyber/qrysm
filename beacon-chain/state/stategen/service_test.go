package stategen

import (
	"context"
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/db"
	testDB "github.com/theQRL/qrysm/beacon-chain/db/testing"
	doublylinkedtree "github.com/theQRL/qrysm/beacon-chain/forkchoice/doubly-linked-tree"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"github.com/theQRL/qrysm/time/slots"
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
	// The migration cursor resumes from the last archived point in the DB; there
	// is none above genesis here.
	assert.Equal(t, primitives.Slot(0), service.finalizedInfo.slot, "Did not get wanted migration cursor")
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

// MigrateToCold runs in the background and its cursor is kept in memory only,
// so archived points between the last completed migration and a restart can be
// missing. Resume must start the cursor at the last archived point in the DB
// rather than at the finalized slot, so the next migration fills the gap.
func TestResume_MigrationCursorFromLastArchivedPoint(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	service := New(beaconDB, doublylinkedtree.New())
	service.slotsPerArchivedPoint = 4

	base, _ := util.DeterministicGenesisStateZond(t, 8)
	genesisRoot := [32]byte{'g'}
	require.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, genesisRoot))
	require.NoError(t, beaconDB.SaveState(ctx, base, genesisRoot))
	// Archived points 4 and 8 were written for canonical blocks; the one at 12
	// was lost to a crash.
	parent := genesisRoot
	for _, slot := range []primitives.Slot{4, 8} {
		parent = saveCanonicalArchivedPoint(t, ctx, beaconDB, base, slot, parent)
	}
	fRoot, fState := saveFinalizedBlock(t, ctx, beaconDB, base, 14, parent)

	_, err := service.Resume(ctx, fState)
	require.NoError(t, err)
	assert.Equal(t, primitives.Slot(8), service.finalizedInfo.slot, "cursor must resume at the last archived point")
	assert.Equal(t, fRoot, service.finalizedInfo.root)
	assert.Equal(t, primitives.Slot(14), service.finalizedInfo.state.Slot())
}

// A state at an archived slot that belongs to a block which lost fork choice
// must not become the migration cursor: the canonical chain may have no block
// at that slot and MigrateToCold never looks below its cursor, so every later
// migration would fail with an unknown block.
func TestResume_MigrationCursorSkipsNonCanonicalArchivedPoint(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	service := New(beaconDB, doublylinkedtree.New())
	service.slotsPerArchivedPoint = 4

	base, _ := util.DeterministicGenesisStateZond(t, 8)
	genesisRoot := [32]byte{'g'}
	require.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, genesisRoot))
	require.NoError(t, beaconDB.SaveState(ctx, base, genesisRoot))
	parent := genesisRoot
	for _, slot := range []primitives.Slot{4, 8} {
		parent = saveCanonicalArchivedPoint(t, ctx, beaconDB, base, slot, parent)
	}
	// A block at the archived slot 12 that lost fork choice: its hot state was
	// saved to the DB during non-finality and survives the dirty-state cleanup.
	orphan := util.NewBeaconBlockZond()
	orphan.Block.Slot = 12
	orphan.Block.ParentRoot = parent[:]
	orphan.Block.Body.Graffiti = []byte("orphan")
	orphan.Block.Body.Graffiti = append(orphan.Block.Body.Graffiti, make([]byte, 32-len(orphan.Block.Body.Graffiti))...)
	wOrphan := util.SaveBlock(t, ctx, beaconDB, orphan)
	orphanRoot, err := wOrphan.Block().HashTreeRoot()
	require.NoError(t, err)
	orphanState := base.Copy()
	require.NoError(t, orphanState.SetSlot(12))
	require.NoError(t, beaconDB.SaveState(ctx, orphanState, orphanRoot))
	// The canonical chain skipped slot 12: the finalized block, in the next
	// epoch and off an archive boundary, descends from 8. (Blocks of the
	// finalized epoch itself carry a "status pending" marker in the index;
	// archived points lie below it.)
	fRoot, fState := saveFinalizedBlock(t, ctx, beaconDB, base, params.BeaconConfig().SlotsPerEpoch+2, parent)

	_, err = service.Resume(ctx, fState)
	require.NoError(t, err)
	assert.Equal(t, primitives.Slot(8), service.finalizedInfo.slot, "cursor must skip the non-canonical archived point")
	// The migration from that cursor completes.
	require.NoError(t, service.MigrateToCold(ctx, fRoot))
}

// saveCanonicalArchivedPoint saves a block at slot with the given parent and
// its state, and returns the block root.
func saveCanonicalArchivedPoint(t *testing.T, ctx context.Context, beaconDB db.Database, base state.BeaconState, slot primitives.Slot, parent [32]byte) [32]byte {
	blk := util.NewBeaconBlockZond()
	blk.Block.Slot = slot
	blk.Block.ParentRoot = parent[:]
	wb := util.SaveBlock(t, ctx, beaconDB, blk)
	root, err := wb.Block().HashTreeRoot()
	require.NoError(t, err)
	st := base.Copy()
	require.NoError(t, st.SetSlot(slot))
	require.NoError(t, beaconDB.SaveState(ctx, st, root))
	return root
}

// saveFinalizedBlock saves a block at slot with the given parent, its state,
// and records it as the finalized checkpoint, which indexes its ancestry as
// finalized. It returns the block root and state.
func saveFinalizedBlock(t *testing.T, ctx context.Context, beaconDB db.Database, base state.BeaconState, slot primitives.Slot, parent [32]byte) ([32]byte, state.BeaconState) {
	blk := util.NewBeaconBlockZond()
	blk.Block.Slot = slot
	blk.Block.ParentRoot = parent[:]
	wb := util.SaveBlock(t, ctx, beaconDB, blk)
	fRoot, err := wb.Block().HashTreeRoot()
	require.NoError(t, err)
	fState := base.Copy()
	require.NoError(t, fState.SetSlot(slot))
	require.NoError(t, beaconDB.SaveState(ctx, fState, fRoot))
	require.NoError(t, beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Epoch: slots.ToEpoch(slot), Root: fRoot[:]}))
	return fRoot, fState
}
