// Package attestationutil contains useful helpers for converting
// attestations into indexed form.
package attestation

import (
	"bytes"
	"context"
	"fmt"
	"runtime"
	"sort"

	"github.com/pkg/errors"
	"github.com/theQRL/go-bitfield"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/v4/config/params"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	"github.com/theQRL/qrysm/v4/crypto/dilithium"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"go.opencensus.io/trace"
	"golang.org/x/sync/errgroup"
)

type signatureSlices struct {
	attestingIndices []uint64
	signatures       [][]byte
}

type sortByValidatorIdx signatureSlices

func (s sortByValidatorIdx) Len() int {
	return len(s.signatures)
}

func (s sortByValidatorIdx) Swap(i, j int) {
	s.attestingIndices[i], s.attestingIndices[j] = s.attestingIndices[j], s.attestingIndices[i]
	s.signatures[i], s.signatures[j] = s.signatures[j], s.signatures[i]
}

func (s sortByValidatorIdx) Less(i, j int) bool {
	return s.attestingIndices[i] < s.attestingIndices[j]
}

// ConvertToIndexed converts attestation to (almost) indexed-verifiable form.
//
// Note about spec pseudocode definition. The state was used by get_attesting_indices to determine
// the attestation committee. Now that we provide this as an argument, we no longer need to provide
// a state.
//
// Spec pseudocode definition:
//
//	def get_indexed_attestation(state: BeaconState, attestation: Attestation) -> IndexedAttestation:
//	 """
//	 Return the indexed attestation corresponding to ``attestation``.
//	 """
//	 attesting_indices = get_attesting_indices(state, attestation.data, attestation.aggregation_bits)
//
//	 return IndexedAttestation(
//	     attesting_indices=sorted(attesting_indices),
//	     data=attestation.data,
//	     signature=attestation.signature,
//	 )
func ConvertToIndexed(ctx context.Context, attestation *zondpb.Attestation, committee []primitives.ValidatorIndex) (*zondpb.IndexedAttestation, error) {
	attIndices, err := AttestingIndices(attestation.AggregationBits, committee)
	if err != nil {
		return nil, err
	}

	if len(attestation.Signatures) != len(attIndices) {
		return nil, fmt.Errorf("signatures length %d is not equal to the attesting participants indices length %d", len(attestation.Signatures), len(attIndices))
	}

	sigsCopy := make([][]byte, len(attestation.Signatures))
	copy(sigsCopy, attestation.Signatures)

	sigSlices := signatureSlices{
		attestingIndices: attIndices,
		signatures:       sigsCopy,
	}
	sort.Sort(sortByValidatorIdx(sigSlices))

	return &zondpb.IndexedAttestation{
		Data:             attestation.Data,
		Signatures:       sigSlices.signatures,
		AttestingIndices: sigSlices.attestingIndices,
	}, nil
}

// AttestingIndices returns the attesting participants indices from the attestation data. The
// committee is provided as an argument rather than a imported implementation from the spec definition.
// Having the committee as an argument allows for re-use of beacon committees when possible.
//
// Spec pseudocode definition:
//
//	def get_attesting_indices(state: BeaconState,
//	                       data: AttestationData,
//	                       bits: Bitlist[MAX_VALIDATORS_PER_COMMITTEE]) -> Set[ValidatorIndex]:
//	 """
//	 Return the set of attesting indices corresponding to ``data`` and ``bits``.
//	 """
//	 committee = get_beacon_committee(state, data.slot, data.index)
//	 return set(index for i, index in enumerate(committee) if bits[i])
func AttestingIndices(bf bitfield.Bitfield, committee []primitives.ValidatorIndex) ([]uint64, error) {
	if bf.Len() != uint64(len(committee)) {
		return nil, fmt.Errorf("bitfield length %d is not equal to committee length %d", bf.Len(), len(committee))
	}
	indices := make([]uint64, 0, bf.Count())
	for _, idx := range bf.BitIndices() {
		if idx < len(committee) {
			indices = append(indices, uint64(committee[idx]))
		}
	}
	return indices, nil
}

// VerifyIndexedAttestationSig this helper function performs the last part of the
// spec indexed attestation validation starting at Verify aggregate signature
// comment.
//
// Spec pseudocode definition:
//
//	def is_valid_indexed_attestation(state: BeaconState, indexed_attestation: IndexedAttestation) -> bool:
//	 """
//	 Check if ``indexed_attestation`` is not empty, has sorted and unique indices and has a valid aggregate signature.
//	 """
//	 # Verify indices are sorted and unique
//	 indices = indexed_attestation.attesting_indices
//	 if len(indices) == 0 or not indices == sorted(set(indices)):
//	     return False
//	 # Verify aggregate signature
//	 pubkeys = [state.validators[i].pubkey for i in indices]
//	 domain = get_domain(state, DOMAIN_BEACON_ATTESTER, indexed_attestation.data.target.epoch)
//	 signing_root = compute_signing_root(indexed_attestation.data, domain)
//	 return bls.FastAggregateVerify(pubkeys, signing_root, indexed_attestation.signature)
func VerifyIndexedAttestationSigs(ctx context.Context, indexedAtt *zondpb.IndexedAttestation, pubKeys []dilithium.PublicKey, domain []byte) error {
	_, span := trace.StartSpan(ctx, "attestationutil.VerifyIndexedAttestationSigs")
	defer span.End()

	if len(indexedAtt.Signatures) != len(pubKeys) {
		return fmt.Errorf("signatures length %d is not equal to pub keys length %d", len(indexedAtt.Signatures), len(pubKeys))
	}

	messageHash, err := signing.ComputeSigningRoot(indexedAtt.Data, domain)
	if err != nil {
		return errors.Wrap(err, "could not get signing root of object")
	}

	n := runtime.GOMAXPROCS(0) - 1
	grp := errgroup.Group{}
	grp.SetLimit(n)
	for i, s := range indexedAtt.Signatures {
		sig, err := dilithium.SignatureFromBytes(s)
		if err != nil {
			return errors.Wrap(err, "could not convert bytes to signature")
		}
		pubKey := pubKeys[i]

		grp.Go(func() error {
			if !sig.Verify(pubKey, messageHash[:]) {
				return signing.ErrSigFailedToVerify
			}

			return nil
		})
	}

	if err := grp.Wait(); err != nil {
		return err
	}

	return nil
}

