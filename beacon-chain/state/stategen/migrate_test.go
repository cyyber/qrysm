package stategen

import (
	"context"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/theQRL/qrysm/beacon-chain/core/blocks"
	testDB "github.com/theQRL/qrysm/beacon-chain/db/testing"
	doublylinkedtree "github.com/theQRL/qrysm/beacon-chain/forkchoice/doubly-linked-tree"
	consensusblocks "github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
)

func TestMigrateToCold_CanSaveFinalizedInfo(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)
	service := New(beaconDB, doublylinkedtree.New())
	beaconState, _ := util.DeterministicGenesisStateZond(t, 32)
	b := util.NewBeaconBlockZond()
	b.Block.Slot = 1
	br, err := b.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b)
	require.NoError(t, service.epochBoundaryStateCache.put(br, beaconState))
	require.NoError(t, service.MigrateToCold(ctx, br))

	wanted := &finalizedInfo{state: beaconState, root: br, slot: 1}
	assert.DeepEqual(t, wanted.root, service.finalizedInfo.root)
	assert.Equal(t, wanted.slot, service.finalizedInfo.slot)
	expectedHTR, err := wanted.state.HashTreeRoot(ctx)
	require.NoError(t, err)
	actualHTR, err := service.finalizedInfo.state.HashTreeRoot(ctx)
	require.NoError(t, err)
	assert.DeepEqual(t, expectedHTR, actualHTR)
}

func TestMigrateToCold_HappyPath(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)

	service := New(beaconDB, doublylinkedtree.New())
	service.slotsPerArchivedPoint = 1
	beaconState, _ := util.DeterministicGenesisStateZond(t, 32)
	stateSlot := primitives.Slot(1)
	require.NoError(t, beaconState.SetSlot(stateSlot))
	b := util.NewBeaconBlockZond()
	b.Block.Slot = 2
	fRoot, err := b.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b)
	require.NoError(t, service.epochBoundaryStateCache.put(fRoot, beaconState))
	// The migration resolves canonicality through the finalized index, which the
	// blockchain service populates before calling MigrateToCold.
	require.NoError(t, service.beaconDB.SaveGenesisBlockRoot(ctx, [32]byte{}))
	require.NoError(t, service.beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Root: fRoot[:]}))
	require.NoError(t, service.MigrateToCold(ctx, fRoot))

	gotState, err := service.beaconDB.State(ctx, fRoot)
	require.NoError(t, err)
	assert.DeepSSZEqual(t, beaconState.ToProtoUnsafe(), gotState.ToProtoUnsafe(), "Did not save state")
	gotRoot := service.beaconDB.ArchivedPointRoot(ctx, stateSlot/service.slotsPerArchivedPoint)
	assert.Equal(t, fRoot, gotRoot, "Did not save archived root")
	lastIndex, err := service.beaconDB.LastArchivedSlot(ctx)
	require.NoError(t, err)
	assert.Equal(t, primitives.Slot(1), lastIndex, "Did not save last archived index")

	require.LogsContain(t, hook, "Saved state in DB")
}

