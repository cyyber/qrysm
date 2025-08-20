package state_native

import (
	"github.com/theQRL/qrysm/beacon-chain/state/state-native/types"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

// SetCurrentSyncCommittee for the beacon state.
func (b *BeaconState) SetCurrentSyncCommittee(val *qrysmpb.SyncCommittee) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.currentSyncCommittee = val
	b.markFieldAsDirty(types.CurrentSyncCommittee)
	return nil
}

// SetNextSyncCommittee for the beacon state.
func (b *BeaconState) SetNextSyncCommittee(val *qrysmpb.SyncCommittee) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	b.nextSyncCommittee = val
	b.markFieldAsDirty(types.NextSyncCommittee)
	return nil
}
