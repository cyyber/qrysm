package apimiddleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/v4/api/gateway/apimiddleware"
	zondpbv1 "github.com/theQRL/qrysm/v4/proto/zond/v1"
)

// https://ethereum.github.io/beacon-APIs/?urls.primaryName=dev#/Beacon/submitPoolBLSToExecutionChange
// expects posting a top-level array. We make it more proto-friendly by wrapping it in a struct.
func wrapDilithiumChangesArray(
	endpoint *apimiddleware.Endpoint,
	_ http.ResponseWriter,
	req *http.Request,
) (apimiddleware.RunDefault, apimiddleware.ErrorJson) {
	if _, ok := endpoint.PostRequest.(*SubmitDilithiumToExecutionChangesRequest); !ok {
		return true, nil
	}
	changes := make([]*SignedDilithiumToExecutionChangeJson, 0)
	if err := json.NewDecoder(req.Body).Decode(&changes); err != nil {
		return false, apimiddleware.InternalServerErrorWithMessage(err, "could not decode body")
	}
	j := &SubmitDilithiumToExecutionChangesRequest{Changes: changes}
	b, err := json.Marshal(j)
	if err != nil {
		return false, apimiddleware.InternalServerErrorWithMessage(err, "could not marshal wrapped body")
	}
	req.Body = io.NopCloser(bytes.NewReader(b))
	return true, nil
}

type capellaPublishBlockRequestJson struct {
	CapellaBlock *SignedBeaconBlockCapellaJson `json:"capella_block"`
}

type capellaPublishBlindedBlockRequestJson struct {
	CapellaBlock *SignedBlindedBeaconBlockCapellaJson `json:"capella_block"`
}

// setInitialPublishBlockPostRequest is triggered before we deserialize the request JSON into a struct.
// We don't know which version of the block got posted, but we can determine it from the slot.
// We know that blocks of all versions have a Message field with a Slot field,
// so we deserialize the request into a struct s, which has the right fields, to obtain the slot.
// Once we know the slot, we can determine what the PostRequest field of the endpoint should be, and we set it appropriately.
func setInitialPublishBlockPostRequest(endpoint *apimiddleware.Endpoint,
	_ http.ResponseWriter,
	req *http.Request,
) (apimiddleware.RunDefault, apimiddleware.ErrorJson) {
	s := struct {
		Slot string
	}{}

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return false, apimiddleware.InternalServerErrorWithMessage(err, "could not read body")
	}

	typeParseMap := make(map[string]json.RawMessage)
	if err := json.Unmarshal(buf, &typeParseMap); err != nil {
		return false, apimiddleware.InternalServerErrorWithMessage(err, "could not parse object")
	}
	if val, ok := typeParseMap["message"]; ok {
		if err := json.Unmarshal(val, &s); err != nil {
			return false, apimiddleware.InternalServerErrorWithMessage(err, "could not unmarshal field 'message' ")
		}
	} else if val, ok := typeParseMap["signed_block"]; ok {
		temp := struct {
			Message struct {
				Slot string
			}
		}{}
		if err := json.Unmarshal(val, &temp); err != nil {
			return false, apimiddleware.InternalServerErrorWithMessage(err, "could not unmarshal field 'signed_block' ")
		}
		s.Slot = temp.Message.Slot
	} else {
		return false, &apimiddleware.DefaultErrorJson{Message: "could not parse slot from request", Code: http.StatusInternalServerError}
	}

	endpoint.PostRequest = &SignedBeaconBlockCapellaJson{}

	req.Body = io.NopCloser(bytes.NewBuffer(buf))
	return true, nil
}

// In preparePublishedBlock we transform the PostRequest.
// gRPC expects an XXX_block field in the JSON object, but we have a message field at this point.
// We do a simple conversion depending on the type of endpoint.PostRequest
// (which was filled out previously in setInitialPublishBlockPostRequest).
func preparePublishedBlock(endpoint *apimiddleware.Endpoint, _ http.ResponseWriter, _ *http.Request) apimiddleware.ErrorJson {
	if block, ok := endpoint.PostRequest.(*SignedBeaconBlockCapellaJson); ok {
		// Prepare post request that can be properly decoded on gRPC side.
		endpoint.PostRequest = &capellaPublishBlockRequestJson{
			CapellaBlock: block,
		}
		return nil
	}
	return apimiddleware.InternalServerError(errors.New("unsupported block type"))
}

