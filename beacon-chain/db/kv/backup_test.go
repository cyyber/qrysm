package kv

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
)

func TestStore_Backup(t *testing.T) {
	db, err := NewKVStore(context.Background(), t.TempDir())
	require.NoError(t, err, "Failed to instantiate DB")
	ctx := context.Background()

	head := util.NewBeaconBlockCapella()
	head.Block.Slot = 5000

	wsb, err := blocks.NewSignedBeaconBlock(head)
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(ctx, wsb))
	root, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	st, err := util.NewBeaconStateCapella()
	require.NoError(t, err)
	require.NoError(t, db.SaveState(ctx, st, root))
	require.NoError(t, db.SaveHeadBlockRoot(ctx, root))

	require.NoError(t, db.Backup(ctx, "", false))

	backupsPath := filepath.Join(db.databasePath, backupsDirectoryName)
	files, err := os.ReadDir(backupsPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(files), "No backups created")
	require.NoError(t, db.Close(), "Failed to close database")

	oldFilePath := filepath.Join(backupsPath, files[0].Name())
	newFilePath := filepath.Join(backupsPath, DatabaseFileName)
	// We rename the file to match the database file name
	// our NewKVStore function expects when opening a database.
	require.NoError(t, os.Rename(oldFilePath, newFilePath))

	backedDB, err := NewKVStore(ctx, backupsPath)
	require.NoError(t, err, "Failed to instantiate DB")
	t.Cleanup(func() {
		require.NoError(t, backedDB.Close(), "Failed to close database")
	})
	require.Equal(t, true, backedDB.HasState(ctx, root))
}

func TestStore_BackupMultipleBuckets(t *testing.T) {
	db, err := NewKVStore(context.Background(), t.TempDir())
	require.NoError(t, err, "Failed to instantiate DB")
	ctx := context.Background()

	startSlot := primitives.Slot(5000)

	for i := startSlot; i < 5200; i++ {
		head := util.NewBeaconBlockCapella()
		head.Block.Slot = i
		wsb, err := blocks.NewSignedBeaconBlock(head)
		require.NoError(t, err)
		require.NoError(t, db.SaveBlock(ctx, wsb))
		root, err := head.Block.HashTreeRoot()
		require.NoError(t, err)
		st, err := util.NewBeaconStateCapella()
		require.NoError(t, st.SetSlot(i))
		require.NoError(t, err)
		require.NoError(t, db.SaveState(ctx, st, root))
		require.NoError(t, db.SaveHeadBlockRoot(ctx, root))
	}

	require.NoError(t, db.Backup(ctx, "", false))

	backupsPath := filepath.Join(db.databasePath, backupsDirectoryName)
	files, err := os.ReadDir(backupsPath)
	require.NoError(t, err)
	require.NotEqual(t, 0, len(files), "No backups created")
	require.NoError(t, db.Close(), "Failed to close database")

	oldFilePath := filepath.Join(backupsPath, files[0].Name())
	newFilePath := filepath.Join(backupsPath, DatabaseFileName)
	// We rename the file to match the database file name
	// our NewKVStore function expects when opening a database.
	require.NoError(t, os.Rename(oldFilePath, newFilePath))

	backedDB, err := NewKVStore(ctx, backupsPath)
	require.NoError(t, err, "Failed to instantiate DB")
	t.Cleanup(func() {
		require.NoError(t, backedDB.Close(), "Failed to close database")
	})
	for i := startSlot; i < 5200; i++ {
		head := util.NewBeaconBlockCapella()
		head.Block.Slot = i
		root, err := head.Block.HashTreeRoot()
		require.NoError(t, err)
		nBlock, err := backedDB.Block(ctx, root)
		require.NoError(t, err)
		require.NotNil(t, nBlock)
		require.Equal(t, nBlock.Block().Slot(), i)
		nState, err := backedDB.State(ctx, root)
		require.NoError(t, err)
		require.NotNil(t, nState)
		require.Equal(t, nState.Slot(), i)
	}
}