func TestMigrateToCold_RegeneratePath(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)

	service := New(beaconDB, doublylinkedtree.New())
	service.slotsPerArchivedPoint = 1
	beaconState, pks := util.DeterministicGenesisStateZond(t, 32)
	genState := beaconState.Copy()
	genesisStateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err)
	genesis := blocks.NewGenesisBlock(genesisStateRoot[:])
	util.SaveBlock(t, ctx, beaconDB, genesis)
	gRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err)
	assert.NoError(t, beaconDB.SaveState(ctx, genState, gRoot))
	assert.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, gRoot))

	b1, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 1)
	require.NoError(t, err)
	wB1, err := consensusblocks.NewSignedBeaconBlock(b1)
	require.NoError(t, err)
	beaconState, err = executeStateTransitionStateGen(ctx, beaconState, wB1)
	require.NoError(t, err)
	r1, err := b1.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b1)
	require.NoError(t, service.beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 1, Root: r1[:]}))

	b4, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 4)
	require.NoError(t, err)
	r4, err := b4.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b4)
	require.NoError(t, service.beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 4, Root: r4[:]}))
	service.finalizedInfo = &finalizedInfo{
		slot:  0,
		root:  genesisStateRoot,
		state: genState,
	}
	// The migration resolves canonicality through the finalized index, which the
	// blockchain service populates before calling MigrateToCold.
	require.NoError(t, service.beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Root: r4[:]}))

	require.NoError(t, service.MigrateToCold(ctx, r4))

	s1, err := service.beaconDB.State(ctx, r1)
	require.NoError(t, err)
	assert.Equal(t, s1.Slot(), primitives.Slot(1), "Did not save state")
	gotRoot := service.beaconDB.ArchivedPointRoot(ctx, 1/service.slotsPerArchivedPoint)
	assert.Equal(t, r1, gotRoot, "Did not save archived root")
	lastIndex, err := service.beaconDB.LastArchivedSlot(ctx)
	require.NoError(t, err)
	assert.Equal(t, primitives.Slot(1), lastIndex, "Did not save last archived index")

	require.LogsContain(t, hook, "Saved state in DB")
}

func TestMigrateToCold_StateExistsInDB(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)

	service := New(beaconDB, doublylinkedtree.New())
	service.slotsPerArchivedPoint = 1
	beaconState, _ := util.DeterministicGenesisStateZond(t, 32)
	stateSlot := primitives.Slot(1)
	require.NoError(t, beaconState.SetSlot(stateSlot))
	b := util.NewBeaconBlockZond()
	b.Block.Slot = 2
	fRoot, err := b.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b)
	require.NoError(t, service.epochBoundaryStateCache.put(fRoot, beaconState))
	require.NoError(t, service.beaconDB.SaveState(ctx, beaconState, fRoot))
	// The migration resolves canonicality through the finalized index, which the
	// blockchain service populates before calling MigrateToCold.
	require.NoError(t, service.beaconDB.SaveGenesisBlockRoot(ctx, [32]byte{}))
	require.NoError(t, service.beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Root: fRoot[:]}))

	service.saveHotStateDB.blockRootsOfSavedStates = [][32]byte{{1}, {2}, {3}, {4}, fRoot}
	require.NoError(t, service.MigrateToCold(ctx, fRoot))
	assert.DeepEqual(t, [][32]byte{{1}, {2}, {3}, {4}}, service.saveHotStateDB.blockRootsOfSavedStates)
	assert.LogsDoNotContain(t, hook, "Saved state in DB")
}

