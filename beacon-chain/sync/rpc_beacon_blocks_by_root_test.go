package sync

import (
	"context"
	"math/big"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/protocol"
	gcache "github.com/patrickmn/go-cache"
	"github.com/theQRL/go-zond/common"
	gzondTypes "github.com/theQRL/go-zond/core/types"
	mock "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	db "github.com/theQRL/qrysm/beacon-chain/db/testing"
	mockExecution "github.com/theQRL/qrysm/beacon-chain/execution/testing"
	"github.com/theQRL/qrysm/beacon-chain/p2p"
	p2ptest "github.com/theQRL/qrysm/beacon-chain/p2p/testing"
	p2pTypes "github.com/theQRL/qrysm/beacon-chain/p2p/types"
	"github.com/theQRL/qrysm/beacon-chain/startup"
	fieldparams "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	leakybucket "github.com/theQRL/qrysm/container/leaky-bucket"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	"github.com/theQRL/qrysm/network/forks"
	enginev1 "github.com/theQRL/qrysm/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
)

func TestRecentBeaconBlocksRPCHandler_ReturnsBlocks(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	p1.Connect(p2)
	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to be connected")
	d := db.SetupDB(t)

	// Start service with 160 as allowed blocks capacity (and almost zero capacity recovery).
	parent := bytesutil.PadTo([]byte("parentHash"), fieldparams.RootLength)
	stateRoot := bytesutil.PadTo([]byte("stateRoot"), fieldparams.RootLength)
	receiptsRoot := bytesutil.PadTo([]byte("receiptsRoot"), fieldparams.RootLength)
	logsBloom := bytesutil.PadTo([]byte("logs"), fieldparams.LogsBloomLength)
	to, err := common.NewAddressFromString("Z095e7baea6a6c7c4c2dfeb977efac326af552d87")
	require.NoError(t, err)
	tx := gzondTypes.NewTx(&gzondTypes.DynamicFeeTx{
		Nonce:     0,
		To:        &to,
		Value:     big.NewInt(0),
		Gas:       0,
		GasFeeCap: big.NewInt(0),
		GasTipCap: big.NewInt(0),
		Data:      nil,
	})
	txs := []*gzondTypes.Transaction{tx}
	encodedBinaryTxs := make([][]byte, 1)
	encodedBinaryTxs[0], err = txs[0].MarshalBinary()
	require.NoError(t, err)
	blockHash := bytesutil.ToBytes32([]byte("foo"))
	payload := &enginev1.ExecutionPayloadCapella{
		ParentHash:    parent,
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     stateRoot,
		ReceiptsRoot:  receiptsRoot,
		LogsBloom:     logsBloom,
		PrevRandao:    blockHash[:],
		BlockNumber:   0,
		GasLimit:      0,
		GasUsed:       0,
		Timestamp:     0,
		ExtraData:     make([]byte, 0),
		BlockHash:     blockHash[:],
		BaseFeePerGas: bytesutil.PadTo([]byte("baseFeePerGas"), fieldparams.RootLength),
		Transactions:  encodedBinaryTxs,
	}

	var blkRoots p2pTypes.BeaconBlockByRootsReq
	// Populate the database with blocks that would match the request.
	for i := primitives.Slot(1); i < 11; i++ {
		blk := util.NewBeaconBlockCapella()
		blk.Block.Slot = i
		blk.Block.Body.ExecutionPayload = payload
		root, err := blk.Block.HashTreeRoot()
		require.NoError(t, err)
		util.SaveBlock(t, context.Background(), d, blk)
		blkRoots = append(blkRoots, root)
	}

	mockEngine := &mockExecution.EngineClient{
		ExecutionPayloadByBlockHash: map[[32]byte]*enginev1.ExecutionPayloadCapella{
			blockHash: payload,
		},
	}

	r := &Service{cfg: &config{p2p: p1, beaconDB: d, clock: startup.NewClock(time.Unix(0, 0), [32]byte{})}, rateLimiter: newRateLimiter(p1)}
	r.cfg.chain = &mock.ChainService{ValidatorsRoot: [32]byte{}}
	r.cfg.executionPayloadReconstructor = mockEngine
	pcl := protocol.ID(p2p.RPCBlocksByRootTopicV2)
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(10000, 10000, time.Second, false)

	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		for i := range blkRoots {
			expectSuccess(t, stream)
			_, err := readContextFromStream(stream)
			assert.NoError(t, err)
			res := util.NewBeaconBlockCapella()
			assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, res))
			if uint64(res.Block.Slot) != uint64(i+1) {
				t.Errorf("Received unexpected block slot %d but wanted %d", res.Block.Slot, i+1)
			}
		}
	})

	stream1, err := p1.BHost.NewStream(context.Background(), p2.BHost.ID(), pcl)
	require.NoError(t, err)
	err = r.beaconBlocksRootRPCHandler(context.Background(), &blkRoots, stream1)
	assert.NoError(t, err)

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}
}

