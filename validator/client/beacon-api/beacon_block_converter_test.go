package beacon_api

import (
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/rpc/apimiddleware"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	test_helpers "github.com/theQRL/qrysm/validator/client/beacon-api/test-helpers"
)

func TestGetBeaconBlockConverter_CapellaValid(t *testing.T) {
	expectedBeaconBlock := test_helpers.GenerateProtoCapellaBeaconBlock()
	beaconBlockConverter := &beaconApiBeaconBlockConverter{}
	beaconBlock, err := beaconBlockConverter.ConvertRESTCapellaBlockToProto(test_helpers.GenerateJsonCapellaBeaconBlock())
	require.NoError(t, err)
	assert.DeepEqual(t, expectedBeaconBlock, beaconBlock)
}

func TestGetBeaconBlockConverter_CapellaError(t *testing.T) {
	testCases := []struct {
		name                 string
		expectedErrorMessage string
		generateData         func() *apimiddleware.BeaconBlockCapellaJson
	}{
		{
			name:                 "nil body",
			expectedErrorMessage: "block body is nil",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body = nil
				return beaconBlock
			},
		},
		{
			name:                 "nil execution data",
			expectedErrorMessage: "execution data is nil",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionData = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad slot",
			expectedErrorMessage: "failed to parse slot `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Slot = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad proposer index",
			expectedErrorMessage: "failed to parse proposer index `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.ProposerIndex = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad parent root",
			expectedErrorMessage: "failed to decode parent root `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.ParentRoot = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad state root",
			expectedErrorMessage: "failed to decode state root `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.StateRoot = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad randao reveal",
			expectedErrorMessage: "failed to decode randao reveal `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.RandaoReveal = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad deposit root",
			expectedErrorMessage: "failed to decode deposit root `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionData.DepositRoot = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad deposit count",
			expectedErrorMessage: "failed to parse deposit count `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionData.DepositCount = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad block hash",
			expectedErrorMessage: "failed to decode block hash `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionData.BlockHash = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad graffiti",
			expectedErrorMessage: "failed to decode graffiti `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.Graffiti = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad proposer slashings",
			expectedErrorMessage: "failed to get proposer slashings",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ProposerSlashings[0] = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad attester slashings",
			expectedErrorMessage: "failed to get attester slashings",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.AttesterSlashings[0] = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad attestations",
			expectedErrorMessage: "failed to get attestations",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.Attestations[0] = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad deposits",
			expectedErrorMessage: "failed to get deposits",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.Deposits[0] = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad voluntary exits",
			expectedErrorMessage: "failed to get voluntary exits",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.VoluntaryExits[0] = nil
				return beaconBlock
			},
		},
		{
			name:                 "nil sync aggregate",
			expectedErrorMessage: "sync aggregate is nil",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.SyncAggregate = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad sync committee bits",
			expectedErrorMessage: "failed to decode sync committee bits `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.SyncAggregate.SyncCommitteeBits = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad sync committee signature",
			expectedErrorMessage: "failed to decode sync committee signature `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.SyncAggregate.SyncCommitteeSignatures = []string{"bar"}
				return beaconBlock
			},
		},
		{
			name:                 "nil execution payload",
			expectedErrorMessage: "execution payload is nil",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload parent hash",
			expectedErrorMessage: "failed to decode execution payload parent hash `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.ParentHash = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload fee recipient",
			expectedErrorMessage: "failed to decode execution payload fee recipient `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.FeeRecipient = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload state root",
			expectedErrorMessage: "failed to decode execution payload state root `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.StateRoot = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload receipts root",
			expectedErrorMessage: "failed to decode execution payload receipts root `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.ReceiptsRoot = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload logs bloom",
			expectedErrorMessage: "failed to decode execution payload logs bloom `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.LogsBloom = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload prev randao",
			expectedErrorMessage: "failed to decode execution payload prev randao `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.PrevRandao = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload block number",
			expectedErrorMessage: "failed to parse execution payload block number `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.BlockNumber = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload gas limit",
			expectedErrorMessage: "failed to parse execution payload gas limit `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.GasLimit = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload gas used",
			expectedErrorMessage: "failed to parse execution payload gas used `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.GasUsed = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload timestamp",
			expectedErrorMessage: "failed to parse execution payload timestamp `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.TimeStamp = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload extra data",
			expectedErrorMessage: "failed to decode execution payload extra data `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.ExtraData = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload base fee per gas",
			expectedErrorMessage: "failed to parse execution payload base fee per gas `bar`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.BaseFeePerGas = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload block hash",
			expectedErrorMessage: "failed to decode execution payload block hash `foo`",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.BlockHash = "foo"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload transactions",
			expectedErrorMessage: "failed to get execution payload transactions",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.Transactions[0] = "bar"
				return beaconBlock
			},
		},
		{
			name:                 "bad exec payload withdrawals",
			expectedErrorMessage: "failed to get withdrawals",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.ExecutionPayload.Withdrawals[0] = nil
				return beaconBlock
			},
		},
		{
			name:                 "bad dilithium execution changes",
			expectedErrorMessage: "failed to get dilithium to execution changes",
			generateData: func() *apimiddleware.BeaconBlockCapellaJson {
				beaconBlock := test_helpers.GenerateJsonCapellaBeaconBlock()
				beaconBlock.Body.DilithiumToExecutionChanges[0] = nil
				return beaconBlock
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			beaconBlockJson := testCase.generateData()

			beaconBlockConverter := &beaconApiBeaconBlockConverter{}
			_, err := beaconBlockConverter.ConvertRESTCapellaBlockToProto(beaconBlockJson)
			assert.ErrorContains(t, testCase.expectedErrorMessage, err)
		})
	}
}
