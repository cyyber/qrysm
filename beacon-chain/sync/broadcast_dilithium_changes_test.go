package sync

import (
	"context"
	"testing"
	"time"

	logTest "github.com/sirupsen/logrus/hooks/test"
	mockChain "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	testingdb "github.com/theQRL/qrysm/beacon-chain/db/testing"
	doublylinkedtree "github.com/theQRL/qrysm/beacon-chain/forkchoice/doubly-linked-tree"
	"github.com/theQRL/qrysm/beacon-chain/operations/dilithiumtoexec"
	mockp2p "github.com/theQRL/qrysm/beacon-chain/p2p/testing"
	"github.com/theQRL/qrysm/beacon-chain/state/stategen"
	mockSync "github.com/theQRL/qrysm/beacon-chain/sync/initial-sync/testing"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
)

func TestBroadcastDilithiumChanges(t *testing.T) {
	params.SetupTestConfigCleanup(t)

	chainService := &mockChain.ChainService{
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{'A'},
	}
	s := NewService(context.Background(),
		WithP2P(mockp2p.NewTestP2P(t)),
		WithInitialSync(&mockSync.Sync{IsSyncing: false}),
		WithChainService(chainService),
		WithOperationNotifier(chainService.OperationNotifier()),
		WithDilithiumToExecPool(dilithiumtoexec.NewPool()),
	)
	var emptySig [96]byte
	s.cfg.dilithiumToExecPool.InsertDilithiumToExecChange(&qrysmpb.SignedDilithiumToExecutionChange{
		Message: &qrysmpb.DilithiumToExecutionChange{
			ValidatorIndex:      10,
			FromDilithiumPubkey: make([]byte, 48),
			ToExecutionAddress:  make([]byte, 20),
		},
		Signature: emptySig[:],
	})

	capellaStart := primitives.Slot(0)
	s.broadcastDilithiumChanges(capellaStart + 1)
}

func TestRateDilithiumChanges(t *testing.T) {
	logHook := logTest.NewGlobal()
	params.SetupTestConfigCleanup(t)

	chainService := &mockChain.ChainService{
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{'A'},
	}
	p1 := mockp2p.NewTestP2P(t)
	s := NewService(context.Background(),
		WithP2P(p1),
		WithInitialSync(&mockSync.Sync{IsSyncing: false}),
		WithChainService(chainService),
		WithOperationNotifier(chainService.OperationNotifier()),
		WithDilithiumToExecPool(dilithiumtoexec.NewPool()),
	)
	beaconDB := testingdb.SetupDB(t)
	s.cfg.stateGen = stategen.New(beaconDB, doublylinkedtree.New())
	s.cfg.beaconDB = beaconDB
	s.initCaches()
	st, keys := util.DeterministicGenesisStateCapella(t, 256)
	s.cfg.chain = &mockChain.ChainService{
		ValidatorsRoot: [32]byte{'A'},
		Genesis:        time.Now().Add(-time.Second * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Duration(10)),
		State:          st,
	}

	for i := 0; i < 200; i++ {
		message := &qrysmpb.DilithiumToExecutionChange{
			ValidatorIndex:      primitives.ValidatorIndex(i),
			FromDilithiumPubkey: keys[i+1].PublicKey().Marshal(),
			ToExecutionAddress:  bytesutil.PadTo([]byte("address"), 20),
		}
		epoch := primitives.Epoch(1)
		domain, err := signing.Domain(st.Fork(), epoch, params.BeaconConfig().DomainDilithiumToExecutionChange, st.GenesisValidatorsRoot())
		assert.NoError(t, err)
		htr, err := signing.SigningData(message.HashTreeRoot, domain)
		assert.NoError(t, err)
		signed := &qrysmpb.SignedDilithiumToExecutionChange{
			Message:   message,
			Signature: keys[i+1].Sign(htr[:]).Marshal(),
		}

		s.cfg.dilithiumToExecPool.InsertDilithiumToExecChange(signed)
	}

	require.Equal(t, false, p1.BroadcastCalled)
	slot := primitives.Slot(0)
	s.broadcastDilithiumChanges(slot)
	time.Sleep(100 * time.Millisecond) // Need a sleep for the go routine to be ready
	require.Equal(t, true, p1.BroadcastCalled)
	require.LogsDoNotContain(t, logHook, "could not")

	p1.BroadcastCalled = false
	time.Sleep(500 * time.Millisecond) // Need a sleep for the second batch to be broadcast
	require.Equal(t, true, p1.BroadcastCalled)
	require.LogsDoNotContain(t, logHook, "could not")
}

func TestBroadcastDilithiumBatch_changes_slice(t *testing.T) {
	message := &qrysmpb.DilithiumToExecutionChange{
		FromDilithiumPubkey: make([]byte, 48),
		ToExecutionAddress:  make([]byte, 20),
	}
	signed := &qrysmpb.SignedDilithiumToExecutionChange{
		Message:   message,
		Signature: make([]byte, 96),
	}
	changes := make([]*qrysmpb.SignedDilithiumToExecutionChange, 200)
	for i := 0; i < len(changes); i++ {
		changes[i] = signed
	}
	p1 := mockp2p.NewTestP2P(t)
	chainService := &mockChain.ChainService{
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{'A'},
	}
	s := NewService(context.Background(),
		WithP2P(p1),
		WithInitialSync(&mockSync.Sync{IsSyncing: false}),
		WithChainService(chainService),
		WithOperationNotifier(chainService.OperationNotifier()),
		WithDilithiumToExecPool(dilithiumtoexec.NewPool()),
	)
	beaconDB := testingdb.SetupDB(t)
	s.cfg.stateGen = stategen.New(beaconDB, doublylinkedtree.New())
	s.cfg.beaconDB = beaconDB
	s.initCaches()
	st, _ := util.DeterministicGenesisStateCapella(t, 32)
	s.cfg.chain = &mockChain.ChainService{
		ValidatorsRoot: [32]byte{'A'},
		Genesis:        time.Now().Add(-time.Second * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Duration(10)),
		State:          st,
	}

	s.broadcastDilithiumBatch(s.ctx, &changes)
	require.Equal(t, 200-128, len(changes))
}
