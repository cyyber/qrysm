package finality

import (
	"testing"

	"github.com/cyyber/qrysm/v4/testing/spectest/shared/capella/finality"
)

func TestMainnet_Capella_Finality(t *testing.T) {
	finality.RunFinalityTest(t, "mainnet")
}
