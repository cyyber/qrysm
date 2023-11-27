package lookup

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	mock "github.com/theQRL/qrysm/v4/beacon-chain/blockchain/testing"
	dbtesting "github.com/theQRL/qrysm/v4/beacon-chain/db/testing"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/testutil"
	"github.com/theQRL/qrysm/v4/consensus-types/blocks"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	zondpbalpha "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
	"github.com/theQRL/qrysm/v4/testing/util"
)

func TestGetBlock(t *testing.T) {
	beaconDB := dbtesting.SetupDB(t)
	ctx := context.Background()

	genBlk, blkContainers := testutil.FillDBWithBlocks(ctx, t, beaconDB)
	canonicalRoots := make(map[[32]byte]bool)

	for _, bContr := range blkContainers {
		canonicalRoots[bytesutil.ToBytes32(bContr.BlockRoot)] = true
	}
	headBlock := blkContainers[len(blkContainers)-1]
	nextSlot := headBlock.GetCapellaBlock().Block.Slot + 1

	b2 := util.NewBeaconBlock()
	b2.Block.Slot = 30
	b2.Block.ParentRoot = bytesutil.PadTo([]byte{1}, 32)
	util.SaveBlock(t, ctx, beaconDB, b2)
	b3 := util.NewBeaconBlock()
	b3.Block.Slot = 30
	b3.Block.ParentRoot = bytesutil.PadTo([]byte{4}, 32)
	util.SaveBlock(t, ctx, beaconDB, b3)
	b4 := util.NewBeaconBlock()
	b4.Block.Slot = nextSlot
	b4.Block.ParentRoot = bytesutil.PadTo([]byte{8}, 32)
	util.SaveBlock(t, ctx, beaconDB, b4)

	wsb, err := blocks.NewSignedBeaconBlock(headBlock.Block.(*zondpbalpha.BeaconBlockContainer_CapellaBlock).CapellaBlock)
	require.NoError(t, err)

	fetcher := &BeaconDbBlocker{
		BeaconDB: beaconDB,
		ChainInfoFetcher: &mock.ChainService{
			DB:                  beaconDB,
			Block:               wsb,
			Root:                headBlock.BlockRoot,
			FinalizedCheckPoint: &zondpbalpha.Checkpoint{Root: blkContainers[64].BlockRoot},
			CanonicalRoots:      canonicalRoots,
		},
	}

	root, err := genBlk.Block.HashTreeRoot()
	require.NoError(t, err)

	tests := []struct {
		name    string
		blockID []byte
		want    *zondpbalpha.SignedBeaconBlock
		wantErr bool
	}{
		{
			name:    "slot",
			blockID: []byte("30"),
			want:    blkContainers[30].Block.(*zondpbalpha.BeaconBlockContainer_CapellaBlock).CapellaBlock,
		},
		{
			name:    "bad formatting",
			blockID: []byte("3bad0"),
			wantErr: true,
		},
		{
			name:    "canonical",
			blockID: []byte("30"),
			want:    blkContainers[30].Block.(*zondpbalpha.BeaconBlockContainer_CapellaBlock).CapellaBlock,
		},
		{
			name:    "non canonical",
			blockID: []byte(fmt.Sprintf("%d", nextSlot)),
			want:    nil,
		},
		{
			name:    "head",
			blockID: []byte("head"),
			want:    headBlock.Block.(*zondpbalpha.BeaconBlockContainer_CapellaBlock).CapellaBlock,
		},
		{
			name:    "finalized",
			blockID: []byte("finalized"),
			want:    blkContainers[64].Block.(*zondpbalpha.BeaconBlockContainer_CapellaBlock).CapellaBlock,
		},
		{
			name:    "genesis",
			blockID: []byte("genesis"),
			want:    genBlk,
		},
		{
			name:    "genesis root",
			blockID: root[:],
			want:    genBlk,
		},
		{
			name:    "root",
			blockID: blkContainers[20].BlockRoot,
			want:    blkContainers[20].Block.(*zondpbalpha.BeaconBlockContainer_CapellaBlock).CapellaBlock,
		},
		{
			name:    "non-existent root",
			blockID: bytesutil.PadTo([]byte("hi there"), 32),
			want:    nil,
		},
		{
			name:    "no block",
			blockID: []byte("105"),
			want:    nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := fetcher.Block(ctx, tt.blockID)
			if tt.wantErr {
				assert.NotEqual(t, err, nil, "no error has been returned")
				return
			}
			if tt.want == nil {
				assert.Equal(t, nil, result)
				return
			}
			require.NoError(t, err)
			pbBlock, err := result.PbCapellaBlock()
			require.NoError(t, err)
			if !reflect.DeepEqual(pbBlock, tt.want) {
				t.Error("Expected blocks to equal")
			}
		})
	}
}
