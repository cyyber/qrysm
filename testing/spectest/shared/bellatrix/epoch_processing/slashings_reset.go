package epoch_processing

import (
	"path"
	"testing"

	"github.com/cyyber/qrysm/v4/beacon-chain/core/epoch"
	"github.com/cyyber/qrysm/v4/beacon-chain/state"
	"github.com/cyyber/qrysm/v4/testing/require"
	"github.com/cyyber/qrysm/v4/testing/spectest/utils"
)

// RunSlashingsResetTests executes "epoch_processing/slashings_reset" tests.
func RunSlashingsResetTests(t *testing.T, config string) {
	require.NoError(t, utils.SetConfig(t, config))

	testFolders, testsFolderPath := utils.TestFolders(t, config, "bellatrix", "epoch_processing/slashings_reset/pyspec_tests")
	if len(testFolders) == 0 {
		t.Fatalf("No test folders found for %s/%s/%s", config, "bellatrix", "epoch_processing/slashings_reset/pyspec_tests")
	}
	for _, folder := range testFolders {
		t.Run(folder.Name(), func(t *testing.T) {
			folderPath := path.Join(testsFolderPath, folder.Name())
			RunEpochOperationTest(t, folderPath, processSlashingsResetWrapper)
		})
	}
}

func processSlashingsResetWrapper(t *testing.T, st state.BeaconState) (state.BeaconState, error) {
	st, err := epoch.ProcessSlashingsReset(st)
	require.NoError(t, err, "Could not process final updates")
	return st, nil
}