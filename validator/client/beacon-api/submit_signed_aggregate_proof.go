package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/beacon-chain/rpc/apimiddleware"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func (c *beaconApiValidatorClient) submitSignedAggregateSelectionProof(ctx context.Context, in *qrysmpb.SignedAggregateSubmitRequest) (*qrysmpb.SignedAggregateSubmitResponse, error) {
	body, err := json.Marshal([]*apimiddleware.SignedAggregateAttestationAndProofJson{jsonifySignedAggregateAndProof(in.SignedAggregateAndProof)})
	if err != nil {
		return nil, errors.Wrap(err, "failed to marshal SignedAggregateAttestationAndProof")
	}

	if _, err := c.jsonRestHandler.PostRestJson(ctx, "/qrl/v1/validator/aggregate_and_proofs", nil, bytes.NewBuffer(body), nil); err != nil {
		return nil, errors.Wrap(err, "failed to send POST data to REST endpoint")
	}

	attestationDataRoot, err := in.SignedAggregateAndProof.Message.Aggregate.Data.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute attestation data root")
	}

	return &qrysmpb.SignedAggregateSubmitResponse{AttestationDataRoot: attestationDataRoot[:]}, nil
}