// setInitialPublishBlindedBlockPostRequest is triggered before we deserialize the request JSON into a struct.
// We don't know which version of the block got posted, but we can determine it from the slot.
// We know that blocks of all versions have a Message field with a Slot field,
// so we deserialize the request into a struct s, which has the right fields, to obtain the slot.
// Once we know the slot, we can determine what the PostRequest field of the endpoint should be, and we set it appropriately.
func setInitialPublishBlindedBlockPostRequest(endpoint *apimiddleware.Endpoint,
	_ http.ResponseWriter,
	req *http.Request,
) (apimiddleware.RunDefault, apimiddleware.ErrorJson) {
	s := struct {
		Slot string
	}{}

	buf, err := io.ReadAll(req.Body)
	if err != nil {
		return false, apimiddleware.InternalServerErrorWithMessage(err, "could not read body")
	}

	typeParseMap := make(map[string]json.RawMessage)
	if err = json.Unmarshal(buf, &typeParseMap); err != nil {
		return false, apimiddleware.InternalServerErrorWithMessage(err, "could not parse object")
	}
	if val, ok := typeParseMap["message"]; ok {
		if err = json.Unmarshal(val, &s); err != nil {
			return false, apimiddleware.InternalServerErrorWithMessage(err, "could not unmarshal field 'message' ")
		}
	} else if val, ok = typeParseMap["signed_blinded_block"]; ok {
		temp := struct {
			Message struct {
				Slot string
			}
		}{}
		if err = json.Unmarshal(val, &temp); err != nil {
			return false, apimiddleware.InternalServerErrorWithMessage(err, "could not unmarshal field 'signed_block' ")
		}
		s.Slot = temp.Message.Slot
	} else {
		return false, &apimiddleware.DefaultErrorJson{Message: "could not parse slot from request", Code: http.StatusInternalServerError}
	}

	endpoint.PostRequest = &SignedBlindedBeaconBlockCapellaJson{}

	req.Body = io.NopCloser(bytes.NewBuffer(buf))
	return true, nil
}

// In preparePublishedBlindedBlock we transform the PostRequest.
// gRPC expects either an XXX_block field in the JSON object, but we have a message field at this point.
// We do a simple conversion depending on the type of endpoint.PostRequest
// (which was filled out previously in setInitialPublishBlockPostRequest).
func preparePublishedBlindedBlock(endpoint *apimiddleware.Endpoint, _ http.ResponseWriter, _ *http.Request) apimiddleware.ErrorJson {
	if block, ok := endpoint.PostRequest.(*SignedBlindedBeaconBlockCapellaJson); ok {
		// Prepare post request that can be properly decoded on gRPC side.
		actualPostReq := &capellaPublishBlindedBlockRequestJson{
			CapellaBlock: &SignedBlindedBeaconBlockCapellaJson{
				Message:   block.Message,
				Signature: block.Signature,
			},
		}
		endpoint.PostRequest = actualPostReq
		return nil
	}
	return apimiddleware.InternalServerError(errors.New("unsupported block type"))
}

type tempSyncCommitteesResponseJson struct {
	Data *tempSyncCommitteeValidatorsJson `json:"data"`
}

type tempSyncCommitteeValidatorsJson struct {
	Validators          []string                              `json:"validators"`
	ValidatorAggregates []*tempSyncSubcommitteeValidatorsJson `json:"validator_aggregates"`
}

type tempSyncSubcommitteeValidatorsJson struct {
	Validators []string `json:"validators"`
}

// https://ethereum.github.io/beacon-APIs/?urls.primaryName=v2.0.0#/Beacon/getEpochSyncCommittees returns validator_aggregates as a nested array.
// grpc-gateway returns a struct with nested fields which we have to transform into a plain 2D array.
func prepareValidatorAggregates(body []byte, responseContainer interface{}) (apimiddleware.RunDefault, apimiddleware.ErrorJson) {
	tempContainer := &tempSyncCommitteesResponseJson{}
	if err := json.Unmarshal(body, tempContainer); err != nil {
		return false, apimiddleware.InternalServerErrorWithMessage(err, "could not unmarshal response into temp container")
	}
	container, ok := responseContainer.(*SyncCommitteesResponseJson)
	if !ok {
		return false, apimiddleware.InternalServerError(errors.New("container is not of the correct type"))
	}

	container.Data = &SyncCommitteeValidatorsJson{}
	container.Data.Validators = tempContainer.Data.Validators
	container.Data.ValidatorAggregates = make([][]string, len(tempContainer.Data.ValidatorAggregates))
	for i, srcValAgg := range tempContainer.Data.ValidatorAggregates {
		dstValAgg := make([]string, len(srcValAgg.Validators))
		copy(dstValAgg, tempContainer.Data.ValidatorAggregates[i].Validators)
		container.Data.ValidatorAggregates[i] = dstValAgg
	}

	return false, nil
}

