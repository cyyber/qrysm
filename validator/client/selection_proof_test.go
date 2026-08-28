package client

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/theQRL/go-bitfield"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"github.com/theQRL/qrysm/validator/client/iface"
	"google.golang.org/protobuf/types/known/emptypb"
)

// The tests in this file cover the selection proof cache: with hedged ML-DSA-87
// signing, every signature over the same message is different, and since
// is_aggregator / is_sync_committee_aggregator are derived from hash(signature),
// the proof RolesAt decided on must be the exact proof later submitted to the
// beacon node.

func TestSelectionProof_SignedOncePerSlotAndReused(t *testing.T) {
	v, m, validatorKey, finish := setup(t)
	defer finish()
	ctx := context.Background()
	pubKey := bytesutil.ToBytes2592(validatorKey.PublicKey().Marshal())
	m.validatorClient.EXPECT().DomainData(gomock.Any(), gomock.Any()).
		Return(&qrysmpb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil).AnyTimes()

	// Test premise: the keymanager signs with hedged ML-DSA-87, so signing the
	// same message twice yields two different (valid) signatures.
	raw1, err := v.newSlotSelectionProof(ctx, pubKey, 1)
	require.NoError(t, err)
	raw2, err := v.newSlotSelectionProof(ctx, pubKey, 1)
	require.NoError(t, err)
	require.DeepNotEqual(t, raw1, raw2, "test premise: hedged signing must produce distinct signatures")

	// The attestation selection proof is signed once per (pubkey, slot).
	proof, err := v.signSlotWithSelectionProof(ctx, pubKey, 1)
	require.NoError(t, err)
	again, err := v.signSlotWithSelectionProof(ctx, pubKey, 1)
	require.NoError(t, err)
	require.DeepEqual(t, proof, again, "selection proof must be reused within a slot")
	otherSlot, err := v.signSlotWithSelectionProof(ctx, pubKey, 2)
	require.NoError(t, err)
	require.DeepNotEqual(t, proof, otherSlot)

	// The sync committee selection proof is signed once per (pubkey, slot, subnet).
	syncProof, err := v.signSyncSelectionData(ctx, pubKey, 0, 1)
	require.NoError(t, err)
	syncAgain, err := v.signSyncSelectionData(ctx, pubKey, 0, 1)
	require.NoError(t, err)
	require.DeepEqual(t, syncProof, syncAgain, "sync selection proof must be reused within a slot")
	otherSubnet, err := v.signSyncSelectionData(ctx, pubKey, 1, 1)
	require.NoError(t, err)
	require.DeepNotEqual(t, syncProof, otherSubnet)
	require.DeepNotEqual(t, proof, syncProof, "attestation and sync proofs for the same slot are distinct")
}

func TestSubmitAggregateAndProof_SubmitsTheSelectionProofRolesAtDecidedOn(t *testing.T) {
	v, m, validatorKey, finish := setup(t)
	defer finish()
	ctx := context.Background()
	slot := primitives.Slot(1)
	pubKeyBytes := validatorKey.PublicKey().Marshal()
	pubKey := bytesutil.ToBytes2592(pubKeyBytes)
	// A committee smaller than 2 * TARGET_AGGREGATORS_PER_COMMITTEE has modulo 1,
	// so the validator is always an aggregator: the test is about which proof
	// bytes get submitted, not about the lottery.
	v.duties = &qrysmpb.DutiesResponse{CurrentEpochDuties: []*qrysmpb.DutiesResponse_Duty{{
		PublicKey:      pubKeyBytes,
		AttesterSlot:   slot,
		CommitteeIndex: 3,
		Committee:      []primitives.ValidatorIndex{0, 1, 2, 3},
		ValidatorIndex: 2,
	}}}
	m.validatorClient.EXPECT().DomainData(gomock.Any(), gomock.Any()).
		Return(&qrysmpb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil).AnyTimes()

	roles, err := v.RolesAt(ctx, slot)
	require.NoError(t, err)
	require.DeepEqual(t, []iface.ValidatorRole{iface.RoleAttester, iface.RoleAggregator}, roles[pubKey])
	// The proof RolesAt derived the aggregator role from (served from the cache).
	decided, err := v.signSlotWithSelectionProof(ctx, pubKey, slot)
	require.NoError(t, err)

	var submitted []byte
	m.validatorClient.EXPECT().SubmitAggregateSelectionProof(gomock.Any(), gomock.AssignableToTypeOf(&qrysmpb.AggregateSelectionRequest{})).
		DoAndReturn(func(_ context.Context, req *qrysmpb.AggregateSelectionRequest) (*qrysmpb.AggregateSelectionResponse, error) {
			submitted = req.SlotSignature
			return &qrysmpb.AggregateSelectionResponse{AggregateAndProof: &qrysmpb.AggregateAttestationAndProof{
				AggregatorIndex: 2,
				Aggregate:       util.HydrateAttestation(&qrysmpb.Attestation{AggregationBits: make([]byte, 1)}),
				SelectionProof:  req.SlotSignature,
			}}, nil
		})
	m.validatorClient.EXPECT().SubmitSignedAggregateSelectionProof(gomock.Any(), gomock.AssignableToTypeOf(&qrysmpb.SignedAggregateSubmitRequest{})).
		Return(&qrysmpb.SignedAggregateSubmitResponse{AttestationDataRoot: make([]byte, 32)}, nil)

	v.SubmitAggregateAndProof(ctx, slot, pubKey)
	require.NotNil(t, submitted)
	require.DeepEqual(t, decided, submitted, "the submitted selection proof must be the one the aggregator role was derived from")
}

