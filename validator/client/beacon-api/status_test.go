package beacon_api

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/qrysm/beacon-chain/rpc/qrl/beacon"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/validator/client/beacon-api/mock"
	test_helpers "github.com/theQRL/qrysm/validator/client/beacon-api/test-helpers"
)

var (
	statusPubKey1 = test_helpers.FillEncodedPubkey(1)
	statusPubKey2 = test_helpers.FillEncodedPubkey(2)
	statusPubKey3 = test_helpers.FillEncodedPubkey(3)
	statusPubKey4 = test_helpers.FillEncodedPubkey(4)
	statusPubKey5 = test_helpers.FillEncodedPubkey(5)
	statusPubKey6 = test_helpers.FillEncodedPubkey(6)
	statusPubKey7 = test_helpers.FillEncodedPubkey(7)
	statusPubKey8 = test_helpers.FillEncodedPubkey(8)
	statusPubKey9 = test_helpers.FillEncodedPubkey(9)
)

func TestValidatorStatus_Nominal(t *testing.T) {
	stringValidatorPubKey := statusPubKey2
	validatorPubKey, err := hexutil.Decode(stringValidatorPubKey)
	require.NoError(t, err)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	stateValidatorsProvider := mock.NewMockstateValidatorsProvider(ctrl)

	stateValidatorsProvider.EXPECT().GetStateValidators(
		ctx,
		[]string{stringValidatorPubKey},
		nil,
		nil,
	).Return(
		&beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{
				{
					Index:  "35000",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey:          stringValidatorPubKey,
						ActivationEpoch: "56",
					},
				},
			},
		},
		nil,
	).Times(1)

	validatorClient := beaconApiValidatorClient{stateValidatorsProvider: stateValidatorsProvider}

	actualValidatorStatusResponse, err := validatorClient.ValidatorStatus(
		ctx,
		&qrysmpb.ValidatorStatusRequest{
			PublicKey: validatorPubKey,
		},
	)

	expectedValidatorStatusResponse := qrysmpb.ValidatorStatusResponse{
		Status:          qrysmpb.ValidatorStatus_ACTIVE,
		ActivationEpoch: 56,
	}

	require.NoError(t, err)
	assert.DeepEqual(t, &expectedValidatorStatusResponse, actualValidatorStatusResponse)
}

func TestValidatorStatus_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	stateValidatorsProvider := mock.NewMockstateValidatorsProvider(ctrl)

	stateValidatorsProvider.EXPECT().GetStateValidators(
		ctx,
		gomock.Any(),
		nil,
		nil,
	).Return(
		&beacon.GetValidatorsResponse{},
		errors.New("a specific error"),
	).Times(1)

	validatorClient := beaconApiValidatorClient{stateValidatorsProvider: stateValidatorsProvider}

	_, err := validatorClient.ValidatorStatus(
		ctx,
		&qrysmpb.ValidatorStatusRequest{
			PublicKey: []byte{},
		},
	)

	require.ErrorContains(t, "failed to get validator status response", err)
}

