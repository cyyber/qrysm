package debug

import (
	"context"
	"testing"

	dbTest "github.com/theQRL/qrysm/v4/beacon-chain/db/testing"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
	"github.com/theQRL/qrysm/v4/testing/util"
)

func TestServer_GetBlock(t *testing.T) {
	db := dbTest.SetupDB(t)
	ctx := context.Background()

	b := util.NewBeaconBlockCapella()
	b.Block.Slot = 100
	util.SaveBlock(t, ctx, db, b)
	blockRoot, err := b.Block.HashTreeRoot()
	require.NoError(t, err)
	bs := &Server{
		BeaconDB: db,
	}
	res, err := bs.GetBlock(ctx, &zondpb.BlockRequestByRoot{
		BlockRoot: blockRoot[:],
	})
	require.NoError(t, err)
	wanted, err := b.MarshalSSZ()
	require.NoError(t, err)
	assert.DeepEqual(t, wanted, res.Encoded)

	// Checking for nil block.
	blockRoot = [32]byte{}
	res, err = bs.GetBlock(ctx, &zondpb.BlockRequestByRoot{
		BlockRoot: blockRoot[:],
	})
	require.NoError(t, err)
	assert.DeepEqual(t, []byte{}, res.Encoded)
}
