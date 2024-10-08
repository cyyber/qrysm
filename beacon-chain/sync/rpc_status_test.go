package sync

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	"github.com/theQRL/go-zond/p2p/enr"
	"github.com/theQRL/qrysm/async/abool"
	mock "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	"github.com/theQRL/qrysm/beacon-chain/db/kv"
	testingDB "github.com/theQRL/qrysm/beacon-chain/db/testing"
	"github.com/theQRL/qrysm/beacon-chain/p2p"
	"github.com/theQRL/qrysm/beacon-chain/p2p/peers"
	p2ptest "github.com/theQRL/qrysm/beacon-chain/p2p/testing"
	p2ptypes "github.com/theQRL/qrysm/beacon-chain/p2p/types"
	"github.com/theQRL/qrysm/beacon-chain/startup"
	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	mockSync "github.com/theQRL/qrysm/beacon-chain/sync/initial-sync/testing"
	"github.com/theQRL/qrysm/config/params"
	consensusblocks "github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/consensus-types/wrapper"
	leakybucket "github.com/theQRL/qrysm/container/leaky-bucket"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	qrysmTime "github.com/theQRL/qrysm/time"
	"google.golang.org/protobuf/proto"
)

func TestStatusRPCHandler_Disconnects_OnForkVersionMismatch(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	p1.Connect(p2)
	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to be connected")
	root := [32]byte{'C'}

	gt := time.Now()
	vr := [32]byte{'A'}
	r := &Service{
		cfg: &config{
			p2p: p1,
			chain: &mock.ChainService{
				Fork: &zondpb.Fork{
					PreviousVersion: params.BeaconConfig().GenesisForkVersion,
					CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
				},
				FinalizedCheckPoint: &zondpb.Checkpoint{
					Epoch: 0,
					Root:  root[:],
				},
				Genesis:        gt,
				ValidatorsRoot: vr,
				Root:           make([]byte, 32),
			},
			clock: startup.NewClock(gt, vr),
		},
		rateLimiter: newRateLimiter(p1),
	}
	pcl := protocol.ID(p2p.RPCStatusTopicV1)
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)

	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		expectSuccess(t, stream)
		out := &zondpb.Status{}
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		assert.DeepEqual(t, root[:], out.FinalizedRoot)
		assert.NoError(t, stream.Close())
	})

	pcl2 := protocol.ID("/eth2/beacon_chain/req/goodbye/1/ssz_snappy")
	topic = string(pcl2)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
	var wg2 sync.WaitGroup
	wg2.Add(1)
	p2.BHost.SetStreamHandler(pcl2, func(stream network.Stream) {
		defer wg2.Done()
		msg := new(primitives.SSZUint64)
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, msg))
		assert.Equal(t, p2ptypes.GoodbyeCodeWrongNetwork, *msg)
		assert.NoError(t, stream.Close())
	})

	stream1, err := p1.BHost.NewStream(context.Background(), p2.BHost.ID(), pcl)
	require.NoError(t, err)
	assert.NoError(t, r.statusRPCHandler(context.Background(), &zondpb.Status{ForkDigest: bytesutil.PadTo([]byte("f"), 4), HeadRoot: make([]byte, 32), FinalizedRoot: make([]byte, 32)}, stream1))

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}
	if util.WaitTimeout(&wg2, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}

	assert.Equal(t, 0, len(p1.BHost.Network().Peers()), "handler did not disconnect peer")
}

func TestStatusRPCHandler_ConnectsOnGenesis(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	p1.Connect(p2)
	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to be connected")
	var root [32]byte

	gt := time.Now()
	vr := [32]byte{'A'}
	r := &Service{
		cfg: &config{
			p2p: p1,
			chain: &mock.ChainService{
				Fork: &zondpb.Fork{
					PreviousVersion: params.BeaconConfig().GenesisForkVersion,
					CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
				},
				FinalizedCheckPoint: &zondpb.Checkpoint{
					Epoch: 0,
					Root:  params.BeaconConfig().ZeroHash[:],
				},
				Genesis:        gt,
				ValidatorsRoot: vr,
				Root:           make([]byte, 32),
			},
			clock: startup.NewClock(gt, vr),
		},
		rateLimiter: newRateLimiter(p1),
	}
	pcl := protocol.ID(p2p.RPCStatusTopicV1)
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)

	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		expectSuccess(t, stream)
		out := &zondpb.Status{}
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		assert.DeepEqual(t, root[:], out.FinalizedRoot)
	})

	stream1, err := p1.BHost.NewStream(context.Background(), p2.BHost.ID(), pcl)
	require.NoError(t, err)
	digest, err := r.currentForkDigest()
	require.NoError(t, err)

	err = r.statusRPCHandler(context.Background(), &zondpb.Status{ForkDigest: digest[:], FinalizedRoot: params.BeaconConfig().ZeroHash[:]}, stream1)
	assert.NoError(t, err)

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}

	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Handler disconnected with peer")
}