type capellaBlockResponseJson struct {
	Version             string                        `json:"version"`
	Data                *SignedBeaconBlockCapellaJson `json:"data"`
	ExecutionOptimistic bool                          `json:"execution_optimistic"`
	Finalized           bool                          `json:"finalized"`
}

type capellaBlindedBlockResponseJson struct {
	Version             string                               `json:"version" enum:"true"`
	Data                *SignedBlindedBeaconBlockCapellaJson `json:"data"`
	ExecutionOptimistic bool                                 `json:"execution_optimistic"`
	Finalized           bool                                 `json:"finalized"`
}

func serializeBlock(response interface{}) (apimiddleware.RunDefault, []byte, apimiddleware.ErrorJson) {
	respContainer, ok := response.(*BlockResponseJson)
	if !ok {
		return false, nil, apimiddleware.InternalServerError(errors.New("container is not of the correct type"))
	}

	var actualRespContainer interface{}
	switch {
	case strings.EqualFold(respContainer.Version, strings.ToLower(zondpbv1.Version_CAPELLA.String())):
		actualRespContainer = &capellaBlockResponseJson{
			Version: respContainer.Version,
			Data: &SignedBeaconBlockCapellaJson{
				Message:   respContainer.Data.CapellaBlock,
				Signature: respContainer.Data.Signature,
			},
			ExecutionOptimistic: respContainer.ExecutionOptimistic,
			Finalized:           respContainer.Finalized,
		}
	default:
		return false, nil, apimiddleware.InternalServerError(fmt.Errorf("unsupported block version '%s'", respContainer.Version))
	}

	j, err := json.Marshal(actualRespContainer)
	if err != nil {
		return false, nil, apimiddleware.InternalServerErrorWithMessage(err, "could not marshal response")
	}
	return false, j, nil
}

func serializeBlindedBlock(response interface{}) (apimiddleware.RunDefault, []byte, apimiddleware.ErrorJson) {
	respContainer, ok := response.(*BlindedBlockResponseJson)
	if !ok {
		return false, nil, apimiddleware.InternalServerError(errors.New("container is not of the correct type"))
	}

	var actualRespContainer interface{}
	switch {
	case strings.EqualFold(respContainer.Version, strings.ToLower(zondpbv1.Version_CAPELLA.String())):
		actualRespContainer = &capellaBlindedBlockResponseJson{
			Version: respContainer.Version,
			Data: &SignedBlindedBeaconBlockCapellaJson{
				Message:   respContainer.Data.CapellaBlock,
				Signature: respContainer.Data.Signature,
			},
			ExecutionOptimistic: respContainer.ExecutionOptimistic,
			Finalized:           respContainer.Finalized,
		}
	default:
		return false, nil, apimiddleware.InternalServerError(fmt.Errorf("unsupported block version '%s'", respContainer.Version))
	}

	j, err := json.Marshal(actualRespContainer)
	if err != nil {
		return false, nil, apimiddleware.InternalServerErrorWithMessage(err, "could not marshal response")
	}
	return false, j, nil
}

type capellaStateResponseJson struct {
	Version string                  `json:"version" enum:"true"`
	Data    *BeaconStateCapellaJson `json:"data"`
}

func serializeState(response interface{}) (apimiddleware.RunDefault, []byte, apimiddleware.ErrorJson) {
	respContainer, ok := response.(*BeaconStateResponseJson)
	if !ok {
		return false, nil, apimiddleware.InternalServerError(errors.New("container is not of the correct type"))
	}

	var actualRespContainer interface{}
	switch {
	case strings.EqualFold(respContainer.Version, strings.ToLower(zondpbv1.Version_CAPELLA.String())):
		actualRespContainer = &capellaStateResponseJson{
			Version: respContainer.Version,
			Data:    respContainer.Data.CapellaState,
		}
	default:
		return false, nil, apimiddleware.InternalServerError(fmt.Errorf("unsupported state version '%s'", respContainer.Version))
	}

	j, err := json.Marshal(actualRespContainer)
	if err != nil {
		return false, nil, apimiddleware.InternalServerErrorWithMessage(err, "could not marshal response")
	}
	return false, j, nil
}

type capellaProduceBlockResponseJson struct {
	Version string                  `json:"version" enum:"true"`
	Data    *BeaconBlockCapellaJson `json:"data"`
}

