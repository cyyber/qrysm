package operations

import (
	"testing"

	"github.com/cyyber/qrysm/v4/testing/spectest/shared/altair/operations"
)

func TestMinimal_Altair_Operations_Deposit(t *testing.T) {
	operations.RunDepositTest(t, "minimal")
}