package state_native

import (
	"github.com/theQRL/qrysm/v4/config/params"
	"github.com/theQRL/qrysm/v4/runtime/version"
)

func (b *BeaconState) ProportionalSlashingMultiplier() (uint64, error) {
	switch b.version {
	case version.Capella:
		return params.BeaconConfig().ProportionalSlashingMultiplierBellatrix, nil
	}
	return 0, errNotSupported("ProportionalSlashingMultiplier()", b.version)
}

func (b *BeaconState) InactivityPenaltyQuotient() (uint64, error) {
	switch b.version {
	case version.Capella:
		return params.BeaconConfig().InactivityPenaltyQuotientBellatrix, nil
	}
	return 0, errNotSupported("InactivityPenaltyQuotient()", b.version)
}
