package initialsync

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/sirupsen/logrus"
	"github.com/theQRL/go-zond/p2p/qnr"
	mock "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	"github.com/theQRL/qrysm/beacon-chain/db"
	dbtest "github.com/theQRL/qrysm/beacon-chain/db/testing"
	"github.com/theQRL/qrysm/beacon-chain/p2p/peers"
	p2pt "github.com/theQRL/qrysm/beacon-chain/p2p/testing"
	p2pTypes "github.com/theQRL/qrysm/beacon-chain/p2p/types"
	"github.com/theQRL/qrysm/beacon-chain/startup"
	beaconsync "github.com/theQRL/qrysm/beacon-chain/sync"
	"github.com/theQRL/qrysm/cmd/beacon-chain/flags"
	"github.com/theQRL/qrysm/config/features"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/container/slice"
	"github.com/theQRL/qrysm/crypto/hash"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	qrysmTime "github.com/theQRL/qrysm/time"
	"github.com/theQRL/qrysm/time/slots"
)

type testCache struct {
	sync.RWMutex
	rootCache       map[primitives.Slot][32]byte
	parentSlotCache map[primitives.Slot]primitives.Slot
}

var cache = &testCache{}

type peerData struct {
	blocks         []primitives.Slot // slots that peer has blocks
	finalizedEpoch primitives.Epoch
	headSlot       primitives.Slot
	failureSlots   []primitives.Slot // slots at which the peer will return an error
	forkedPeer     bool
}

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(io.Discard)

	resetCfg := features.InitWithReset(&features.Flags{
		EnablePeerScorer: true,
	})
	defer resetCfg()

	resetFlags := flags.Get()
	flags.Init(&flags.GlobalFlags{
		BlockBatchLimit:            64,
		BlockBatchLimitBurstFactor: 10,
	})
	defer func() {
		flags.Init(resetFlags)
	}()

	m.Run()
}

func initializeTestServices(t *testing.T, slots []primitives.Slot, peers []*peerData) (*mock.ChainService, *p2pt.TestP2P, db.Database) {
	cache.initializeRootCache(slots, t)
	beaconDB := dbtest.SetupDB(t)

	p := p2pt.NewTestP2P(t)
	connectPeers(t, p, peers, p.Peers())
	cache.RLock()
	genesisRoot := cache.rootCache[0]
	cache.RUnlock()

	util.SaveBlock(t, context.Background(), beaconDB, util.NewBeaconBlockCapella())

	st, err := util.NewBeaconStateCapella()
	require.NoError(t, err)

	return &mock.ChainService{
		State: st,
		Root:  genesisRoot[:],
		DB:    beaconDB,
		FinalizedCheckPoint: &zondpb.Checkpoint{
			Epoch: 0,
		},
		Genesis:        time.Now(),
		ValidatorsRoot: [32]byte{},
	}, p, beaconDB
}

// makeGenesisTime where now is the current slot.
func makeGenesisTime(currentSlot primitives.Slot) time.Time {
	return qrysmTime.Now().Add(-1 * time.Second * time.Duration(currentSlot) * time.Duration(params.BeaconConfig().SecondsPerSlot))
}

// sanity test on helper function
func TestMakeGenesisTime(t *testing.T) {
	currentSlot := primitives.Slot(64)
	gt := makeGenesisTime(currentSlot)
	require.Equal(t, currentSlot, slots.Since(gt))
}

// helper function for sequences of block slots
func makeSequence(start, end primitives.Slot) []primitives.Slot {
	if end < start {
		panic("cannot make sequence where end is before start")
	}
	seq := make([]primitives.Slot, 0, end-start+1)
	for i := start; i <= end; i++ {
		seq = append(seq, i)
	}
	return seq
}

func (c *testCache) initializeRootCache(reqSlots []primitives.Slot, t *testing.T) {
	c.Lock()
	defer c.Unlock()

	c.rootCache = make(map[primitives.Slot][32]byte)
	c.parentSlotCache = make(map[primitives.Slot]primitives.Slot)
	parentSlot := primitives.Slot(0)

	genesisBlock := util.NewBeaconBlockCapella().Block
	genesisRoot, err := genesisBlock.HashTreeRoot()
	require.NoError(t, err)
	c.rootCache[0] = genesisRoot
	parentRoot := genesisRoot
	for _, slot := range reqSlots {
		currentBlock := util.NewBeaconBlockCapella().Block
		currentBlock.Slot = slot
		currentBlock.ParentRoot = parentRoot[:]
		parentRoot, err = currentBlock.HashTreeRoot()
		require.NoError(t, err)
		c.rootCache[slot] = parentRoot
		c.parentSlotCache[slot] = parentSlot
		parentSlot = slot
	}
}

