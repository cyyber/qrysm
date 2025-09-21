package altair_test

import (
	"context"
	"testing"
	"time"

	"github.com/theQRL/qrysm/beacon-chain/core/altair"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/beacon-chain/state"
	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/ml_dsa_87"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	qrysmTime "github.com/theQRL/qrysm/time"
)

func TestSyncCommitteeIndices_CanGet(t *testing.T) {
	getState := func(t *testing.T, count uint64) state.BeaconState {
		validators := make([]*qrysmpb.Validator, count)
		for i := 0; i < len(validators); i++ {
			validators[i] = &qrysmpb.Validator{
				ExitEpoch:        params.BeaconConfig().FarFutureEpoch,
				EffectiveBalance: params.BeaconConfig().MinDepositAmount,
			}
		}
		st, err := state_native.InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{
			Validators:  validators,
			RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		})
		require.NoError(t, err)
		return st
	}

	type args struct {
		state state.BeaconState
		epoch primitives.Epoch
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		errString string
	}{
		{
			name: "genesis validator count, epoch 0",
			args: args{
				state: getState(t, params.BeaconConfig().MinGenesisActiveValidatorCount),
				epoch: 0,
			},
			wantErr: false,
		},
		{
			name: "genesis validator count, epoch 100",
			args: args{
				state: getState(t, params.BeaconConfig().MinGenesisActiveValidatorCount),
				epoch: 100,
			},
			wantErr: false,
		},
		{
			name: "less than optimal validator count, epoch 100",
			args: args{
				state: getState(t, params.BeaconConfig().MaxValidatorsPerCommittee),
				epoch: 100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()
			got, err := altair.NextSyncCommitteeIndices(context.Background(), tt.args.state)
			if tt.wantErr {
				require.ErrorContains(t, tt.errString, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, int(params.BeaconConfig().SyncCommitteeSize), len(got))
			}
		})
	}
}

func TestSyncCommitteeIndices_DifferentPeriods(t *testing.T) {
	helpers.ClearCache()
	getState := func(t *testing.T, count uint64) state.BeaconState {
		validators := make([]*qrysmpb.Validator, count)
		for i := 0; i < len(validators); i++ {
			validators[i] = &qrysmpb.Validator{
				ExitEpoch:        params.BeaconConfig().FarFutureEpoch,
				EffectiveBalance: params.BeaconConfig().MinDepositAmount,
			}
		}
		st, err := state_native.InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{
			Validators:  validators,
			RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		})
		require.NoError(t, err)
		return st
	}

	st := getState(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	got1, err := altair.NextSyncCommitteeIndices(context.Background(), st)
	require.NoError(t, err)
	require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch))
	got2, err := altair.NextSyncCommitteeIndices(context.Background(), st)
	require.NoError(t, err)
	require.DeepNotEqual(t, got1, got2)
	require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch*primitives.Slot(params.BeaconConfig().EpochsPerSyncCommitteePeriod)))
	got2, err = altair.NextSyncCommitteeIndices(context.Background(), st)
	require.NoError(t, err)
	require.DeepNotEqual(t, got1, got2)
	require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch*primitives.Slot(2*params.BeaconConfig().EpochsPerSyncCommitteePeriod)))
	got2, err = altair.NextSyncCommitteeIndices(context.Background(), st)
	require.NoError(t, err)
	require.DeepNotEqual(t, got1, got2)
}