type capellaProduceBlindedBlockResponseJson struct {
	Version string                         `json:"version" enum:"true"`
	Data    *BlindedBeaconBlockCapellaJson `json:"data"`
}

func serializeProducedBlock(response interface{}) (apimiddleware.RunDefault, []byte, apimiddleware.ErrorJson) {
	respContainer, ok := response.(*ProduceBlockResponseJson)
	if !ok {
		return false, nil, apimiddleware.InternalServerError(errors.New("container is not of the correct type"))
	}

	var actualRespContainer interface{}
	switch {
	case strings.EqualFold(respContainer.Version, strings.ToLower(zondpbv1.Version_CAPELLA.String())):
		actualRespContainer = &capellaProduceBlockResponseJson{
			Version: respContainer.Version,
			Data:    respContainer.Data.CapellaBlock,
		}
	default:
		return false, nil, apimiddleware.InternalServerError(fmt.Errorf("unsupported block version '%s'", respContainer.Version))
	}

	j, err := json.Marshal(actualRespContainer)
	if err != nil {
		return false, nil, apimiddleware.InternalServerErrorWithMessage(err, "could not marshal response")
	}
	return false, j, nil
}

func serializeProducedBlindedBlock(response interface{}) (apimiddleware.RunDefault, []byte, apimiddleware.ErrorJson) {
	respContainer, ok := response.(*ProduceBlindedBlockResponseJson)
	if !ok {
		return false, nil, apimiddleware.InternalServerError(errors.New("container is not of the correct type"))
	}

	var actualRespContainer interface{}
	switch {
	case strings.EqualFold(respContainer.Version, strings.ToLower(zondpbv1.Version_CAPELLA.String())):
		actualRespContainer = &capellaProduceBlindedBlockResponseJson{
			Version: respContainer.Version,
			Data:    respContainer.Data.CapellaBlock,
		}
	default:
		return false, nil, apimiddleware.InternalServerError(fmt.Errorf("unsupported block version '%s'", respContainer.Version))
	}

	j, err := json.Marshal(actualRespContainer)
	if err != nil {
		return false, nil, apimiddleware.InternalServerErrorWithMessage(err, "could not marshal response")
	}
	return false, j, nil
}

func prepareForkChoiceResponse(response interface{}) (apimiddleware.RunDefault, []byte, apimiddleware.ErrorJson) {
	dump, ok := response.(*ForkChoiceDumpJson)
	if !ok {
		return false, nil, apimiddleware.InternalServerError(errors.New("response is not of the correct type"))
	}

	nodes := make([]*ForkChoiceNodeResponseJson, len(dump.ForkChoiceNodes))
	for i, n := range dump.ForkChoiceNodes {
		nodes[i] = &ForkChoiceNodeResponseJson{
			Slot:               n.Slot,
			BlockRoot:          n.BlockRoot,
			ParentRoot:         n.ParentRoot,
			JustifiedEpoch:     n.JustifiedEpoch,
			FinalizedEpoch:     n.FinalizedEpoch,
			Weight:             n.Weight,
			Validity:           n.Validity,
			ExecutionBlockHash: n.ExecutionBlockHash,
			ExtraData: &ForkChoiceNodeExtraDataJson{
				UnrealizedJustifiedEpoch: n.UnrealizedJustifiedEpoch,
				UnrealizedFinalizedEpoch: n.UnrealizedFinalizedEpoch,
				Balance:                  n.Balance,
				ExecutionOptimistic:      n.ExecutionOptimistic,
				TimeStamp:                n.TimeStamp,
			},
		}
	}
	forkChoice := &ForkChoiceResponseJson{
		JustifiedCheckpoint: dump.JustifiedCheckpoint,
		FinalizedCheckpoint: dump.FinalizedCheckpoint,
		ForkChoiceNodes:     nodes,
		ExtraData: &ForkChoiceResponseExtraDataJson{
			BestJustifiedCheckpoint:       dump.BestJustifiedCheckpoint,
			UnrealizedJustifiedCheckpoint: dump.UnrealizedJustifiedCheckpoint,
			UnrealizedFinalizedCheckpoint: dump.UnrealizedFinalizedCheckpoint,
			ProposerBoostRoot:             dump.ProposerBoostRoot,
			PreviousProposerBoostRoot:     dump.PreviousProposerBoostRoot,
			HeadRoot:                      dump.HeadRoot,
		},
	}

	result, err := json.Marshal(forkChoice)
	if err != nil {
		return false, nil, apimiddleware.InternalServerError(errors.New("could not marshal fork choice to JSON"))
	}
	return false, result, nil
}
