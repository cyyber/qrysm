package validator

import (
	"context"
	"testing"
	"time"

	mock "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	"github.com/theQRL/qrysm/beacon-chain/cache"
	"github.com/theQRL/qrysm/beacon-chain/rpc/core"
	mockSync "github.com/theQRL/qrysm/beacon-chain/sync/initial-sync/testing"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"github.com/theQRL/qrysm/time/slots"
	"google.golang.org/protobuf/proto"
)

func TestAttestationDataAtSlot_HandlesFarAwayJustifiedEpoch(t *testing.T) {
	// Scenario:
	//
	// State slot = 10000
	// Last justified slot = epoch start of 1500
	// HistoricalRootsLimit = 8192
	//
	// More background: https://github.com/prysmaticlabs/prysm/issues/2153
	// This test breaks if it doesn't use mainnet config

	// Ensure HistoricalRootsLimit matches scenario
	params.SetupTestConfigCleanup(t)
	cfg := params.MainnetConfig().Copy()
	cfg.HistoricalRootsLimit = 8192
	params.OverrideBeaconConfig(cfg)

	block := util.NewBeaconBlockCapella()
	block.Block.Slot = 40000
	epochBoundaryBlock := util.NewBeaconBlockCapella()
	var err error
	epochBoundaryBlock.Block.Slot, err = slots.EpochStart(slots.ToEpoch(40000))
	require.NoError(t, err)
	justifiedBlock := util.NewBeaconBlockCapella()
	justifiedBlock.Block.Slot, err = slots.EpochStart(slots.ToEpoch(1500))
	require.NoError(t, err)
	justifiedBlock.Block.Slot -= 2 // Imagine two skip block
	blockRoot, err := block.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash beacon block")
	justifiedBlockRoot, err := justifiedBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash justified block")
	epochBoundaryRoot, err := epochBoundaryBlock.Block.HashTreeRoot()
	require.NoError(t, err, "Could not hash justified block")
	slot := primitives.Slot(40000)

	beaconState, err := util.NewBeaconStateCapella()
	require.NoError(t, err)
	require.NoError(t, beaconState.SetSlot(slot))
	err = beaconState.SetCurrentJustifiedCheckpoint(&zondpb.Checkpoint{
		Epoch: slots.ToEpoch(1500),
		Root:  justifiedBlockRoot[:],
	})
	require.NoError(t, err)
	blockRoots := beaconState.BlockRoots()
	blockRoots[1] = blockRoot[:]
	blockRoots[1*params.BeaconConfig().SlotsPerEpoch] = epochBoundaryRoot[:]
	blockRoots[2*params.BeaconConfig().SlotsPerEpoch] = justifiedBlockRoot[:]
	require.NoError(t, beaconState.SetBlockRoots(blockRoots))
	offset := int64(slot.Mul(params.BeaconConfig().SecondsPerSlot))
	attesterServer := &Server{
		SyncChecker:           &mockSync.Sync{IsSyncing: false},
		OptimisticModeFetcher: &mock.ChainService{Optimistic: false},
		TimeFetcher:           &mock.ChainService{Genesis: time.Now().Add(time.Duration(-1*offset) * time.Second)},
		CoreService: &core.Service{
			AttestationCache:   cache.NewAttestationCache(),
			HeadFetcher:        &mock.ChainService{State: beaconState, Root: blockRoot[:]},
			GenesisTimeFetcher: &mock.ChainService{Genesis: time.Now().Add(time.Duration(-1*offset) * time.Second)},
		},
	}

	req := &zondpb.AttestationDataRequest{
		CommitteeIndex: 0,
		Slot:           40000,
	}
	res, err := attesterServer.GetAttestationData(context.Background(), req)
	require.NoError(t, err, "Could not get attestation info at slot")

	expectedInfo := &zondpb.AttestationData{
		Slot:            req.Slot,
		BeaconBlockRoot: blockRoot[:],
		Source: &zondpb.Checkpoint{
			Epoch: slots.ToEpoch(1500),
			Root:  justifiedBlockRoot[:],
		},
		Target: &zondpb.Checkpoint{
			Epoch: 312,
			Root:  blockRoot[:],
		},
	}

	if !proto.Equal(res, expectedInfo) {
		t.Errorf("Expected attestation info to match, received %v, wanted %v", res, expectedInfo)
	}
}
