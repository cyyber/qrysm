package beacon

import (
	"testing"

	"github.com/theQRL/qrysm/cmd/beacon-chain/flags"
	"github.com/theQRL/qrysm/config/params"
)

func TestMain(m *testing.M) {
	// SSZ encoding is always emitted for the mainnet preset (proto/qrysm
	// generated.ssz.go has hardcoded 16-byte SyncCommitteeBits, 1024 Slashings,
	// etc.) so the runtime config has to match — otherwise every state read
	// trips an "expected N got M" length mismatch. Pin mainnet here even
	// though it's slower than the minimal preset; switching back to minimal
	// requires regenerating the SSZ stubs with `-tags minimal`.
	prevConfig := params.BeaconConfig().Copy()
	defer params.OverrideBeaconConfig(prevConfig)
	params.OverrideBeaconConfig(params.MainnetConfig())

	resetFlags := flags.Get()
	flags.Init(&flags.GlobalFlags{
		MinimumSyncPeers: 30,
	})
	defer func() {
		flags.Init(resetFlags)
	}()

	m.Run()
}