func TestRecentBeaconBlocksRPCHandler_ReturnsBlocks_ReconstructsPayload(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	p1.Connect(p2)
	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to be connected")
	d := db.SetupDB(t)

	// Start service with 160 as allowed blocks capacity (and almost zero capacity recovery).
	parent := bytesutil.PadTo([]byte("parentHash"), fieldparams.RootLength)
	stateRoot := bytesutil.PadTo([]byte("stateRoot"), fieldparams.RootLength)
	receiptsRoot := bytesutil.PadTo([]byte("receiptsRoot"), fieldparams.RootLength)
	logsBloom := bytesutil.PadTo([]byte("logs"), fieldparams.LogsBloomLength)
	to, err := common.NewAddressFromString("Z095e7baea6a6c7c4c2dfeb977efac326af552d87")
	require.NoError(t, err)
	tx := gzondTypes.NewTx(&gzondTypes.DynamicFeeTx{
		Nonce:     0,
		To:        &to,
		Value:     big.NewInt(0),
		Gas:       0,
		GasFeeCap: big.NewInt(0),
		GasTipCap: big.NewInt(0),
		Data:      nil,
	})
	txs := []*gzondTypes.Transaction{tx}
	encodedBinaryTxs := make([][]byte, 1)
	encodedBinaryTxs[0], err = txs[0].MarshalBinary()
	require.NoError(t, err)
	blockHash := bytesutil.ToBytes32([]byte("foo"))
	payload := &enginev1.ExecutionPayloadCapella{
		ParentHash:    parent,
		FeeRecipient:  make([]byte, fieldparams.FeeRecipientLength),
		StateRoot:     stateRoot,
		ReceiptsRoot:  receiptsRoot,
		LogsBloom:     logsBloom,
		PrevRandao:    blockHash[:],
		BlockNumber:   0,
		GasLimit:      0,
		GasUsed:       0,
		Timestamp:     0,
		ExtraData:     make([]byte, 0),
		BlockHash:     blockHash[:],
		BaseFeePerGas: bytesutil.PadTo([]byte("baseFeePerGas"), fieldparams.RootLength),
		Transactions:  encodedBinaryTxs,
	}
	wrappedPayload, err := blocks.WrappedExecutionPayloadCapella(payload, 0)
	require.NoError(t, err)
	header, err := blocks.PayloadToHeaderCapella(wrappedPayload)
	require.NoError(t, err)

	var blkRoots p2pTypes.BeaconBlockByRootsReq
	// Populate the database with blocks that would match the request.
	for i := primitives.Slot(1); i < 11; i++ {
		blk := util.NewBlindedBeaconBlockCapella()
		blk.Block.Body.ExecutionPayloadHeader = header
		blk.Block.Slot = i
		root, err := blk.Block.HashTreeRoot()
		require.NoError(t, err)
		wsb, err := blocks.NewSignedBeaconBlock(blk)
		require.NoError(t, err)
		require.NoError(t, d.SaveBlock(context.Background(), wsb))
		blkRoots = append(blkRoots, root)
	}

	mockEngine := &mockExecution.EngineClient{
		ExecutionPayloadByBlockHash: map[[32]byte]*enginev1.ExecutionPayloadCapella{
			blockHash: payload,
		},
	}
	r := &Service{cfg: &config{
		p2p:                           p1,
		beaconDB:                      d,
		executionPayloadReconstructor: mockEngine,
		chain:                         &mock.ChainService{ValidatorsRoot: [32]byte{}},
		clock:                         startup.NewClock(time.Unix(0, 0), [32]byte{}),
	}, rateLimiter: newRateLimiter(p1)}
	pcl := protocol.ID(p2p.RPCBlocksByRootTopicV2)
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(10000, 10000, time.Second, false)

	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		for i := range blkRoots {
			expectSuccess(t, stream)
			_, err := readContextFromStream(stream)
			assert.NoError(t, err)
			res := util.NewBeaconBlockCapella()
			assert.NoError(t, r.cfg.p2p.Encoding().DecodeWithMaxLength(stream, res))
			if uint64(res.Block.Slot) != uint64(i+1) {
				t.Errorf("Received unexpected block slot %d but wanted %d", res.Block.Slot, i+1)
			}
		}
		require.Equal(t, uint64(10), mockEngine.NumReconstructedPayloads)
	})

	stream1, err := p1.BHost.NewStream(context.Background(), p2.BHost.ID(), pcl)
	require.NoError(t, err)
	err = r.beaconBlocksRootRPCHandler(context.Background(), &blkRoots, stream1)
	assert.NoError(t, err)

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}
}