func TestMultipleValidatorStatus_Nominal(t *testing.T) {
	stringValidatorsPubKey := []string{
		statusPubKey1, // existing
		statusPubKey2, // existing
	}

	ctx := context.Background()
	validatorsPubKey := make([][]byte, len(stringValidatorsPubKey))

	for i, stringValidatorPubKey := range stringValidatorsPubKey {
		validatorPubKey, err := hexutil.Decode(stringValidatorPubKey)
		require.NoError(t, err)
		validatorsPubKey[i] = validatorPubKey
	}

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stateValidatorsProvider := mock.NewMockstateValidatorsProvider(ctrl)

	stateValidatorsProvider.EXPECT().GetStateValidators(
		ctx,
		stringValidatorsPubKey,
		nil,
		nil,
	).Return(
		&beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{
				{
					Index:  "11111",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey1,
						ActivationEpoch: "12",
					},
				},
				{
					Index:  "22222",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey2,
						ActivationEpoch: "34",
					},
				},
			},
		},
		nil,
	).Times(1)

	validatorClient := beaconApiValidatorClient{stateValidatorsProvider: stateValidatorsProvider}

	expectedValidatorStatusResponse := qrysmpb.MultipleValidatorStatusResponse{
		PublicKeys: validatorsPubKey,
		Indices: []primitives.ValidatorIndex{
			11111,
			22222,
		},
		Statuses: []*qrysmpb.ValidatorStatusResponse{
			{
				Status:          qrysmpb.ValidatorStatus_ACTIVE,
				ActivationEpoch: 12,
			},
			{
				Status:          qrysmpb.ValidatorStatus_ACTIVE,
				ActivationEpoch: 34,
			},
		},
	}

	actualValidatorStatusResponse, err := validatorClient.MultipleValidatorStatus(
		ctx,
		&qrysmpb.MultipleValidatorStatusRequest{
			PublicKeys: validatorsPubKey,
		},
	)
	require.NoError(t, err)
	assert.DeepEqual(t, &expectedValidatorStatusResponse, actualValidatorStatusResponse)
}

func TestMultipleValidatorStatus_No_Keys(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	stateValidatorsProvider := mock.NewMockstateValidatorsProvider(ctrl)

	validatorClient := beaconApiValidatorClient{stateValidatorsProvider: stateValidatorsProvider}

	resp, err := validatorClient.MultipleValidatorStatus(
		ctx,
		&qrysmpb.MultipleValidatorStatusRequest{
			PublicKeys: [][]byte{},
		},
	)
	require.NoError(t, err)
	require.DeepEqual(t, &qrysmpb.MultipleValidatorStatusResponse{}, resp)
}