func TestStatusRPCHandler_ReturnsHelloMessage(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	p1.Connect(p2)
	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to be connected")
	db := testingDB.SetupDB(t)

	// Set up a head state with data we expect.
	head := util.NewBeaconBlockCapella()
	head.Block.Slot = 111
	headRoot, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	blkSlot := 3 * params.BeaconConfig().SlotsPerEpoch
	finalized := util.NewBeaconBlockCapella()
	finalized.Block.Slot = blkSlot
	finalizedRoot, err := finalized.Block.HashTreeRoot()
	require.NoError(t, err)
	genesisState, _ := util.DeterministicGenesisStateCapella(t, 1)
	require.NoError(t, genesisState.SetSlot(111))
	require.NoError(t, genesisState.UpdateBlockRootAtIndex(111%uint64(params.BeaconConfig().SlotsPerHistoricalRoot), headRoot))
	util.SaveBlock(t, context.Background(), db, finalized)
	require.NoError(t, db.SaveGenesisBlockRoot(context.Background(), finalizedRoot))
	finalizedCheckpt := &zondpb.Checkpoint{
		Epoch: 3,
		Root:  finalizedRoot[:],
	}
	totalSec := int64(params.BeaconConfig().SlotsPerEpoch.Mul(5 * params.BeaconConfig().SecondsPerSlot))
	genTime := time.Now().Unix() - totalSec

	gt := time.Unix(genTime, 0)
	vr := [32]byte{'A'}
	r := &Service{
		cfg: &config{
			p2p: p1,
			chain: &mock.ChainService{
				State:               genesisState,
				FinalizedCheckPoint: finalizedCheckpt,
				Root:                headRoot[:],
				Fork: &zondpb.Fork{
					PreviousVersion: params.BeaconConfig().GenesisForkVersion,
					CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
				},
				ValidatorsRoot: vr,
				Genesis:        gt,
				FinalizedRoots: map[[32]byte]bool{
					finalizedRoot: true,
				},
			},
			clock:    startup.NewClock(gt, vr),
			beaconDB: db,
		},
		rateLimiter: newRateLimiter(p1),
	}
	digest, err := r.currentForkDigest()
	require.NoError(t, err)

	// Setup streams
	pcl := protocol.ID(p2p.RPCStatusTopicV1)
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		expectSuccess(t, stream)
		out := &zondpb.Status{}
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		expected := &zondpb.Status{
			ForkDigest:     digest[:],
			HeadSlot:       genesisState.Slot(),
			HeadRoot:       headRoot[:],
			FinalizedEpoch: 3,
			FinalizedRoot:  finalizedRoot[:],
		}
		if !proto.Equal(out, expected) {
			t.Errorf("Did not receive expected message. Got %+v wanted %+v", out, expected)
		}
	})
	stream1, err := p1.BHost.NewStream(context.Background(), p2.BHost.ID(), pcl)
	require.NoError(t, err)

	err = r.statusRPCHandler(context.Background(), &zondpb.Status{
		ForkDigest:     digest[:],
		FinalizedRoot:  finalizedRoot[:],
		FinalizedEpoch: 3,
	}, stream1)
	assert.NoError(t, err)

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}
}