func TestSyncCommittee_CanGet(t *testing.T) {
	getState := func(t *testing.T, count uint64) state.BeaconState {
		validators := make([]*qrysmpb.Validator, count)
		for i := 0; i < len(validators); i++ {
			mlDSA87Key, err := ml_dsa_87.RandKey()
			require.NoError(t, err)
			validators[i] = &qrysmpb.Validator{
				ExitEpoch:        params.BeaconConfig().FarFutureEpoch,
				EffectiveBalance: params.BeaconConfig().MinDepositAmount,
				PublicKey:        mlDSA87Key.PublicKey().Marshal(),
			}
		}
		st, err := state_native.InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{
			Validators:  validators,
			RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
		})
		require.NoError(t, err)
		return st
	}

	type args struct {
		state state.BeaconState
		epoch primitives.Epoch
	}
	tests := []struct {
		name      string
		args      args
		wantErr   bool
		errString string
	}{
		{
			name: "genesis validator count, epoch 0",
			args: args{
				state: getState(t, params.BeaconConfig().MinGenesisActiveValidatorCount),
				epoch: 0,
			},
			wantErr: false,
		},
		{
			name: "genesis validator count, epoch 100",
			args: args{
				state: getState(t, params.BeaconConfig().MinGenesisActiveValidatorCount),
				epoch: 100,
			},
			wantErr: false,
		},
		{
			name: "less than optimal validator count, epoch 100",
			args: args{
				state: getState(t, params.BeaconConfig().MaxValidatorsPerCommittee),
				epoch: 100,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			helpers.ClearCache()
			if !tt.wantErr {
				require.NoError(t, tt.args.state.SetSlot(primitives.Slot(tt.args.epoch)*params.BeaconConfig().SlotsPerEpoch))
			}
			got, err := altair.NextSyncCommittee(context.Background(), tt.args.state)
			if tt.wantErr {
				require.ErrorContains(t, tt.errString, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, int(params.BeaconConfig().SyncCommitteeSize), len(got.Pubkeys))
			}
		})
	}
}

