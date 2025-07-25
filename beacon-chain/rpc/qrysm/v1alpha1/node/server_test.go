package node

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/theQRL/go-zond/common"
	"github.com/theQRL/go-zond/crypto"
	"github.com/theQRL/go-zond/p2p/qnode"
	mock "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	dbutil "github.com/theQRL/qrysm/beacon-chain/db/testing"
	"github.com/theQRL/qrysm/beacon-chain/p2p"
	mockP2p "github.com/theQRL/qrysm/beacon-chain/p2p/testing"
	"github.com/theQRL/qrysm/beacon-chain/rpc/testutil"
	mockSync "github.com/theQRL/qrysm/beacon-chain/sync/initial-sync/testing"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/runtime/version"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestNodeServer_GetSyncStatus(t *testing.T) {
	mSync := &mockSync.Sync{IsSyncing: false}
	ns := &Server{
		SyncChecker: mSync,
	}
	res, err := ns.GetSyncStatus(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, false, res.Syncing)
	ns.SyncChecker = &mockSync.Sync{IsSyncing: true}
	res, err = ns.GetSyncStatus(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, true, res.Syncing)
}

func TestNodeServer_GetGenesis(t *testing.T) {
	db := dbutil.SetupDB(t)
	ctx := context.Background()
	addr := common.Address{1, 2, 3}
	require.NoError(t, db.SaveDepositContractAddress(ctx, addr))
	st, err := util.NewBeaconStateCapella()
	require.NoError(t, err)
	genValRoot := bytesutil.ToBytes32([]byte("I am root"))
	ns := &Server{
		BeaconDB:           db,
		GenesisTimeFetcher: &mock.ChainService{},
		GenesisFetcher: &mock.ChainService{
			State:          st,
			ValidatorsRoot: genValRoot,
		},
	}
	res, err := ns.GetGenesis(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.DeepEqual(t, addr.Bytes(), res.DepositContractAddress)
	pUnix := timestamppb.New(time.Unix(0, 0))
	assert.Equal(t, res.GenesisTime.Seconds, pUnix.Seconds)
	assert.DeepEqual(t, genValRoot[:], res.GenesisValidatorsRoot)

	ns.GenesisTimeFetcher = &mock.ChainService{Genesis: time.Unix(10, 0)}
	res, err = ns.GetGenesis(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	pUnix = timestamppb.New(time.Unix(10, 0))
	assert.Equal(t, res.GenesisTime.Seconds, pUnix.Seconds)
}

func TestNodeServer_GetVersion(t *testing.T) {
	v := version.Version()
	ns := &Server{}
	res, err := ns.GetVersion(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, v, res.Version)
}

func TestNodeServer_GetImplementedServices(t *testing.T) {
	server := grpc.NewServer()
	ns := &Server{
		Server: server,
	}
	qrysmpb.RegisterNodeServer(server, ns)
	reflection.Register(server)

	res, err := ns.ListImplementedServices(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	// We verify the services include the node service + the registered reflection service.
	assert.Equal(t, 2, len(res.Services))
}

func TestNodeServer_GetHost(t *testing.T) {
	server := grpc.NewServer()
	peersProvider := &mockP2p.MockPeersProvider{}
	mP2P := mockP2p.NewTestP2P(t)
	key, err := crypto.GenerateKey()
	require.NoError(t, err)
	db, err := qnode.OpenDB("")
	require.NoError(t, err)
	lNode := qnode.NewLocalNode(db, key)
	record := lNode.Node().Record()
	stringQNR, err := p2p.SerializeQNR(record)
	require.NoError(t, err)
	ns := &Server{
		PeerManager:  &mockP2p.MockPeerManager{BHost: mP2P.BHost, Qnr: record, PID: mP2P.BHost.ID()},
		PeersFetcher: peersProvider,
	}
	qrysmpb.RegisterNodeServer(server, ns)
	reflection.Register(server)
	h, err := ns.GetHost(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, mP2P.PeerID().String(), h.PeerId)
	assert.Equal(t, stringQNR, h.Qnr)
}

func TestNodeServer_GetPeer(t *testing.T) {
	server := grpc.NewServer()
	peersProvider := &mockP2p.MockPeersProvider{}
	ns := &Server{
		PeersFetcher: peersProvider,
	}
	qrysmpb.RegisterNodeServer(server, ns)
	reflection.Register(server)
	firstPeer := peersProvider.Peers().All()[0]

	res, err := ns.GetPeer(context.Background(), &qrysmpb.PeerRequest{PeerId: firstPeer.String()})
	require.NoError(t, err)
	assert.Equal(t, firstPeer.String(), res.PeerId, "Unexpected peer ID")
	assert.Equal(t, int(qrysmpb.PeerDirection_INBOUND), int(res.Direction), "Expected 1st peer to be an inbound connection")
	assert.Equal(t, qrysmpb.ConnectionState_CONNECTED, res.ConnectionState, "Expected peer to be connected")
}

func TestNodeServer_ListPeers(t *testing.T) {
	server := grpc.NewServer()
	peersProvider := &mockP2p.MockPeersProvider{}
	ns := &Server{
		PeersFetcher: peersProvider,
	}
	qrysmpb.RegisterNodeServer(server, ns)
	reflection.Register(server)

	res, err := ns.ListPeers(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, 2, len(res.Peers))
	assert.Equal(t, int(qrysmpb.PeerDirection_INBOUND), int(res.Peers[0].Direction))
	assert.Equal(t, qrysmpb.PeerDirection_OUTBOUND, res.Peers[1].Direction)
}

func TestNodeServer_GetExecutionNodeConnectionStatus(t *testing.T) {
	server := grpc.NewServer()
	ep := "foo"
	err := errors.New("error1")
	errStr := "error1"
	mockFetcher := &testutil.MockExecutionChainInfoFetcher{
		CurrEndpoint: ep,
		CurrError:    err,
	}
	ns := &Server{
		ExecutionChainInfoFetcher: mockFetcher,
	}
	qrysmpb.RegisterNodeServer(server, ns)
	reflection.Register(server)

	res, err := ns.GetExecutionNodeConnectionStatus(context.Background(), &emptypb.Empty{})
	require.NoError(t, err)
	assert.Equal(t, ep, res.CurrentAddress)
	assert.Equal(t, errStr, res.CurrentConnectionError)
}
