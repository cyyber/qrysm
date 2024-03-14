package state_native_test

import (
	"testing"

	state_native "github.com/theQRL/qrysm/v4/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/v4/config/params"
	zond "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/require"
)

func BenchmarkAppendEth1DataVotes(b *testing.B) {
	st, err := state_native.InitializeFromProtoCapella(&zondpb.BeaconStateCapella{})
	require.NoError(b, err)

	max := params.BeaconConfig().Eth1DataVotesLength()

	if max < 2 {
		b.Fatalf("Eth1DataVotesLength is less than 2")
	}

	for i := uint64(0); i < max-2; i++ {
		err := st.AppendEth1DataVotes(&zondpb.Eth1Data{
			DepositCount: i,
			DepositRoot:  make([]byte, 64),
			BlockHash:    make([]byte, 64),
		})
		require.NoError(b, err)
	}

	ref := st.Copy()

	for i := 0; i < b.N; i++ {
		err := ref.AppendEth1DataVotes(&zond.Eth1Data{DepositCount: uint64(i)})
		require.NoError(b, err)
		ref = st.Copy()
	}
}
