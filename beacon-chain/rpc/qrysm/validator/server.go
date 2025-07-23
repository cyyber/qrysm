package validator

import (
	"github.com/theQRL/qrysm/beacon-chain/blockchain"
	"github.com/theQRL/qrysm/beacon-chain/db"
	"github.com/theQRL/qrysm/beacon-chain/rpc/core"
	"github.com/theQRL/qrysm/beacon-chain/rpc/lookup"
	"github.com/theQRL/qrysm/beacon-chain/sync"
)

// Server defines a server implementation for HTTP endpoints, providing
// access data relevant to the QRL Beacon Chain.
type Server struct {
	GenesisTimeFetcher    blockchain.TimeFetcher
	SyncChecker           sync.Checker
	HeadFetcher           blockchain.HeadFetcher
	CoreService           *core.Service
	OptimisticModeFetcher blockchain.OptimisticModeFetcher
	Stater                lookup.Stater
	ChainInfoFetcher      blockchain.ChainInfoFetcher
	BeaconDB              db.ReadOnlyDatabase
	FinalizationFetcher   blockchain.FinalizationFetcher
}
