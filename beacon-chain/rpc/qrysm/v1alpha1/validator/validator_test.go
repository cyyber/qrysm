package validator

import (
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/theQRL/qrysm/config/params"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(io.Discard)
	// This package's test fixtures (slot numbers, proposer indices, vote
	// counts) were tuned against the minimal preset, so we keep that
	// runtime config here. The SSZ stubs in proto/qrysm/v1alpha1, however,
	// are generated for the mainnet preset (1024 Slashings, 16-byte
	// SyncCommitteeBits, 65536 EpochsPerHistoricalVector). Reconciling the
	// two requires running with `-tags minimal` after regenerating the
	// .ssz.go stubs under that build tag (see `bazel test //...` for the
	// blessed setup). Under plain `go test ./...` these tests are expected
	// to fail with `expected 1024 and 64 found`-style errors — track that
	// in a follow-up regen task instead of papering over the mismatch
	// here.
	prevConfig := params.BeaconConfig().Copy()
	defer params.OverrideBeaconConfig(prevConfig)
	params.OverrideBeaconConfig(params.MinimalSpecConfig())

	m.Run()
}
