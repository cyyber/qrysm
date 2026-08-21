package blockchain

import (
	"context"
	"testing"

	testDB "github.com/theQRL/qrysm/beacon-chain/db/testing"
	doublylinkedtree "github.com/theQRL/qrysm/beacon-chain/forkchoice/doubly-linked-tree"
	forkchoicetypes "github.com/theQRL/qrysm/beacon-chain/forkchoice/types"
	fieldparams "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"github.com/theQRL/qrysm/time/slots"
)

func TestService_VerifyWeakSubjectivityRoot(t *testing.T) {
	b := util.NewBeaconBlockZond()
	b.Block.Slot = 1792480
	r, err := b.Block.HashTreeRoot()
	require.NoError(t, err)

	// A block at the same slot that never becomes part of the finalized
	// canonical chain.
	forkBlock := util.NewBeaconBlockZond()
	forkBlock.Block.Slot = b.Block.Slot
	forkBlock.Block.ProposerIndex = 1
	forkRoot, err := forkBlock.Block.HashTreeRoot()
	require.NoError(t, err)

	blockEpoch := slots.ToEpoch(b.Block.Slot)
	childSlot, err := slots.EpochStart(blockEpoch + 1)
	require.NoError(t, err)
	childBlock := util.NewBeaconBlockZond()
	childBlock.Block.Slot = childSlot
	childBlock.Block.ParentRoot = r[:]
	childRoot, err := childBlock.Block.HashTreeRoot()
	require.NoError(t, err)
	tests := []struct {
		wsVerified     bool
		disabled       bool
		wantErr        error
		checkpt        *qrysmpb.Checkpoint
		finalizedEpoch primitives.Epoch
		name           string
	}{
		{
			name:     "nil root and epoch",
			disabled: true,
		},
		{
			name:           "not yet to verify, ws epoch higher than finalized epoch",
			checkpt:        &qrysmpb.Checkpoint{Root: bytesutil.PadTo([]byte{'a'}, 32), Epoch: blockEpoch},
			finalizedEpoch: blockEpoch - 1,
		},
		{
			name:           "can't find the block in DB",
			checkpt:        &qrysmpb.Checkpoint{Root: bytesutil.PadTo([]byte{'a'}, fieldparams.RootLength), Epoch: 1},
			finalizedEpoch: blockEpoch + 1,
			wantErr:        errWSBlockNotFound,
		},
		{
			name:           "can't find the block corresponds to ws epoch in DB",
			checkpt:        &qrysmpb.Checkpoint{Root: r[:], Epoch: blockEpoch - 2}, // Root belongs in epoch 1.
			finalizedEpoch: blockEpoch - 1,
			wantErr:        errWSBlockNotFoundInEpoch,
		},
		{
			name:           "block in db but not canonical",
			checkpt:        &qrysmpb.Checkpoint{Root: forkRoot[:], Epoch: blockEpoch},
			finalizedEpoch: blockEpoch + 1,
			wantErr:        errWSBlockNotCanonical,
		},
		{
			name:           "canonical block from next epoch fails epoch range",
			checkpt:        &qrysmpb.Checkpoint{Root: childRoot[:], Epoch: blockEpoch},
			finalizedEpoch: blockEpoch + 1,
			wantErr:        errWSBlockNotFoundInEpoch,
		},
		{
			name:           "can verify and pass",
			checkpt:        &qrysmpb.Checkpoint{Root: r[:], Epoch: blockEpoch},
			finalizedEpoch: blockEpoch + 1,
		},
		{
			name:           "not yet to verify, equal epoch",
			checkpt:        &qrysmpb.Checkpoint{Root: r[:], Epoch: blockEpoch},
			finalizedEpoch: blockEpoch,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			beaconDB := testDB.SetupDB(t)
			ctx := context.Background()
			util.SaveBlock(t, ctx, beaconDB, b)
			util.SaveBlock(t, ctx, beaconDB, forkBlock)
			util.SaveBlock(t, ctx, beaconDB, childBlock)
			require.NoError(t, beaconDB.SaveGenesisBlockRoot(ctx, bytesutil.ToBytes32(b.Block.ParentRoot)))
			// Finalizing through the child indexes b and childBlock as part of
			// the finalized canonical chain; forkBlock stays non-canonical.
			require.NoError(t, beaconDB.SaveFinalizedCheckpoint(ctx, &qrysmpb.Checkpoint{Root: childRoot[:], Epoch: blockEpoch + 1}))
			wv, err := NewWeakSubjectivityVerifier(tt.checkpt, beaconDB)
			require.NoError(t, err)
			require.Equal(t, !tt.disabled, wv.enabled)
			fcs := doublylinkedtree.New()
			s := &Service{
				cfg:        &config{BeaconDB: beaconDB, WeakSubjectivityCheckpt: tt.checkpt, ForkChoiceStore: fcs},
				wsVerifier: wv,
			}
			require.NoError(t, fcs.UpdateFinalizedCheckpoint(&forkchoicetypes.Checkpoint{Epoch: tt.finalizedEpoch}))
			cp := s.cfg.ForkChoiceStore.FinalizedCheckpoint()
			err = s.wsVerifier.VerifyWeakSubjectivity(context.Background(), cp.Epoch)
			if tt.wantErr == nil {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}