func TestGetValidatorsStatusResponse_Nominal_SomeActiveValidators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()
	stringValidatorsPubKey := []string{
		statusPubKey1, // existing
		statusPubKey2, // existing
		statusPubKey3, // NOT existing
		statusPubKey4, // existing
		statusPubKey5, // NOT existing
		statusPubKey6, // existing
	}

	validatorsPubKey := make([][]byte, len(stringValidatorsPubKey))

	for i, stringValidatorPubKey := range stringValidatorsPubKey {
		validatorPubKey, err := hexutil.Decode(stringValidatorPubKey)
		require.NoError(t, err)
		validatorsPubKey[i] = validatorPubKey
	}

	validatorsIndex := []int64{
		12345, // NOT existing
		33333, // existing
	}

	extraStringValidatorKey := statusPubKey7

	stateValidatorsProvider := mock.NewMockstateValidatorsProvider(ctrl)

	stateValidatorsProvider.EXPECT().GetStateValidators(
		ctx,
		stringValidatorsPubKey,
		validatorsIndex,
		nil,
	).Return(
		&beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{
				{
					Index:  "11111",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey1,
						ActivationEpoch: "12",
					},
				},
				{
					Index:  "22222",
					Status: "active_exiting",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey6,
						ActivationEpoch: "34",
					},
				},
				{
					Index:  "33333",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey:          extraStringValidatorKey,
						ActivationEpoch: "56",
					},
				},
				{
					Index:  "40000",
					Status: "pending_queued",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey2,
						ActivationEpoch: fmt.Sprintf("%d", params.BeaconConfig().FarFutureEpoch),
					},
				},
				{
					Index:  "50000",
					Status: "pending_queued",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey4,
						ActivationEpoch: fmt.Sprintf("%d", params.BeaconConfig().FarFutureEpoch),
					},
				},
			},
		},
		nil,
	).Times(1)

	stateValidatorsProvider.EXPECT().GetStateValidators(
		ctx,
		nil,
		nil,
		[]string{"active"},
	).Return(
		&beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{
				{
					Index:  "35000",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey8,
						ActivationEpoch: "56",
					},
				},
				{
					Index:  "39000",
					Status: "active_ongoing",
					Validator: &beacon.Validator{
						Pubkey:          statusPubKey9,
						ActivationEpoch: "56",
					},
				},
			},
		},
		nil,
	).Times(1)

	wantedStringValidatorsPubkey := []string{
		statusPubKey1,           // existing
		statusPubKey6,           // existing,
		extraStringValidatorKey, // existing,
		statusPubKey2,           // existing,
		statusPubKey4,           // existing
		statusPubKey3,           // NOT existing
		statusPubKey5,           // NOT existing
	}

	wantedValidatorsPubKey := make([][]byte, len(wantedStringValidatorsPubkey))
	for i, stringValidatorPubKey := range wantedStringValidatorsPubkey {
		validatorPubKey, err := hexutil.Decode(stringValidatorPubKey)
		require.NoError(t, err)

		wantedValidatorsPubKey[i] = validatorPubKey
	}

	wantedValidatorsIndex := []primitives.ValidatorIndex{
		11111,
		22222,
		33333,
		40000,
		50000,
		primitives.ValidatorIndex(^uint64(0)),
		primitives.ValidatorIndex(^uint64(0)),
	}

	wantedValidatorsStatusResponse := []*qrysmpb.ValidatorStatusResponse{
		{
			Status:          qrysmpb.ValidatorStatus_ACTIVE,
			ActivationEpoch: 12,
		},
		{
			Status:          qrysmpb.ValidatorStatus_EXITING,
			ActivationEpoch: 34,
		},
		{
			Status:          qrysmpb.ValidatorStatus_ACTIVE,
			ActivationEpoch: 56,
		},
		{
			Status:                    qrysmpb.ValidatorStatus_PENDING,
			ActivationEpoch:           params.BeaconConfig().FarFutureEpoch,
			PositionInActivationQueue: 1000,
		},
		{
			Status:                    qrysmpb.ValidatorStatus_PENDING,
			ActivationEpoch:           params.BeaconConfig().FarFutureEpoch,
			PositionInActivationQueue: 11000,
		},
		{
			Status:          qrysmpb.ValidatorStatus_UNKNOWN_STATUS,
			ActivationEpoch: params.BeaconConfig().FarFutureEpoch,
		},
		{
			Status:          qrysmpb.ValidatorStatus_UNKNOWN_STATUS,
			ActivationEpoch: params.BeaconConfig().FarFutureEpoch,
		},
	}

	validatorClient := beaconApiValidatorClient{stateValidatorsProvider: stateValidatorsProvider}
	actualValidatorsPubKey, actualValidatorsIndex, actualValidatorsStatusResponse, err := validatorClient.getValidatorsStatusResponse(ctx, validatorsPubKey, validatorsIndex)

	require.NoError(t, err)
	assert.DeepEqual(t, wantedValidatorsPubKey, actualValidatorsPubKey)
	assert.DeepEqual(t, wantedValidatorsIndex, actualValidatorsIndex)
	assert.DeepEqual(t, wantedValidatorsStatusResponse, actualValidatorsStatusResponse)
}

