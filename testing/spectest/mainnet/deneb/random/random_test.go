package random

import (
	"testing"

	"github.com/theQRL/qrysm/v4/testing/spectest/shared/deneb/sanity"
)

func TestMainnet_Deneb_Random(t *testing.T) {
	sanity.RunBlockProcessingTest(t, "mainnet", "random/random/pyspec_tests")
}