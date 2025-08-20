package validator

import (
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func (vs *Server) getExits(head state.BeaconState, slot primitives.Slot) []*qrysmpb.SignedVoluntaryExit {
	exits, err := vs.ExitPool.ExitsForInclusion(head, slot)
	if err != nil {
		log.WithError(err).Error("Could not get exits")
		return []*qrysmpb.SignedVoluntaryExit{}
	}
	return exits
}
