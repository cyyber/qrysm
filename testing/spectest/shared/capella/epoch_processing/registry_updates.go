package epoch_processing

import (
	"context"
	"path"
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/core/epoch"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/spectest/utils"
)

// RunRegistryUpdatesTests executes "epoch_processing/registry_updates" tests.
func RunRegistryUpdatesTests(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testFolders, testsFolderPath := utils.TestFolders(t, config, "capella", "epoch_processing/registry_updates/pyspec_tests")
	if len(testFolders) == 0 {
		t.Fatalf("No test folders found for %s/%s/%s", config, "capella", "epoch_processing/registry_updates/pyspec_tests")
	}
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			// Important to clear cache for every test or else the old value of active validator count gets reused.
			helpers.ClearCache()
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processRegistryUpdatesWrapper)
		})
	}
}

func processRegistryUpdatesWrapper(_ *testing.T, state state.BeaconState) (state.BeaconState, error) {
	return epoch.ProcessRegistryUpdates(context.Background(), state)
}
