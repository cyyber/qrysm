package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/apimiddleware"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
	"github.com/theQRL/qrysm/v4/validator/client/beacon-api/mock"
	test_helpers "github.com/theQRL/qrysm/v4/validator/client/beacon-api/test-helpers"
)

func TestSubmitSignedAggregateSelectionProof_Valid(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signedAggregateAndProof := generateSignedAggregateAndProofJson()
	marshalledSignedAggregateSignedAndProof, err := json.Marshal([]*apimiddleware.SignedAggregateAttestationAndProofJson{jsonifySignedAggregateAndProof(signedAggregateAndProof)})
	require.NoError(t, err)

	ctx := context.Background()

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		ctx,
		"/zond/v1/validator/aggregate_and_proofs",
		nil,
		bytes.NewBuffer(marshalledSignedAggregateSignedAndProof),
		nil,
	).Return(
		nil,
		nil,
	).Times(1)

	attestationDataRoot, err := signedAggregateAndProof.Message.Aggregate.Data.HashTreeRoot()
	require.NoError(t, err)

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	resp, err := validatorClient.submitSignedAggregateSelectionProof(ctx, &zondpb.SignedAggregateSubmitRequest{
		SignedAggregateAndProof: signedAggregateAndProof,
	})
	require.NoError(t, err)
	assert.DeepEqual(t, attestationDataRoot[:], resp.AttestationDataRoot)
}

func TestSubmitSignedAggregateSelectionProof_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	signedAggregateAndProof := generateSignedAggregateAndProofJson()
	marshalledSignedAggregateSignedAndProof, err := json.Marshal([]*apimiddleware.SignedAggregateAttestationAndProofJson{jsonifySignedAggregateAndProof(signedAggregateAndProof)})
	require.NoError(t, err)

	ctx := context.Background()
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		ctx,
		"/zond/v1/validator/aggregate_and_proofs",
		nil,
		bytes.NewBuffer(marshalledSignedAggregateSignedAndProof),
		nil,
	).Return(
		nil,
		errors.New("bad request"),
	).Times(1)

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	_, err = validatorClient.submitSignedAggregateSelectionProof(ctx, &zondpb.SignedAggregateSubmitRequest{
		SignedAggregateAndProof: signedAggregateAndProof,
	})
	assert.ErrorContains(t, "failed to send POST data to REST endpoint", err)
	assert.ErrorContains(t, "bad request", err)
}

func generateSignedAggregateAndProofJson() *zondpb.SignedAggregateAttestationAndProof {
	return &zondpb.SignedAggregateAttestationAndProof{
		Message: &zondpb.AggregateAttestationAndProof{
			AggregatorIndex: 72,
			Aggregate: &zondpb.Attestation{
				ParticipationBits: test_helpers.FillByteSlice(4, 74),
				Data: &zondpb.AttestationData{
					Slot:            75,
					CommitteeIndex:  76,
					BeaconBlockRoot: test_helpers.FillByteSlice(32, 38),
					Source: &zondpb.Checkpoint{
						Epoch: 78,
						Root:  test_helpers.FillByteSlice(32, 79),
					},
					Target: &zondpb.Checkpoint{
						Epoch: 80,
						Root:  test_helpers.FillByteSlice(32, 81),
					},
				},
				Signatures: test_helpers.FillByteSlice(96, 82),
			},
			SelectionProof: test_helpers.FillByteSlice(96, 82),
		},
		Signature: test_helpers.FillByteSlice(96, 82),
	}
}