func TestHandshakeHandlers_Roundtrip(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Scenario is that p1 and p2 connect, exchange handshakes.
	// p2 disconnects and p1 should forget the handshake status.
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	db := testingDB.SetupDB(t)

	p1.LocalMetadata = wrapper.WrappedMetadataV0(&zondpb.MetaDataV0{
		SeqNumber: 2,
		Attnets:   bytesutil.PadTo([]byte{'A', 'B'}, 8),
	})

	p2.LocalMetadata = wrapper.WrappedMetadataV0(&zondpb.MetaDataV0{
		SeqNumber: 2,
		Attnets:   bytesutil.PadTo([]byte{'C', 'D'}, 8),
	})

	st, err := state_native.InitializeFromProtoCapella(&zondpb.BeaconStateCapella{
		Slot: 5,
	})
	require.NoError(t, err)
	blk := util.NewBeaconBlockCapella()
	blk.Block.Slot = 0
	util.SaveBlock(t, ctx, db, blk)
	finalizedRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)
	require.NoError(t, db.SaveGenesisBlockRoot(ctx, finalizedRoot))
	chain := &mock.ChainService{
		State:               st,
		FinalizedCheckPoint: &zondpb.Checkpoint{Epoch: 0, Root: finalizedRoot[:]},
		Fork: &zondpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{'A'},
		Root:           make([]byte, 32),
		FinalizedRoots: map[[32]byte]bool{
			finalizedRoot: true,
		},
	}
	cw := startup.NewClockSynchronizer()
	r := &Service{
		ctx: ctx,
		cfg: &config{
			p2p:      p1,
			chain:    chain,
			clock:    startup.NewClock(chain.Genesis, chain.ValidatorsRoot),
			beaconDB: db,
		},
		rateLimiter:  newRateLimiter(p1),
		clockWaiter:  cw,
		chainStarted: abool.New(),
	}
	p1.Digest, err = r.currentForkDigest()
	require.NoError(t, err)

	chain2 := &mock.ChainService{
		FinalizedCheckPoint: &zondpb.Checkpoint{Epoch: 0, Root: finalizedRoot[:]},
	}
	r2 := &Service{
		ctx: ctx,
		cfg: &config{
			chain: chain2,
			clock: startup.NewClock(chain2.Genesis, chain2.ValidatorsRoot),
			p2p:   p2,
		},
		rateLimiter: newRateLimiter(p2),
	}
	p2.Digest, err = r.currentForkDigest()
	require.NoError(t, err)

	go r.Start()

	// Setup streams
	pcl := protocol.ID("/eth2/beacon_chain/req/status/1/ssz_snappy")
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		out := &zondpb.Status{}
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		log.WithField("status", out).Warn("received status")
		resp := &zondpb.Status{HeadSlot: 100, HeadRoot: make([]byte, 32), ForkDigest: p2.Digest[:],
			FinalizedRoot: finalizedRoot[:], FinalizedEpoch: 0}
		_, err := stream.Write([]byte{responseCodeSuccess})
		assert.NoError(t, err)
		_, err = r.cfg.p2p.Encoding().EncodeWithMaxLength(stream, resp)
		assert.NoError(t, err)
		log.WithField("status", out).Warn("sending status")
		if err := stream.Close(); err != nil {
			t.Log(err)
		}
	})

	pcl = "/eth2/beacon_chain/req/ping/1/ssz_snappy"
	topic = string(pcl)
	r2.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
	var wg2 sync.WaitGroup
	wg2.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg2.Done()
		out := new(primitives.SSZUint64)
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		assert.Equal(t, uint64(2), uint64(*out))
		assert.NoError(t, r2.pingHandler(ctx, out, stream))
		assert.NoError(t, stream.Close())
	})

	numInactive1 := len(p1.Peers().Inactive())
	numActive1 := len(p1.Peers().Active())

	require.NoError(t, cw.SetClock(startup.NewClock(chain.Genesis, chain.ValidatorsRoot)))
	p1.Connect(p2)

	p1.Peers().Add(new(enr.Record), p2.BHost.ID(), p2.BHost.Addrs()[0], network.DirUnknown)
	p1.Peers().SetMetadata(p2.BHost.ID(), p2.LocalMetadata)

	p2.Peers().Add(new(enr.Record), p1.BHost.ID(), p1.BHost.Addrs()[0], network.DirUnknown)
	p2.Peers().SetMetadata(p1.BHost.ID(), p1.LocalMetadata)

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}
	if util.WaitTimeout(&wg2, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}

	// Wait for stream buffer to be read.
	time.Sleep(200 * time.Millisecond)

	numInactive2 := len(p1.Peers().Inactive())
	numActive2 := len(p1.Peers().Active())

	assert.Equal(t, numInactive1, numInactive1, "Number of inactive peers changed unexpectedly")
	assert.Equal(t, numActive1+1, numActive2, "Number of active peers unexpected")

	require.NoError(t, p2.Disconnect(p1.PeerID()))
	p1.Peers().SetConnectionState(p2.PeerID(), peers.PeerDisconnected)

	// Wait for disconnect event to trigger.
	time.Sleep(200 * time.Millisecond)

	numInactive3 := len(p1.Peers().Inactive())
	numActive3 := len(p1.Peers().Active())
	assert.Equal(t, numInactive2+1, numInactive3, "Number of inactive peers unexpected")
	assert.Equal(t, numActive2-1, numActive3, "Number of active peers unexpected")
}

