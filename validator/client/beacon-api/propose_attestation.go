package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/pkg/errors"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
)

func (c beaconApiValidatorClient) proposeAttestation(ctx context.Context, attestation *qrysmpb.Attestation) (*qrysmpb.AttestResponse, error) {
	if err := checkNilAttestation(attestation); err != nil {
		return nil, err
	}

	marshalledAttestation, err := json.Marshal(jsonifyAttestations([]*qrysmpb.Attestation{attestation}))
	if err != nil {
		return nil, err
	}

	if _, err := c.jsonRestHandler.PostRestJson(ctx, "/qrl/v1/beacon/pool/attestations", nil, bytes.NewBuffer(marshalledAttestation), nil); err != nil {
		return nil, errors.Wrap(err, "failed to send POST data to REST endpoint")
	}

	attestationDataRoot, err := attestation.Data.HashTreeRoot()
	if err != nil {
		return nil, errors.Wrap(err, "failed to compute attestation data root")
	}

	return &qrysmpb.AttestResponse{AttestationDataRoot: attestationDataRoot[:]}, nil
}

// checkNilAttestation returns error if attestation or any field of attestation is nil.
func checkNilAttestation(attestation *qrysmpb.Attestation) error {
	if attestation == nil {
		return errors.New("attestation is nil")
	}

	if attestation.Data == nil {
		return errors.New("attestation data is nil")
	}

	if attestation.Data.Source == nil || attestation.Data.Target == nil {
		return errors.New("source/target in attestation data is nil")
	}

	if len(attestation.AggregationBits) == 0 {
		return errors.New("attestation aggregation bits is empty")
	}

	if len(attestation.Signatures) == 0 {
		return errors.New("attestation signatures slice is empty")
	}

	return nil
}
