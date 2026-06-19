package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/pkg/errors"
	"github.com/theQRL/go-qrl/common/hexutil"
	"github.com/theQRL/qrysm/beacon-chain/rpc/apimiddleware"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/validator/client/beacon-api/mock"
	test_helpers "github.com/theQRL/qrysm/validator/client/beacon-api/test-helpers"
)

const proposeExitTestEndpoint = "/qrl/v1/beacon/pool/voluntary_exits"

func TestProposeExit_Valid(t *testing.T) {
	signature := test_helpers.FillEncodedSignature(1)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	jsonSignedVoluntaryExit := apimiddleware.SignedVoluntaryExitJson{
		Exit: &apimiddleware.VoluntaryExitJson{
			Epoch:          "1",
			ValidatorIndex: "2",
		},
		Signature: signature,
	}

	marshalledVoluntaryExit, err := json.Marshal(jsonSignedVoluntaryExit)
	require.NoError(t, err)

	ctx := context.Background()

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		ctx,
		proposeExitTestEndpoint,
		nil,
		bytes.NewBuffer(marshalledVoluntaryExit),
		nil,
	).Return(
		nil,
		nil,
	).Times(1)

	decodedSignature, err := hexutil.Decode(signature)
	require.NoError(t, err)

	protoSignedVoluntaryExit := &qrysmpb.SignedVoluntaryExit{
		Exit: &qrysmpb.VoluntaryExit{
			Epoch:          1,
			ValidatorIndex: 2,
		},
		Signature: decodedSignature,
	}

	expectedExitRoot, err := protoSignedVoluntaryExit.Exit.HashTreeRoot()
	require.NoError(t, err)

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	exitResponse, err := validatorClient.proposeExit(ctx, protoSignedVoluntaryExit)
	require.NoError(t, err)
	assert.DeepEqual(t, expectedExitRoot[:], exitResponse.ExitRoot)
}

func TestProposeExit_NilSignedVoluntaryExit(t *testing.T) {
	validatorClient := &beaconApiValidatorClient{}
	_, err := validatorClient.proposeExit(context.Background(), nil)
	assert.ErrorContains(t, "signed voluntary exit is nil", err)
}

func TestProposeExit_NilExit(t *testing.T) {
	validatorClient := &beaconApiValidatorClient{}
	_, err := validatorClient.proposeExit(context.Background(), &qrysmpb.SignedVoluntaryExit{})
	assert.ErrorContains(t, "exit is nil", err)
}

func TestProposeExit_BadRequest(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	ctx := context.Background()

	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)
	jsonRestHandler.EXPECT().PostRestJson(
		ctx,
		proposeExitTestEndpoint,
		nil,
		gomock.Any(),
		nil,
	).Return(
		nil,
		errors.New("foo error"),
	).Times(1)

	protoSignedVoluntaryExit := &qrysmpb.SignedVoluntaryExit{
		Exit: &qrysmpb.VoluntaryExit{
			Epoch:          1,
			ValidatorIndex: 2,
		},
		Signature: []byte{3},
	}

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	_, err := validatorClient.proposeExit(ctx, protoSignedVoluntaryExit)
	assert.ErrorContains(t, "failed to send POST data to REST endpoint", err)
	assert.ErrorContains(t, "foo error", err)
}
