package beacon_api

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/apimiddleware"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
	"github.com/theQRL/qrysm/v4/validator/client/beacon-api/mock"
	test_helpers "github.com/theQRL/qrysm/v4/validator/client/beacon-api/test-helpers"
)

func TestProposeBeaconBlock_Capella(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	jsonRestHandler := mock.NewMockjsonRestHandler(ctrl)

	capellaBlock := generateSignedCapellaBlock()

	genericSignedBlock := &zondpb.GenericSignedBeaconBlock{}
	genericSignedBlock.Block = capellaBlock

	jsonCapellaBlock := &apimiddleware.SignedBeaconBlockCapellaJson{
		Signature: hexutil.Encode(capellaBlock.Capella.Signature),
		Message: &apimiddleware.BeaconBlockCapellaJson{
			ParentRoot:    hexutil.Encode(capellaBlock.Capella.Block.ParentRoot),
			ProposerIndex: uint64ToString(capellaBlock.Capella.Block.ProposerIndex),
			Slot:          uint64ToString(capellaBlock.Capella.Block.Slot),
			StateRoot:     hexutil.Encode(capellaBlock.Capella.Block.StateRoot),
			Body: &apimiddleware.BeaconBlockBodyCapellaJson{
				Attestations:      jsonifyAttestations(capellaBlock.Capella.Block.Body.Attestations),
				AttesterSlashings: jsonifyAttesterSlashings(capellaBlock.Capella.Block.Body.AttesterSlashings),
				Deposits:          jsonifyDeposits(capellaBlock.Capella.Block.Body.Deposits),
				Zond1Data:         jsonifyZond1Data(capellaBlock.Capella.Block.Body.Zond1Data),
				Graffiti:          hexutil.Encode(capellaBlock.Capella.Block.Body.Graffiti),
				ProposerSlashings: jsonifyProposerSlashings(capellaBlock.Capella.Block.Body.ProposerSlashings),
				RandaoReveal:      hexutil.Encode(capellaBlock.Capella.Block.Body.RandaoReveal),
				VoluntaryExits:    JsonifySignedVoluntaryExits(capellaBlock.Capella.Block.Body.VoluntaryExits),
				SyncAggregate:     JsonifySignedSyncAggregate(capellaBlock.Capella.Block.Body.SyncAggregate),
				ExecutionPayload: &apimiddleware.ExecutionPayloadCapellaJson{
					BaseFeePerGas: bytesutil.LittleEndianBytesToBigInt(capellaBlock.Capella.Block.Body.ExecutionPayload.BaseFeePerGas).String(),
					BlockHash:     hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.BlockHash),
					BlockNumber:   uint64ToString(capellaBlock.Capella.Block.Body.ExecutionPayload.BlockNumber),
					ExtraData:     hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.ExtraData),
					FeeRecipient:  hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.FeeRecipient),
					GasLimit:      uint64ToString(capellaBlock.Capella.Block.Body.ExecutionPayload.GasLimit),
					GasUsed:       uint64ToString(capellaBlock.Capella.Block.Body.ExecutionPayload.GasUsed),
					LogsBloom:     hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.LogsBloom),
					ParentHash:    hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.ParentHash),
					PrevRandao:    hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.PrevRandao),
					ReceiptsRoot:  hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.ReceiptsRoot),
					StateRoot:     hexutil.Encode(capellaBlock.Capella.Block.Body.ExecutionPayload.StateRoot),
					TimeStamp:     uint64ToString(capellaBlock.Capella.Block.Body.ExecutionPayload.Timestamp),
					Transactions:  jsonifyTransactions(capellaBlock.Capella.Block.Body.ExecutionPayload.Transactions),
					Withdrawals:   jsonifyWithdrawals(capellaBlock.Capella.Block.Body.ExecutionPayload.Withdrawals),
				},
				DilithiumToExecutionChanges: jsonifyDilithiumToExecutionChanges(capellaBlock.Capella.Block.Body.DilithiumToExecutionChanges),
			},
		},
	}

	marshalledBlock, err := json.Marshal(jsonCapellaBlock)
	require.NoError(t, err)

	// Make sure that what we send in the POST body is the marshalled version of the protobuf block
	headers := map[string]string{"Zond-Consensus-Version": "capella"}
	jsonRestHandler.EXPECT().PostRestJson(
		context.Background(),
		"/zond/v1/beacon/blocks",
		headers,
		bytes.NewBuffer(marshalledBlock),
		nil,
	)

	validatorClient := &beaconApiValidatorClient{jsonRestHandler: jsonRestHandler}
	proposeResponse, err := validatorClient.proposeBeaconBlock(context.Background(), genericSignedBlock)
	assert.NoError(t, err)
	require.NotNil(t, proposeResponse)

	expectedBlockRoot, err := capellaBlock.Capella.Block.HashTreeRoot()
	require.NoError(t, err)

	// Make sure that the block root is set
	assert.DeepEqual(t, expectedBlockRoot[:], proposeResponse.BlockRoot)
}

func generateSignedCapellaBlock() *zondpb.GenericSignedBeaconBlock_Capella {
	return &zondpb.GenericSignedBeaconBlock_Capella{
		Capella: &zondpb.SignedBeaconBlock{
			Block:     test_helpers.GenerateProtoCapellaBeaconBlock(),
			Signature: test_helpers.FillByteSlice(4595, 127),
		},
	}
}