func TestStatusRPCRequest_RequestSent(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)

	// Set up a head state with data we expect.
	head := util.NewBeaconBlockCapella()
	head.Block.Slot = 111
	headRoot, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	finalized := util.NewBeaconBlockCapella()
	finalized.Block.Slot = 40
	finalizedRoot, err := finalized.Block.HashTreeRoot()
	require.NoError(t, err)
	genesisState, _ := util.DeterministicGenesisStateCapella(t, 1)
	require.NoError(t, genesisState.SetSlot(111))
	require.NoError(t, genesisState.UpdateBlockRootAtIndex(111%uint64(params.BeaconConfig().SlotsPerHistoricalRoot), headRoot))
	finalizedCheckpt := &zondpb.Checkpoint{
		Epoch: 5,
		Root:  finalizedRoot[:],
	}

	chain := &mock.ChainService{
		State:               genesisState,
		FinalizedCheckPoint: finalizedCheckpt,
		Root:                headRoot[:],
		Fork: &zondpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{'A'},
	}
	r := &Service{
		cfg: &config{
			p2p:   p1,
			chain: chain,
			clock: startup.NewClock(chain.Genesis, chain.ValidatorsRoot),
		},
		ctx:         context.Background(),
		rateLimiter: newRateLimiter(p1),
	}

	// Setup streams
	pcl := protocol.ID("/eth2/beacon_chain/req/status/1/ssz_snappy")
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		out := &zondpb.Status{}
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		digest, err := r.currentForkDigest()
		assert.NoError(t, err)
		expected := &zondpb.Status{
			ForkDigest:     digest[:],
			HeadSlot:       genesisState.Slot(),
			HeadRoot:       headRoot[:],
			FinalizedEpoch: 5,
			FinalizedRoot:  finalizedRoot[:],
		}
		if !proto.Equal(out, expected) {
			t.Errorf("Did not receive expected message. Got %+v wanted %+v", out, expected)
		}
	})

	p1.AddConnectionHandler(r.sendRPCStatusRequest, nil)
	p1.Connect(p2)

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}

	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to continue being connected")
}