func TestGetValidatorsStatusResponse_Nominal_NoActiveValidators(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	stringValidatorPubKey := statusPubKey2
	validatorPubKey, err := hexutil.Decode(stringValidatorPubKey)
	require.NoError(t, err)

	ctx := context.Background()
	stateValidatorsProvider := mock.NewMockstateValidatorsProvider(ctrl)

	stateValidatorsProvider.EXPECT().GetStateValidators(
		ctx,
		[]string{stringValidatorPubKey},
		nil,
		nil,
	).Return(
		&beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{
				{
					Index:  "40000",
					Status: "pending_queued",
					Validator: &beacon.Validator{
						Pubkey:          stringValidatorPubKey,
						ActivationEpoch: fmt.Sprintf("%d", params.BeaconConfig().FarFutureEpoch),
					},
				},
			},
		},
		nil,
	).Times(1)

	stateValidatorsProvider.EXPECT().GetStateValidators(
		ctx,
		nil,
		nil,
		[]string{"active"},
	).Return(
		&beacon.GetValidatorsResponse{
			Data: []*beacon.ValidatorContainer{},
		},
		nil,
	).Times(1)

	wantedValidatorsPubKey := [][]byte{validatorPubKey}
	wantedValidatorsIndex := []primitives.ValidatorIndex{40000}
	wantedValidatorsStatusResponse := []*qrysmpb.ValidatorStatusResponse{
		{
			Status:                    qrysmpb.ValidatorStatus_PENDING,
			ActivationEpoch:           params.BeaconConfig().FarFutureEpoch,
			PositionInActivationQueue: 40000,
		},
	}

	validatorClient := beaconApiValidatorClient{stateValidatorsProvider: stateValidatorsProvider}
	actualValidatorsPubKey, actualValidatorsIndex, actualValidatorsStatusResponse, err := validatorClient.getValidatorsStatusResponse(ctx, wantedValidatorsPubKey, nil)

	require.NoError(t, err)
	require.NoError(t, err)
	assert.DeepEqual(t, wantedValidatorsPubKey, actualValidatorsPubKey)
	assert.DeepEqual(t, wantedValidatorsIndex, actualValidatorsIndex)
	assert.DeepEqual(t, wantedValidatorsStatusResponse, actualValidatorsStatusResponse)
}

type getStateValidatorsInterface struct {
	// Inputs
	inputStringPubKeys []string
	inputIndexes       []int64
	inputStatuses      []string

	// Outputs
	outputStateValidatorsResponseJson *beacon.GetValidatorsResponse
	outputErr                         error
}

