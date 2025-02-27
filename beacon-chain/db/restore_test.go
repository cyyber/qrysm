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
	head := util.NewBeaconBlockCapella()
	head.Block.Slot = 5000
	wsb, err := blocks.NewSignedBeaconBlock(head)
	require.NoError(t, err)
	require.NoError(t, backupDb.SaveBlock(ctx, wsb))
	root, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	st, err := util.NewBeaconStateCapella()
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