func TestStatusRPCRequest_FinalizedBlockExists(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	db := testingDB.SetupDB(t)

	// Set up a head state with data we expect.
	head := util.NewBeaconBlockCapella()
	head.Block.Slot = 111
	headRoot, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	blkSlot := 3 * params.BeaconConfig().SlotsPerEpoch
	finalized := util.NewBeaconBlockCapella()
	finalized.Block.Slot = blkSlot
	finalizedRoot, err := finalized.Block.HashTreeRoot()
	require.NoError(t, err)
	genesisState, _ := util.DeterministicGenesisStateCapella(t, 1)
	require.NoError(t, genesisState.SetSlot(111))
	require.NoError(t, genesisState.UpdateBlockRootAtIndex(111%uint64(params.BeaconConfig().SlotsPerHistoricalRoot), headRoot))
	blk := util.NewBeaconBlockCapella()
	blk.Block.Slot = blkSlot
	util.SaveBlock(t, context.Background(), db, blk)
	require.NoError(t, db.SaveGenesisBlockRoot(context.Background(), finalizedRoot))
	finalizedCheckpt := &zondpb.Checkpoint{
		Epoch: 3,
		Root:  finalizedRoot[:],
	}
	totalSec := int64(params.BeaconConfig().SlotsPerEpoch.Mul(5 * params.BeaconConfig().SecondsPerSlot))
	genTime := time.Now().Unix() - totalSec
	chain := &mock.ChainService{
		State:               genesisState,
		FinalizedCheckPoint: finalizedCheckpt,
		Root:                headRoot[:],
		Fork: &zondpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
		Genesis:        time.Unix(genTime, 0),
		ValidatorsRoot: [32]byte{'A'},
		FinalizedRoots: map[[32]byte]bool{
			finalizedRoot: true,
		},
	}
	r := &Service{
		cfg: &config{
			p2p:   p1,
			chain: chain,
			clock: startup.NewClock(chain.Genesis, chain.ValidatorsRoot),
		},
		ctx:         context.Background(),
		rateLimiter: newRateLimiter(p1),
	}
	chain2 := &mock.ChainService{
		State:               genesisState,
		FinalizedCheckPoint: finalizedCheckpt,
		Root:                headRoot[:],
		Fork: &zondpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
		Genesis:        time.Unix(genTime, 0),
		ValidatorsRoot: [32]byte{'A'},
		FinalizedRoots: map[[32]byte]bool{
			finalizedRoot: true,
		},
	}
	r2 := &Service{
		cfg: &config{
			p2p:      p1,
			chain:    chain2,
			clock:    startup.NewClock(chain2.Genesis, chain2.ValidatorsRoot),
			beaconDB: db,
		},
		ctx:         context.Background(),
		rateLimiter: newRateLimiter(p1),
	}

	// Setup streams
	pcl := protocol.ID("/eth2/beacon_chain/req/status/1/ssz_snappy")
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		out := &zondpb.Status{}
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		assert.NoError(t, r2.validateStatusMessage(context.Background(), out))
	})

	p1.AddConnectionHandler(r.sendRPCStatusRequest, nil)
	p1.Connect(p2)

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}

	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to continue being connected")
}

