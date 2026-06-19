package beacon_api

import (
	"context"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/qrysm/beacon-chain/rpc/qrl/beacon"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/validator/client/beacon-api/mock"
	test_helpers "github.com/theQRL/qrysm/validator/client/beacon-api/test-helpers"
)

var stringPubKey = test_helpers.FillEncodedPubkey(1)

func getPubKeyAndURL(t *testing.T) ([]byte, string) {
	baseUrl := "/qrl/v1/beacon/states/head/validators"
	url := fmt.Sprintf("%s?id=%s", baseUrl, stringPubKey)

	pubKey, err := hexutil.Decode(stringPubKey)
	require.NoError(t, err)

	return pubKey, url
}

func TestIndex_Nominal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)
	ctx := context.Background()

	stateValidatorsResponseJson := beacon.GetValidatorsResponse{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		ctx,
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		nil,
	).SetArg(
		2,
		beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{
				{
					Index:  "55293",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey: stringPubKey,
					},
				},
			},
		},
	).Times(1)

	validatorClient := beaconApiValidatorClient{
		stateValidatorsProvider: beaconApiStateValidatorsProvider{
			jsonRestHandler: jsonRestHandler,
		},
	}

	validatorIndex, err := validatorClient.ValidatorIndex(
		ctx,
		&qrysmpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	require.NoError(t, err)
	assert.Equal(t, primitives.ValidatorIndex(55293), validatorIndex.Index)
}

func TestIndex_UnexistingValidator(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)
	ctx := context.Background()

	stateValidatorsResponseJson := beacon.GetValidatorsResponse{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		ctx,
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		nil,
	).SetArg(
		2,
		beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{},
		},
	).Times(1)

	validatorClient := beaconApiValidatorClient{
		stateValidatorsProvider: beaconApiStateValidatorsProvider{
			jsonRestHandler: jsonRestHandler,
		},
	}

	_, err := validatorClient.ValidatorIndex(
		ctx,
		&qrysmpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	wanted := fmt.Sprintf("could not find validator index for public key `%s`", stringPubKey)
	assert.ErrorContains(t, wanted, err)
}

func TestIndex_BadIndexError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)
	ctx := context.Background()

	stateValidatorsResponseJson := beacon.GetValidatorsResponse{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		ctx,
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		nil,
	).SetArg(
		2,
		beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{
				{
					Index:  "This is not an index",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey: stringPubKey,
					},
				},
			},
		},
	).Times(1)

	validatorClient := beaconApiValidatorClient{
		stateValidatorsProvider: beaconApiStateValidatorsProvider{
			jsonRestHandler: jsonRestHandler,
		},
	}

	_, err := validatorClient.ValidatorIndex(
		ctx,
		&qrysmpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	assert.ErrorContains(t, "failed to parse validator index", err)
}

func TestIndex_JsonResponseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	pubKey, url := getPubKeyAndURL(t)
	ctx := context.Background()

	stateValidatorsResponseJson := beacon.GetValidatorsResponse{}
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	jsonRestHandler.EXPECT().GetRestJsonResponse(
		ctx,
		url,
		&stateValidatorsResponseJson,
	).Return(
		nil,
		errors.New("some specific json error"),
	).Times(1)

	validatorClient := beaconApiValidatorClient{
		stateValidatorsProvider: beaconApiStateValidatorsProvider{
			jsonRestHandler: jsonRestHandler,
		},
	}

	_, err := validatorClient.ValidatorIndex(
		ctx,
		&qrysmpb.ValidatorIndexRequest{
			PublicKey: pubKey,
		},
	)

	assert.ErrorContains(t, "failed to get state validator", err)
}
