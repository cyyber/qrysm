package ssz_static

import (
	"testing"

	"github.com/theQRL/qrysm/v4/testing/spectest/shared/phase0/ssz_static"
)

func TestMainnet_Phase0_SSZStatic(t *testing.T) {
	ssz_static.RunSSZStaticTests(t, "mainnet")
}