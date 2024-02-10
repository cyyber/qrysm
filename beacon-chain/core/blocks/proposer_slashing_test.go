package blocks_test

import (
	"context"
	"fmt"
	"testing"

	dilithium2 "github.com/theQRL/go-qrllib/dilithium"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/blocks"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/signing"
	v "github.com/theQRL/qrysm/v4/beacon-chain/core/validators"
	"github.com/theQRL/qrysm/v4/beacon-chain/state"
	state_native "github.com/theQRL/qrysm/v4/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/v4/config/params"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	"github.com/theQRL/qrysm/v4/crypto/dilithium"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
	"github.com/theQRL/qrysm/v4/testing/util"
	"github.com/theQRL/qrysm/v4/time/slots"
)

func TestProcessProposerSlashings_UnmatchedHeaderSlots(t *testing.T) {

	beaconState, _ := util.DeterministicGenesisStateCapella(t, 20)
	currentSlot := primitives.Slot(0)
	slashings := []*zondpb.ProposerSlashing{
		{
			Header_1: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					ProposerIndex: 1,
					Slot:          params.BeaconConfig().SlotsPerEpoch + 1,
				},
			},
			Header_2: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					ProposerIndex: 1,
					Slot:          0,
				},
			},
		},
	}
	require.NoError(t, beaconState.SetSlot(currentSlot))

	b := util.NewBeaconBlockCapella()
	b.Block = &zondpb.BeaconBlockCapella{
		Body: &zondpb.BeaconBlockBodyCapella{
			ProposerSlashings: slashings,
		},
	}
	want := "mismatched header slots"
	_, err := blocks.ProcessProposerSlashings(context.Background(), beaconState, b.Block.Body.ProposerSlashings, v.SlashValidator)
	assert.ErrorContains(t, want, err)
}

func TestProcessProposerSlashings_SameHeaders(t *testing.T) {

	beaconState, _ := util.DeterministicGenesisStateCapella(t, 2)
	currentSlot := primitives.Slot(0)
	slashings := []*zondpb.ProposerSlashing{
		{
			Header_1: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					ProposerIndex: 1,
					Slot:          0,
				},
			},
			Header_2: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					ProposerIndex: 1,
					Slot:          0,
				},
			},
		},
	}

	require.NoError(t, beaconState.SetSlot(currentSlot))
	b := util.NewBeaconBlockCapella()
	b.Block = &zondpb.BeaconBlockCapella{
		Body: &zondpb.BeaconBlockBodyCapella{
			ProposerSlashings: slashings,
		},
	}
	want := "expected slashing headers to differ"
	_, err := blocks.ProcessProposerSlashings(context.Background(), beaconState, b.Block.Body.ProposerSlashings, v.SlashValidator)
	assert.ErrorContains(t, want, err)
}

func TestProcessProposerSlashings_ValidatorNotSlashable(t *testing.T) {
	registry := []*zondpb.Validator{
		{
			PublicKey:         []byte("key"),
			Slashed:           true,
			ActivationEpoch:   0,
			WithdrawableEpoch: 0,
		},
	}
	currentSlot := primitives.Slot(0)
	slashings := []*zondpb.ProposerSlashing{
		{
			Header_1: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					ProposerIndex: 0,
					Slot:          0,
					BodyRoot:      []byte("foo"),
				},
				Signature: bytesutil.PadTo([]byte("A"), dilithium2.CryptoBytes),
			},
			Header_2: &zondpb.SignedBeaconBlockHeader{
				Header: &zondpb.BeaconBlockHeader{
					ProposerIndex: 0,
					Slot:          0,
					BodyRoot:      []byte("bar"),
				},
				Signature: bytesutil.PadTo([]byte("B"), dilithium2.CryptoBytes),
			},
		},
	}

	beaconState, err := state_native.InitializeFromProtoCapella(&zondpb.BeaconStateCapella{
		Validators: registry,
		Slot:       currentSlot,
	})
	require.NoError(t, err)
	b := util.NewBeaconBlockCapella()
	b.Block = &zondpb.BeaconBlockCapella{
		Body: &zondpb.BeaconBlockBodyCapella{
			ProposerSlashings: slashings,
		},
	}
	want := fmt.Sprintf(
		"validator with key %#x is not slashable",
		bytesutil.ToBytes48(beaconState.Validators()[0].PublicKey),
	)
	_, err = blocks.ProcessProposerSlashings(context.Background(), beaconState, b.Block.Body.ProposerSlashings, v.SlashValidator)
	assert.ErrorContains(t, want, err)
}