// sanity test on helper function
func TestMakeSequence(t *testing.T) {
	got := makeSequence(3, 5)
	want := []primitives.Slot{3, 4, 5}
	require.DeepEqual(t, want, got)
}

// Connect peers with local host. This method sets up peer statuses and the appropriate handlers
// for each test peer.
func connectPeers(t *testing.T, host *p2pt.TestP2P, data []*peerData, peerStatus *peers.Status) {
	for _, d := range data {
		connectPeer(t, host, d, peerStatus)
	}
}

// connectPeer connects a peer to a local host.
func connectPeer(t *testing.T, host *p2pt.TestP2P, datum *peerData, peerStatus *peers.Status) peer.ID {
	const topic = "/eth2/beacon_chain/req/beacon_blocks_by_range/2/ssz_snappy"
	p := p2pt.NewTestP2P(t)
	p.SetStreamHandler(topic, func(stream network.Stream) {
		defer func() {
			assert.NoError(t, stream.Close())
		}()

		req := &zondpb.BeaconBlocksByRangeRequest{}
		assert.NoError(t, p.Encoding().DecodeWithMaxLength(stream, req))

		requestedBlocks := makeSequence(req.StartSlot, req.StartSlot.Add((req.Count-1)*req.Step))

		// Expected failure range
		if len(slice.IntersectionSlot(datum.failureSlots, requestedBlocks)) > 0 {
			_, err := stream.Write([]byte{0x01})
			assert.NoError(t, err)
			msg := p2pTypes.ErrorMessage("bad")
			_, err = p.Encoding().EncodeWithMaxLength(stream, &msg)
			assert.NoError(t, err)
			return
		}

		// Determine the correct subset of blocks to return as dictated by the test scenario.
		ss := slice.IntersectionSlot(datum.blocks, requestedBlocks)

		ret := make([]*zondpb.SignedBeaconBlockCapella, 0)
		for _, slot := range ss {
			if (slot - req.StartSlot).Mod(req.Step) != 0 {
				continue
			}
			cache.RLock()
			parentRoot := cache.rootCache[cache.parentSlotCache[slot]]
			cache.RUnlock()
			blk := util.NewBeaconBlockCapella()
			blk.Block.Slot = slot
			blk.Block.ParentRoot = parentRoot[:]
			// If forked peer, give a different parent root.
			if datum.forkedPeer {
				newRoot := hash.Hash(parentRoot[:])
				blk.Block.ParentRoot = newRoot[:]
			}
			ret = append(ret, blk)
			currRoot, err := blk.Block.HashTreeRoot()
			require.NoError(t, err)
			logrus.Tracef("block with slot %d , signing root %#x and parent root %#x", slot, currRoot, parentRoot)
		}

		if uint64(len(ret)) > req.Count {
			ret = ret[:req.Count]
		}

		for i := 0; i < len(ret); i++ {
			wsb, err := blocks.NewSignedBeaconBlock(ret[i])
			require.NoError(t, err)
			assert.NoError(t, beaconsync.WriteBlockChunk(stream, startup.NewClock(time.Now(), [32]byte{}), p.Encoding(), wsb))
		}
	})

	p.Connect(host)

	peerStatus.Add(new(qnr.Record), p.PeerID(), nil, network.DirOutbound)
	peerStatus.SetConnectionState(p.PeerID(), peers.PeerConnected)
	peerStatus.SetChainState(p.PeerID(), &zondpb.Status{
		ForkDigest:     params.BeaconConfig().GenesisForkVersion,
		FinalizedRoot:  []byte(fmt.Sprintf("finalized_root %d", datum.finalizedEpoch)),
		FinalizedEpoch: datum.finalizedEpoch,
		HeadRoot:       bytesutil.PadTo([]byte("head_root"), 32),
		HeadSlot:       datum.headSlot,
	})

	return p.PeerID()
}

