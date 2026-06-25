package genesis_test

import (
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/state/genesis"
	fieldparams "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/testing/require"
)

func TestGenesisState(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: params.MainnetName,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			st, err := genesis.State(tt.name)
			if err != nil {
				t.Fatal(err)
			}
			if st == nil {
				t.Fatal("nil state")
			}
			if st.NumValidators() <= 0 {
				t.Error("No validators present in state")
			}
			header, err := st.LatestExecutionPayloadHeader()
			require.NoError(t, err)
			require.Equal(t, fieldparams.FeeRecipientLength, len(header.FeeRecipient()))
			for i, validator := range st.Validators() {
				require.Equal(t, fieldparams.WithdrawalCredentialsLength, len(validator.WithdrawalCredentials), "validator %d withdrawal credentials", i)
			}
		})
	}
}
