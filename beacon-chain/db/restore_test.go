package db

import (
	"context"
	"flag"
	"os"
	"path"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"
	"github.com/theQRL/qrysm/beacon-chain/db/kv"
	"github.com/theQRL/qrysm/cmd"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"github.com/urfave/cli/v2"
)

func TestRestore(t *testing.T) {
	logHook := logTest.NewGlobal()
	ctx := context.Background()

	backupDb, err := kv.NewKVStore(context.Background(), t.TempDir())
	require.NoError(t, err)
	head := util.NewBeaconBlockZond()
	head.Block.Slot = 5000
	wsb, err := blocks.NewSignedBeaconBlock(head)
	require.NoError(t, err)
	require.NoError(t, backupDb.SaveBlock(ctx, wsb))
	root, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	st, err := util.NewBeaconStateZond()
	require.NoError(t, err)
	require.NoError(t, backupDb.SaveState(ctx, st, root))
	require.NoError(t, backupDb.SaveHeadBlockRoot(ctx, root))
	require.NoError(t, err)
	require.NoError(t, backupDb.Close())
	// We rename the backup file so that we can later verify
	// whether the restored db has been renamed correctly.
	require.NoError(t, os.Rename(
		path.Join(backupDb.DatabasePath(), kv.DatabaseFileName),
		path.Join(backupDb.DatabasePath(), "backup.db")))

	restoreDir := t.TempDir()
	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String(cmd.RestoreSourceFileFlag.Name, "", "")
	set.String(cmd.RestoreTargetDirFlag.Name, "", "")
	require.NoError(t, set.Set(cmd.RestoreSourceFileFlag.Name, path.Join(backupDb.DatabasePath(), "backup.db")))
	require.NoError(t, set.Set(cmd.RestoreTargetDirFlag.Name, restoreDir))
	cliCtx := cli.NewContext(&app, set, nil)

	assert.NoError(t, Restore(cliCtx))

	files, err := os.ReadDir(path.Join(restoreDir, kv.BeaconNodeDbDirName))
	require.NoError(t, err)
	assert.Equal(t, 1, len(files))
	assert.Equal(t, kv.DatabaseFileName, files[0].Name())
	restoredDb, err := kv.NewKVStore(context.Background(), path.Join(restoreDir, kv.BeaconNodeDbDirName))
	defer func() {
		require.NoError(t, restoredDb.Close())
	}()
	require.NoError(t, err)
	headBlock, err := restoredDb.HeadBlock(ctx)
	require.NoError(t, err)
	assert.Equal(t, primitives.Slot(5000), headBlock.Block().Slot(), "Restored database has incorrect data")
	assert.LogsContain(t, logHook, "Restore completed successfully")

}

// Restoring a database onto itself used to truncate it before a byte was read.
func TestRestore_RefusesToCopyDatabaseOntoItself(t *testing.T) {
	dataDir := t.TempDir()
	dbDir := path.Join(dataDir, kv.BeaconNodeDbDirName)
	store, err := kv.NewKVStore(context.Background(), dbDir)
	require.NoError(t, err)
	require.NoError(t, store.Close())
	dbFile := path.Join(dbDir, kv.DatabaseFileName)

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String(cmd.RestoreSourceFileFlag.Name, "", "")
	set.String(cmd.RestoreTargetDirFlag.Name, "", "")
	require.NoError(t, set.Set(cmd.RestoreSourceFileFlag.Name, dbFile))
	require.NoError(t, set.Set(cmd.RestoreTargetDirFlag.Name, dataDir))
	cliCtx := cli.NewContext(&app, set, nil)

	require.ErrorContains(t, "the restore source is the database file in the target directory", Restore(cliCtx))

	// The database is intact.
	store, err = kv.NewKVStore(context.Background(), dbDir)
	require.NoError(t, err)
	require.NoError(t, store.Close())
}

// A source living in the target directory under a temporary-looking name (a
// leftover of an interrupted restore) must be copied, not truncated: the
// restore uses a fresh temporary file of its own.

func TestRestore_SourceInTargetDirectoryIsNotTruncated(t *testing.T) {
	dataDir := t.TempDir()
	dbDir := path.Join(dataDir, kv.BeaconNodeDbDirName)
	store, err := kv.NewKVStore(context.Background(), dbDir)
	require.NoError(t, err)
	head := util.NewBeaconBlockZond()
	head.Block.Slot = 77
	wsb, err := blocks.NewSignedBeaconBlock(head)
	require.NoError(t, err)
	require.NoError(t, store.SaveBlock(context.Background(), wsb))
	root, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	st, err := util.NewBeaconStateZond()
	require.NoError(t, err)
	require.NoError(t, store.SaveState(context.Background(), st, root))
	require.NoError(t, store.SaveHeadBlockRoot(context.Background(), root))
	require.NoError(t, store.Close())
	// The only copy of the database sits under a leftover temporary name.
	source := path.Join(dbDir, kv.DatabaseFileName+".restore.tmp")
	require.NoError(t, os.Rename(path.Join(dbDir, kv.DatabaseFileName), source))

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String(cmd.RestoreSourceFileFlag.Name, "", "")
	set.String(cmd.RestoreTargetDirFlag.Name, "", "")
	require.NoError(t, set.Set(cmd.RestoreSourceFileFlag.Name, source))
	require.NoError(t, set.Set(cmd.RestoreTargetDirFlag.Name, dataDir))
	cliCtx := cli.NewContext(&app, set, nil)
	require.NoError(t, Restore(cliCtx))

	restored, err := kv.NewKVStore(context.Background(), dbDir)
	require.NoError(t, err)
	defer func() { require.NoError(t, restored.Close()) }()
	headBlock, err := restored.HeadBlock(context.Background())
	require.NoError(t, err)
	require.Equal(t, primitives.Slot(77), headBlock.Block().Slot())
	info, err := os.Stat(source)
	require.NoError(t, err)
	require.NotEqual(t, int64(0), info.Size(), "the source was truncated")
}

// A source that is not a database is refused before the target is touched.

func TestRestore_RefusesNonDatabaseSource(t *testing.T) {
	dataDir := t.TempDir()
	source := path.Join(t.TempDir(), "not-a-db")
	require.NoError(t, os.WriteFile(source, []byte("hello"), 0o600))

	app := cli.App{}
	set := flag.NewFlagSet("test", 0)
	set.String(cmd.RestoreSourceFileFlag.Name, "", "")
	set.String(cmd.RestoreTargetDirFlag.Name, "", "")
	require.NoError(t, set.Set(cmd.RestoreSourceFileFlag.Name, source))
	require.NoError(t, set.Set(cmd.RestoreTargetDirFlag.Name, dataDir))
	cliCtx := cli.NewContext(&app, set, nil)

	require.ErrorContains(t, "not a readable database file", Restore(cliCtx))
}
