package helpers

import (
	"fmt"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/container/trie"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

// UpdateGenesisEth1Data updates eth1 data for genesis state.
func UpdateGenesisEth1Data(state state.BeaconState, deposits []*zondpb.Deposit, eth1Data *zondpb.Eth1Data) (state.BeaconState, error) {
	if eth1Data == nil {
		return nil, errors.New("no eth1data provided for genesis state")
	}

	leaves := make([][]byte, 0, len(deposits))
	for _, deposit := range deposits {
		if deposit == nil || deposit.Data == nil {
			return nil, fmt.Errorf("nil deposit or deposit with nil data cannot be processed: %v", deposit)
		}
		hash, err := deposit.Data.HashTreeRoot()
		if err != nil {
			return nil, err
		}
		leaves = append(leaves, hash[:])
	}
	var t *trie.SparseMerkleTrie
	var err error
	if len(leaves) > 0 {
		t, err = trie.GenerateTrieFromItems(leaves, params.BeaconConfig().DepositContractTreeDepth)
		if err != nil {
			return nil, err
		}
	} else {
		t, err = trie.NewTrie(params.BeaconConfig().DepositContractTreeDepth)
		if err != nil {
			return nil, err
		}
	}

	depositRoot, err := t.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	eth1Data.DepositRoot = depositRoot[:]
	err = state.SetEth1Data(eth1Data)
	if err != nil {
		return nil, err
	}
	return state, nil
}