func TestRecentBeaconBlocks_RPCRequestSent(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	p1.DelaySend = true

	blockA := util.NewBeaconBlockCapella()
	blockA.Block.Slot = 111
	blockB := util.NewBeaconBlockCapella()
	blockB.Block.Slot = 40
	// Set up a head state with data we expect.
	blockARoot, err := blockA.Block.HashTreeRoot()
	require.NoError(t, err)
	blockBRoot, err := blockB.Block.HashTreeRoot()
	require.NoError(t, err)
	genesisState, _ := util.DeterministicGenesisStateCapella(t, 1)
	require.NoError(t, genesisState.SetSlot(111))
	require.NoError(t, genesisState.UpdateBlockRootAtIndex(111%uint64(params.BeaconConfig().SlotsPerHistoricalRoot), blockARoot))
	finalizedCheckpt := &zondpb.Checkpoint{
		Epoch: 5,
		Root:  blockBRoot[:],
	}

	expectedRoots := p2pTypes.BeaconBlockByRootsReq{blockBRoot, blockARoot}

	chain := &mock.ChainService{
		State:               genesisState,
		FinalizedCheckPoint: finalizedCheckpt,
		Root:                blockARoot[:],
		Genesis:             time.Now(),
		ValidatorsRoot:      [32]byte{},
	}
	r := &Service{
		cfg: &config{
			p2p:   p1,
			chain: chain,
			clock: startup.NewClock(chain.Genesis, chain.ValidatorsRoot),
		},
		slotToPendingBlocks: gcache.New(time.Second, 2*time.Second),
		seenPendingBlocks:   make(map[[32]byte]bool),
		ctx:                 context.Background(),
		rateLimiter:         newRateLimiter(p1),
	}

	// Setup streams
	pcl := protocol.ID("/eth2/beacon_chain/req/beacon_blocks_by_root/2/ssz_snappy")
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(10000, 10000, time.Second, false)

	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		out := new(p2pTypes.BeaconBlockByRootsReq)
		assert.NoError(t, p2.Encoding().DecodeWithMaxLength(stream, out))
		assert.DeepEqual(t, &expectedRoots, out, "Did not receive expected message")
		response := []*zondpb.SignedBeaconBlockCapella{blockB, blockA}
		for _, blk := range response {
			_, err := stream.Write([]byte{responseCodeSuccess})
			assert.NoError(t, err, "Could not write to stream")
			vRoot := r.cfg.clock.GenesisValidatorsRoot()
			digest, err := forks.ForkDigestFromEpoch(params.BeaconConfig().GenesisEpoch, vRoot[:])
			assert.NoError(t, err)
			assert.NoError(t, writeContextToStream(digest[:], stream))
			_, err = p2.Encoding().EncodeWithMaxLength(stream, blk)
			assert.NoError(t, err, "Could not send response back")
		}
		assert.NoError(t, stream.Close())
	})

	p1.Connect(p2)
	require.NoError(t, r.sendRecentBeaconBlocksRequest(context.Background(), &expectedRoots, p2.PeerID()))

	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}
}

func TestRecentBeaconBlocksRPCHandler_HandleZeroBlocks(t *testing.T) {
	p1 := p2ptest.NewTestP2P(t)
	p2 := p2ptest.NewTestP2P(t)
	p1.Connect(p2)
	assert.Equal(t, 1, len(p1.BHost.Network().Peers()), "Expected peers to be connected")
	d := db.SetupDB(t)

	r := &Service{cfg: &config{p2p: p1, beaconDB: d}, rateLimiter: newRateLimiter(p1)}
	pcl := protocol.ID(p2p.RPCBlocksByRootTopicV2)
	topic := string(pcl)
	r.rateLimiter.limiterMap[topic] = leakybucket.NewCollector(1, 1, time.Second, false)

	var wg sync.WaitGroup
	wg.Add(1)
	p2.BHost.SetStreamHandler(pcl, func(stream network.Stream) {
		defer wg.Done()
		expectFailure(t, 1, "no block roots provided in request", stream)
	})

	stream1, err := p1.BHost.NewStream(context.Background(), p2.BHost.ID(), pcl)
	require.NoError(t, err)
	err = r.beaconBlocksRootRPCHandler(context.Background(), &p2pTypes.BeaconBlockByRootsReq{}, stream1)
	assert.ErrorContains(t, "no block roots provided", err)
	if util.WaitTimeout(&wg, 1*time.Second) {
		t.Fatal("Did not receive stream within 1 sec")
	}

	r.rateLimiter.RLock() // retrieveCollector requires a lock to be held.
	defer r.rateLimiter.RUnlock()
	lter, err := r.rateLimiter.retrieveCollector(topic)
	require.NoError(t, err)
	assert.Equal(t, 1, int(lter.Count(stream1.Conn().RemotePeer().String())))
}