func TestSubmitSignedContributionAndProof_SubmitsTheSelectionProofRolesAtDecidedOn(t *testing.T) {
	forceSyncCommitteeAggregatorSelection(t)
	v, m, validatorKey, finish := setup(t)
	defer finish()
	ctx := context.Background()
	slot := primitives.Slot(10)
	pubKeyBytes := validatorKey.PublicKey().Marshal()
	pubKey := bytesutil.ToBytes2592(pubKeyBytes)
	v.duties = &qrysmpb.DutiesResponse{CurrentEpochDuties: []*qrysmpb.DutiesResponse_Duty{{
		PublicKey:      pubKeyBytes,
		ValidatorIndex: 7,
	}}}
	m.validatorClient.EXPECT().DomainData(gomock.Any(), gomock.Any()).
		Return(&qrysmpb.DomainResponse{SignatureDomain: make([]byte, 32)}, nil).AnyTimes()
	// Once for the RolesAt-side eligibility check, once for the submission.
	m.validatorClient.EXPECT().GetSyncSubcommitteeIndex(gomock.Any(), &qrysmpb.SyncSubcommitteeIndexRequest{PublicKey: pubKeyBytes, Slot: slot}).
		Return(&qrysmpb.SyncSubcommitteeIndexResponse{Indices: []primitives.CommitteeIndex{1}}, nil).Times(2)

	isAggregator, err := v.isSyncCommitteeAggregator(ctx, slot, pubKey)
	require.NoError(t, err)
	require.Equal(t, true, isAggregator)
	// Subcommittee index 1 lives in subnet 0; this is the proof the role was
	// derived from (served from the cache).
	decided, err := v.signSyncSelectionData(ctx, pubKey, 0, slot)
	require.NoError(t, err)

	aggBits := bitfield.NewBitvector128()
	aggBits.SetBitAt(0, true)
	m.validatorClient.EXPECT().GetSyncCommitteeContribution(gomock.Any(), &qrysmpb.SyncCommitteeContributionRequest{Slot: slot, PublicKey: pubKeyBytes, SubnetId: 0}).
		Return(&qrysmpb.SyncCommitteeContribution{BlockRoot: make([]byte, field_params.RootLength), Signatures: [][]byte{}, AggregationBits: aggBits}, nil)
	var submitted []byte
	m.validatorClient.EXPECT().SubmitSignedContributionAndProof(gomock.Any(), gomock.AssignableToTypeOf(&qrysmpb.SignedContributionAndProof{})).
		DoAndReturn(func(_ context.Context, req *qrysmpb.SignedContributionAndProof) (*emptypb.Empty, error) {
			submitted = req.Message.SelectionProof
			return &emptypb.Empty{}, nil
		})

	v.SubmitSignedContributionAndProof(ctx, slot, pubKey)
	require.NotNil(t, submitted, "contribution must be submitted when RolesAt selected the validator as sync aggregator")
	require.DeepEqual(t, decided, submitted, "the submitted selection proof must be the one the sync aggregator role was derived from")
}

func TestRolesAt_PrunesSelectionProofsOfPastSlots(t *testing.T) {
	v, _, _, finish := setup(t)
	defer finish()
	v.duties = &qrysmpb.DutiesResponse{}
	var pubKey [field_params.MLDSA87PubkeyLength]byte
	v.selectionProofCache = map[selectionProofKey][]byte{
		{pubKey: pubKey, slot: 3}:                        {3},
		{pubKey: pubKey, slot: 4, sync: true, subnet: 1}: {4},
		{pubKey: pubKey, slot: 5}:                        {5},
		{pubKey: pubKey, slot: 5, sync: true, subnet: 0}: {5},
		{pubKey: pubKey, slot: 130}:                      {130}, // next-epoch duty signed ahead by subscribeToSubnets
	}

	_, err := v.RolesAt(context.Background(), 5)
	require.NoError(t, err)

	require.DeepEqual(t, map[selectionProofKey][]byte{
		{pubKey: pubKey, slot: 5}:                        {5},
		{pubKey: pubKey, slot: 5, sync: true, subnet: 0}: {5},
		{pubKey: pubKey, slot: 130}:                      {130},
	}, v.selectionProofCache)
}
