// Package interop contains deterministic utilities for generating
// genesis states and keys.
package interop

import (
	"context"

	"github.com/pkg/errors"
	coreState "github.com/theQRL/qrysm/beacon-chain/core/transition"
	statenative "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/container/trie"
	enginev1 "github.com/theQRL/qrysm/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/time"
)

// GenerateGenesisStateCapella deterministically given a genesis time and number of validators.
// If a genesis time of 0 is supplied it is set to the current time.
func GenerateGenesisStateCapella(ctx context.Context, genesisTime, numValidators uint64, ep *enginev1.ExecutionPayloadCapella, ed *zondpb.ExecutionNodeData) (*zondpb.BeaconStateCapella, []*zondpb.Deposit, error) {
	privKeys, pubKeys, err := DeterministicallyGenerateKeys(0 /*startIndex*/, numValidators)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "could not deterministically generate keys for %d validators", numValidators)
	}
	depositDataItems, depositDataRoots, err := DepositDataFromKeys(privKeys, pubKeys)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate deposit data from keys")
	}
	return GenerateGenesisStateCapellaFromDepositData(ctx, genesisTime, depositDataItems, depositDataRoots, ep, ed)
}

// GenerateGenesisStateCapellaFromDepositData creates a genesis state given a list of
// deposit data items and their corresponding roots.
func GenerateGenesisStateCapellaFromDepositData(
	ctx context.Context, genesisTime uint64, depositData []*zondpb.Deposit_Data, depositDataRoots [][]byte, ep *enginev1.ExecutionPayloadCapella, e1d *zondpb.ExecutionNodeData,
) (*zondpb.BeaconStateCapella, []*zondpb.Deposit, error) {
	t, err := trie.GenerateTrieFromItems(depositDataRoots, params.BeaconConfig().DepositContractTreeDepth)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate Merkle trie for deposit proofs")
	}
	deposits, err := GenerateDepositsFromData(depositData, t)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate deposits from the deposit data provided")
	}
	if genesisTime == 0 {
		genesisTime = uint64(time.Now().Unix())
	}
	beaconState, err := coreState.GenesisBeaconStateCapella(ctx, deposits, genesisTime, e1d, ep)
	if err != nil {
		return nil, nil, errors.Wrap(err, "could not generate genesis state")
	}
	bsi := beaconState.ToProtoUnsafe()
	pbc, ok := bsi.(*zondpb.BeaconStateCapella)
	if !ok {
		return nil, nil, errors.New("unexpected BeaconState version")
	}
	pbState, err := statenative.ProtobufBeaconStateCapella(pbc)
	if err != nil {
		return nil, nil, err
	}
	return pbState, deposits, nil
}