func TestMigrateToCold_ParallelCalls(t *testing.T) {
	hook := logTest.NewGlobal()
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)

	service := New(beaconDB, doublylinkedtree.New())
	service.slotsPerArchivedPoint = 1
	beaconState, pks := util.DeterministicGenesisStateZond(t, 32)
	genState := beaconState.Copy()
	genesisStateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err)
	genesis := blocks.NewGenesisBlock(genesisStateRoot[:])
	util.SaveBlock(t, ctx, beaconDB, genesis)
	gRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err)
	assert.NoError(t, beaconDB.SaveState(ctx, beaconState, gRoot))
	assert.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, gRoot))

	b1, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 1)
	require.NoError(t, err)
	wB1, err := consensusblocks.NewSignedBeaconBlock(b1)
	require.NoError(t, err)
	beaconState, err = executeStateTransitionStateGen(ctx, beaconState, wB1)
	assert.NoError(t, err)
	r1, err := b1.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b1)
	require.NoError(t, service.beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 1, Root: r1[:]}))

	b4, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 4)
	require.NoError(t, err)
	wB4, err := consensusblocks.NewSignedBeaconBlock(b4)
	require.NoError(t, err)
	beaconState, err = executeStateTransitionStateGen(ctx, beaconState, wB4)
	assert.NoError(t, err)
	r4, err := b4.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b4)
	require.NoError(t, service.beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 4, Root: r4[:]}))

	b7, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 7)
	require.NoError(t, err)
	wB7, err := consensusblocks.NewSignedBeaconBlock(b7)
	require.NoError(t, err)
	beaconState, err = executeStateTransitionStateGen(ctx, beaconState, wB7)
	assert.NoError(t, err)
	r7, err := b7.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, service.beaconDB, b7)
	require.NoError(t, service.beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 7, Root: r7[:]}))

	// The migration resolves canonicality through the finalized index, which the
	// blockchain service populates before calling MigrateToCold.
	require.NoError(t, service.beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Root: r7[:]}))

	service.finalizedInfo = &finalizedInfo{
		slot:  0,
		root:  genesisStateRoot,
		state: genState,
	}
	service.saveHotStateDB.blockRootsOfSavedStates = [][32]byte{r1, r4, r7}

	// Run the migration routines concurrently for 2 different finalized roots.
	go func() {
		require.NoError(t, service.MigrateToCold(ctx, r4))
	}()

	require.NoError(t, service.MigrateToCold(ctx, r7))

	s1, err := service.beaconDB.State(ctx, r1)
	require.NoError(t, err)
	assert.Equal(t, s1.Slot(), primitives.Slot(1), "Did not save state")
	s4, err := service.beaconDB.State(ctx, r4)
	require.NoError(t, err)
	assert.Equal(t, s4.Slot(), primitives.Slot(4), "Did not save state")

	gotRoot := service.beaconDB.ArchivedPointRoot(ctx, 1/service.slotsPerArchivedPoint)
	assert.Equal(t, r1, gotRoot, "Did not save archived root")
	gotRoot = service.beaconDB.ArchivedPointRoot(ctx, 4)
	assert.Equal(t, r4, gotRoot, "Did not save archived root")
	lastIndex, err := service.beaconDB.LastArchivedSlot(ctx)
	require.NoError(t, err)
	assert.Equal(t, primitives.Slot(4), lastIndex, "Did not save last archived index")
	assert.DeepEqual(t, [][32]byte{r7}, service.saveHotStateDB.blockRootsOfSavedStates, "Did not remove all saved hot state roots")
	require.LogsContain(t, hook, "Saved state in DB")
}