func TestStatusRPCRequest_FinalizedBlockSkippedSlots(t *testing.T) {
	db, err := kv.NewKVStore(context.Background(), t.TempDir())
	require.NoError(t, err)
	bState, _ := util.DeterministicGenesisStateCapella(t, 1)

	blk := util.NewBeaconBlockCapella()
	blk.Block.Slot = 0
	genRoot, err := blk.Block.HashTreeRoot()
	require.NoError(t, err)

	wsb, err := consensusblocks.NewSignedBeaconBlock(blk)
	require.NoError(t, err)
	require.NoError(t, db.SaveBlock(context.Background(), wsb))
	require.NoError(t, db.SaveGenesisBlockRoot(context.Background(), genRoot))
	blocksTillHead := makeBlocks(t, 1, 1000, genRoot)
	require.NoError(t, db.SaveBlocks(context.Background(), blocksTillHead))

	stateSummaries := make([]*zondpb.StateSummary, len(blocksTillHead))
	for i, b := range blocksTillHead {
		bRoot, err := b.Block().HashTreeRoot()
		require.NoError(t, err)
		stateSummaries[i] = &zondpb.StateSummary{
			Slot: b.Block().Slot(),
			Root: bRoot[:],
		}
	}
	require.NoError(t, db.SaveStateSummaries(context.Background(), stateSummaries))

	rootFetcher := func(slot primitives.Slot) [32]byte {
		rt, err := blocksTillHead[slot-1].Block().HashTreeRoot()
		require.NoError(t, err)
		return rt
	}
	tests := []struct {
		name                   string
		expectedFinalizedEpoch primitives.Epoch
		expectedFinalizedRoot  [32]byte
		headSlot               primitives.Slot
		remoteFinalizedEpoch   primitives.Epoch
		remoteFinalizedRoot    [32]byte
		remoteHeadSlot         primitives.Slot
		expectError            bool
	}{
		{
			name:                   "valid finalized epoch",
			expectedFinalizedEpoch: 3,
			expectedFinalizedRoot:  rootFetcher(3 * params.BeaconConfig().SlotsPerEpoch),
			headSlot:               111,
			remoteFinalizedEpoch:   3,
			remoteFinalizedRoot:    rootFetcher(3 * params.BeaconConfig().SlotsPerEpoch),
			remoteHeadSlot:         100,
			expectError:            false,
		},
		{
			name:                   "invalid finalized epoch",
			expectedFinalizedEpoch: 3,
			expectedFinalizedRoot:  rootFetcher(3 * params.BeaconConfig().SlotsPerEpoch),
			headSlot:               111,
			remoteFinalizedEpoch:   3,
			// give an incorrect root relative to the finalized epoch.
			remoteFinalizedRoot: rootFetcher(2 * params.BeaconConfig().SlotsPerEpoch),
			remoteHeadSlot:      120,
			expectError:         true,
		},
		{
			name:                   "invalid finalized root",
			expectedFinalizedEpoch: 3,
			expectedFinalizedRoot:  rootFetcher(3 * params.BeaconConfig().SlotsPerEpoch),
			headSlot:               111,
			remoteFinalizedEpoch:   3,
			// give a bad finalized root, and the beacon node verifies that
			// it is indeed incorrect.
			remoteFinalizedRoot: [32]byte{'a', 'b', 'c'},
			remoteHeadSlot:      120,
			expectError:         true,
		},
	}

	for _, tt := range tests {
		p1 := p2ptest.NewTestP2P(t)
		p2 := p2ptest.NewTestP2P(t)

		expectedFinalizedEpoch := tt.expectedFinalizedEpoch
		headSlot := tt.headSlot

		nState := bState.Copy()
		// Set up a head state with data we expect.
		head := blocksTillHead[len(blocksTillHead)-1]
		headRoot, err := head.Block().HashTreeRoot()
		require.NoError(t, err)

		rHead := blocksTillHead[tt.remoteHeadSlot-1]
		rHeadRoot, err := rHead.Block().HashTreeRoot()
		require.NoError(t, err)

		require.NoError(t, nState.SetSlot(headSlot))
		require.NoError(t, nState.UpdateBlockRootAtIndex(uint64(headSlot.ModSlot(params.BeaconConfig().SlotsPerHistoricalRoot)), headRoot))

		finalizedCheckpt := &zondpb.Checkpoint{
			Epoch: expectedFinalizedEpoch,
			Root:  tt.expectedFinalizedRoot[:],
		}

		remoteFinalizedChkpt := &zondpb.Checkpoint{
			Epoch: tt.remoteFinalizedEpoch,
			Root:  tt.remoteFinalizedRoot[:],
		}
		require.NoError(t, db.SaveFinalizedCheckpoint(context.Background(), finalizedCheckpt))

		epoch := expectedFinalizedEpoch.Add(2)
		totalSec := uint64(params.BeaconConfig().SlotsPerEpoch.Mul(uint64(epoch) * params.BeaconConfig().SecondsPerSlot))
		gt := time.Unix(time.Now().Unix()-int64(totalSec), 0)
		vr := [32]byte{'A'}
		chain := &mock.ChainService{
			State:               nState,
			FinalizedCheckPoint: remoteFinalizedChkpt,
			Root:                rHeadRoot[:],
			Fork: &zondpb.Fork{
				PreviousVersion: params.BeaconConfig().GenesisForkVersion,
				CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			},
			Genesis:        gt,
			ValidatorsRoot: vr,
			FinalizedRoots: map[[32]byte]bool{
				tt.expectedFinalizedRoot: true,
				tt.remoteFinalizedRoot:   true,
			},
		}
		r := &Service{
			cfg: &config{
				p2p:   p1,
				chain: chain,
				clock: startup.NewClock(chain.Genesis, chain.ValidatorsRoot),
			},
			ctx:         context.Background(),
			rateLimiter: newRateLimiter(p1),
		}
		chain2 := &mock.ChainService{
			State:               nState,
			FinalizedCheckPoint: finalizedCheckpt,
			Root:                headRoot[:],
			Fork: &zondpb.Fork{
				PreviousVersion: params.BeaconConfig().GenesisForkVersion,
				CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
			},
			Genesis:        gt,
			ValidatorsRoot: vr,
			FinalizedRoots: map[[32]byte]bool{
				tt.expectedFinalizedRoot: true,
				tt.remoteFinalizedRoot:   true,
			},
		}
		r2 := &Service{
			cfg: &config{
				p2p:      p2,
				chain:    chain2,
				clock:    startup.NewClock(chain2.Genesis, chain2.ValidatorsRoot),
				beaconDB: db,
			},

			ctx:         context.Background(),
			rateLimiter: newRateLimiter(p1),
		}

		// Setup streams
		pcl := protocol.ID("/eth2/beacon_chain/req/status/1/ssz_snappy")
		topic := string(pcl)
		r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
		var wg sync.WaitGroup
		wg.Add(1)
		p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
			defer wg.Done()
			out := &zondpb.Status{}
			assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
			assert.Equal(t, tt.expectError, r2.validateStatusMessage(context.Background(), out) != nil)
		})

		p1.AddConnectionHandler(r.sendRPCStatusRequest, nil)
		p1.Connect(p2)

		if util.WaitTimeout(&wg, 1*time.Second) {
			t.Fatal("Did not receive stream within 1 sec")
		}

		assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to continue being connected")
		assert.NoError(t, p1.Disconnect(p2.PeerID()))
	}
	assert.NoError(t, db.Close())
}