func TestValidateNilSyncContribution(t *testing.T) {
	tests := []struct {
		name    string
		s       *qrysmpb.SignedContributionAndProof
		wantErr bool
	}{
		{
			name:    "nil object",
			s:       nil,
			wantErr: true,
		},
		{
			name:    "nil message",
			s:       &qrysmpb.SignedContributionAndProof{},
			wantErr: true,
		},
		{
			name:    "nil contribution",
			s:       &qrysmpb.SignedContributionAndProof{Message: &qrysmpb.ContributionAndProof{}},
			wantErr: true,
		},
		{
			name: "nil bitfield",
			s: &qrysmpb.SignedContributionAndProof{
				Message: &qrysmpb.ContributionAndProof{
					Contribution: &qrysmpb.SyncCommitteeContribution{},
				}},
			wantErr: true,
		},
		{
			name: "non nil sync contribution",
			s: &qrysmpb.SignedContributionAndProof{
				Message: &qrysmpb.ContributionAndProof{
					Contribution: &qrysmpb.SyncCommitteeContribution{
						AggregationBits: []byte{},
					},
				}},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := altair.ValidateNilSyncContribution(tt.s); (err != nil) != tt.wantErr {
				t.Errorf("ValidateNilSyncContribution() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSyncSubCommitteePubkeys_CanGet(t *testing.T) {
	helpers.ClearCache()
	st := getState(t, params.BeaconConfig().MaxValidatorsPerCommittee)
	com, err := altair.NextSyncCommittee(context.Background(), st)
	require.NoError(t, err)
	sub, err := altair.SyncSubCommitteePubkeys(com, 0)
	require.NoError(t, err)
	subCommSize := params.BeaconConfig().SyncCommitteeSize / params.BeaconConfig().SyncCommitteeSubnetCount
	require.Equal(t, int(subCommSize), len(sub))
	require.DeepSSZEqual(t, com.Pubkeys[0:subCommSize], sub)
}

func Test_ValidateSyncMessageTime(t *testing.T) {
	if params.BeaconNetworkConfig().MaximumGossipClockDisparity < 200*time.Millisecond {
		t.Fatal("This test expects the maximum clock disparity to be at least 200ms")
	}

	type args struct {
		syncMessageSlot primitives.Slot
		genesisTime     time.Time
	}
	tests := []struct {
		name      string
		args      args
		wantedErr string
	}{
		{
			name: "sync_message.slot == current_slot",
			args: args{
				syncMessageSlot: 15,
				genesisTime:     qrysmTime.Now().Add(-15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
		},
		{
			name: "sync_message.slot == current_slot, received in middle of slot",
			args: args{
				syncMessageSlot: 15,
				genesisTime: qrysmTime.Now().Add(
					-15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second,
				).Add(-(time.Duration(params.BeaconConfig().SecondsPerSlot/2) * time.Second)),
			},
		},
		{
			name: "sync_message.slot == current_slot, received 200ms early",
			args: args{
				syncMessageSlot: 16,
				genesisTime: qrysmTime.Now().Add(
					-16 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second,
				).Add(-200 * time.Millisecond),
			},
		},
		{
			name: "sync_message.slot > current_slot",
			args: args{
				syncMessageSlot: 16,
				genesisTime:     qrysmTime.Now().Add(-(15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second)),
			},
			wantedErr: "(message slot 16) not within allowable range of",
		},
		{
			name: "sync_message.slot == current_slot+CLOCK_DISPARITY",
			args: args{
				syncMessageSlot: 100,
				genesisTime:     qrysmTime.Now().Add(-(100*time.Duration(params.BeaconConfig().SecondsPerSlot)*time.Second - params.BeaconNetworkConfig().MaximumGossipClockDisparity)),
			},
			wantedErr: "",
		},
		{
			name: "sync_message.slot == current_slot+CLOCK_DISPARITY-1000ms",
			args: args{
				syncMessageSlot: 100,
				genesisTime:     qrysmTime.Now().Add(-(100 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second) + params.BeaconNetworkConfig().MaximumGossipClockDisparity + 1000*time.Millisecond),
			},
			wantedErr: "(message slot 100) not within allowable range of",
		},
		{
			name: "sync_message.slot == current_slot-CLOCK_DISPARITY",
			args: args{
				syncMessageSlot: 100,
				genesisTime:     qrysmTime.Now().Add(-(100*time.Duration(params.BeaconConfig().SecondsPerSlot)*time.Second + params.BeaconNetworkConfig().MaximumGossipClockDisparity)),
			},
			wantedErr: "",
		},
		{
			name: "sync_message.slot > current_slot+CLOCK_DISPARITY",
			args: args{
				syncMessageSlot: 101,
				genesisTime:     qrysmTime.Now().Add(-(100*time.Duration(params.BeaconConfig().SecondsPerSlot)*time.Second + params.BeaconNetworkConfig().MaximumGossipClockDisparity)),
			},
			wantedErr: "(message slot 101) not within allowable range of",
		},
		{
			name: "sync_message.slot is well beyond current slot",
			args: args{
				syncMessageSlot: 1 << 32,
				genesisTime:     qrysmTime.Now().Add(-15 * time.Duration(params.BeaconConfig().SecondsPerSlot) * time.Second),
			},
			wantedErr: "which exceeds max allowed value relative to the local clock",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := altair.ValidateSyncMessageTime(tt.args.syncMessageSlot, tt.args.genesisTime,
				params.BeaconNetworkConfig().MaximumGossipClockDisparity)
			if tt.wantedErr != "" {
				assert.ErrorContains(t, tt.wantedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func getState(t *testing.T, count uint64) state.BeaconState {
	validators := make([]*qrysmpb.Validator, count)
	for i := 0; i < len(validators); i++ {
		mlDSA87Key, err := ml_dsa_87.RandKey()
		require.NoError(t, err)
		validators[i] = &qrysmpb.Validator{
			ExitEpoch:        params.BeaconConfig().FarFutureEpoch,
			EffectiveBalance: params.BeaconConfig().MinDepositAmount,
			PublicKey:        mlDSA87Key.PublicKey().Marshal(),
		}
	}
	st, err := state_native.InitializeFromProtoCapella(&qrysmpb.BeaconStateCapella{
		Validators:  validators,
		RandaoMixes: make([][]byte, params.BeaconConfig().EpochsPerHistoricalVector),
	})
	require.NoError(t, err)
	return st
}
