package slasher

import (
	"reflect"
	"testing"

	logTest "github.com/sirupsen/logrus/hooks/test"
	slashertypes "github.com/theQRL/qrysm/beacon-chain/slasher/types"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
)

func TestService_groupByValidatorChunkIndex(t *testing.T) {
	tests := []struct {
		name   string
		params *Parameters
		atts   []*slashertypes.IndexedAttestationWrapper
		want   map[uint64][]*slashertypes.IndexedAttestationWrapper
	}{
		{
			name:   "No attestations returns empty map",
			params: DefaultParams(),
			atts:   make([]*slashertypes.IndexedAttestationWrapper, 0),
			want:   make(map[uint64][]*slashertypes.IndexedAttestationWrapper),
		},
		{
			name: "Groups multiple attestations belonging to single validator chunk",
			params: &Parameters{
				validatorChunkSize: 2,
			},
			atts: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 0, []uint64{0, 1}, nil),
				createAttestationWrapper(t, 0, 0, []uint64{0, 1}, nil),
			},
			want: map[uint64][]*slashertypes.IndexedAttestationWrapper{
				0: {
					createAttestationWrapper(t, 0, 0, []uint64{0, 1}, nil),
					createAttestationWrapper(t, 0, 0, []uint64{0, 1}, nil),
				},
			},
		},
		{
			name: "Groups single attestation belonging to multiple validator chunk",
			params: &Parameters{
				validatorChunkSize: 2,
			},
			atts: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 0, []uint64{0, 2, 4}, nil),
			},
			want: map[uint64][]*slashertypes.IndexedAttestationWrapper{
				0: {
					createAttestationWrapper(t, 0, 0, []uint64{0, 2, 4}, nil),
				},
				1: {
					createAttestationWrapper(t, 0, 0, []uint64{0, 2, 4}, nil),
				},
				2: {
					createAttestationWrapper(t, 0, 0, []uint64{0, 2, 4}, nil),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				params: tt.params,
			}
			if got := s.groupByValidatorChunkIndex(tt.atts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupByValidatorChunkIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_groupByChunkIndex(t *testing.T) {
	tests := []struct {
		name   string
		params *Parameters
		atts   []*slashertypes.IndexedAttestationWrapper
		want   map[uint64][]*slashertypes.IndexedAttestationWrapper
	}{
		{
			name:   "No attestations returns empty map",
			params: DefaultParams(),
			atts:   make([]*slashertypes.IndexedAttestationWrapper, 0),
			want:   make(map[uint64][]*slashertypes.IndexedAttestationWrapper),
		},
		{
			name: "Groups multiple attestations belonging to single chunk",
			params: &Parameters{
				chunkSize:     2,
				historyLength: 3,
			},
			atts: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 0, nil, nil),
				createAttestationWrapper(t, 1, 0, nil, nil),
			},
			want: map[uint64][]*slashertypes.IndexedAttestationWrapper{
				0: {
					createAttestationWrapper(t, 0, 0, nil, nil),
					createAttestationWrapper(t, 1, 0, nil, nil),
				},
			},
		},
		{
			name: "Groups multiple attestations belonging to multiple chunks",
			params: &Parameters{
				chunkSize:     2,
				historyLength: 3,
			},
			atts: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 0, nil, nil),
				createAttestationWrapper(t, 1, 0, nil, nil),
				createAttestationWrapper(t, 2, 0, nil, nil),
			},
			want: map[uint64][]*slashertypes.IndexedAttestationWrapper{
				0: {
					createAttestationWrapper(t, 0, 0, nil, nil),
					createAttestationWrapper(t, 1, 0, nil, nil),
				},
				1: {
					createAttestationWrapper(t, 2, 0, nil, nil),
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Service{
				params: tt.params,
			}
			if got := s.groupByChunkIndex(tt.atts); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("groupByChunkIndex() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_filterAttestations(t *testing.T) {
	tests := []struct {
		name           string
		input          []*slashertypes.IndexedAttestationWrapper
		inputEpoch     primitives.Epoch
		wantedValid    []*slashertypes.IndexedAttestationWrapper
		wantedDeferred []*slashertypes.IndexedAttestationWrapper
		wantedDropped  int
	}{
		{
			name:          "Nil attestation input gets dropped",
			input:         make([]*slashertypes.IndexedAttestationWrapper, 1),
			inputEpoch:    0,
			wantedDropped: 1,
		},
		{
			name: "Nil attestation data gets dropped",
			input: []*slashertypes.IndexedAttestationWrapper{
				{
					IndexedAttestation: &qrysmpb.IndexedAttestation{},
				},
			},
			inputEpoch:    0,
			wantedDropped: 1,
		},
		{
			name: "Nil attestation source and target gets dropped",
			input: []*slashertypes.IndexedAttestationWrapper{
				{
					IndexedAttestation: &qrysmpb.IndexedAttestation{
						Data: &qrysmpb.AttestationData{},
					},
				},
			},
			inputEpoch:    0,
			wantedDropped: 1,
		},
		{
			name: "Nil attestation source and good target gets dropped",
			input: []*slashertypes.IndexedAttestationWrapper{
				{
					IndexedAttestation: &qrysmpb.IndexedAttestation{
						Data: &qrysmpb.AttestationData{
							Target: &qrysmpb.Checkpoint{},
						},
					},
				},
			},
			inputEpoch:    0,
			wantedDropped: 1,
		},
		{
			name: "Nil attestation target and good source gets dropped",
			input: []*slashertypes.IndexedAttestationWrapper{
				{
					IndexedAttestation: &qrysmpb.IndexedAttestation{
						Data: &qrysmpb.AttestationData{
							Source: &qrysmpb.Checkpoint{},
						},
					},
				},
			},
			inputEpoch:    0,
			wantedDropped: 1,
		},
		{
			name: "Source > target gets dropped",
			input: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 1, 0, []uint64{1}, make([]byte, 32)),
			},
			inputEpoch:    0,
			wantedDropped: 1,
		},
		{
			name: "Source < target is valid",
			input: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 1, []uint64{1}, make([]byte, 32)),
			},
			inputEpoch: 1,
			wantedValid: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 1, []uint64{1}, make([]byte, 32)),
			},
			wantedDropped: 0,
		},
		{
			name: "Source == target is valid",
			input: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 0, []uint64{1}, make([]byte, 32)),
			},
			inputEpoch: 1,
			wantedValid: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 0, []uint64{1}, make([]byte, 32)),
			},
			wantedDropped: 0,
		},
		{
			name: "Attestation from the future is deferred",
			input: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 2, []uint64{1}, make([]byte, 32)),
			},
			inputEpoch: 1,
			wantedDeferred: []*slashertypes.IndexedAttestationWrapper{
				createAttestationWrapper(t, 0, 2, []uint64{1}, make([]byte, 32)),
			},
			wantedDropped: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := &Service{
				params: DefaultParams(),
			}
			valid, deferred, numDropped := srv.filterAttestations(tt.input, tt.inputEpoch)
			if len(tt.wantedValid) > 0 {
				require.DeepEqual(t, tt.wantedValid, valid)
			}
			if len(tt.wantedDeferred) > 0 {
				require.DeepEqual(t, tt.wantedDeferred, deferred)
			}
			require.DeepEqual(t, tt.wantedDropped, numDropped)
		})
	}
}

