package forks

import (
	"math"
	"reflect"
	"testing"
	"time"

	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
)

func TestFork(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()

	tests := []struct {
		name        string
		targetEpoch primitives.Epoch
		want        *zondpb.Fork
		wantErr     bool
		setConfg    func()
	}{
		{
			name:        "no fork",
			targetEpoch: 0,
			want: &zondpb.Fork{
				Epoch:           0,
				CurrentVersion:  []byte{'A', 'B', 'C', 'D'},
				PreviousVersion: []byte{'A', 'B', 'C', 'D'},
			},
			wantErr: false,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{}
				params.OverrideBeaconConfig(cfg)
			},
		},
		{
			name:        "genesis fork",
			targetEpoch: 0,
			want: &zondpb.Fork{
				Epoch:           0,
				CurrentVersion:  []byte{'A', 'B', 'C', 'D'},
				PreviousVersion: []byte{'A', 'B', 'C', 'D'},
			},
			wantErr: false,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
					{'A', 'B', 'C', 'D'}: 0,
				}
				params.OverrideBeaconConfig(cfg)
			},
		},
		{
			name:        "3 forks, pre-fork",
			targetEpoch: 20,
			want: &zondpb.Fork{
				Epoch:           10,
				CurrentVersion:  []byte{'A', 'B', 'C', 'F'},
				PreviousVersion: []byte{'A', 'B', 'C', 'D'},
			},
			wantErr: false,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
					{'A', 'B', 'C', 'D'}: 0,
					{'A', 'B', 'C', 'F'}: 10,
					{'A', 'B', 'C', 'Z'}: 100,
				}
				params.OverrideBeaconConfig(cfg)
			},
		},
		{
			name:        "3 forks, on fork",
			targetEpoch: 100,
			want: &zondpb.Fork{
				Epoch:           100,
				CurrentVersion:  []byte{'A', 'B', 'C', 'Z'},
				PreviousVersion: []byte{'A', 'B', 'C', 'F'},
			},
			wantErr: false,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
					{'A', 'B', 'C', 'D'}: 0,
					{'A', 'B', 'C', 'F'}: 10,
					{'A', 'B', 'C', 'Z'}: 100,
				}
				params.OverrideBeaconConfig(cfg)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setConfg()
			got, err := Fork(tt.targetEpoch)
			if (err != nil) != tt.wantErr {
				t.Errorf("Fork() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Fork() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRetrieveForkDataFromDigest(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
	cfg.GenesisEpoch = 0

	cfg.InitializeForkSchedule()
	params.OverrideBeaconConfig(cfg)
	genValRoot := [32]byte{'A', 'B', 'C', 'D'}
	digest, err := signing.ComputeForkDigest([]byte{'A', 'B', 'C', 'D'}, genValRoot[:])
	assert.NoError(t, err)

	version, epoch, err := RetrieveForkDataFromDigest(digest, genValRoot[:])
	assert.NoError(t, err)
	assert.Equal(t, [4]byte{'A', 'B', 'C', 'D'}, version)
	assert.Equal(t, epoch, primitives.Epoch(0))

}

func TestIsForkNextEpoch(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
	cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
		{'A', 'B', 'C', 'D'}: 0,
		{'A', 'B', 'C', 'F'}: 10,
		{'A', 'B', 'C', 'Z'}: 100,
	}
	params.OverrideBeaconConfig(cfg)
	genTimeCreator := func(epoch primitives.Epoch) time.Time {
		return time.Now().Add(-time.Duration(uint64(params.BeaconConfig().SlotsPerEpoch)*uint64(epoch)*params.BeaconConfig().SecondsPerSlot) * time.Second)
	}
	// Is at Fork Epoch
	genesisTime := genTimeCreator(10)
	genRoot := [32]byte{'A'}

	isFork, err := IsForkNextEpoch(genesisTime, genRoot[:])
	assert.NoError(t, err)
	assert.Equal(t, false, isFork)

	// Is right before fork epoch
	genesisTime = genTimeCreator(9)

	isFork, err = IsForkNextEpoch(genesisTime, genRoot[:])
	assert.NoError(t, err)
	assert.Equal(t, true, isFork)

	// Is at fork epoch
	genesisTime = genTimeCreator(100)

	isFork, err = IsForkNextEpoch(genesisTime, genRoot[:])
	assert.NoError(t, err)
	assert.Equal(t, false, isFork)

	genesisTime = genTimeCreator(99)

	// Is right before fork epoch.
	isFork, err = IsForkNextEpoch(genesisTime, genRoot[:])
	assert.NoError(t, err)
	assert.Equal(t, true, isFork)
}

func TestNextForkData(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	cfg := params.BeaconConfig().Copy()
	tests := []struct {
		name              string
		setConfg          func()
		currEpoch         primitives.Epoch
		wantedForkVerison [4]byte
		wantedEpoch       primitives.Epoch
	}{
		{
			name:              "genesis fork",
			currEpoch:         0,
			wantedForkVerison: [4]byte{'A', 'B', 'C', 'D'},
			wantedEpoch:       math.MaxUint64,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
					{'A', 'B', 'C', 'D'}: 0,
				}
				params.OverrideBeaconConfig(cfg)
			},
		},
		{
			name:              "3 forks, pre-fork, 1st fork",
			currEpoch:         5,
			wantedForkVerison: [4]byte{'A', 'B', 'C', 'F'},
			wantedEpoch:       10,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
					{'A', 'B', 'C', 'D'}: 0,
					{'A', 'B', 'C', 'F'}: 10,
					{'A', 'B', 'C', 'Z'}: 100,
				}
				params.OverrideBeaconConfig(cfg)
			},
		},
		{
			name:              "3 forks, pre-fork, 2nd fork",
			currEpoch:         50,
			wantedForkVerison: [4]byte{'A', 'B', 'C', 'Z'},
			wantedEpoch:       100,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
					{'A', 'B', 'C', 'D'}: 0,
					{'A', 'B', 'C', 'F'}: 10,
					{'A', 'B', 'C', 'Z'}: 100,
				}
				params.OverrideBeaconConfig(cfg)
			},
		},
		{
			name:              "3 forks, on fork",
			currEpoch:         100,
			wantedForkVerison: [4]byte{'A', 'B', 'C', 'Z'},
			wantedEpoch:       math.MaxUint64,
			setConfg: func() {
				cfg = cfg.Copy()
				cfg.GenesisForkVersion = []byte{'A', 'B', 'C', 'D'}
				cfg.ForkVersionSchedule = map[[4]byte]primitives.Epoch{
					{'A', 'B', 'C', 'D'}: 0,
					{'A', 'B', 'C', 'F'}: 10,
					{'A', 'B', 'C', 'Z'}: 100,
				}
				params.OverrideBeaconConfig(cfg)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.setConfg()
			fVersion, fEpoch, err := NextForkData(tt.currEpoch)
			assert.NoError(t, err)
			if fVersion != tt.wantedForkVerison {
				t.Errorf("NextForkData() fork version = %v, want %v", fVersion, tt.wantedForkVerison)
			}
			if fEpoch != tt.wantedEpoch {
				t.Errorf("NextForkData() fork epoch = %v, want %v", fEpoch, tt.wantedEpoch)
			}
		})
	}
}
