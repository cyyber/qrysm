package genesis

import (
	_ "embed"
	"fmt"

	"github.com/golang/snappy"
	"github.com/theQRL/qrysm/v4/beacon-chain/state"
	state_native "github.com/theQRL/qrysm/v4/beacon-chain/state/state-native"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

var embeddedStates = map[string]*[]byte{}

// State returns a copy of the genesis state from a hardcoded value.
func State(name string) (state.BeaconState, error) {
	sb, exists := embeddedStates[name]
	if exists {
		return load(*sb)
	}
	return nil, nil
}

// load a compressed ssz state file into a beacon state struct.
func load(b []byte) (state.BeaconState, error) {
	st := &zondpb.BeaconState{}
	b, err := snappy.Decode(nil /*dst*/, b)
	if err != nil {
		return nil, err
	}
	if err := st.UnmarshalSSZ(b); err != nil {
		fmt.Println("UnmarshalSSZ")
		return nil, err
	}
	//return state_native.InitializeFromProtoUnsafePhase0(st)
	return state_native.InitializeFromProtoUnsafeCapella(st)
}
