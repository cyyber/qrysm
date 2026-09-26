package blockchain

import (
	"context"
	"errors"
	"testing"
	"testing/synctest"

	"github.com/theQRL/go-qrl/common"
	"github.com/theQRL/qrysm/beacon-chain/core/feed"
	statefeed "github.com/theQRL/qrysm/beacon-chain/core/feed/state"
	"github.com/theQRL/qrysm/beacon-chain/db"
	"github.com/theQRL/qrysm/beacon-chain/execution"
	mockExecution "github.com/theQRL/qrysm/beacon-chain/execution/testing"
	"github.com/theQRL/qrysm/beacon-chain/state/stategen"
	"github.com/theQRL/qrysm/beacon-chain/verification"
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
)

type overlappingImportDB struct {
	db.HeadAccessDatabase
	root    [32]byte
	failure error
}

func (d *overlappingImportDB) SaveStateSummary(ctx context.Context, summary *qrysmpb.StateSummary) error {
	if bytesutil.ToBytes32(summary.Root) == d.root && d.failure != nil {
		return d.failure
	}
	return d.HeadAccessDatabase.SaveStateSummary(ctx, summary)
}

type overlappingImportEngine struct {
	*mockExecution.EngineClient
	hash, lastValid  [32]byte
	response         error
	entered, release chan struct{}
	waiting          bool
}

func (e *overlappingImportEngine) NewPayload(ctx context.Context, payload interfaces.ExecutionData, hashes []common.Hash, root *common.Hash) ([]byte, error) {
	if bytesutil.ToBytes32(payload.BlockHash()) == e.hash && !e.waiting {
		e.waiting = true
		close(e.entered)
		<-e.release
		return e.lastValid[:], e.response
	}
	return e.EngineClient.NewPayload(ctx, payload, hashes, root)
}

func TestReceiveBlock_OverlappingBatchImport(t *testing.T) {
	for _, tc := range []struct {
		name       string
		response   error
		failWrites bool
		batchEnd   int
	}{
		{name: "SYNCING", response: execution.ErrAcceptedSyncingPayloadStatus, batchEnd: 4},
		{name: "summary write fails", response: execution.ErrAcceptedSyncingPayloadStatus, failWrites: true, batchEnd: 4},
		{name: "VALID updates ancestors", failWrites: true, batchEnd: 4},
		{name: "VALID updates published head", failWrites: true, batchEnd: 3},
		{name: "INVALID removes accepted batch", response: execution.ErrInvalidPayloadStatus, batchEnd: 4},
		{name: "RPC failure preserves batch", response: errors.New("temporary execution RPC failure"), batchEnd: 4},
	} {
		t.Run(tc.name, func(t *testing.T) {
			f := newBatchExecutionFixture(t, 4)
			root := f.blks[2].Root()
			d := &overlappingImportDB{HeadAccessDatabase: f.s.cfg.BeaconDB, root: root}
			f.s.cfg.BeaconDB = d
			f.s.cfg.StateGen = stategen.New(d, f.s.cfg.ForkChoiceStore)
			for i := 0; i < 2; i++ {
				require.NoError(t, f.s.cfg.StateGen.SaveState(f.ctx, f.blks[i].Root(), f.states[i+1].Copy()))
			}
			payload, err := f.blks[2].Block().Body().Execution()
			require.NoError(t, err)
			parentPayload, err := f.blks[1].Block().Body().Execution()
			require.NoError(t, err)
			e := &overlappingImportEngine{
				EngineClient: f.engine, hash: bytesutil.ToBytes32(payload.BlockHash()),
				lastValid: bytesutil.ToBytes32(parentPayload.BlockHash()), response: tc.response,
			}
			f.s.cfg.ExecutionEngineCaller = e
			synctest.Test(t, func(t *testing.T) {
				t.Cleanup(synctest.Wait)
				driftGenesisTime(f.s, 5, 0)
				events := make(chan *feed.Event, 32)
				sub := f.s.cfg.StateNotifier.StateFeed().Subscribe(events)
				defer sub.Unsubscribe()
				e.entered, e.release = make(chan struct{}), make(chan struct{})
				result := make(chan error, 1)
				go func() { result <- f.s.ReceiveBlock(f.ctx, f.blks[2], root) }()
				<-e.entered
				synctest.Wait()
				// The batch accepts C3 (and optionally D4) before gossip's
				// payload response arrives. C3 is not cached as a state when
				// D4 ends the batch, so repeating its save can fail here.
				require.NoError(t, f.s.ReceiveBlockBatch(f.ctx, f.blks[2:tc.batchEnd]))
				require.NoError(t, f.s.saveInitSyncBlocks(f.ctx, true))
				require.Equal(t, true, f.s.cfg.ForkChoiceStore.HasNode(root))
				require.Equal(t, true, d.HasBlock(f.ctx, root))
				if tc.failWrites {
					d.failure = errors.New("temporary summary persistence failure")
				}
				close(e.release)
				err = <-result
				synctest.Wait()
				invalid := tc.response == execution.ErrInvalidPayloadStatus
				switch tc.response {
				case nil, execution.ErrAcceptedSyncingPayloadStatus:
					require.NoError(t, err, "a completed batch import must not repeat persistence")
				case execution.ErrInvalidPayloadStatus:
					require.Equal(t, true, IsInvalidBlock(err), "duplicate detection must not discard INVALID")
					require.Equal(t, root, InvalidBlockRoot(err))
					require.Equal(t, e.lastValid, InvalidBlockLVH(err))
				default:
					require.ErrorContains(t, tc.response.Error(), err)
				}
				require.Equal(t, invalid, errors.Is(err, verification.ErrInvalid))
				for _, b := range f.blks[2:tc.batchEnd] {
					assert.Equal(t, !invalid, f.s.cfg.ForkChoiceStore.HasNode(b.Root()))
					assert.Equal(t, !invalid, d.HasBlock(f.ctx, b.Root()))
					assert.Equal(t, !invalid, d.HasStateSummary(f.ctx, b.Root()))
				}
				if tc.response == nil {
					for _, b := range f.blks[1:3] {
						optimistic, err := f.s.IsOptimisticForRoot(f.ctx, b.Root())
						require.NoError(t, err)
						assert.Equal(t, false, optimistic, "retain the delayed VALID verdict")
					}
					optimistic, err := f.s.IsOptimistic(f.ctx)
					require.NoError(t, err)
					assert.Equal(t, tc.batchEnd == 4, optimistic, "refresh head validation without validating descendants")
				}
				if !invalid {
					d.failure = nil
					require.NoError(t, f.s.ReceiveBlock(f.ctx, f.blks[2], root))
					assert.Equal(t, true, d.HasBlock(f.ctx, root), "a duplicate must leave the stored block intact")
				}
				processed := make(map[[32]byte]int)
				for len(events) > 0 {
					event := <-events
					if event.Type == statefeed.BlockProcessed {
						processed[event.Data.(*statefeed.BlockProcessedData).BlockRoot]++
					}
				}
				for _, b := range f.blks[2:tc.batchEnd] {
					assert.Equal(t, 1, processed[b.Root()], "the overlapping import must not announce the block twice")
				}
			})
		})
	}
}
