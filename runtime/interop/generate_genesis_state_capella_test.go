package interop

import (
	"context"
	"testing"

	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/container/trie"
	enginev1 "github.com/theQRL/qrysm/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
)

func TestGenerateGenesisStateCapella(t *testing.T) {
	ep := &enginev1.ExecutionPayloadCapella{
		ParentHash:    make([]byte, 32),
		FeeRecipient:  make([]byte, 20),
		StateRoot:     make([]byte, 32),
		ReceiptsRoot:  make([]byte, 32),
		LogsBloom:     make([]byte, 256),
		PrevRandao:    make([]byte, 32),
		BlockNumber:   0,
		GasLimit:      0,
		GasUsed:       0,
		Timestamp:     0,
		ExtraData:     make([]byte, 32),
		BaseFeePerGas: make([]byte, 32),
		BlockHash:     make([]byte, 32),
		Transactions:  make([][]byte, 0),
	}
	e1d := &zondpb.ExecutionNodeData{
		DepositRoot:  make([]byte, 32),
		DepositCount: 0,
		BlockHash:    make([]byte, 32),
	}
	g, _, err := GenerateGenesisStateCapella(context.Background(), 0, params.BeaconConfig().MinGenesisActiveValidatorCount, ep, e1d)
	require.NoError(t, err)

	tr, err := trie.NewTrie(params.BeaconConfig().DepositContractTreeDepth)
	require.NoError(t, err)
	dr, err := tr.HashTreeRoot()
	require.NoError(t, err)
	g.ExecutionNodeData.DepositRoot = dr[:]
	g.ExecutionNodeData.BlockHash = make([]byte, 32)
	st, err := state_native.InitializeFromProtoUnsafeCapella(g)
	require.NoError(t, err)
	_, err = st.MarshalSSZ()
	require.NoError(t, err)
}
