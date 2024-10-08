package rewards

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/theQRL/go-bitfield"
	mock "github.com/theQRL/qrysm/beacon-chain/blockchain/testing"
	"github.com/theQRL/qrysm/beacon-chain/core/altair"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/beacon-chain/core/signing"
	"github.com/theQRL/qrysm/beacon-chain/rpc/testutil"
	"github.com/theQRL/qrysm/beacon-chain/state"
	mockstategen "github.com/theQRL/qrysm/beacon-chain/state/stategen/mock"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	fieldparams "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/consensus-types/interfaces"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/crypto/dilithium"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	http2 "github.com/theQRL/qrysm/network/http"
	zond "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
)

func TestBlockRewards(t *testing.T) {
	helpers.ClearCache()

	valCount := 256

	st, err := util.NewBeaconStateCapella()
	require.NoError(t, st.SetSlot(1))
	require.NoError(t, err)
	validators := make([]*zond.Validator, 0, valCount)
	balances := make([]uint64, 0, valCount)
	secretKeys := make([]dilithium.DilithiumKey, 0, valCount)
	for i := 0; i < valCount; i++ {
		dilithiumKey, err := dilithium.RandKey()
		require.NoError(t, err)
		secretKeys = append(secretKeys, dilithiumKey)
		validators = append(validators, &zond.Validator{
			PublicKey:         dilithiumKey.PublicKey().Marshal(),
			ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
			WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			EffectiveBalance:  params.BeaconConfig().MaxEffectiveBalance,
		})
		balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
	}
	require.NoError(t, st.SetValidators(validators))
	require.NoError(t, st.SetBalances(balances))
	require.NoError(t, st.SetCurrentParticipationBits(make([]byte, valCount)))
	syncCommittee, err := altair.NextSyncCommittee(context.Background(), st)
	require.NoError(t, err)
	require.NoError(t, st.SetCurrentSyncCommittee(syncCommittee))
	slot0bRoot := bytesutil.PadTo([]byte("slot0root"), 32)
	bRoots := make([][]byte, fieldparams.BlockRootsLength)
	bRoots[0] = slot0bRoot
	require.NoError(t, st.SetBlockRoots(bRoots))

	b := util.HydrateSignedBeaconBlockCapella(util.NewBeaconBlockCapella())
	b.Block.Slot = 2
	// we have to set the proposer index to the value that will be randomly chosen (fortunately it's deterministic)
	b.Block.ProposerIndex = 12
	b.Block.Body.Attestations = []*zond.Attestation{
		{
			AggregationBits: bitfield.Bitlist{0b00000111},
			Data:            util.HydrateAttestationData(&zond.AttestationData{}),
			Signatures:      [][]byte{make([]byte, field_params.DilithiumSignatureLength), make([]byte, field_params.DilithiumSignatureLength)},
		},
		{
			AggregationBits: bitfield.Bitlist{0b00000111},
			Data:            util.HydrateAttestationData(&zond.AttestationData{}),
			Signatures:      [][]byte{make([]byte, field_params.DilithiumSignatureLength), make([]byte, field_params.DilithiumSignatureLength)},
		},
	}
	attData1 := util.HydrateAttestationData(&zond.AttestationData{BeaconBlockRoot: bytesutil.PadTo([]byte("root1"), 32)})
	attData2 := util.HydrateAttestationData(&zond.AttestationData{BeaconBlockRoot: bytesutil.PadTo([]byte("root2"), 32)})
	domain, err := signing.Domain(st.Fork(), 0, params.BeaconConfig().DomainBeaconAttester, st.GenesisValidatorsRoot())
	require.NoError(t, err)
	sigRoot1, err := signing.ComputeSigningRoot(attData1, domain)
	require.NoError(t, err)
	sigRoot2, err := signing.ComputeSigningRoot(attData2, domain)
	require.NoError(t, err)
	b.Block.Body.AttesterSlashings = []*zond.AttesterSlashing{
		{
			Attestation_1: &zond.IndexedAttestation{
				AttestingIndices: []uint64{0},
				Data:             attData1,
				Signatures:       [][]byte{secretKeys[0].Sign(sigRoot1[:]).Marshal()},
			},
			Attestation_2: &zond.IndexedAttestation{
				AttestingIndices: []uint64{0},
				Data:             attData2,
				Signatures:       [][]byte{secretKeys[0].Sign(sigRoot2[:]).Marshal()},
			},
		},
	}
	header1 := &zond.BeaconBlockHeader{
		Slot:          0,
		ProposerIndex: 1,
		ParentRoot:    bytesutil.PadTo([]byte("root1"), 32),
		StateRoot:     bytesutil.PadTo([]byte("root1"), 32),
		BodyRoot:      bytesutil.PadTo([]byte("root1"), 32),
	}
	header2 := &zond.BeaconBlockHeader{
		Slot:          0,
		ProposerIndex: 1,
		ParentRoot:    bytesutil.PadTo([]byte("root2"), 32),
		StateRoot:     bytesutil.PadTo([]byte("root2"), 32),
		BodyRoot:      bytesutil.PadTo([]byte("root2"), 32),
	}
	domain, err = signing.Domain(st.Fork(), 0, params.BeaconConfig().DomainBeaconProposer, st.GenesisValidatorsRoot())
	require.NoError(t, err)
	sigRoot1, err = signing.ComputeSigningRoot(header1, domain)
	require.NoError(t, err)
	sigRoot2, err = signing.ComputeSigningRoot(header2, domain)
	require.NoError(t, err)
	b.Block.Body.ProposerSlashings = []*zond.ProposerSlashing{
		{
			Header_1: &zond.SignedBeaconBlockHeader{
				Header:    header1,
				Signature: secretKeys[1].Sign(sigRoot1[:]).Marshal(),
			},
			Header_2: &zond.SignedBeaconBlockHeader{
				Header:    header2,
				Signature: secretKeys[1].Sign(sigRoot2[:]).Marshal(),
			},
		},
	}
	scBits := bitfield.NewBitvector16()
	scBits.SetBitAt(10, true)
	scBits.SetBitAt(12, true)
	domain, err = signing.Domain(st.Fork(), 0, params.BeaconConfig().DomainSyncCommittee, st.GenesisValidatorsRoot())
	require.NoError(t, err)
	sszBytes := primitives.SSZBytes(slot0bRoot)
	r, err := signing.ComputeSigningRoot(&sszBytes, domain)
	require.NoError(t, err)
	// Bits set in sync committee bits determine which validators will be treated as participating in sync committee.
	// These validators have to sign the message.
	sig1 := secretKeys[149].Sign(r[:]).Marshal()
	sig2 := secretKeys[48].Sign(r[:]).Marshal()
	b.Block.Body.SyncAggregate = &zond.SyncAggregate{SyncCommitteeBits: scBits, SyncCommitteeSignatures: [][]byte{sig1, sig2}}

	sbb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	require.NoError(t, err)
	mockChainService := &mock.ChainService{Optimistic: true}
	s := &Server{
		Blocker: &testutil.MockBlocker{SlotBlockMap: map[primitives.Slot]interfaces.ReadOnlySignedBeaconBlock{
			2: sbb,
		}},
		OptimisticModeFetcher: mockChainService,
		FinalizationFetcher:   mockChainService,
		ReplayerBuilder:       mockstategen.NewMockReplayerBuilder(mockstategen.WithMockState(st)),
	}

	t.Run("ok", func(t *testing.T) {
		url := "http://only.the.slot.number.at.the.end.is.important/2"
		request := httptest.NewRequest("GET", url, nil)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.BlockRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &BlockRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		assert.Equal(t, "12", resp.Data.ProposerIndex)
		assert.Equal(t, "28214", resp.Data.Total)
		assert.Equal(t, "0", resp.Data.Attestations)
		assert.Equal(t, "28214", resp.Data.SyncAggregate)
		assert.Equal(t, "0", resp.Data.AttesterSlashings)
		assert.Equal(t, "0", resp.Data.ProposerSlashings)
		assert.Equal(t, true, resp.ExecutionOptimistic)
		assert.Equal(t, false, resp.Finalized)
	})
}