// IsValidAttestationIndices this helper function performs the first part of the
// spec indexed attestation validation starting at Check if “indexed_attestation“
// comment and ends at Verify aggregate signature comment.
//
// Spec pseudocode definition:
//
//	def is_valid_indexed_attestation(state: BeaconState, indexed_attestation: IndexedAttestation) -> bool:
//	  """
//	  Check if ``indexed_attestation`` is not empty, has sorted and unique indices and has a valid aggregate signature.
//	  """
//	  # Verify indices are sorted and unique
//	  indices = indexed_attestation.attesting_indices
//	  if len(indices) == 0 or not indices == sorted(set(indices)):
//	      return False
//	  # Verify aggregate signature
//	  pubkeys = [state.validators[i].pubkey for i in indices]
//	  domain = get_domain(state, DOMAIN_BEACON_ATTESTER, indexed_attestation.data.target.epoch)
//	  signing_root = compute_signing_root(indexed_attestation.data, domain)
//	  return bls.FastAggregateVerify(pubkeys, signing_root, indexed_attestation.signature)
func IsValidAttestationIndices(ctx context.Context, indexedAttestation *zondpb.IndexedAttestation) error {
	_, span := trace.StartSpan(ctx, "attestationutil.IsValidAttestationIndices")
	defer span.End()

	if indexedAttestation == nil || indexedAttestation.Data == nil || indexedAttestation.Data.Target == nil || indexedAttestation.AttestingIndices == nil {
		return errors.New("nil or missing indexed attestation data")
	}
	indices := indexedAttestation.AttestingIndices
	if len(indices) == 0 {
		return errors.New("expected non-empty attesting indices")
	}
	if uint64(len(indices)) > params.BeaconConfig().MaxValidatorsPerCommittee {
		return fmt.Errorf("validator indices count exceeds MAX_VALIDATORS_PER_COMMITTEE, %d > %d", len(indices), params.BeaconConfig().MaxValidatorsPerCommittee)
	}
	for i := 1; i < len(indices); i++ {
		if indices[i-1] >= indices[i] {
			return errors.New("attesting indices is not uniquely sorted")
		}
	}
	return nil
}

// AttDataIsEqual this function performs an equality check between 2 attestation data, if they're unequal, it will return false.
func AttDataIsEqual(attData1, attData2 *zondpb.AttestationData) bool {
	if attData1.Slot != attData2.Slot {
		return false
	}
	if attData1.CommitteeIndex != attData2.CommitteeIndex {
		return false
	}
	if !bytes.Equal(attData1.BeaconBlockRoot, attData2.BeaconBlockRoot) {
		return false
	}
	if attData1.Source.Epoch != attData2.Source.Epoch {
		return false
	}
	if !bytes.Equal(attData1.Source.Root, attData2.Source.Root) {
		return false
	}
	if attData1.Target.Epoch != attData2.Target.Epoch {
		return false
	}
	if !bytes.Equal(attData1.Target.Root, attData2.Target.Root) {
		return false
	}
	return true
}

// CheckPointIsEqual performs an equality check between 2 check points, returns false if unequal.
func CheckPointIsEqual(checkPt1, checkPt2 *zondpb.Checkpoint) bool {
	if checkPt1.Epoch != checkPt2.Epoch {
		return false
	}
	if !bytes.Equal(checkPt1.Root, checkPt2.Root) {
		return false
	}
	return true
}

func SearchInsertIdxWithOffset(arr []int, initialIdx int, target int) (int, error) {
	arrLen := len(arr)

	if arrLen == 0 {
		return 0, nil
	}

	if initialIdx > (arrLen - 1) {
		return 0, fmt.Errorf("invalid initial index %d for slice length %d", initialIdx, arrLen)
	}

	if target <= arr[initialIdx] {
		return initialIdx, nil
	}

	if target > arr[arrLen-1] {
		return arrLen, nil
	}

	low := initialIdx
	high := arrLen - 1

	for low <= high {
		mid := (low + high) / 2
		if arr[mid] == target {
			return mid + 1, nil
		}
		if arr[mid] > target {
			high = mid - 1
		} else {
			low = mid + 1
		}
	}

	return low, nil
}