// Regression test for the qrysm port of upstream PR #17371: at an archived
// point, the migration must not archive the state of a block that lost fork
// choice. Three mechanisms are covered in one run: a stale epoch-boundary
// cache entry keyed by a reorged block, an orphan sitting at the highest
// populated slot below the boundary, and two equivocating blocks at the same
// slot (which used to abort the migration with errUnknownBlock forever).
func TestMigrateToCold_SkipsReorgedBoundaryBlocks(t *testing.T) {
	ctx := context.Background()
	beaconDB := testDB.SetupDB(t)

	service := New(beaconDB, doublylinkedtree.New())
	service.slotsPerArchivedPoint = 3
	beaconState, pks := util.DeterministicGenesisStateZond(t, 32)
	genState := beaconState.Copy()
	genesisStateRoot, err := beaconState.HashTreeRoot(ctx)
	require.NoError(t, err)
	genesis := blocks.NewGenesisBlock(genesisStateRoot[:])
	util.SaveBlock(t, ctx, beaconDB, genesis)
	gRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, beaconDB.SaveState(ctx, genState, gRoot))
	require.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, gRoot))

	// Canonical block at slot 1.
	b1, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 1)
	require.NoError(t, err)
	wB1, err := consensusblocks.NewSignedBeaconBlock(b1)
	require.NoError(t, err)
	beaconState, err = executeStateTransitionStateGen(ctx, beaconState, wB1)
	require.NoError(t, err)
	r1, err := b1.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, beaconDB, b1)
	require.NoError(t, beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 1, Root: r1[:]}))

	// Two equivocating children of slot 1 at slot 2; both lose fork choice and
	// are never deleted from the DB.
	bO1, err := util.GenerateFullBlockZond(beaconState.Copy(), pks, util.DefaultBlockGenConfig(), 2)
	require.NoError(t, err)
	rO1, err := bO1.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, beaconDB, bO1)
	require.NoError(t, beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 2, Root: rO1[:]}))
	bO2, err := util.GenerateFullBlockZond(beaconState.Copy(), pks, util.DefaultBlockGenConfig(), 2)
	require.NoError(t, err)
	graffiti := make([]byte, 32)
	graffiti[0] = 'o'
	bO2.Block.Body.Graffiti = graffiti
	rO2, err := bO2.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NotEqual(t, rO1, rO2, "the equivocating blocks must differ for this test to mean anything")
	util.SaveBlock(t, ctx, beaconDB, bO2)
	require.NoError(t, beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 2, Root: rO2[:]}))

	// The canonical chain skips slot 2: slot 3 builds on slot 1.
	b3, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 3)
	require.NoError(t, err)
	require.DeepEqual(t, r1[:], b3.Block.ParentRoot, "slot 3 must build on slot 1, orphaning slot 2")
	wB3, err := consensusblocks.NewSignedBeaconBlock(b3)
	require.NoError(t, err)
	beaconState, err = executeStateTransitionStateGen(ctx, beaconState, wB3)
	require.NoError(t, err)
	r3, err := b3.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, beaconDB, b3)
	require.NoError(t, beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 3, Root: r3[:]}))

	// The new finalized block at slot 4.
	b4, err := util.GenerateFullBlockZond(beaconState, pks, util.DefaultBlockGenConfig(), 4)
	require.NoError(t, err)
	r4, err := b4.Block.HashTreeRoot()
	require.NoError(t, err)
	util.SaveBlock(t, ctx, beaconDB, b4)
	require.NoError(t, beaconDB.SaveStateSummary(ctx, &qrysmpb.StateSummary{Slot: 4, Root: r4[:]}))

	// The blockchain service saves the finalized checkpoint before calling
	// MigrateToCold, so the index holds 4 -> 3 -> 1 -> genesis, never the orphans.
	// The checkpoint epoch is 1: the finalized index is canonical-exact only for
	// epochs older than the checkpoint epoch (blocks in the checkpoint epoch
	// itself are all indexed as finalized-but-not-canonical), and the migration
	// processes slots from the previous finalized epoch, which is older.
	require.NoError(t, beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Root: r4[:], Epoch: 1}))
	require.Equal(t, false, beaconDB.IsFinalizedBlock(ctx, rO1))
	require.Equal(t, false, beaconDB.IsFinalizedBlock(ctx, rO2))

	// A stale epoch-boundary cache entry at the archived slot, keyed by a
	// reorged block: processed first, it holds the slot key until evicted.
	orphanCacheState := genState.Copy()
	require.NoError(t, orphanCacheState.SetSlot(3))
	require.NoError(t, service.epochBoundaryStateCache.put(rO1, orphanCacheState))

	service.finalizedInfo = &finalizedInfo{slot: 0, root: gRoot, state: genState}
	require.NoError(t, service.MigrateToCold(ctx, r4))

	// The archived state at the boundary belongs to the canonical slot-1 block,
	// not to either reorged block.
	s1, err := beaconDB.State(ctx, r1)
	require.NoError(t, err)
	require.NotNil(t, s1)
	assert.Equal(t, primitives.Slot(1), s1.Slot(), "Did not archive the canonical state")
	assert.Equal(t, false, beaconDB.HasState(ctx, rO1), "Archived a reorged block's state")
	assert.Equal(t, false, beaconDB.HasState(ctx, rO2), "Archived a reorged block's state")
}
