//go:build !minimal

package field_params_test

import (
	"testing"

	fieldparams "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/testing/require"
)

func TestFieldParametersValues(t *testing.T) {
	min, err := params.ByName(params.MainnetName)
	require.NoError(t, err)
	undo, err := params.SetActiveWithUndo(min)
	require.NoError(t, err)
	defer func() {
		require.NoError(t, undo())
	}()
	require.Equal(t, "mainnet", fieldparams.Preset)
	testFieldParametersMatchConfig(t)
}
