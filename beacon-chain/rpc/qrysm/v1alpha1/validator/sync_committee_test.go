package validator

import (
	"context"
	"testing"
	"time"

	mock "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	"github.com/theQRL/qrysm/beacon-chain/core/feed"
	opfeed "github.com/theQRL/qrysm/beacon-chain/core/feed/operation"
	"github.com/theQRL/qrysm/beacon-chain/core/transition"
	"github.com/theQRL/qrysm/beacon-chain/operations/synccommittee"
	mockp2p "github.com/theQRL/qrysm/beacon-chain/p2p/testing"
	"github.com/theQRL/qrysm/beacon-chain/rpc/core"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/dilithium"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestGetSyncMessageBlockRoot_OK(t *testing.T) {
	r := []byte{'a'}
	server := &Server{
		HeadFetcher:           &mock.ChainService{Root: r},
		TimeFetcher:           &mock.ChainService{Genesis: time.Now()},
		OptimisticModeFetcher: &mock.ChainService{},
	}
	res, err := server.GetSyncMessageBlockRoot(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	require.DeepEqual(t, r, res.Root)
}

func TestGetSyncMessageBlockRoot_Optimistic(t *testing.T) {
	params.SetupTestConfigCleanup(t)

	server := &Server{
		HeadFetcher:           &mock.ChainService{},
		TimeFetcher:           &mock.ChainService{Genesis: time.Now()},
		OptimisticModeFetcher: &mock.ChainService{Optimistic: true},
	}
	_, err := server.GetSyncMessageBlockRoot(context.Background(), &emptypb.Empty{})
	s, ok := status.FromError(err)
	require.Equal(t, true, ok)
	require.DeepEqual(t, codes.Unavailable, s.Code())
	require.ErrorContains(t, errOptimisticMode.Error(), err)

	server = &Server{
		HeadFetcher:           &mock.ChainService{},
		TimeFetcher:           &mock.ChainService{Genesis: time.Now()},
		OptimisticModeFetcher: &mock.ChainService{Optimistic: false},
	}
	_, err = server.GetSyncMessageBlockRoot(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
}

func TestSubmitSyncMessage_OK(t *testing.T) {
	st, _ := util.DeterministicGenesisStateCapella(t, 10)
	server := &Server{
		CoreService: &core.Service{
			SyncCommitteePool: synccommittee.NewStore(),
			P2P:               &mockp2p.MockBroadcaster{},
			HeadFetcher: &mock.ChainService{
				State: st,
			},
		},
	}
	msg := &qrysmpb.SyncCommitteeMessage{
		Slot:           1,
		ValidatorIndex: 2,
	}
	_, err := server.SubmitSyncMessage(context.Background(), msg)
	require.NoError(t, err)
	savedMsgs, err := server.CoreService.SyncCommitteePool.SyncCommitteeMessages(1)
	require.NoError(t, err)
	require.DeepEqual(t, []*qrysmpb.SyncCommitteeMessage{msg}, savedMsgs)
}

func TestGetSyncSubcommitteeIndex_Ok(t *testing.T) {
	transition.SkipSlotCache.Disable()
	defer transition.SkipSlotCache.Enable()

	server := &Server{
		HeadFetcher: &mock.ChainService{
			SyncCommitteeIndices: []primitives.CommitteeIndex{0},
		},
	}
	var pubKey [field_params.DilithiumPubkeyLength]byte
	// Request slot 0, should get the index 0 for validator 0.
	res, err := server.GetSyncSubcommitteeIndex(context.Background(), &qrysmpb.SyncSubcommitteeIndexRequest{
		PublicKey: pubKey[:], Slot: primitives.Slot(0),
	})
	require.NoError(t, err)
	require.DeepEqual(t, []primitives.CommitteeIndex{0}, res.Indices)
}

func TestGetSyncCommitteeContribution_FiltersDuplicates(t *testing.T) {
	st, _ := util.DeterministicGenesisStateCapella(t, 10)
	syncCommitteePool := synccommittee.NewStore()
	headFetcher := &mock.ChainService{
		State:                st,
		SyncCommitteeIndices: []primitives.CommitteeIndex{10},
	}
	server := &Server{
		CoreService: &core.Service{
			SyncCommitteePool: syncCommitteePool,
			HeadFetcher:       headFetcher,
			P2P:               &mockp2p.MockBroadcaster{},
		},
		SyncCommitteePool:     syncCommitteePool,
		HeadFetcher:           headFetcher,
		P2P:                   &mockp2p.MockBroadcaster{},
		TimeFetcher:           &mock.ChainService{Genesis: time.Now()},
		OptimisticModeFetcher: &mock.ChainService{},
	}
	secKey, err := dilithium.RandKey()
	require.NoError(t, err)
	sig := secKey.Sign([]byte{'A'}).Marshal()
	msg := &qrysmpb.SyncCommitteeMessage{
		Slot:           1,
		ValidatorIndex: 2,
		BlockRoot:      make([]byte, 32),
		Signature:      sig,
	}
	_, err = server.SubmitSyncMessage(context.Background(), msg)
	require.NoError(t, err)
	_, err = server.SubmitSyncMessage(context.Background(), msg)
	require.NoError(t, err)
	val, err := st.ValidatorAtIndex(2)
	require.NoError(t, err)

	contr, err := server.GetSyncCommitteeContribution(context.Background(),
		&qrysmpb.SyncCommitteeContributionRequest{
			Slot:      1,
			PublicKey: val.PublicKey,
			SubnetId:  0})
	require.NoError(t, err)
	assert.DeepEqual(t, sig, contr.Signatures[0])
}

func TestSubmitSignedContributionAndProof_OK(t *testing.T) {
	server := &Server{
		CoreService: &core.Service{
			SyncCommitteePool: synccommittee.NewStore(),
			Broadcaster:       &mockp2p.MockBroadcaster{},
			OperationNotifier: (&mock.ChainService{}).OperationNotifier(),
		},
	}
	contribution := &qrysmpb.SignedContributionAndProof{
		Message: &qrysmpb.ContributionAndProof{
			Contribution: &qrysmpb.SyncCommitteeContribution{
				Slot:              1,
				SubcommitteeIndex: 2,
				Signatures:        [][]byte{},
			},
		},
	}
	_, err := server.SubmitSignedContributionAndProof(context.Background(), contribution)
	require.NoError(t, err)
	savedMsgs, err := server.CoreService.SyncCommitteePool.SyncCommitteeContributions(1)
	require.NoError(t, err)
	require.DeepEqual(t, []*qrysmpb.SyncCommitteeContribution{contribution.Message.Contribution}, savedMsgs)
}

func TestSubmitSignedContributionAndProof_Notification(t *testing.T) {
	server := &Server{
		CoreService: &core.Service{
			SyncCommitteePool: synccommittee.NewStore(),
			Broadcaster:       &mockp2p.MockBroadcaster{},
			OperationNotifier: (&mock.ChainService{}).OperationNotifier(),
		},
	}

	// Subscribe to operation notifications.
	opChannel := make(chan *feed.Event, 1024)
	opSub := server.CoreService.OperationNotifier.OperationFeed().Subscribe(opChannel)
	defer opSub.Unsubscribe()

	contribution := &qrysmpb.SignedContributionAndProof{
		Message: &qrysmpb.ContributionAndProof{
			Contribution: &qrysmpb.SyncCommitteeContribution{
				Slot:              1,
				SubcommitteeIndex: 2,
			},
		},
	}
	_, err := server.SubmitSignedContributionAndProof(context.Background(), contribution)
	require.NoError(t, err)

	// Ensure the state notification was broadcast.
	notificationFound := false
	for !notificationFound {
		select {
		case event := <-opChannel:
			if event.Type == opfeed.SyncCommitteeContributionReceived {
				notificationFound = true
				data, ok := event.Data.(*opfeed.SyncCommitteeContributionReceivedData)
				assert.Equal(t, true, ok, "Entity is of the wrong type")
				assert.NotNil(t, data.Contribution)
			}
		case <-opSub.Err():
			t.Error("Subscription to state notifier failed")
			return
		}
	}
}
