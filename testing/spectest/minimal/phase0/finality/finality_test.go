package finality

import (
	"testing"

	"github.com/cyyber/qrysm/v4/testing/spectest/shared/phase0/finality"
)

func TestMinimal_Phase0_Finality(t *testing.T) {
	finality.RunFinalityTest(t, "minimal")
}