package simulator

import (
	"context"
	"math"

	"github.com/sirupsen/logrus"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/rand"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/time/slots"
)

func (s *Simulator) generateAttestationsForSlot(
	ctx context.Context, slot primitives.Slot,
) ([]*qrysmpb.IndexedAttestation, []*qrysmpb.AttesterSlashing, error) {
	attestations := make([]*qrysmpb.IndexedAttestation, 0)
	slashings := make([]*qrysmpb.AttesterSlashing, 0)
	currentEpoch := slots.ToEpoch(slot)

	committeesPerSlot := helpers.SlotCommitteeCount(s.srvConfig.Params.NumValidators)
	valsPerCommittee := s.srvConfig.Params.NumValidators /
		(committeesPerSlot * uint64(s.srvConfig.Params.SlotsPerEpoch))
	valsPerSlot := committeesPerSlot * valsPerCommittee

	if currentEpoch < 2 {
		return nil, nil, nil
	}
	sourceEpoch := currentEpoch - 1

	var slashedIndices []uint64
	startIdx := valsPerSlot * uint64(slot%s.srvConfig.Params.SlotsPerEpoch)
	endIdx := startIdx + valsPerCommittee
	for c := primitives.CommitteeIndex(0); uint64(c) < committeesPerSlot; c++ {
		attData := &qrysmpb.AttestationData{
			Slot:            slot,
			CommitteeIndex:  c,
			BeaconBlockRoot: bytesutil.PadTo([]byte("block"), 32),
			Source: &qrysmpb.Checkpoint{
				Epoch: sourceEpoch,
				Root:  bytesutil.PadTo([]byte("source"), 32),
			},
			Target: &qrysmpb.Checkpoint{
				Epoch: currentEpoch,
				Root:  bytesutil.PadTo([]byte("target"), 32),
			},
		}

		valsPerAttestation := uint64(math.Floor(s.srvConfig.Params.AggregationPercent * float64(valsPerCommittee)))
		for i := startIdx; i < endIdx; i += valsPerAttestation {
			attEndIdx := i + valsPerAttestation
			if attEndIdx >= endIdx {
				attEndIdx = endIdx
			}
			indices := make([]uint64, 0, valsPerAttestation)
			for idx := i; idx < attEndIdx; idx++ {
				indices = append(indices, idx)
			}
			att := &qrysmpb.IndexedAttestation{
				AttestingIndices: indices,
				Data:             attData,
				Signatures:       [][]byte{},
			}
			beaconState, err := s.srvConfig.AttestationStateFetcher.AttestationTargetState(ctx, att.Data.Target)
			if err != nil {
				return nil, nil, err
			}

			// Sign the attestation with a valid signature.
			sigs, err := s.sigsForAttestation(beaconState, att)
			if err != nil {
				return nil, nil, err
			}
			att.Signatures = sigs

			attestations = append(attestations, att)
			if rand.NewGenerator().Float64() < s.srvConfig.Params.AttesterSlashingProbab {
				slashableAtt := makeSlashableFromAtt(att, []uint64{indices[0]})
				sigs, err := s.sigsForAttestation(beaconState, slashableAtt)
				if err != nil {
					return nil, nil, err
				}
				slashableAtt.Signatures = sigs
				slashedIndices = append(slashedIndices, slashableAtt.AttestingIndices...)
				slashings = append(slashings, &qrysmpb.AttesterSlashing{
					Attestation_1: att,
					Attestation_2: slashableAtt,
				})
				attestations = append(attestations, slashableAtt)
			}
		}
		startIdx += valsPerCommittee
		endIdx += valsPerCommittee
	}
	if len(slashedIndices) > 0 {
		log.WithFields(logrus.Fields{
			"amount":  len(slashedIndices),
			"indices": slashedIndices,
		}).Infof("Slashable attestation made")
	}
	return attestations, slashings, nil
}

func (s *Simulator) sigsForAttestation(
	beaconState state.ReadOnlyBeaconState, att *qrysmpb.IndexedAttestation,
) ([][]byte, error) {
	domain, err := signing.Domain(
		beaconState.Fork(),
		att.Data.Target.Epoch,
		params.BeaconConfig().DomainBeaconAttester,
		beaconState.GenesisValidatorsRoot(),
	)
	if err != nil {
		return nil, err
	}
	signingRoot, err := signing.ComputeSigningRoot(att.Data, domain)
	if err != nil {
		return nil, err
	}
	sigs := make([][]byte, len(att.AttestingIndices))
	for i, validatorIndex := range att.AttestingIndices {
		privKey := s.srvConfig.PrivateKeysByValidatorIndex[primitives.ValidatorIndex(validatorIndex)]
		sigs[i] = privKey.Sign(signingRoot[:]).Marshal()
	}

	return sigs, nil
}

func makeSlashableFromAtt(att *qrysmpb.IndexedAttestation, indices []uint64) *qrysmpb.IndexedAttestation {
	if att.Data.Source.Epoch <= 2 {
		return makeDoubleVoteFromAtt(att, indices)
	}
	attData := &qrysmpb.AttestationData{
		Slot:            att.Data.Slot,
		CommitteeIndex:  att.Data.CommitteeIndex,
		BeaconBlockRoot: att.Data.BeaconBlockRoot,
		Source: &qrysmpb.Checkpoint{
			Epoch: att.Data.Source.Epoch - 3,
			Root:  att.Data.Source.Root,
		},
		Target: &qrysmpb.Checkpoint{
			Epoch: att.Data.Target.Epoch,
			Root:  att.Data.Target.Root,
		},
	}
	return &qrysmpb.IndexedAttestation{
		AttestingIndices: indices,
		Data:             attData,
		Signatures:       [][]byte{params.BeaconConfig().EmptyDilithiumSignature[:]},
	}
}

func makeDoubleVoteFromAtt(att *qrysmpb.IndexedAttestation, indices []uint64) *qrysmpb.IndexedAttestation {
	attData := &qrysmpb.AttestationData{
		Slot:            att.Data.Slot,
		CommitteeIndex:  att.Data.CommitteeIndex,
		BeaconBlockRoot: bytesutil.PadTo([]byte("slash me"), 32),
		Source: &qrysmpb.Checkpoint{
			Epoch: att.Data.Source.Epoch,
			Root:  att.Data.Source.Root,
		},
		Target: &qrysmpb.Checkpoint{
			Epoch: att.Data.Target.Epoch,
			Root:  att.Data.Target.Root,
		},
	}
	return &qrysmpb.IndexedAttestation{
		AttestingIndices: indices,
		Data:             attData,
		Signatures:       [][]byte{params.BeaconConfig().EmptyDilithiumSignature[:]},
	}
}