func TestAttestationRewards(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	helpers.ClearCache()

	valCount := 64

	st, err := util.NewBeaconStateCapella()
	require.NoError(t, err)
	require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch*3-1))
	validators := make([]*zond.Validator, 0, valCount)
	balances := make([]uint64, 0, valCount)
	secretKeys := make([]dilithium.DilithiumKey, 0, valCount)
	for i := 0; i < valCount; i++ {
		dilithiumKey, err := dilithium.RandKey()
		require.NoError(t, err)
		secretKeys = append(secretKeys, dilithiumKey)
		validators = append(validators, &zond.Validator{
			PublicKey:         dilithiumKey.PublicKey().Marshal(),
			ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
			WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			EffectiveBalance:  params.BeaconConfig().MaxEffectiveBalance / 64 * uint64(i+1),
		})
		balances = append(balances, params.BeaconConfig().MaxEffectiveBalance/64*uint64(i+1))
	}
	require.NoError(t, st.SetValidators(validators))
	require.NoError(t, st.SetBalances(balances))
	require.NoError(t, st.SetInactivityScores(make([]uint64, len(validators))))
	participation := make([]byte, len(validators))
	for i := range participation {
		participation[i] = 0b111
	}
	require.NoError(t, st.SetCurrentParticipationBits(participation))
	require.NoError(t, st.SetPreviousParticipationBits(participation))

	currentSlot := params.BeaconConfig().SlotsPerEpoch * 3
	mockChainService := &mock.ChainService{Optimistic: true, Slot: &currentSlot}
	s := &Server{
		Stater: &testutil.MockStater{StatesBySlot: map[primitives.Slot]state.BeaconState{
			params.BeaconConfig().SlotsPerEpoch*3 - 1: st,
		}},
		TimeFetcher:           mockChainService,
		OptimisticModeFetcher: mockChainService,
		FinalizationFetcher:   mockChainService,
	}

	t.Run("ok - ideal rewards", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/1"
		request := httptest.NewRequest("POST", url, nil)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &AttestationRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		require.Equal(t, 32, len(resp.Data.IdealRewards))
		sum := uint64(0)
		for _, r := range resp.Data.IdealRewards {
			hr, err := strconv.ParseUint(r.Head, 10, 64)
			require.NoError(t, err)
			sr, err := strconv.ParseUint(r.Source, 10, 64)
			require.NoError(t, err)
			tr, err := strconv.ParseUint(r.Target, 10, 64)
			require.NoError(t, err)
			sum += hr + sr + tr
		}
		// assert.Equal(t, uint64(20756849), sum)
		assert.Equal(t, uint64(1452726516), sum)
	})

	t.Run("ok - filtered vals", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/1"
		var body bytes.Buffer
		pubkey := fmt.Sprintf("%#x", secretKeys[10].PublicKey().Marshal())
		valIds, err := json.Marshal([]string{"20", pubkey})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &AttestationRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		require.Equal(t, 2, len(resp.Data.TotalRewards))
		sum := uint64(0)
		for _, r := range resp.Data.TotalRewards {
			hr, err := strconv.ParseUint(r.Head, 10, 64)
			require.NoError(t, err)
			sr, err := strconv.ParseUint(r.Source, 10, 64)
			require.NoError(t, err)
			tr, err := strconv.ParseUint(r.Target, 10, 64)
			require.NoError(t, err)
			sum += hr + sr + tr
		}
		// assert.Equal(t, uint64(794265), sum)
		assert.Equal(t, uint64(29953122), sum)
	})
	t.Run("ok - all vals", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/1"
		request := httptest.NewRequest("POST", url, nil)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &AttestationRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		require.Equal(t, 64, len(resp.Data.TotalRewards))
		sum := uint64(0)
		for _, r := range resp.Data.TotalRewards {
			hr, err := strconv.ParseUint(r.Head, 10, 64)
			require.NoError(t, err)
			sr, err := strconv.ParseUint(r.Source, 10, 64)
			require.NoError(t, err)
			tr, err := strconv.ParseUint(r.Target, 10, 64)
			require.NoError(t, err)
			sum += hr + sr + tr
		}
		// assert.Equal(t, uint64(54221955), sum)
		assert.Equal(t, uint64(1946953032), sum)
	})
	t.Run("ok - penalty", func(t *testing.T) {
		st, err := util.NewBeaconStateCapella()
		require.NoError(t, err)
		require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch*3-1))
		validators := make([]*zond.Validator, 0, valCount)
		balances := make([]uint64, 0, valCount)
		secretKeys := make([]dilithium.DilithiumKey, 0, valCount)
		for i := 0; i < valCount; i++ {
			dilithiumKey, err := dilithium.RandKey()
			require.NoError(t, err)
			secretKeys = append(secretKeys, dilithiumKey)
			validators = append(validators, &zond.Validator{
				PublicKey:         dilithiumKey.PublicKey().Marshal(),
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
				EffectiveBalance:  params.BeaconConfig().MaxEffectiveBalance / 64 * uint64(i),
			})
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance/64*uint64(i))
		}
		validators[63].Slashed = true
		require.NoError(t, st.SetValidators(validators))
		require.NoError(t, st.SetBalances(balances))
		require.NoError(t, st.SetInactivityScores(make([]uint64, len(validators))))
		participation := make([]byte, len(validators))
		for i := range participation {
			participation[i] = 0b111
		}
		require.NoError(t, st.SetCurrentParticipationBits(participation))
		require.NoError(t, st.SetPreviousParticipationBits(participation))

		currentSlot := params.BeaconConfig().SlotsPerEpoch * 3
		mockChainService := &mock.ChainService{Optimistic: true, Slot: &currentSlot}
		s := &Server{
			Stater: &testutil.MockStater{StatesBySlot: map[primitives.Slot]state.BeaconState{
				params.BeaconConfig().SlotsPerEpoch*3 - 1: st,
			}},
			TimeFetcher:           mockChainService,
			OptimisticModeFetcher: mockChainService,
			FinalizationFetcher:   mockChainService,
		}

		url := "http://only.the.epoch.number.at.the.end.is.important/1"
		var body bytes.Buffer
		valIds, err := json.Marshal([]string{"63"})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &AttestationRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		assert.Equal(t, "0", resp.Data.TotalRewards[0].Head)
		// assert.Equal(t, "-432270", resp.Data.TotalRewards[0].Source)
		assert.Equal(t, "-15521132", resp.Data.TotalRewards[0].Source)
		// assert.Equal(t, "-802788", resp.Data.TotalRewards[0].Target)
		assert.Equal(t, "-28824960", resp.Data.TotalRewards[0].Target)
	})
	t.Run("invalid validator index/pubkey", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/1"
		var body bytes.Buffer
		valIds, err := json.Marshal([]string{"10", "foo"})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusBadRequest, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusBadRequest, e.Code)
		assert.Equal(t, "foo is not a validator index or pubkey", e.Message)
	})
	t.Run("unknown validator pubkey", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/1"
		var body bytes.Buffer
		privkey, err := dilithium.RandKey()
		require.NoError(t, err)
		pubkey := fmt.Sprintf("%#x", privkey.PublicKey().Marshal())
		valIds, err := json.Marshal([]string{"10", pubkey})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusBadRequest, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusBadRequest, e.Code)
		assert.Equal(t, "No validator index found for pubkey "+pubkey, e.Message)
	})
	t.Run("validator index too large", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/1"
		var body bytes.Buffer
		valIds, err := json.Marshal([]string{"10", "999"})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusBadRequest, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusBadRequest, e.Code)
		assert.Equal(t, "Validator index 999 is too large. Maximum allowed index is 63", e.Message)
	})
	t.Run("invalid epoch", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/foo"
		request := httptest.NewRequest("POST", url, nil)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusBadRequest, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusBadRequest, e.Code)
		assert.Equal(t, true, strings.Contains(e.Message, "Could not decode epoch"))
	})
	t.Run("previous epoch", func(t *testing.T) {
		url := "http://only.the.epoch.number.at.the.end.is.important/2"
		request := httptest.NewRequest("POST", url, nil)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.AttestationRewards(writer, request)
		assert.Equal(t, http.StatusNotFound, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusNotFound, e.Code)
		assert.Equal(t, "Attestation rewards are available after two epoch transitions to ensure all attestations have a chance of inclusion", e.Message)
	})
}