func Test_logSlashingEvent(t *testing.T) {
	tests := []struct {
		name     string
		slashing *qrysmpb.AttesterSlashing
	}{
		{
			name: "Surrounding vote",
			slashing: &qrysmpb.AttesterSlashing{
				Attestation_1: createAttestationWrapper(t, 0, 0, nil, nil).IndexedAttestation,
				Attestation_2: createAttestationWrapper(t, 0, 0, nil, nil).IndexedAttestation,
			},
		},
		{
			name: "Surrounded vote",
			slashing: &qrysmpb.AttesterSlashing{
				Attestation_1: createAttestationWrapper(t, 0, 0, nil, nil).IndexedAttestation,
				Attestation_2: createAttestationWrapper(t, 0, 0, nil, nil).IndexedAttestation,
			},
		},
		{
			name: "Double vote",
			slashing: &qrysmpb.AttesterSlashing{
				Attestation_1: createAttestationWrapper(t, 0, 0, nil, nil).IndexedAttestation,
				Attestation_2: createAttestationWrapper(t, 0, 0, nil, nil).IndexedAttestation,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hook := logTest.NewGlobal()
			logAttesterSlashing(tt.slashing)
			require.LogsContain(t, hook, "")
		})
	}
}

func Test_validateAttestationIntegrity(t *testing.T) {
	tests := []struct {
		name string
		att  *qrysmpb.IndexedAttestation
		want bool
	}{
		{
			name: "Nil attestation returns false",
			att:  nil,
			want: false,
		},
		{
			name: "Nil attestation data returns false",
			att:  &qrysmpb.IndexedAttestation{},
			want: false,
		},
		{
			name: "Nil attestation source and target returns false",
			att: &qrysmpb.IndexedAttestation{
				Data: &qrysmpb.AttestationData{},
			},
			want: false,
		},
		{
			name: "Nil attestation source and good target returns false",
			att: &qrysmpb.IndexedAttestation{
				Data: &qrysmpb.AttestationData{
					Target: &qrysmpb.Checkpoint{},
				},
			},
			want: false,
		},
		{
			name: "Nil attestation target and good source returns false",
			att: &qrysmpb.IndexedAttestation{
				Data: &qrysmpb.AttestationData{
					Source: &qrysmpb.Checkpoint{},
				},
			},
			want: false,
		},
		{
			name: "Source > target returns false",
			att: &qrysmpb.IndexedAttestation{
				Data: &qrysmpb.AttestationData{
					Source: &qrysmpb.Checkpoint{
						Epoch: 1,
					},
					Target: &qrysmpb.Checkpoint{
						Epoch: 0,
					},
				},
			},
			want: false,
		},
		{
			name: "Source == target returns false",
			att: &qrysmpb.IndexedAttestation{
				Data: &qrysmpb.AttestationData{
					Source: &qrysmpb.Checkpoint{
						Epoch: 1,
					},
					Target: &qrysmpb.Checkpoint{
						Epoch: 1,
					},
				},
			},
			want: false,
		},
		{
			name: "Source < target returns true",
			att: &qrysmpb.IndexedAttestation{
				Data: &qrysmpb.AttestationData{
					Source: &qrysmpb.Checkpoint{
						Epoch: 1,
					},
					Target: &qrysmpb.Checkpoint{
						Epoch: 2,
					},
				},
			},
			want: true,
		},
		{
			name: "Source 0 target 0 returns true (genesis epoch attestations)",
			att: &qrysmpb.IndexedAttestation{
				Data: &qrysmpb.AttestationData{
					Source: &qrysmpb.Checkpoint{
						Epoch: 0,
					},
					Target: &qrysmpb.Checkpoint{
						Epoch: 0,
					},
				},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateAttestationIntegrity(tt.att); got != tt.want {
				t.Errorf("validateAttestationIntegrity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_validateBlockHeaderIntegrity(t *testing.T) {
	type args struct {
		header *qrysmpb.SignedBeaconBlockHeader
	}
	fakeSig := make([]byte, field_params.DilithiumSignatureLength)
	copy(fakeSig, "hi")
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "nil header",
			args: args{
				header: nil,
			},
			want: false,
		},
		{
			name: "nil inner header",
			args: args{
				header: &qrysmpb.SignedBeaconBlockHeader{},
			},
			want: false,
		},
		{
			name: "bad signature 1",
			args: args{
				header: &qrysmpb.SignedBeaconBlockHeader{Header: &qrysmpb.BeaconBlockHeader{}, Signature: []byte("hi")},
			},
			want: false,
		},
		{
			name: "bad signature 2",
			args: args{
				header: &qrysmpb.SignedBeaconBlockHeader{
					Header:    &qrysmpb.BeaconBlockHeader{},
					Signature: make([]byte, field_params.DilithiumSignatureLength+1),
				},
			},
			want: false,
		},
		{
			name: "empty signature",
			args: args{
				header: &qrysmpb.SignedBeaconBlockHeader{Header: &qrysmpb.BeaconBlockHeader{}},
			},
			want: false,
		},
		{
			name: "OK",
			args: args{
				header: &qrysmpb.SignedBeaconBlockHeader{Header: &qrysmpb.BeaconBlockHeader{}, Signature: fakeSig},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := validateBlockHeaderIntegrity(tt.args.header); got != tt.want {
				t.Errorf("validateBlockHeaderIntegrity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_isDoubleProposal(t *testing.T) {
	type args struct {
		incomingSigningRoot [32]byte
		existingSigningRoot [32]byte
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Existing signing root empty returns false",
			args: args{
				incomingSigningRoot: [32]byte{1},
				existingSigningRoot: params.BeaconConfig().ZeroHash,
			},
			want: false,
		},
		{
			name: "Existing signing root non-empty and equal to incoming returns false",
			args: args{
				incomingSigningRoot: [32]byte{1},
				existingSigningRoot: [32]byte{1},
			},
			want: false,
		},
		{
			name: "Existing signing root non-empty and not-equal to incoming returns true",
			args: args{
				incomingSigningRoot: [32]byte{1},
				existingSigningRoot: [32]byte{2},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isDoubleProposal(tt.args.incomingSigningRoot, tt.args.existingSigningRoot); got != tt.want {
				t.Errorf("isDoubleProposal() = %v, want %v", got, tt.want)
			}
		})
	}
}