// extendBlockSequence extends block chain sequentially (creating genesis block, if necessary).
func extendBlockSequence(t *testing.T, inSeq []*zondpb.SignedBeaconBlockCapella, size int) []*zondpb.SignedBeaconBlockCapella {
	// Start from the original sequence.
	outSeq := make([]*zondpb.SignedBeaconBlockCapella, len(inSeq)+size)
	copy(outSeq, inSeq)

	// See if genesis block needs to be created.
	startSlot := len(inSeq)
	if len(inSeq) == 0 {
		outSeq[0] = util.NewBeaconBlockCapella()
		outSeq[0].Block.StateRoot = util.Random32Bytes(t)
		startSlot++
		outSeq = append(outSeq, nil)
	}

	// Extend block chain sequentially.
	for slot := startSlot; slot < len(outSeq); slot++ {
		outSeq[slot] = util.NewBeaconBlockCapella()
		outSeq[slot].Block.Slot = primitives.Slot(slot)
		parentRoot, err := outSeq[slot-1].Block.HashTreeRoot()
		require.NoError(t, err)
		outSeq[slot].Block.ParentRoot = parentRoot[:]
		// Make sure that blocks having the same slot number, produce different hashes.
		// That way different branches/forks will have different blocks for the same slots.
		outSeq[slot].Block.StateRoot = util.Random32Bytes(t)
	}

	return outSeq
}

// connectPeerHavingBlocks connect host with a peer having provided blocks.
func connectPeerHavingBlocks(
	t *testing.T, host *p2pt.TestP2P, blks []*zondpb.SignedBeaconBlockCapella, finalizedSlot primitives.Slot,
	peerStatus *peers.Status,
) peer.ID {
	p := p2pt.NewTestP2P(t)

	p.SetStreamHandler("/eth2/beacon_chain/req/beacon_blocks_by_range/2/ssz_snappy", func(stream network.Stream) {
		defer func() {
			_err := stream.Close()
			_ = _err
		}()

		req := &zondpb.BeaconBlocksByRangeRequest{}
		assert.NoError(t, p.Encoding().DecodeWithMaxLength(stream, req))

		for i := req.StartSlot; i < req.StartSlot.Add(req.Count*req.Step); i += primitives.Slot(req.Step) {
			if uint64(i) >= uint64(len(blks)) {
				break
			}
			wsb, err := blocks.NewSignedBeaconBlock(blks[i])
			require.NoError(t, err)
			require.NoError(t, beaconsync.WriteBlockChunk(stream, startup.NewClock(time.Now(), [32]byte{}), p.Encoding(), wsb))
		}
	})

	p.SetStreamHandler("/eth2/beacon_chain/req/beacon_blocks_by_root/2/ssz_snappy", func(stream network.Stream) {
		defer func() {
			_err := stream.Close()
			_ = _err
		}()

		req := new(p2pTypes.BeaconBlockByRootsReq)
		assert.NoError(t, p.Encoding().DecodeWithMaxLength(stream, req))
		if len(*req) == 0 {
			return
		}
		for _, expectedRoot := range *req {
			for _, blk := range blks {
				if root, err := blk.Block.HashTreeRoot(); err == nil && expectedRoot == root {
					log.Printf("Found blocks_by_root: %#x for slot: %v", root, blk.Block.Slot)
					wsb, err := blocks.NewSignedBeaconBlock(blk)
					require.NoError(t, err)
					require.NoError(t, beaconsync.WriteBlockChunk(stream, startup.NewClock(time.Now(), [32]byte{}), p.Encoding(), wsb))
				}
			}
		}
	})

	p.Connect(host)

	finalizedEpoch := slots.ToEpoch(finalizedSlot)
	headRoot, err := blks[len(blks)-1].Block.HashTreeRoot()
	require.NoError(t, err)

	peerStatus.Add(new(qnr.Record), p.PeerID(), nil, network.DirOutbound)
	peerStatus.SetConnectionState(p.PeerID(), peers.PeerConnected)
	peerStatus.SetChainState(p.PeerID(), &zondpb.Status{
		ForkDigest:     params.BeaconConfig().GenesisForkVersion,
		FinalizedRoot:  []byte(fmt.Sprintf("finalized_root %d", finalizedEpoch)),
		FinalizedEpoch: finalizedEpoch,
		HeadRoot:       headRoot[:],
		HeadSlot:       blks[len(blks)-1].Block.Slot,
	})

	return p.PeerID()
}