func TestSyncCommiteeRewards(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	helpers.ClearCache()

	const valCount = 1024
	// we have to set the proposer index to the value that will be randomly chosen (fortunately it's deterministic)
	const proposerIndex = 7

	st, err := util.NewBeaconStateCapella()
	require.NoError(t, err)
	require.NoError(t, st.SetSlot(params.BeaconConfig().SlotsPerEpoch-1))
	validators := make([]*zond.Validator, 0, valCount)
	secretKeys := make([]dilithium.DilithiumKey, 0, valCount)
	for i := 0; i < valCount; i++ {
		dilithiumKey, err := dilithium.RandKey()
		require.NoError(t, err)
		secretKeys = append(secretKeys, dilithiumKey)
		validators = append(validators, &zond.Validator{
			PublicKey:         dilithiumKey.PublicKey().Marshal(),
			ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
			WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			EffectiveBalance:  params.BeaconConfig().MaxEffectiveBalance,
		})
	}
	require.NoError(t, st.SetValidators(validators))
	require.NoError(t, st.SetInactivityScores(make([]uint64, len(validators))))
	syncCommitteePubkeys := make([][]byte, fieldparams.SyncCommitteeLength)
	for i := 0; i < fieldparams.SyncCommitteeLength; i++ {
		syncCommitteePubkeys[i] = secretKeys[i].PublicKey().Marshal()
	}
	require.NoError(t, err)
	require.NoError(t, st.SetCurrentSyncCommittee(&zond.SyncCommittee{
		Pubkeys: syncCommitteePubkeys,
	}))

	b := util.HydrateSignedBeaconBlockCapella(util.NewBeaconBlockCapella())
	b.Block.Slot = 128
	b.Block.ProposerIndex = proposerIndex
	scBits := bitfield.NewBitvector16()
	// last 10 sync committee members didn't perform their duty
	for i := uint64(0); i < fieldparams.SyncCommitteeLength-2; i++ {
		scBits.SetBitAt(i, true)
	}
	domain, err := signing.Domain(st.Fork(), 0, params.BeaconConfig().DomainSyncCommittee, st.GenesisValidatorsRoot())
	require.NoError(t, err)
	sszBytes := primitives.SSZBytes("")
	r, err := signing.ComputeSigningRoot(&sszBytes, domain)
	require.NoError(t, err)
	// Bits set in sync committee bits determine which validators will be treated as participating in sync committee.
	// These validators have to sign the message.
	sigs := make([][]byte, fieldparams.SyncCommitteeLength-2)
	for i := range sigs {
		sigs[i] = secretKeys[i].Sign(r[:]).Marshal()
	}
	b.Block.Body.SyncAggregate = &zond.SyncAggregate{SyncCommitteeBits: scBits, SyncCommitteeSignatures: sigs}
	sbb, err := blocks.NewSignedBeaconBlock(b)
	require.NoError(t, err)
	require.NoError(t, err)

	currentSlot := params.BeaconConfig().SlotsPerEpoch
	mockChainService := &mock.ChainService{Optimistic: true, Slot: &currentSlot}
	s := &Server{
		Blocker: &testutil.MockBlocker{SlotBlockMap: map[primitives.Slot]interfaces.ReadOnlySignedBeaconBlock{
			128: sbb,
		}},
		OptimisticModeFetcher: mockChainService,
		FinalizationFetcher:   mockChainService,
		ReplayerBuilder:       mockstategen.NewMockReplayerBuilder(mockstategen.WithMockState(st)),
	}

	t.Run("ok - filtered vals", func(t *testing.T) {
		balances := make([]uint64, 0, valCount)
		for i := 0; i < valCount; i++ {
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
		}
		require.NoError(t, st.SetBalances(balances))

		url := "http://only.the.slot.number.at.the.end.is.important/128"
		var body bytes.Buffer
		pubkey := fmt.Sprintf("%#x", secretKeys[10].PublicKey().Marshal())
		valIds, err := json.Marshal([]string{"5", pubkey})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.SyncCommitteeRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &SyncCommitteeRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		require.Equal(t, 2, len(resp.Data))
		sum := uint64(0)
		for _, scReward := range resp.Data {
			r, err := strconv.ParseUint(scReward.Reward, 10, 64)
			require.NoError(t, err)
			sum += r
		}
		// assert.Equal(t, uint64(1396), sum)
		assert.Equal(t, uint64(395000), sum)
		assert.Equal(t, true, resp.ExecutionOptimistic)
		assert.Equal(t, false, resp.Finalized)
	})
	t.Run("ok - all vals", func(t *testing.T) {
		balances := make([]uint64, 0, valCount)
		for i := 0; i < valCount; i++ {
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
		}
		require.NoError(t, st.SetBalances(balances))

		url := "http://only.the.slot.number.at.the.end.is.important/128"
		request := httptest.NewRequest("POST", url, nil)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.SyncCommitteeRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &SyncCommitteeRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		require.Equal(t, 16, len(resp.Data))
		sum := 0
		for _, scReward := range resp.Data {
			r, err := strconv.Atoi(scReward.Reward)
			require.NoError(t, err)
			sum += r
		}
		// assert.Equal(t, 343416, sum)
		assert.Equal(t, 1975004, sum)
	})
	t.Run("ok - validator outside sync committee is ignored", func(t *testing.T) {
		balances := make([]uint64, 0, valCount)
		for i := 0; i < valCount; i++ {
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
		}
		require.NoError(t, st.SetBalances(balances))

		url := "http://only.the.slot.number.at.the.end.is.important/128"
		var body bytes.Buffer
		pubkey := fmt.Sprintf("%#x", secretKeys[10].PublicKey().Marshal())
		valIds, err := json.Marshal([]string{"12", "999", pubkey})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.SyncCommitteeRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &SyncCommitteeRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		require.Equal(t, 2, len(resp.Data))
		sum := 0
		for _, scReward := range resp.Data {
			r, err := strconv.Atoi(scReward.Reward)
			require.NoError(t, err)
			sum += r
		}
		// assert.Equal(t, 1396, sum)
		assert.Equal(t, 395000, sum)
	})

	t.Run("ok - proposer reward is deducted", func(t *testing.T) {
		balances := make([]uint64, 0, valCount)
		for i := 0; i < valCount; i++ {
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
		}
		require.NoError(t, st.SetBalances(balances))

		url := "http://only.the.slot.number.at.the.end.is.important/128"
		var body bytes.Buffer
		pubkey := fmt.Sprintf("%#x", secretKeys[10].PublicKey().Marshal())
		valIds, err := json.Marshal([]string{"5", "7", pubkey})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.SyncCommitteeRewards(writer, request)
		assert.Equal(t, http.StatusOK, writer.Code)
		resp := &SyncCommitteeRewardsResponse{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), resp))
		require.Equal(t, 3, len(resp.Data))
		sum := 0
		for _, scReward := range resp.Data {
			r, err := strconv.Atoi(scReward.Reward)
			require.NoError(t, err)
			sum += r
		}
		// assert.Equal(t, 2094, sum)
		assert.Equal(t, 197504, sum)
	})
	t.Run("invalid validator index/pubkey", func(t *testing.T) {
		balances := make([]uint64, 0, valCount)
		for i := 0; i < valCount; i++ {
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
		}
		require.NoError(t, st.SetBalances(balances))

		url := "http://only.the.slot.number.at.the.end.is.important/128"
		var body bytes.Buffer
		valIds, err := json.Marshal([]string{"10", "foo"})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.SyncCommitteeRewards(writer, request)
		assert.Equal(t, http.StatusBadRequest, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusBadRequest, e.Code)
		assert.Equal(t, "foo is not a validator index or pubkey", e.Message)
	})

	t.Run("unknown validator pubkey", func(t *testing.T) {
		balances := make([]uint64, 0, valCount)
		for i := 0; i < valCount; i++ {
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
		}
		require.NoError(t, st.SetBalances(balances))

		url := "http://only.the.slot.number.at.the.end.is.important/128"
		var body bytes.Buffer
		privkey, err := dilithium.RandKey()
		require.NoError(t, err)
		pubkey := fmt.Sprintf("%#x", privkey.PublicKey().Marshal())
		valIds, err := json.Marshal([]string{"10", pubkey})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.SyncCommitteeRewards(writer, request)
		assert.Equal(t, http.StatusBadRequest, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusBadRequest, e.Code)
		assert.Equal(t, "No validator index found for pubkey "+pubkey, e.Message)
	})
	t.Run("validator index too large", func(t *testing.T) {
		balances := make([]uint64, 0, valCount)
		for i := 0; i < valCount; i++ {
			balances = append(balances, params.BeaconConfig().MaxEffectiveBalance)
		}
		require.NoError(t, st.SetBalances(balances))

		url := "http://only.the.slot.number.at.the.end.is.important/128"
		var body bytes.Buffer
		valIds, err := json.Marshal([]string{"10", "9999"})
		require.NoError(t, err)
		_, err = body.Write(valIds)
		require.NoError(t, err)
		request := httptest.NewRequest("POST", url, &body)
		writer := httptest.NewRecorder()
		writer.Body = &bytes.Buffer{}

		s.SyncCommitteeRewards(writer, request)
		assert.Equal(t, http.StatusBadRequest, writer.Code)
		e := &http2.DefaultErrorJson{}
		require.NoError(t, json.Unmarshal(writer.Body.Bytes(), e))
		assert.Equal(t, http.StatusBadRequest, e.Code)
		assert.Equal(t, "Validator index 9999 is too large. Maximum allowed index is 1023", e.Message)
	})
}
