package beacon_api

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/qrysm/beacon-chain/rpc/qrl/beacon"
	"github.com/theQRL/qrysm/config/params"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/validator/client/beacon-api/mock"
	test_helpers "github.com/theQRL/qrysm/validator/client/beacon-api/test-helpers"
)

func TestComputeWaitElements_LastRecvTimeZero(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	lastRecvTime := time.Time{}

	waitDuration, nextRecvTime := computeWaitElements(now, lastRecvTime)

	assert.Equal(t, time.Duration(0), waitDuration)
	assert.Equal(t, now, nextRecvTime)
}

func TestComputeWaitElements_LastRecvTimeNotZero(t *testing.T) {
	delay := 10
	now := time.Date(2022, 1, 1, 0, 0, delay, 0, time.UTC)
	lastRecvTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)
	secondsPerSlot := params.BeaconConfig().SecondsPerSlot

	waitDuration, nextRecvTime := computeWaitElements(now, lastRecvTime)

	assert.Equal(t, time.Duration(secondsPerSlot-uint64(delay))*time.Second, waitDuration)
	assert.Equal(t, time.Date(2022, 1, 1, 0, 0, int(secondsPerSlot), 0, time.UTC), nextRecvTime)
}

func TestComputeWaitElements_Longest(t *testing.T) {
	now := time.Date(2022, 1, 1, 0, 1, 0, 0, time.UTC)
	lastRecvTime := time.Date(2022, 1, 1, 0, 0, 0, 0, time.UTC)

	waitDuration, nextRecvTime := computeWaitElements(now, lastRecvTime)

	assert.Equal(t, 0*time.Second, waitDuration)
	assert.Equal(t, now, nextRecvTime)
}

func TestActivation_Nominal(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stringPubKeys := []string{
		test_helpers.FillEncodedPubkey(1), // active_ongoing
		test_helpers.FillEncodedPubkey(2), // active_exiting
		test_helpers.FillEncodedPubkey(3), // does not exist
		test_helpers.FillEncodedPubkey(4), // exited_slashed
	}

	pubKeys := make([][]byte, len(stringPubKeys))

	url := strings.Join([]string{
		"/qrl/v1/beacon/states/head/validators?",
		"id=" + stringPubKeys[0] + "&",
		"id=" + stringPubKeys[1] + "&",
		"id=" + stringPubKeys[2] + "&",
		"id=" + stringPubKeys[3],
	}, "")

	for i, stringPubKey := range stringPubKeys {
		pubKey, err := hexutil.Decode(stringPubKey)
		require.NoError(t, err)

		pubKeys[i] = pubKey
	}

	wantedStatuses := []*qrysmpb.ValidatorActivationResponse_Status{
		{
			PublicKey: pubKeys[0],
			Index:     55293,
			Status: &qrysmpb.ValidatorStatusResponse{
				Status: qrysmpb.ValidatorStatus_ACTIVE,
			},
		},
		{
			PublicKey: pubKeys[1],
			Index:     11877,
			Status: &qrysmpb.ValidatorStatusResponse{
				Status: qrysmpb.ValidatorStatus_EXITING,
			},
		},
		{
			PublicKey: pubKeys[3],
			Index:     210439,
			Status: &qrysmpb.ValidatorStatusResponse{
				Status: qrysmpb.ValidatorStatus_EXITED,
			},
		},
		{
			PublicKey: pubKeys[2],
			Index:     18446744073709551615,
			Status: &qrysmpb.ValidatorStatusResponse{
				Status: qrysmpb.ValidatorStatus_UNKNOWN_STATUS,
			},
		},
	}

	stateValidatorsResponseJson := beacon.GetValidatorsResponse{}

	// Instantiate a cancellable context.
	ctx, cancel := context.WithCancel(context.Background())

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	// GetRestJsonResponse does not return any result for non existing key
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
						Pubkey: stringPubKeys[0],
					},
				},
				{
					Index:  "11877",
					Status: "active_exiting",
					Validator: &beacon.Validator{
						Pubkey: stringPubKeys[1],
					},
				},
				{
					Index:  "210439",
					Status: "exited_slashed",
					Validator: &beacon.Validator{
						Pubkey: stringPubKeys[3],
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

	waitForActivation, err := validatorClient.WaitForActivation(
		ctx,
		&qrysmpb.ValidatorActivationRequest{
			PublicKeys: pubKeys,
		},
	)
	assert.NoError(t, err)

	// This first call to `Recv` should return immediately
	resp, err := waitForActivation.Recv()
	require.NoError(t, err)
	assert.DeepEqual(t, wantedStatuses, resp.Statuses)

	// Cancel the context after 1 second
	go func(ctx context.Context) {
		time.Sleep(time.Second)
		cancel()
	}(ctx)

	// This second call to `Recv` should return after ~12 seconds, but is interrupted by the cancel
	_, err = waitForActivation.Recv()

	assert.ErrorContains(t, "context canceled", err)
}

func TestActivation_InvalidData(t *testing.T) {
	testCases := []struct {
		name                 string
		data                 []*beacon.ValidatorContainer
		expectedErrorMessage string
	}{
		{
			name: "bad validator public key",
			data: []*beacon.ValidatorContainer{
				{
					Index:  "55293",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey: "NotAPubKey",
					},
				},
			},
			expectedErrorMessage: "failed to parse validator public key",
		},
		{
			name: "bad validator index",
			data: []*beacon.ValidatorContainer{
				{
					Index:  "NotAnIndex",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey: stringPubKey,
					},
				},
			},
			expectedErrorMessage: "failed to parse validator index",
		},
		{
			name: "invalid validator status",
			data: []*beacon.ValidatorContainer{
				{
					Index:  "12345",
					Status: "NotAStatus",
					Validator: &beacon.Validator{
						Pubkey: stringPubKey,
					},
				},
			},
			expectedErrorMessage: "invalid validator status: NotAStatus",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name,
			func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				ctx := context.Background()

				jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
				jsonRestHandler.EXPECT().GetRestJsonResponse(
					ctx,
					gomock.Any(),
					gomock.Any(),
				).Return(
					nil,
					nil,
				).SetArg(
					2,
					beacon.GetValidatorsResponse{
						Data: testCase.data,
					},
				).Times(1)

				validatorClient := beaconApiValidatorClient{
					stateValidatorsProvider: beaconApiStateValidatorsProvider{
						jsonRestHandler: jsonRestHandler,
					},
				}

				waitForActivation, err := validatorClient.WaitForActivation(
					ctx,
					&qrysmpb.ValidatorActivationRequest{},
				)
				assert.NoError(t, err)

				_, err = waitForActivation.Recv()
				assert.ErrorContains(t, testCase.expectedErrorMessage, err)
			},
		)
	}
}

func TestActivation_JsonResponseError(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().GetRestJsonResponse(
		ctx,
		gomock.Any(),
		gomock.Any(),
	).Return(
		nil,
		errors.New("some specific json error"),
	).Times(1)

	validatorClient := beaconApiValidatorClient{
		stateValidatorsProvider: beaconApiStateValidatorsProvider{
			jsonRestHandler: jsonRestHandler,
		},
	}

	waitForActivation, err := validatorClient.WaitForActivation(
		ctx,
		&qrysmpb.ValidatorActivationRequest{},
	)
	assert.NoError(t, err)

	_, err = waitForActivation.Recv()
	assert.ErrorContains(t, "failed to get state validators", err)
}