func TestValidatorStatusResponse_InvalidData(t *testing.T) {
	stringPubKey := statusPubKey2
	pubKey, err := hexutil.Decode(stringPubKey)
	require.NoError(t, err)

	testCases := []struct {
		name string

		// Inputs
		inputPubKeys                      [][]byte
		inputIndexes                      []int64
		inputGetStateValidatorsInterfaces []getStateValidatorsInterface

		// Outputs
		outputErrMessage string
	}{
		{
			name: "failed getStateValidators",

			inputPubKeys: [][]byte{pubKey},
			inputIndexes: nil,
			inputGetStateValidatorsInterfaces: []getStateValidatorsInterface{
				{
					inputStringPubKeys: []string{stringPubKey},
					inputIndexes:       nil,
					inputStatuses:      nil,

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{},
					outputErr:                         errors.New("a specific error"),
				},
			},
			outputErrMessage: "failed to get state validators",
		},
		{
			name: "failed to parse validator public key NotAPublicKey",

			inputPubKeys: [][]byte{pubKey},
			inputIndexes: nil,
			inputGetStateValidatorsInterfaces: []getStateValidatorsInterface{
				{
					inputStringPubKeys: []string{stringPubKey},
					inputIndexes:       nil,
					inputStatuses:      nil,

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{
						Data: []*beacon.ValidatorContainer{
							{
								Validator: &beacon.Validator{
									Pubkey: "NotAPublicKey",
								},
							},
						},
					},
					outputErr: nil,
				},
			},
			outputErrMessage: "failed to parse validator public key",
		},
		{
			name: "failed to parse validator index NotAnIndex",

			inputPubKeys: [][]byte{pubKey},
			inputIndexes: nil,
			inputGetStateValidatorsInterfaces: []getStateValidatorsInterface{
				{
					inputStringPubKeys: []string{stringPubKey},
					inputIndexes:       nil,
					inputStatuses:      nil,

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{
						Data: []*beacon.ValidatorContainer{
							{
								Index: "NotAnIndex",
								Validator: &beacon.Validator{
									Pubkey: stringPubKey,
								},
							},
						},
					},
					outputErr: nil,
				},
			},
			outputErrMessage: "failed to parse validator index",
		},
		{
			name: "invalid validator status",

			inputPubKeys: [][]byte{pubKey},
			inputIndexes: nil,
			inputGetStateValidatorsInterfaces: []getStateValidatorsInterface{
				{
					inputStringPubKeys: []string{stringPubKey},
					inputIndexes:       nil,
					inputStatuses:      nil,

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{
						Data: []*beacon.ValidatorContainer{
							{
								Index:  "12345",
								Status: "NotAStatus",
								Validator: &beacon.Validator{
									Pubkey: stringPubKey,
								},
							},
						},
					},
					outputErr: nil,
				},
			},
			outputErrMessage: "invalid validator status NotAStatus",
		},
		{
			name: "failed to parse activation epoch",

			inputPubKeys: [][]byte{pubKey},
			inputIndexes: nil,
			inputGetStateValidatorsInterfaces: []getStateValidatorsInterface{
				{
					inputStringPubKeys: []string{stringPubKey},
					inputIndexes:       nil,
					inputStatuses:      nil,

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{
						Data: []*beacon.ValidatorContainer{
							{
								Index:  "12345",
								Status: "active_ongoing",
								Validator: &beacon.Validator{
									Pubkey:          stringPubKey,
									ActivationEpoch: "NotAnEpoch",
								},
							},
						},
					},
					outputErr: nil,
				},
			},
			outputErrMessage: "failed to parse activation epoch NotAnEpoch",
		},
		{
			name: "failed to get state validators",

			inputPubKeys: [][]byte{pubKey},
			inputIndexes: nil,
			inputGetStateValidatorsInterfaces: []getStateValidatorsInterface{
				{
					inputStringPubKeys: []string{stringPubKey},
					inputIndexes:       nil,
					inputStatuses:      nil,

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{
						Data: []*beacon.ValidatorContainer{
							{
								Index:  "12345",
								Status: "pending_queued",
								Validator: &beacon.Validator{
									Pubkey:          stringPubKey,
									ActivationEpoch: "10",
								},
							},
						},
					},
					outputErr: nil,
				},
				{
					inputStringPubKeys: nil,
					inputIndexes:       nil,
					inputStatuses:      []string{"active"},

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{},
					outputErr:                         errors.New("a specific error"),
				},
			},
			outputErrMessage: "failed to get state validators",
		},
		{
			name: "failed to parse last validator index",

			inputPubKeys: [][]byte{pubKey},
			inputIndexes: nil,
			inputGetStateValidatorsInterfaces: []getStateValidatorsInterface{
				{
					inputStringPubKeys: []string{stringPubKey},
					inputIndexes:       nil,
					inputStatuses:      nil,

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{
						Data: []*beacon.ValidatorContainer{
							{
								Index:  "12345",
								Status: "pending_queued",
								Validator: &beacon.Validator{
									Pubkey:          stringPubKey,
									ActivationEpoch: "10",
								},
							},
						},
					},
					outputErr: nil,
				},
				{
					inputStringPubKeys: nil,
					inputIndexes:       nil,
					inputStatuses:      []string{"active"},

					outputStateValidatorsResponseJson: &beacon.GetValidatorsResponse{
						Data: []*beacon.ValidatorContainer{
							{
								Index: "NotAnIndex",
							},
						},
					},
					outputErr: nil,
				},
			},
			outputErrMessage: "failed to parse last validator index NotAnIndex",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name,
			func(t *testing.T) {
				ctrl := gomock.NewController(t)
				defer ctrl.Finish()

				ctx := context.Background()
				stateValidatorsProvider := mock.NewMockstateValidatorsProvider(ctrl)

				for _, aa := range testCase.inputGetStateValidatorsInterfaces {
					stateValidatorsProvider.EXPECT().GetStateValidators(
						ctx,
						aa.inputStringPubKeys,
						aa.inputIndexes,
						aa.inputStatuses,
					).Return(
						aa.outputStateValidatorsResponseJson,
						aa.outputErr,
					).Times(1)
				}

				validatorClient := beaconApiValidatorClient{stateValidatorsProvider: stateValidatorsProvider}

				_, _, _, err := validatorClient.getValidatorsStatusResponse(
					ctx,
					testCase.inputPubKeys,
					testCase.inputIndexes,
				)

				assert.ErrorContains(t, testCase.outputErrMessage, err)
			},
		)
	}
}