func TestStatusRPCRequest_BadPeerHandshake(t *testing.T) {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)

	// Set up a head state with data we expect.
	head := util.NewBeaconBlockCapella()
	head.Block.Slot = 111
	headRoot, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	finalized := util.NewBeaconBlockCapella()
	finalizedRoot, err := finalized.Block.HashTreeRoot()
	require.NoError(t, err)
	genesisState, _ := util.DeterministicGenesisStateCapella(t, 1)
	require.NoError(t, genesisState.SetSlot(111))
	require.NoError(t, genesisState.UpdateBlockRootAtIndex(111%uint64(params.BeaconConfig().SlotsPerHistoricalRoot), headRoot))
	finalizedCheckpt := &zondpb.Checkpoint{
		Epoch: 5,
		Root:  finalizedRoot[:],
	}
	chain := &mock.ChainService{
		State:               genesisState,
		FinalizedCheckPoint: finalizedCheckpt,
		Root:                headRoot[:],
		Fork: &zondpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{'A'},
	}
	cw := startup.NewClockSynchronizer()

	r := &Service{
		cfg: &config{
			p2p:   p1,
			chain: chain,
		},

		ctx:          ctx,
		rateLimiter:  newRateLimiter(p1),
		clockWaiter:  cw,
		chainStarted: abool.New(),
	}

	go r.Start()

	// Setup streams
	pcl := protocol.ID("/eth2/beacon_chain/req/status/1/ssz_snappy")
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)
	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		out := &zondpb.Status{}
		assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, out))
		expected := &zondpb.Status{
			ForkDigest:     []byte{1, 1, 1, 1},
			HeadSlot:       genesisState.Slot(),
			HeadRoot:       headRoot[:],
			FinalizedEpoch: 5,
			FinalizedRoot:  finalizedRoot[:],
		}
		if _, err := stream.Write([]byte{responseCodeSuccess}); err != nil {
			log.WithError(err).Debug("Could not write to stream")
		}
		_, err := r.cfg.p2p.Encoding().EncodeWithMaxLength(stream, expected)
		assert.NoError(t, err)
	})

	require.NoError(t, cw.SetClock(startup.NewClock(chain.Genesis, chain.ValidatorsRoot)))

	assert.Equal(t, false, p1.Peers().Scorers().IsBadPeer(p2.PeerID()), "Peer is marked as bad")
	p1.Connect(p2)

	if util.WaitTimeout(&wg, time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}
	time.Sleep(100 * time.Millisecond)

	connectionState, err := p1.Peers().ConnectionState(p2.PeerID())
	require.NoError(t, err, "Could not obtain peer connection state")
	assert.Equal(t, peers.PeerDisconnected, connectionState, "Expected peer to be disconnected")

	assert.Equal(t, true, p1.Peers().Scorers().IsBadPeer(p2.PeerID()), "Peer is not marked as bad")
}

