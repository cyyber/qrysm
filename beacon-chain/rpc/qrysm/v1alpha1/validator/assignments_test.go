package validator

import (
	"context"
	"encoding/binary"
	"testing"
	"time"

	"github.com/theQRL/go-qrllib/dilithium"
	mockChain "github.com/theQRL/qrysm/v4/beacon-chain/blockchain/testing"
	"github.com/theQRL/qrysm/v4/beacon-chain/cache"
	"github.com/theQRL/qrysm/v4/beacon-chain/cache/depositcache"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/v4/beacon-chain/core/transition"
	mockExecution "github.com/theQRL/qrysm/v4/beacon-chain/execution/testing"
	mockSync "github.com/theQRL/qrysm/v4/beacon-chain/sync/initial-sync/testing"
	fieldparams "github.com/theQRL/qrysm/v4/config/fieldparams"
	"github.com/theQRL/qrysm/v4/config/params"
	"github.com/theQRL/qrysm/v4/consensus-types/primitives"
	"github.com/theQRL/qrysm/v4/encoding/bytesutil"
	enginev1 "github.com/theQRL/qrysm/v4/proto/engine/v1"
	zondpb "github.com/theQRL/qrysm/v4/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/v4/testing/assert"
	"github.com/theQRL/qrysm/v4/testing/require"
	"github.com/theQRL/qrysm/v4/testing/util"
)

// pubKey is a helper to generate a well-formed public key.
func pubKey(i uint64) []byte {
	pubKey := make([]byte, dilithium.CryptoPublicKeyBytes)
	binary.LittleEndian.PutUint64(pubKey, i)
	return pubKey
}

