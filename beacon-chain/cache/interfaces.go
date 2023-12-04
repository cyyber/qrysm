package cache

import (
	"context"
	"math/big"

	"github.com/theQRL/go-zond/common"
	zondpb "github.com/theQRL/qrysm/v4/proto/prysm/v1alpha1"
)

// DepositCache combines the interfaces for retrieving and inserting deposit information.
type DepositCache interface {
	DepositFetcher
	DepositInserter
}

// DepositFetcher defines a struct which can retrieve deposit information from a store.
type DepositFetcher interface {
	AllDeposits(ctx context.Context, untilBlk *big.Int) []*zondpb.Deposit
	AllDepositContainers(ctx context.Context) []*zondpb.DepositContainer
	DepositByPubkey(ctx context.Context, pubKey []byte) (*zondpb.Deposit, *big.Int)
	DepositsNumberAndRootAtHeight(ctx context.Context, blockHeight *big.Int) (uint64, [32]byte)
	InsertPendingDeposit(ctx context.Context, d *zondpb.Deposit, blockNum uint64, index int64, depositRoot [32]byte)
	PendingDeposits(ctx context.Context, untilBlk *big.Int) []*zondpb.Deposit
	PendingContainers(ctx context.Context, untilBlk *big.Int) []*zondpb.DepositContainer
	PrunePendingDeposits(ctx context.Context, merkleTreeIndex int64)
	PruneProofs(ctx context.Context, untilDepositIndex int64) error
	FinalizedFetcher
}

// DepositInserter defines a struct which can insert deposit information from a store.
type DepositInserter interface {
	InsertDeposit(ctx context.Context, d *zondpb.Deposit, blockNum uint64, index int64, depositRoot [32]byte) error
	InsertDepositContainers(ctx context.Context, ctrs []*zondpb.DepositContainer)
	InsertFinalizedDeposits(ctx context.Context, eth1DepositIndex int64, executionHash common.Hash, executionNumber uint64) error
}

// FinalizedFetcher is a smaller interface defined to be the bare minimum to satisfy “Service”.
// It extends the "DepositFetcher" interface with additional methods for fetching finalized deposits.
type FinalizedFetcher interface {
	FinalizedDeposits(ctx context.Context) (FinalizedDeposits, error)
	NonFinalizedDeposits(ctx context.Context, lastFinalizedIndex int64, untilBlk *big.Int) []*zondpb.Deposit
}

// FinalizedDeposits defines a method to access a merkle tree containing deposits and their indexes.
type FinalizedDeposits interface {
	Deposits() MerkleTree
	MerkleTrieIndex() int64
}

// MerkleTree defines methods for constructing and manipulating a merkle tree.
type MerkleTree interface {
	HashTreeRoot() ([32]byte, error)
	NumOfItems() int
	Insert(item []byte, index int) error
	MerkleProof(index int) ([][]byte, error)
}