func TestStatusRPC_ValidGenesisMessage(t *testing.T) {
	// Set up a head state with data we expect.
	head := util.NewBeaconBlockCapella()
	head.Block.Slot = 111
	headRoot, err := head.Block.HashTreeRoot()
	require.NoError(t, err)
	blkSlot := 3 * params.BeaconConfig().SlotsPerEpoch
	finalized := util.NewBeaconBlockCapella()
	finalized.Block.Slot = blkSlot
	finalizedRoot, err := finalized.Block.HashTreeRoot()
	require.NoError(t, err)
	genesisState, _ := util.DeterministicGenesisStateCapella(t, 1)
	require.NoError(t, genesisState.SetSlot(111))
	require.NoError(t, genesisState.UpdateBlockRootAtIndex(111%uint64(params.BeaconConfig().SlotsPerHistoricalRoot), headRoot))
	finalizedCheckpt := &zondpb.Checkpoint{
		Epoch: 5,
		Root:  finalizedRoot[:],
	}
	chain := &mock.ChainService{
		State:               genesisState,
		FinalizedCheckPoint: finalizedCheckpt,
		Root:                headRoot[:],
		Fork: &zondpb.Fork{
			PreviousVersion: params.BeaconConfig().GenesisForkVersion,
			CurrentVersion:  params.BeaconConfig().GenesisForkVersion,
		},
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{'A'},
	}
	r := &Service{
		cfg: &config{
			chain: chain,
			clock: startup.NewClock(chain.Genesis, chain.ValidatorsRoot),
		},
		ctx: context.Background(),
	}
	digest, err := r.currentForkDigest()
	require.NoError(t, err)
	// There should be no error for a status message
	// with a genesis checkpoint.
	err = r.validateStatusMessage(r.ctx, &zondpb.Status{
		ForkDigest:     digest[:],
		FinalizedRoot:  params.BeaconConfig().ZeroHash[:],
		FinalizedEpoch: 0,
		HeadRoot:       headRoot[:],
		HeadSlot:       111,
	})
	require.NoError(t, err)
}

func TestShouldResync(t *testing.T) {
	type args struct {
		genesis  time.Time
		syncing  bool
		headSlot primitives.Slot
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "genesis epoch should not resync when syncing is true",
			args: args{
				headSlot: 127,
				genesis:  qrysmTime.Now(),
				syncing:  true,
			},
			want: false,
		},
		{
			name: "genesis epoch should not resync when syncing is false",
			args: args{
				headSlot: 127,
				genesis:  qrysmTime.Now(),
				syncing:  false,
			},
			want: false,
		},
		{
			name: "two epochs behind, resync ok",
			args: args{
				headSlot: 127,
				genesis:  qrysmTime.Now().Add(-1 * 384 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
				syncing:  false,
			},
			want: true,
		},
		{
			name: "two epochs behind, already syncing",
			args: args{
				headSlot: 31,
				genesis:  qrysmTime.Now().Add(-1 * 384 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
				syncing:  true,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		headState, _ := util.DeterministicGenesisStateCapella(t, 1)
		require.NoError(t, headState.SetSlot(tt.args.headSlot))
		chain := &mock.ChainService{
			State:   headState,
			Genesis: tt.args.genesis,
		}
		r := &Service{
			cfg: &config{
				chain:       chain,
				clock:       startup.NewClock(chain.Genesis, chain.ValidatorsRoot),
				initialSync: &mockSync.Sync{IsSyncing: tt.args.syncing},
			},
			ctx: context.Background(),
		}
		t.Run(tt.name, func(t *testing.T) {
			if got := r.shouldReSync(); got != tt.want {
				t.Errorf("shouldReSync() = %v, want %v", got, tt.want)
			}
		})
	}
}

func makeBlocks(t *testing.T, i, n uint64, previousRoot [32]byte) []interfaces.ReadOnlySignedBeaconBlock {
	blocks := make([]*zondpb.SignedBeaconBlockCapella, n)
	ifaceBlocks := make([]interfaces.ReadOnlySignedBeaconBlock, n)
	for j := i; j < n+i; j++ {
		parentRoot := make([]byte, 32)
		copy(parentRoot, previousRoot[:])
		blocks[j-i] = util.NewBeaconBlockCapella()
		blocks[j-i].Block.Slot = primitives.Slot(j + 1)
		blocks[j-i].Block.ParentRoot = parentRoot
		var err error
		previousRoot, err = blocks[j-i].Block.HashTreeRoot()
		require.NoError(t, err)
		ifaceBlocks[j-i], err = consensusblocks.NewSignedBeaconBlock(blocks[j-i])
		require.NoError(t, err)
	}
	return ifaceBlocks
}