func TestProcessProposerSlashings_AppliesCorrectStatusCapella(t *testing.T) {
	// We test the case when data is correct and verify the validator
	// registry has been updated.
	beaconState, privKeys := util.DeterministicGenesisStateCapella(t, 100)
	proposerIdx := primitives.ValidatorIndex(1)

	header1 := &zondpb.SignedBeaconBlockHeader{
		Header: util.HydrateBeaconHeader(&zondpb.BeaconBlockHeader{
			ProposerIndex: proposerIdx,
			StateRoot:     bytesutil.PadTo([]byte("A"), 32),
		}),
	}
	var err error
	header1.Signature, err = signing.ComputeDomainAndSign(beaconState, 0, header1.Header, params.BeaconConfig().DomainBeaconProposer, privKeys[proposerIdx])
	require.NoError(t, err)

	header2 := util.HydrateSignedBeaconHeader(&zondpb.SignedBeaconBlockHeader{
		Header: &zondpb.BeaconBlockHeader{
			ProposerIndex: proposerIdx,
			StateRoot:     bytesutil.PadTo([]byte("B"), 32),
		},
	})
	header2.Signature, err = signing.ComputeDomainAndSign(beaconState, 0, header2.Header, params.BeaconConfig().DomainBeaconProposer, privKeys[proposerIdx])
	require.NoError(t, err)

	slashings := []*zondpb.ProposerSlashing{
		{
			Header_1: header1,
			Header_2: header2,
		},
	}

	block := util.NewBeaconBlockCapella()
	block.Block.Body.ProposerSlashings = slashings

	newState, err := blocks.ProcessProposerSlashings(context.Background(), beaconState, block.Block.Body.ProposerSlashings, v.SlashValidator)
	require.NoError(t, err)

	newStateVals := newState.Validators()
	if newStateVals[1].ExitEpoch != beaconState.Validators()[1].ExitEpoch {
		t.Errorf("Proposer with index 1 did not correctly exit,"+"wanted slot:%d, got:%d",
			newStateVals[1].ExitEpoch, beaconState.Validators()[1].ExitEpoch)
	}

	require.Equal(t, uint64(31000000000), newState.Balances()[1])
	require.Equal(t, uint64(32000000000), newState.Balances()[2])
}

func TestVerifyProposerSlashing(t *testing.T) {
	type args struct {
		beaconState state.BeaconState
		slashing    *zondpb.ProposerSlashing
	}

	beaconState, sks := util.DeterministicGenesisStateCapella(t, 2)
	currentSlot := primitives.Slot(0)
	require.NoError(t, beaconState.SetSlot(currentSlot))
	rand1, err := dilithium.RandKey()
	require.NoError(t, err)
	sig1 := rand1.Sign([]byte("foo")).Marshal()

	rand2, err := dilithium.RandKey()
	require.NoError(t, err)
	sig2 := rand2.Sign([]byte("bar")).Marshal()

	tests := []struct {
		name    string
		args    args
		wantErr string
	}{
		{
			name: "same header, same slot as state",
			args: args{
				slashing: &zondpb.ProposerSlashing{
					Header_1: util.HydrateSignedBeaconHeader(&zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							ProposerIndex: 1,
							Slot:          currentSlot,
						},
					}),
					Header_2: util.HydrateSignedBeaconHeader(&zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							ProposerIndex: 1,
							Slot:          currentSlot,
						},
					}),
				},
				beaconState: beaconState,
			},
			wantErr: "expected slashing headers to differ",
		},
		{ // Regression test for https://github.com/sigp/beacon-fuzz/issues/74
			name: "same header, different signatures",
			args: args{
				slashing: &zondpb.ProposerSlashing{
					Header_1: util.HydrateSignedBeaconHeader(&zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							ProposerIndex: 1,
						},
						Signature: sig1,
					}),
					Header_2: util.HydrateSignedBeaconHeader(&zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							ProposerIndex: 1,
						},
						Signature: sig2,
					}),
				},
				beaconState: beaconState,
			},
			wantErr: "expected slashing headers to differ",
		},
		{
			name: "slashing in future epoch",
			args: args{
				slashing: &zondpb.ProposerSlashing{
					Header_1: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							ProposerIndex: 1,
							Slot:          65,
							StateRoot:     bytesutil.PadTo([]byte{}, 32),
							BodyRoot:      bytesutil.PadTo([]byte{}, 32),
							ParentRoot:    bytesutil.PadTo([]byte("foo"), 32),
						},
					},
					Header_2: &zondpb.SignedBeaconBlockHeader{
						Header: &zondpb.BeaconBlockHeader{
							ProposerIndex: 1,
							Slot:          65,
							StateRoot:     bytesutil.PadTo([]byte{}, 32),
							BodyRoot:      bytesutil.PadTo([]byte{}, 32),
							ParentRoot:    bytesutil.PadTo([]byte("bar"), 32),
						},
					},
				},
				beaconState: beaconState,
			},
			wantErr: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			sk := sks[tt.args.slashing.Header_1.Header.ProposerIndex]
			d, err := signing.Domain(tt.args.beaconState.Fork(), slots.ToEpoch(tt.args.slashing.Header_1.Header.Slot), params.BeaconConfig().DomainBeaconProposer, tt.args.beaconState.GenesisValidatorsRoot())
			require.NoError(t, err)
			if tt.args.slashing.Header_1.Signature == nil {
				sr, err := signing.ComputeSigningRoot(tt.args.slashing.Header_1.Header, d)
				require.NoError(t, err)
				tt.args.slashing.Header_1.Signature = sk.Sign(sr[:]).Marshal()
			}
			if tt.args.slashing.Header_2.Signature == nil {
				sr, err := signing.ComputeSigningRoot(tt.args.slashing.Header_2.Header, d)
				require.NoError(t, err)
				tt.args.slashing.Header_2.Signature = sk.Sign(sr[:]).Marshal()
			}
			if err := blocks.VerifyProposerSlashing(tt.args.beaconState, tt.args.slashing); (err != nil || tt.wantErr != "") && err.Error() != tt.wantErr {
				t.Errorf("VerifyProposerSlashing() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
