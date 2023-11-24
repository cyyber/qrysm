package validator

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/theQRL/go-zond/common/hexutil"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/core"
	"github.com/theQRL/qrysm/v4/beacon-chain/rpc/zond/shared"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	http2 "github.com/theQRL/qrysm/v4/network/http"
	zondpbalpha "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
)

// GetAggregateAttestation aggregates all attestations matching the given attestation data root and slot, returning the aggregated result.
func (s *Server) GetAggregateAttestation(w http.ResponseWriter, r *http.Request) {
	attDataRoot := r.URL.Query().Get("attestation_data_root")
	valid := shared.ValidateHex(w, "Attestation data root", attDataRoot)
	if !valid {
		return
	}
	rawSlot := r.URL.Query().Get("slot")
	slot, valid := shared.ValidateUint(w, "Slot", rawSlot)
	if !valid {
		return
	}

	if err := s.AttestationsPool.AggregateUnaggregatedAttestations(r.Context()); err != nil {
		http2.HandleError(w, "Could not aggregate unaggregated attestations: "+err.Error(), http.StatusBadRequest)
		return
	}

	allAtts := s.AttestationsPool.AggregatedAttestations()
	var bestMatchingAtt *zondpbalpha.Attestation
	for _, att := range allAtts {
		if att.Data.Slot == primitives.Slot(slot) {
			root, err := att.Data.HashTreeRoot()
			if err != nil {
				http2.HandleError(w, "Could not get attestation data root: "+err.Error(), http.StatusInternalServerError)
				return
			}
			attDataRootBytes, err := hexutil.Decode(attDataRoot)
			if err != nil {
				http2.HandleError(w, "Could not decode attestation data root into bytes: "+err.Error(), http.StatusBadRequest)
				return
			}
			if bytes.Equal(root[:], attDataRootBytes) {
				if bestMatchingAtt == nil || len(att.ParticipationBits) > len(bestMatchingAtt.ParticipationBits) {
					bestMatchingAtt = att
				}
			}
		}
	}
	if bestMatchingAtt == nil {
		http2.HandleError(w, "No matching attestation found", http.StatusNotFound)
		return
	}

	signatures := make([]string, len(bestMatchingAtt.Signatures))
	for i, sig := range bestMatchingAtt.Signatures {
		signatures[i] = hexutil.Encode(sig)
	}

	response := &AggregateAttestationResponse{
		Data: &shared.Attestation{
			ParticipationBits: hexutil.Encode(bestMatchingAtt.ParticipationBits),
			Data: &shared.AttestationData{
				Slot:            strconv.FormatUint(uint64(bestMatchingAtt.Data.Slot), 10),
				CommitteeIndex:  strconv.FormatUint(uint64(bestMatchingAtt.Data.CommitteeIndex), 10),
				BeaconBlockRoot: hexutil.Encode(bestMatchingAtt.Data.BeaconBlockRoot),
				Source: &shared.Checkpoint{
					Epoch: strconv.FormatUint(uint64(bestMatchingAtt.Data.Source.Epoch), 10),
					Root:  hexutil.Encode(bestMatchingAtt.Data.Source.Root),
				},
				Target: &shared.Checkpoint{
					Epoch: strconv.FormatUint(uint64(bestMatchingAtt.Data.Target.Epoch), 10),
					Root:  hexutil.Encode(bestMatchingAtt.Data.Target.Root),
				},
			},
			Signatures: signatures,
		}}
	http2.WriteJson(w, response)
}

// SubmitContributionAndProofs publishes multiple signed sync committee contribution and proofs.
func (s *Server) SubmitContributionAndProofs(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		http2.HandleError(w, "No data submitted", http.StatusBadRequest)
		return
	}

	var req SubmitContributionAndProofsRequest
	if err := json.NewDecoder(r.Body).Decode(&req.Data); err != nil {
		http2.HandleError(w, "Could not decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http2.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	for _, item := range req.Data {
		consensusItem, err := item.ToConsensus()
		if err != nil {
			http2.HandleError(w, "Could not convert request contribution to consensus contribution: "+err.Error(), http.StatusBadRequest)
			return
		}
		rpcError := s.CoreService.SubmitSignedContributionAndProof(r.Context(), consensusItem)
		if rpcError != nil {
			http2.HandleError(w, rpcError.Err.Error(), core.ErrorReasonToHTTP(rpcError.Reason))
		}
	}
}

// SubmitAggregateAndProofs verifies given aggregate and proofs and publishes them on appropriate gossipsub topic.
func (s *Server) SubmitAggregateAndProofs(w http.ResponseWriter, r *http.Request) {
	if r.Body == http.NoBody {
		http2.HandleError(w, "No data submitted", http.StatusBadRequest)
		return
	}

	var req SubmitAggregateAndProofsRequest
	if err := json.NewDecoder(r.Body).Decode(&req.Data); err != nil {
		http2.HandleError(w, "Could not decode request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	validate := validator.New()
	if err := validate.Struct(req); err != nil {
		http2.HandleError(w, err.Error(), http.StatusBadRequest)
		return
	}

	broadcastFailed := false
	for _, item := range req.Data {
		consensusItem, err := item.ToConsensus()
		if err != nil {
			http2.HandleError(w, "Could not convert request aggregate to consensus aggregate: "+err.Error(), http.StatusBadRequest)
			return
		}
		rpcError := s.CoreService.SubmitSignedAggregateSelectionProof(
			r.Context(),
			&zondpbalpha.SignedAggregateSubmitRequest{SignedAggregateAndProof: consensusItem},
		)
		if rpcError != nil {
			_, ok := rpcError.Err.(*core.AggregateBroadcastFailedError)
			if ok {
				broadcastFailed = true
			} else {
				http2.HandleError(w, rpcError.Err.Error(), core.ErrorReasonToHTTP(rpcError.Reason))
				return
			}
		}
	}

	if broadcastFailed {
		http2.HandleError(w, "Could not broadcast one or more signed aggregated attestations", http.StatusInternalServerError)
	}
}
