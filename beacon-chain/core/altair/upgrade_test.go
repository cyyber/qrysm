package altair_test

import (
	"context"
	"testing"

	"github.com/theQRL/go-bitfield"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/altair"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/time"
	"github.com/theQRL/qrysm/v4/config/params"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1/attestation"
	"github.com/theQRL/qrysm/v4/testing/require"
	"github.com/theQRL/qrysm/v4/testing/util"
)

func TestTranslateParticipation(t *testing.T) {
	ctx := context.Background()
	s, _ := util.DeterministicGenesisStateAltair(t, 64)
	require.NoError(t, s.SetSlot(s.Slot()+params.BeaconConfig().MinAttestationInclusionDelay))

	var err error
	newState, err := altair.TranslateParticipation(ctx, s, nil)
	require.NoError(t, err)
	participation, err := newState.PreviousEpochParticipation()
	require.NoError(t, err)
	require.DeepSSZEqual(t, make([]byte, 64), participation)

	aggBits := bitfield.NewBitlist(2)
	aggBits.SetBitAt(0, true)
	aggBits.SetBitAt(1, true)
	r, err := helpers.BlockRootAtSlot(s, 0)
	require.NoError(t, err)
	var pendingAtts []*zondpb.PendingAttestation
	for i := 0; i < 3; i++ {
		pendingAtts = append(pendingAtts, &zondpb.PendingAttestation{
			Data: &zondpb.AttestationData{
				CommitteeIndex:  primitives.CommitteeIndex(i),
				BeaconBlockRoot: r,
				Source:          &zondpb.Checkpoint{Epoch: 0, Root: make([]byte, 32)},
				Target:          &zondpb.Checkpoint{Epoch: 0, Root: make([]byte, 32)},
			},
			AggregationBits: aggBits,
			InclusionDelay:  1,
		})
	}

	newState, err = altair.TranslateParticipation(ctx, newState, pendingAtts)
	require.NoError(t, err)
	participation, err = newState.PreviousEpochParticipation()
	require.NoError(t, err)
	require.DeepNotSSZEqual(t, make([]byte, 64), participation)

	committee, err := helpers.BeaconCommitteeFromState(ctx, s, pendingAtts[0].Data.Slot, pendingAtts[0].Data.CommitteeIndex)
	require.NoError(t, err)
	indices, err := attestation.AttestingIndices(pendingAtts[0].AggregationBits, committee)
	require.NoError(t, err)
	for _, index := range indices {
		has, err := altair.HasValidatorFlag(participation[index], params.BeaconConfig().TimelySourceFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
		has, err = altair.HasValidatorFlag(participation[index], params.BeaconConfig().TimelyTargetFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
		has, err = altair.HasValidatorFlag(participation[index], params.BeaconConfig().TimelyHeadFlagIndex)
		require.NoError(t, err)
		require.Equal(t, true, has)
	}
}

func TestUpgradeToAltair(t *testing.T) {
	st, _ := util.DeterministicGenesisState(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	preForkState := st.Copy()
	aState, err := altair.UpgradeToAltair(context.Background(), st)
	require.NoError(t, err)

	require.Equal(t, preForkState.GenesisTime(), aState.GenesisTime())
	require.DeepSSZEqual(t, preForkState.GenesisValidatorsRoot(), aState.GenesisValidatorsRoot())
	require.Equal(t, preForkState.Slot(), aState.Slot())
	require.DeepSSZEqual(t, preForkState.LatestBlockHeader(), aState.LatestBlockHeader())
	require.DeepSSZEqual(t, preForkState.BlockRoots(), aState.BlockRoots())
	require.DeepSSZEqual(t, preForkState.StateRoots(), aState.StateRoots())
	r1, err := preForkState.HistoricalRoots()
	require.NoError(t, err)
	r2, err := aState.HistoricalRoots()
	require.NoError(t, err)
	require.DeepSSZEqual(t, r1, r2)
	require.DeepSSZEqual(t, preForkState.ZondData(), aState.ZondData())
	require.DeepSSZEqual(t, preForkState.ZondDataVotes(), aState.ZondDataVotes())
	require.DeepSSZEqual(t, preForkState.ZondDepositIndex(), aState.ZondDepositIndex())
	require.DeepSSZEqual(t, preForkState.Validators(), aState.Validators())
	require.DeepSSZEqual(t, preForkState.Balances(), aState.Balances())
	require.DeepSSZEqual(t, preForkState.RandaoMixes(), aState.RandaoMixes())
	require.DeepSSZEqual(t, preForkState.Slashings(), aState.Slashings())
	require.DeepSSZEqual(t, preForkState.JustificationBits(), aState.JustificationBits())
	require.DeepSSZEqual(t, preForkState.PreviousJustifiedCheckpoint(), aState.PreviousJustifiedCheckpoint())
	require.DeepSSZEqual(t, preForkState.CurrentJustifiedCheckpoint(), aState.CurrentJustifiedCheckpoint())
	require.DeepSSZEqual(t, preForkState.FinalizedCheckpoint(), aState.FinalizedCheckpoint())
	numValidators := aState.NumValidators()
	p, err := aState.PreviousEpochParticipation()
	require.NoError(t, err)
	require.DeepSSZEqual(t, make([]byte, numValidators), p)
	p, err = aState.CurrentEpochParticipation()
	require.NoError(t, err)
	require.DeepSSZEqual(t, make([]byte, numValidators), p)
	s, err := aState.InactivityScores()
	require.NoError(t, err)
	require.DeepSSZEqual(t, make([]uint64, numValidators), s)

	f := aState.Fork()
	require.DeepSSZEqual(t, &zondpb.Fork{
		PreviousVersion: st.Fork().CurrentVersion,
		CurrentVersion:  params.BeaconConfig().AltairForkVersion,
		Epoch:           time.CurrentEpoch(st),
	}, f)
	csc, err := aState.CurrentSyncCommittee()
	require.NoError(t, err)
	nsc, err := aState.NextSyncCommittee()
	require.NoError(t, err)
	require.DeepSSZEqual(t, nsc, csc)
}
