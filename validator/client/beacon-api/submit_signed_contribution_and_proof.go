package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"strconv"

	"github.com/pkg/errors"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/apimiddleware"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

func (c beaconApiValidatorClient) submitSignedContributionAndProof(ctx context.Context, in *zondpb.SignedContributionAndProof) error {
	if in == nil {
		return errors.New("signed contribution and proof is nil")
	}

	if in.Message == nil {
		return errors.New("signed contribution and proof message is nil")
	}

	if in.Message.Contribution == nil {
		return errors.New("signed contribution and proof contribution is nil")
	}

	signatures := make([]string, len(in.Message.Contribution.Signatures))
	for i, sig := range in.Message.Contribution.Signatures {
		signatures[i] = hexutil.Encode(sig)
	}

	signaturesIdxToParticipationIdx := make([]string, len(in.Message.Contribution.SignaturesIdxToParticipationIdx))
	for i, participationIdx := range in.Message.Contribution.SignaturesIdxToParticipationIdx {
		signaturesIdxToParticipationIdx[i] = uint64ToString(participationIdx)
	}

	jsonContributionAndProofs := []apimiddleware.SignedContributionAndProofJson{
		{
			Message: &apimiddleware.ContributionAndProofJson{
				AggregatorIndex: strconv.FormatUint(uint64(in.Message.AggregatorIndex), 10),
				Contribution: &apimiddleware.SyncCommitteeContributionJson{
					Slot:                            strconv.FormatUint(uint64(in.Message.Contribution.Slot), 10),
					BeaconBlockRoot:                 hexutil.Encode(in.Message.Contribution.BlockRoot),
					SubcommitteeIndex:               strconv.FormatUint(in.Message.Contribution.SubcommitteeIndex, 10),
					ParticipationBits:               hexutil.Encode(in.Message.Contribution.ParticipationBits),
					Signatures:                      signatures,
					SignaturesIdxToParticipationIdx: signaturesIdxToParticipationIdx,
				},
				SelectionProof: hexutil.Encode(in.Message.SelectionProof),
			},
			Signature: hexutil.Encode(in.Signature),
		},
	}

	jsonContributionAndProofsBytes, err := json.Marshal(jsonContributionAndProofs)
	if err != nil {
		return errors.Wrap(err, "failed to marshall signed contribution and proof")
	}

	if _, err := c.jsonRestHandler.PostRestJson(ctx, "/zond/v1/validator/contribution_and_proofs", nil, bytes.NewBuffer(jsonContributionAndProofsBytes), nil); err != nil {
		return errors.Wrap(err, "failed to send POST data to REST endpoint")
	}

	return nil
}
