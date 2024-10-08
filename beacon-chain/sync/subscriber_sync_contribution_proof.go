package sync

import (
	"context"
	"errors"
	"fmt"

	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"google.golang.org/protobuf/proto"
)

// syncContributionAndProofSubscriber forwards the incoming validated sync contributions and proof to the
// contribution pool for processing.
// skipcq: SCC-U1000
func (s *Service) syncContributionAndProofSubscriber(_ context.Context, msg proto.Message) error {
	sContr, ok := msg.(*zondpb.SignedContributionAndProof)
	if !ok {
		return fmt.Errorf("message was not type *zondpb.SignedContributionAndProof, type=%T", msg)
	}

	if sContr.Message == nil || sContr.Message.Contribution == nil {
		return errors.New("nil contribution")
	}

	return s.cfg.syncCommsPool.SaveSyncCommitteeContribution(sContr.Message.Contribution)
}
