package state_native

import (
	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/v4/beacon-chain/state/state-native/types"
	consensusblocks "github.com/theQRL/qrysm/v4/consensus-types/blocks"
	"github.com/theQRL/qrysm/v4/consensus-types/interfaces"
	enginev1 "github.com/theQRL/qrysm/v4/proto/engine/v1"
	_ "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

// SetLatestExecutionPayloadHeader for the beacon state.
func (b *BeaconState) SetLatestExecutionPayloadHeader(val interfaces.ExecutionData) error {
	b.lock.Lock()
	defer b.lock.Unlock()

	switch header := val.Proto().(type) {
	case *enginev1.ExecutionPayload:
		latest, err := consensusblocks.PayloadToHeader(val)
		if err != nil {
			return errors.Wrap(err, "could not convert payload to header")
		}
		b.latestExecutionPayloadHeader = latest
		b.markFieldAsDirty(types.LatestExecutionPayloadHeaderCapella)
		return nil
	case *enginev1.ExecutionPayloadHeader:
		b.latestExecutionPayloadHeader = header
		b.markFieldAsDirty(types.LatestExecutionPayloadHeaderCapella)
		return nil
	default:
		return errors.New("value must be an execution payload header")
	}
}