func TestGetDuties_OK(t *testing.T) {
	genesis := util.NewBeaconBlock()
	depChainStart := params.BeaconConfig().MinGenesisActiveValidatorCount
	deposits, _, err := util.DeterministicDepositsAndKeys(depChainStart)
	require.NoError(t, err)
	zond1Data, err := util.DeterministicZond1Data(len(deposits))
	require.NoError(t, err)
	bs, err := transition.GenesisBeaconState(context.Background(), deposits, 0, zond1Data, &enginev1.ExecutionPayload{})
	require.NoError(t, err, "Could not setup genesis bs")
	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	pubKeys := make([][]byte, len(deposits))
	indices := make([]uint64, len(deposits))
	for i := 0; i < len(deposits); i++ {
		pubKeys[i] = deposits[i].Data.PublicKey
		indices[i] = uint64(i)
	}

	chain := &mockChain.ChainService{
		State: bs, Root: genesisRoot[:], Genesis: time.Now(),
	}
	vs := &Server{
		HeadFetcher:            chain,
		TimeFetcher:            chain,
		SyncChecker:            &mockSync.Sync{IsSyncing: false},
		ProposerSlotIndexCache: cache.NewProposerPayloadIDsCache(),
	}

	// Test the first validator in registry.
	req := &zondpb.DutiesRequest{
		PublicKeys: [][]byte{deposits[0].Data.PublicKey},
	}
	res, err := vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	if res.CurrentEpochDuties[0].AttesterSlot > bs.Slot()+params.BeaconConfig().SlotsPerEpoch {
		t.Errorf("Assigned slot %d can't be higher than %d",
			res.CurrentEpochDuties[0].AttesterSlot, bs.Slot()+params.BeaconConfig().SlotsPerEpoch)
	}

	// Test the last validator in registry.
	lastValidatorIndex := depChainStart - 1
	req = &zondpb.DutiesRequest{
		PublicKeys: [][]byte{deposits[lastValidatorIndex].Data.PublicKey},
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	if res.CurrentEpochDuties[0].AttesterSlot > bs.Slot()+params.BeaconConfig().SlotsPerEpoch {
		t.Errorf("Assigned slot %d can't be higher than %d",
			res.CurrentEpochDuties[0].AttesterSlot, bs.Slot()+params.BeaconConfig().SlotsPerEpoch)
	}

	// We request for duties for all validators.
	req = &zondpb.DutiesRequest{
		PublicKeys: pubKeys,
		Epoch:      0,
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	for i := 0; i < len(res.CurrentEpochDuties); i++ {
		assert.Equal(t, primitives.ValidatorIndex(i), res.CurrentEpochDuties[i].ValidatorIndex)
	}
}

/*
func TestGetAltairDuties_SyncCommitteeOK(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	//cfg := params.BeaconConfig().Copy()
	//cfg.AltairForkEpoch = primitives.Epoch(0)
	//params.OverrideBeaconConfig(cfg)

	genesis := util.NewBeaconBlock()
	deposits, _, err := util.DeterministicDepositsAndKeys(params.BeaconConfig().SyncCommitteeSize)
	require.NoError(t, err)
	zond1Data, err := util.DeterministicZond1Data(len(deposits))
	require.NoError(t, err)
	bs, err := util.GenesisBeaconState(context.Background(), deposits, 0, zond1Data)
	require.NoError(t, err, "Could not setup genesis bs")
	h := &zondpb.BeaconBlockHeader{
		StateRoot:  bytesutil.PadTo([]byte{'a'}, fieldparams.RootLength),
		ParentRoot: bytesutil.PadTo([]byte{'b'}, fieldparams.RootLength),
		BodyRoot:   bytesutil.PadTo([]byte{'c'}, fieldparams.RootLength),
	}
	require.NoError(t, bs.SetLatestBlockHeader(h))
	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	syncCommittee, err := altair.NextSyncCommittee(context.Background(), bs)
	require.NoError(t, err)
	require.NoError(t, bs.SetCurrentSyncCommittee(syncCommittee))
	pubKeys := make([][]byte, len(deposits))
	indices := make([]uint64, len(deposits))
	for i := 0; i < len(deposits); i++ {
		pubKeys[i] = deposits[i].Data.PublicKey
		indices[i] = uint64(i)
	}
	require.NoError(t, bs.SetSlot(params.BeaconConfig().SlotsPerEpoch*primitives.Slot(params.BeaconConfig().EpochsPerSyncCommitteePeriod)-1))
	require.NoError(t, helpers.UpdateSyncCommitteeCache(bs))

	pubkeysAs2592ByteType := make([][dilithium.CryptoPublicKeyBytes]byte, len(pubKeys))
	for i, pk := range pubKeys {
		pubkeysAs2592ByteType[i] = bytesutil.ToBytes2592(pk)
	}

	slot := uint64(params.BeaconConfig().SlotsPerEpoch) * uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod) * params.BeaconConfig().SecondsPerSlot
	chain := &mockChain.ChainService{
		State: bs, Root: genesisRoot[:], Genesis: time.Now().Add(time.Duration(-1*int64(slot-1)) * time.Second),
	}
	vs := &Server{
		HeadFetcher:            chain,
		TimeFetcher:            chain,
		Zond1InfoFetcher:       &mockExecution.Chain{},
		SyncChecker:            &mockSync.Sync{IsSyncing: false},
		ProposerSlotIndexCache: cache.NewProposerPayloadIDsCache(),
	}

	// Test the first validator in registry.
	req := &zondpb.DutiesRequest{
		PublicKeys: [][]byte{deposits[0].Data.PublicKey},
	}
	res, err := vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	if res.CurrentEpochDuties[0].AttesterSlot > bs.Slot()+params.BeaconConfig().SlotsPerEpoch {
		t.Errorf("Assigned slot %d can't be higher than %d",
			res.CurrentEpochDuties[0].AttesterSlot, bs.Slot()+params.BeaconConfig().SlotsPerEpoch)
	}

	// Test the last validator in registry.
	lastValidatorIndex := params.BeaconConfig().SyncCommitteeSize - 1
	req = &zondpb.DutiesRequest{
		PublicKeys: [][]byte{deposits[lastValidatorIndex].Data.PublicKey},
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	if res.CurrentEpochDuties[0].AttesterSlot > bs.Slot()+params.BeaconConfig().SlotsPerEpoch {
		t.Errorf("Assigned slot %d can't be higher than %d",
			res.CurrentEpochDuties[0].AttesterSlot, bs.Slot()+params.BeaconConfig().SlotsPerEpoch)
	}

	// We request for duties for all validators.
	req = &zondpb.DutiesRequest{
		PublicKeys: pubKeys,
		Epoch:      0,
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	for i := 0; i < len(res.CurrentEpochDuties); i++ {
		require.Equal(t, primitives.ValidatorIndex(i), res.CurrentEpochDuties[i].ValidatorIndex)
	}
	for i := 0; i < len(res.CurrentEpochDuties); i++ {
		require.Equal(t, true, res.CurrentEpochDuties[i].IsSyncCommittee)
		// Current epoch and next epoch duties should be equal before the sync period epoch boundary.
		require.Equal(t, res.CurrentEpochDuties[i].IsSyncCommittee, res.NextEpochDuties[i].IsSyncCommittee)
	}

	// Current epoch and next epoch duties should not be equal at the sync period epoch boundary.
	req = &zondpb.DutiesRequest{
		PublicKeys: pubKeys,
		Epoch:      params.BeaconConfig().EpochsPerSyncCommitteePeriod - 1,
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	for i := 0; i < len(res.CurrentEpochDuties); i++ {
		require.NotEqual(t, res.CurrentEpochDuties[i].IsSyncCommittee, res.NextEpochDuties[i].IsSyncCommittee)
	}
}
*/

// TODO(rgeraldes24) fix
/*
func TestGetBellatrixDuties_SyncCommitteeOK(t *testing.T) {
	params.SetupTestConfigCleanup(t)

	genesis := util.NewBeaconBlock()
	deposits, _, err := util.DeterministicDepositsAndKeys(params.BeaconConfig().SyncCommitteeSize)
	require.NoError(t, err)
	zond1Data, err := util.DeterministicZond1Data(len(deposits))
	require.NoError(t, err)
	bs, err := util.GenesisBeaconState(context.Background(), deposits, 0, zond1Data)
	h := &zondpb.BeaconBlockHeader{
		StateRoot:  bytesutil.PadTo([]byte{'a'}, fieldparams.RootLength),
		ParentRoot: bytesutil.PadTo([]byte{'b'}, fieldparams.RootLength),
		BodyRoot:   bytesutil.PadTo([]byte{'c'}, fieldparams.RootLength),
	}
	require.NoError(t, bs.SetLatestBlockHeader(h))
	require.NoError(t, err, "Could not setup genesis bs")
	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	syncCommittee, err := altair.NextSyncCommittee(context.Background(), bs)
	require.NoError(t, err)
	require.NoError(t, bs.SetCurrentSyncCommittee(syncCommittee))
	pubKeys := make([][]byte, len(deposits))
	indices := make([]uint64, len(deposits))
	for i := 0; i < len(deposits); i++ {
		pubKeys[i] = deposits[i].Data.PublicKey
		indices[i] = uint64(i)
	}
	require.NoError(t, bs.SetSlot(params.BeaconConfig().SlotsPerEpoch*primitives.Slot(params.BeaconConfig().EpochsPerSyncCommitteePeriod)-1))
	require.NoError(t, helpers.UpdateSyncCommitteeCache(bs))

	pubkeysAs2592ByteType := make([][dilithium.CryptoPublicKeyBytes]byte, len(pubKeys))
	for i, pk := range pubKeys {
		pubkeysAs2592ByteType[i] = bytesutil.ToBytes2592(pk)
	}

	slot := uint64(params.BeaconConfig().SlotsPerEpoch) * uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod) * params.BeaconConfig().SecondsPerSlot
	chain := &mockChain.ChainService{
		State: bs, Root: genesisRoot[:], Genesis: time.Now().Add(time.Duration(-1*int64(slot-1)) * time.Second),
	}
	vs := &Server{
		HeadFetcher:            chain,
		TimeFetcher:            chain,
		Zond1InfoFetcher:       &mockExecution.Chain{},
		SyncChecker:            &mockSync.Sync{IsSyncing: false},
		ProposerSlotIndexCache: cache.NewProposerPayloadIDsCache(),
	}

	// Test the first validator in registry.
	req := &zondpb.DutiesRequest{
		PublicKeys: [][]byte{deposits[0].Data.PublicKey},
	}
	res, err := vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	if res.CurrentEpochDuties[0].AttesterSlot > bs.Slot()+params.BeaconConfig().SlotsPerEpoch {
		t.Errorf("Assigned slot %d can't be higher than %d",
			res.CurrentEpochDuties[0].AttesterSlot, bs.Slot()+params.BeaconConfig().SlotsPerEpoch)
	}

	// Test the last validator in registry.
	lastValidatorIndex := params.BeaconConfig().SyncCommitteeSize - 1
	req = &zondpb.DutiesRequest{
		PublicKeys: [][]byte{deposits[lastValidatorIndex].Data.PublicKey},
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	if res.CurrentEpochDuties[0].AttesterSlot > bs.Slot()+params.BeaconConfig().SlotsPerEpoch {
		t.Errorf("Assigned slot %d can't be higher than %d",
			res.CurrentEpochDuties[0].AttesterSlot, bs.Slot()+params.BeaconConfig().SlotsPerEpoch)
	}

	// We request for duties for all validators.
	req = &zondpb.DutiesRequest{
		PublicKeys: pubKeys,
		Epoch:      0,
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	for i := 0; i < len(res.CurrentEpochDuties); i++ {
		assert.Equal(t, primitives.ValidatorIndex(i), res.CurrentEpochDuties[i].ValidatorIndex)
	}
	for i := 0; i < len(res.CurrentEpochDuties); i++ {
		assert.Equal(t, true, res.CurrentEpochDuties[i].IsSyncCommittee)
		// Current epoch and next epoch duties should be equal before the sync period epoch boundary.
		assert.Equal(t, res.CurrentEpochDuties[i].IsSyncCommittee, res.NextEpochDuties[i].IsSyncCommittee)
	}

	// Current epoch and next epoch duties should not be equal at the sync period epoch boundary.
	req = &zondpb.DutiesRequest{
		PublicKeys: pubKeys,
		Epoch:      params.BeaconConfig().EpochsPerSyncCommitteePeriod - 1,
	}
	res, err = vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	for i := 0; i < len(res.CurrentEpochDuties); i++ {
		require.NotEqual(t, res.CurrentEpochDuties[i].IsSyncCommittee, res.NextEpochDuties[i].IsSyncCommittee)
	}
}
*/

func TestGetAltairDuties_UnknownPubkey(t *testing.T) {
	params.SetupTestConfigCleanup(t)
	//cfg := params.BeaconConfig().Copy()
	//cfg.AltairForkEpoch = primitives.Epoch(0)
	//params.OverrideBeaconConfig(cfg)

	genesis := util.NewBeaconBlock()
	deposits, _, err := util.DeterministicDepositsAndKeys(params.BeaconConfig().SyncCommitteeSize)
	require.NoError(t, err)
	zond1Data, err := util.DeterministicZond1Data(len(deposits))
	require.NoError(t, err)
	bs, err := util.GenesisBeaconState(context.Background(), deposits, 0, zond1Data)
	require.NoError(t, err)
	h := &zondpb.BeaconBlockHeader{
		StateRoot:  bytesutil.PadTo([]byte{'a'}, fieldparams.RootLength),
		ParentRoot: bytesutil.PadTo([]byte{'b'}, fieldparams.RootLength),
		BodyRoot:   bytesutil.PadTo([]byte{'c'}, fieldparams.RootLength),
	}
	require.NoError(t, bs.SetLatestBlockHeader(h))
	require.NoError(t, err, "Could not setup genesis bs")
	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	require.NoError(t, bs.SetSlot(params.BeaconConfig().SlotsPerEpoch*primitives.Slot(params.BeaconConfig().EpochsPerSyncCommitteePeriod)-1))
	require.NoError(t, helpers.UpdateSyncCommitteeCache(bs))

	slot := uint64(params.BeaconConfig().SlotsPerEpoch) * uint64(params.BeaconConfig().EpochsPerSyncCommitteePeriod) * params.BeaconConfig().SecondsPerSlot
	chain := &mockChain.ChainService{
		State: bs, Root: genesisRoot[:], Genesis: time.Now().Add(time.Duration(-1*int64(slot-1)) * time.Second),
	}
	depositCache, err := depositcache.New()
	require.NoError(t, err)

	vs := &Server{
		HeadFetcher:            chain,
		TimeFetcher:            chain,
		Zond1InfoFetcher:       &mockExecution.Chain{},
		SyncChecker:            &mockSync.Sync{IsSyncing: false},
		DepositFetcher:         depositCache,
		ProposerSlotIndexCache: cache.NewProposerPayloadIDsCache(),
	}

	unknownPubkey := bytesutil.PadTo([]byte{'u'}, 2592)
	req := &zondpb.DutiesRequest{
		PublicKeys: [][]byte{unknownPubkey},
	}
	res, err := vs.GetDuties(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, false, res.CurrentEpochDuties[0].IsSyncCommittee)
	assert.Equal(t, false, res.NextEpochDuties[0].IsSyncCommittee)
}

func TestGetDuties_SlotOutOfUpperBound(t *testing.T) {
	chain := &mockChain.ChainService{
		Genesis: time.Now(),
	}
	vs := &Server{
		TimeFetcher: chain,
	}
	req := &zondpb.DutiesRequest{
		Epoch: primitives.Epoch(chain.CurrentSlot()/params.BeaconConfig().SlotsPerEpoch + 2),
	}
	_, err := vs.duties(context.Background(), req)
	require.ErrorContains(t, "can not be greater than next epoch", err)
}

func TestGetDuties_CurrentEpoch_ShouldNotFail(t *testing.T) {
	genesis := util.NewBeaconBlock()
	depChainStart := params.BeaconConfig().MinGenesisActiveValidatorCount
	deposits, _, err := util.DeterministicDepositsAndKeys(depChainStart)
	require.NoError(t, err)
	zond1Data, err := util.DeterministicZond1Data(len(deposits))
	require.NoError(t, err)
	bState, err := transition.GenesisBeaconState(context.Background(), deposits, 0, zond1Data, &enginev1.ExecutionPayload{})
	require.NoError(t, err, "Could not setup genesis state")
	// Set state to non-epoch start slot.
	require.NoError(t, bState.SetSlot(5))

	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	pubKeys := make([][dilithium.CryptoPublicKeyBytes]byte, len(deposits))
	indices := make([]uint64, len(deposits))
	for i := 0; i < len(deposits); i++ {
		pubKeys[i] = bytesutil.ToBytes2592(deposits[i].Data.PublicKey)
		indices[i] = uint64(i)
	}

	chain := &mockChain.ChainService{
		State: bState, Root: genesisRoot[:], Genesis: time.Now(),
	}
	vs := &Server{
		HeadFetcher:            chain,
		TimeFetcher:            chain,
		SyncChecker:            &mockSync.Sync{IsSyncing: false},
		ProposerSlotIndexCache: cache.NewProposerPayloadIDsCache(),
	}

	// Test the first validator in registry.
	req := &zondpb.DutiesRequest{
		PublicKeys: [][]byte{deposits[0].Data.PublicKey},
	}
	res, err := vs.GetDuties(context.Background(), req)
	require.NoError(t, err)
	assert.Equal(t, 1, len(res.CurrentEpochDuties), "Expected 1 assignment")
}

func TestGetDuties_MultipleKeys_OK(t *testing.T) {
	genesis := util.NewBeaconBlock()
	depChainStart := uint64(64)

	deposits, _, err := util.DeterministicDepositsAndKeys(depChainStart)
	require.NoError(t, err)
	zond1Data, err := util.DeterministicZond1Data(len(deposits))
	require.NoError(t, err)
	bs, err := transition.GenesisBeaconState(context.Background(), deposits, 0, zond1Data, &enginev1.ExecutionPayload{})
	require.NoError(t, err, "Could not setup genesis bs")
	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(t, err, "Could not get signing root")

	pubKeys := make([][dilithium.CryptoPublicKeyBytes]byte, len(deposits))
	indices := make([]uint64, len(deposits))
	for i := 0; i < len(deposits); i++ {
		pubKeys[i] = bytesutil.ToBytes2592(deposits[i].Data.PublicKey)
		indices[i] = uint64(i)
	}

	chain := &mockChain.ChainService{
		State: bs, Root: genesisRoot[:], Genesis: time.Now(),
	}
	vs := &Server{
		HeadFetcher:            chain,
		TimeFetcher:            chain,
		SyncChecker:            &mockSync.Sync{IsSyncing: false},
		ProposerSlotIndexCache: cache.NewProposerPayloadIDsCache(),
	}

	pubkey0 := deposits[0].Data.PublicKey
	pubkey1 := deposits[1].Data.PublicKey

	// Test the first validator in registry.
	req := &zondpb.DutiesRequest{
		PublicKeys: [][]byte{pubkey0, pubkey1},
	}
	res, err := vs.GetDuties(context.Background(), req)
	require.NoError(t, err, "Could not call epoch committee assignment")
	assert.Equal(t, 2, len(res.CurrentEpochDuties))
	// TODO(rgeraldes24) double check
	assert.Equal(t, primitives.Slot(5), res.CurrentEpochDuties[0].AttesterSlot)
	assert.Equal(t, primitives.Slot(5), res.CurrentEpochDuties[0].AttesterSlot)
	// assert.Equal(t, primitives.Slot(4), res.CurrentEpochDuties[0].AttesterSlot)
	// assert.Equal(t, primitives.Slot(4), res.CurrentEpochDuties[1].AttesterSlot)
}

func TestGetDuties_SyncNotReady(t *testing.T) {
	vs := &Server{
		SyncChecker: &mockSync.Sync{IsSyncing: true},
	}
	_, err := vs.GetDuties(context.Background(), &zondpb.DutiesRequest{})
	assert.ErrorContains(t, "Syncing to latest head", err)
}

func TestAssignValidatorToSubnet(t *testing.T) {
	k := pubKey(3)

	assignValidatorToSubnet(k, zondpb.ValidatorStatus_ACTIVE)
	coms, ok, exp := cache.SubnetIDs.GetPersistentSubnets(k)
	require.Equal(t, true, ok, "No cache entry found for validator")
	assert.Equal(t, params.BeaconConfig().RandomSubnetsPerValidator, uint64(len(coms)))
	epochDuration := time.Duration(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot))
	totalTime := time.Duration(params.BeaconConfig().EpochsPerRandomSubnetSubscription) * epochDuration * time.Second
	receivedTime := time.Until(exp.Round(time.Second))
	if receivedTime < totalTime {
		t.Fatalf("Expiration time of %f was less than expected duration of %f ", receivedTime.Seconds(), totalTime.Seconds())
	}
}

func TestAssignValidatorToSyncSubnet(t *testing.T) {
	k := pubKey(3)
	committee := make([][]byte, 0)

	for i := 0; i < 100; i++ {
		committee = append(committee, pubKey(uint64(i)))
	}
	sCommittee := &zondpb.SyncCommittee{
		Pubkeys: committee,
	}
	registerSyncSubnet(0, 0, k, sCommittee, zondpb.ValidatorStatus_ACTIVE)
	coms, _, ok, exp := cache.SyncSubnetIDs.GetSyncCommitteeSubnets(k, 0)
	require.Equal(t, true, ok, "No cache entry found for validator")
	assert.Equal(t, uint64(1), uint64(len(coms)))
	epochDuration := time.Duration(params.BeaconConfig().SlotsPerEpoch.Mul(params.BeaconConfig().SecondsPerSlot))
	totalTime := time.Duration(params.BeaconConfig().EpochsPerSyncCommitteePeriod) * epochDuration * time.Second
	receivedTime := time.Until(exp.Round(time.Second)).Round(time.Second)
	if receivedTime < totalTime {
		t.Fatalf("Expiration time of %f was less than expected duration of %f ", receivedTime.Seconds(), totalTime.Seconds())
	}
}

func BenchmarkCommitteeAssignment(b *testing.B) {
	genesis := util.NewBeaconBlock()
	depChainStart := uint64(8192 * 2)
	deposits, _, err := util.DeterministicDepositsAndKeys(depChainStart)
	require.NoError(b, err)
	zond1Data, err := util.DeterministicZond1Data(len(deposits))
	require.NoError(b, err)
	bs, err := transition.GenesisBeaconState(context.Background(), deposits, 0, zond1Data, &enginev1.ExecutionPayload{})
	require.NoError(b, err, "Could not setup genesis bs")
	genesisRoot, err := genesis.Block.HashTreeRoot()
	require.NoError(b, err, "Could not get signing root")

	pubKeys := make([][dilithium.CryptoPublicKeyBytes]byte, len(deposits))
	indices := make([]uint64, len(deposits))
	for i := 0; i < len(deposits); i++ {
		pubKeys[i] = bytesutil.ToBytes2592(deposits[i].Data.PublicKey)
		indices[i] = uint64(i)
	}

	vs := &Server{
		HeadFetcher: &mockChain.ChainService{State: bs, Root: genesisRoot[:]},
		SyncChecker: &mockSync.Sync{IsSyncing: false},
	}

	// Create request for all validators in the system.
	pks := make([][]byte, len(deposits))
	for i, deposit := range deposits {
		pks[i] = deposit.Data.PublicKey
	}
	req := &zondpb.DutiesRequest{
		PublicKeys: pks,
		Epoch:      0,
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := vs.GetDuties(context.Background(), req)
		assert.NoError(b, err)
	}
}