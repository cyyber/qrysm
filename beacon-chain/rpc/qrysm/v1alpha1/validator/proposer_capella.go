package validator

import (
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

// Sets the dilithium to exec data for a block.
func (vs *Server) setDilithiumToExecData(blk interfaces.SignedBeaconBlock, headState state.BeaconState) {
	if err := blk.SetDilithiumToExecutionChanges([]*qrysmpb.SignedDilithiumToExecutionChange{}); err != nil {
		log.WithError(err).Error("Could not set dilithium to execution data in block")
		return
	}
	changes, err := vs.DilithiumChangesPool.DilithiumToExecChangesForInclusion(headState)
	if err != nil {
		log.WithError(err).Error("Could not get dilithium to execution changes")
		return
	} else {
		if err := blk.SetDilithiumToExecutionChanges(changes); err != nil {
			log.WithError(err).Error("Could not set dilithium to execution changes")
			return
		}
	}
}
