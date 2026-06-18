package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/qrysm/beacon-chain/rpc/qrl/shared"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/validator/client/beacon-api/mock"
)

const prepareBeaconProposerTestEndpoint = "/qrl/v1/validator/prepare_beacon_proposer"

func TestPrepareBeaconProposer_Valid(t *testing.T) {
	const feeRecipient1 = "Q0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const feeRecipient2 = "Qfedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
	const feeRecipient3 = "Q00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	jsonRecipients := []*shared.FeeRecipient{
		{
			ValidatorIndex: "1",
			FeeRecipient:   feeRecipient1,
		},
		{
			ValidatorIndex: "2",
			FeeRecipient:   feeRecipient2,
		},
		{
			ValidatorIndex: "3",
			FeeRecipient:   feeRecipient3,
		},
	}

	marshalledJsonRecipients, err := json.Marshal(jsonRecipients)
	require.NoError(t, err)

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		ctx,
		prepareBeaconProposerTestEndpoint,
		nil,
		bytes.NewBuffer(marshalledJsonRecipients),
		nil,
	).Return(
		nil,
		nil,
	).Times(1)

	decodedFeeRecipient1, err := hexutil.DecodeQ(feeRecipient1)
	require.NoError(t, err)
	decodedFeeRecipient2, err := hexutil.DecodeQ(feeRecipient2)
	require.NoError(t, err)
	decodedFeeRecipient3, err := hexutil.DecodeQ(feeRecipient3)
	require.NoError(t, err)

	protoRecipients := []*qrysmpb.PrepareBeaconProposerRequest_FeeRecipientContainer{
		{
			ValidatorIndex: 1,
			FeeRecipient:   decodedFeeRecipient1,
		},
		{
			ValidatorIndex: 2,
			FeeRecipient:   decodedFeeRecipient2,
		},
		{
			ValidatorIndex: 3,
			FeeRecipient:   decodedFeeRecipient3,
		},
	}

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	err = validatorClient.prepareBeaconProposer(ctx, protoRecipients)
	require.NoError(t, err)
}

func TestPrepareBeaconProposer_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		ctx,
		prepareBeaconProposerTestEndpoint,
		nil,
		gomock.Any(),
		nil,
	).Return(
		nil,
		errors.New("foo error"),
	).Times(1)

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	err := validatorClient.prepareBeaconProposer(ctx, nil)
	assert.ErrorContains(t, "failed to send POST data to REST endpoint", err)
	assert.ErrorContains(t, "foo error", err)
}
