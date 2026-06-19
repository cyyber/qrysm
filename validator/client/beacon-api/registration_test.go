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
	test_helpers "github.com/theQRL/qrysm/validator/client/beacon-api/test-helpers"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestRegistration_Valid(t *testing.T) {
	const feeRecipient1 = "Q0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const feeRecipient2 = "Qfedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210fedcba9876543210"
	const feeRecipient3 = "Q00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff00112233445566778899aabbccddeeff"
	pubKey1 := test_helpers.FillEncodedPubkey(1)
	signature1 := test_helpers.FillEncodedSignature(11)
	pubKey2 := test_helpers.FillEncodedPubkey(2)
	signature2 := test_helpers.FillEncodedSignature(12)
	pubKey3 := test_helpers.FillEncodedPubkey(3)
	signature3 := test_helpers.FillEncodedSignature(13)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonRegistrations := []*shared.SignedValidatorRegistration{
		{
			Message: &shared.ValidatorRegistration{
				FeeRecipient: feeRecipient1,
				GasLimit:     "100",
				Timestamp:    "1000",
				Pubkey:       pubKey1,
			},
			Signature: signature1,
		},
		{
			Message: &shared.ValidatorRegistration{
				FeeRecipient: feeRecipient2,
				GasLimit:     "200",
				Timestamp:    "2000",
				Pubkey:       pubKey2,
			},
			Signature: signature2,
		},
		{
			Message: &shared.ValidatorRegistration{
				FeeRecipient: feeRecipient3,
				GasLimit:     "300",
				Timestamp:    "3000",
				Pubkey:       pubKey3,
			},
			Signature: signature3,
		},
	}

	marshalledJsonRegistrations, err := json.Marshal(jsonRegistrations)
	require.NoError(t, err)

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		context.Background(),
		"/qrl/v1/validator/register_validator",
		nil,
		bytes.NewBuffer(marshalledJsonRegistrations),
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

	decodedPubkey1, err := hexutil.Decode(pubKey1)
	require.NoError(t, err)
	decodedPubkey2, err := hexutil.Decode(pubKey2)
	require.NoError(t, err)
	decodedPubkey3, err := hexutil.Decode(pubKey3)
	require.NoError(t, err)

	decodedSignature1, err := hexutil.Decode(signature1)
	require.NoError(t, err)
	decodedSignature2, err := hexutil.Decode(signature2)
	require.NoError(t, err)
	decodedSignature3, err := hexutil.Decode(signature3)
	require.NoError(t, err)

	protoRegistrations := qrysmpb.SignedValidatorRegistrationsV1{
		Messages: []*qrysmpb.SignedValidatorRegistrationV1{
			{
				Message: &qrysmpb.ValidatorRegistrationV1{
					FeeRecipient: decodedFeeRecipient1,
					GasLimit:     100,
					Timestamp:    1000,
					Pubkey:       decodedPubkey1,
				},
				Signature: decodedSignature1,
			},
			{
				Message: &qrysmpb.ValidatorRegistrationV1{
					FeeRecipient: decodedFeeRecipient2,
					GasLimit:     200,
					Timestamp:    2000,
					Pubkey:       decodedPubkey2,
				},
				Signature: decodedSignature2,
			},
			{
				Message: &qrysmpb.ValidatorRegistrationV1{
					FeeRecipient: decodedFeeRecipient3,
					GasLimit:     300,
					Timestamp:    3000,
					Pubkey:       decodedPubkey3,
				},
				Signature: decodedSignature3,
			},
		},
	}

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	res, err := validatorClient.SubmitValidatorRegistrations(context.Background(), &protoRegistrations)

	assert.DeepEqual(t, new(emptypb.Empty), res)
	require.NoError(t, err)
}

func TestRegistration_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		context.Background(),
		"/qrl/v1/validator/register_validator",
		nil,
		gomock.Any(),
		nil,
	).Return(
		nil,
		errors.New("foo error"),
	).Times(1)

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	_, err := validatorClient.SubmitValidatorRegistrations(context.Background(), &qrysmpb.SignedValidatorRegistrationsV1{})
	assert.ErrorContains(t, "failed to send POST data to `/qrl/v1/validator/register_validator` REST endpoint", err)
	assert.ErrorContains(t, "foo error", err)
}
