package reward_and_penalty

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strconv"
	"strings"
	"testing"

	"github.com/theQRL/go-bitfield"
	"github.com/theQRL/qrysm/beacon-chain/core/altair"
	e "github.com/theQRL/qrysm/beacon-chain/core/epoch"
	"github.com/theQRL/qrysm/beacon-chain/core/epoch/precompute"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/beacon-chain/state"
	state_native "github.com/theQRL/qrysm/beacon-chain/state/state-native"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
	"github.com/theQRL/qrysm/time/slots"
)

const (
	FlagsFullParticipation = 0xff
	FlagsNoParticipation   = 0x00
	RootLength             = 32
	PublicKeyLength        = 2592
	SignatureLength        = 4627
)

type MissStrategy func(epoch int) float64

type SimulationMetrics struct {
	ProposerRewards   uint64
	ProposerPenalties uint64

	AttesterRewards   uint64
	AttesterPenalties uint64

	SyncCommitteeRewards   uint64
	SyncCommitteePenalties uint64
}

func (m *SimulationMetrics) TotalIssued() uint64 {
	return m.ProposerRewards + m.AttesterRewards + m.SyncCommitteeRewards
}

func (m *SimulationMetrics) TotalBurned() uint64 {
	return m.ProposerPenalties + m.AttesterPenalties + m.SyncCommitteePenalties
}

func (m *SimulationMetrics) AddEpochData(attesterReward, attesterPenalty, propReward, propPenalty, syncReward, syncPenalty uint64) {
	m.ProposerRewards += propReward
	m.ProposerPenalties += propPenalty
	m.AttesterRewards += attesterReward
	m.AttesterPenalties += attesterPenalty
	m.SyncCommitteeRewards += syncReward
	m.SyncCommitteePenalties += syncPenalty
}

func formatWithCommas(n int64) string {
	sign := "+"
	if n < 0 {
		sign = "-"
		n = -n
	}

	str := strconv.FormatInt(n, 10)
	var result []string
	for i, c := range reverse(str) {
		if i > 0 && i%3 == 0 {
			result = append(result, ",")
		}
		result = append(result, string(c))
	}

	return sign + strings.Join(reverse(strings.Join(result, "")), "")
}

func formatWithCommasUint(n uint64) string {
	return formatWithCommas(int64(n))[1:]
}

func reverse(s string) []string {
	chars := strings.Split(s, "")
	for i, j := 0, len(chars)-1; i < j; i, j = i+1, j-1 {
		chars[i], chars[j] = chars[j], chars[i]
	}
	return chars
}

func (m *SimulationMetrics) printFinalDetailedReport() {
	toZond := func(val uint64) float64 { return float64(val) / 1e9 }

	fmt.Printf("\nBREAKDOWN:\n")

	totalIssued := m.TotalIssued()
	fmt.Printf("1. TOTAL ISSUED:   %20s Gwei (%.4f Zond)\n", formatWithCommas(int64(totalIssued)), toZond(totalIssued))

	fmt.Printf("   ├─ Proposers:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(m.ProposerRewards)), toZond(m.ProposerRewards), percentage(m.ProposerRewards, totalIssued))
	fmt.Printf("   ├─ Attesters:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(m.AttesterRewards)), toZond(m.AttesterRewards), percentage(m.AttesterRewards, totalIssued))
	fmt.Printf("   └─ Sync Comm:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(m.SyncCommitteeRewards)), toZond(m.SyncCommitteeRewards), percentage(m.SyncCommitteeRewards, totalIssued))

	totalBurned := m.TotalBurned()
	burnedStr := strings.Replace(formatWithCommas(int64(totalBurned)), "+", "-", 1)

	fmt.Printf("2. TOTAL BURNED:   %20s Gwei (%.4f Zond)\n", burnedStr, toZond(totalBurned))
	fmt.Printf("   ├─ Proposers:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		strings.Replace(formatWithCommas(int64(m.ProposerPenalties)), "+", "-", 1), toZond(m.ProposerPenalties), percentage(m.ProposerPenalties, totalBurned))
	fmt.Printf("   ├─ Attesters:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		strings.Replace(formatWithCommas(int64(m.AttesterPenalties)), "+", "-", 1), toZond(m.AttesterPenalties), percentage(m.AttesterPenalties, totalBurned))
	fmt.Printf("   └─ Sync Comm:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		strings.Replace(formatWithCommas(int64(m.SyncCommitteePenalties)), "+", "-", 1), toZond(m.SyncCommitteePenalties), percentage(m.SyncCommitteePenalties, totalBurned))

	fmt.Printf("========================================\n")
}
func percentage(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return (float64(part) / float64(total)) * 100
}

func WithMockValidators(numValidators int) func(*zondpb.BeaconStateCapella) error {
	return func(s *zondpb.BeaconStateCapella) error {
		cfg := params.BeaconConfig()
		maxBalance := cfg.MaxEffectiveBalance

		s.Validators = make([]*zondpb.Validator, numValidators)
		s.Balances = make([]uint64, numValidators)
		s.PreviousEpochParticipation = make([]byte, numValidators)
		s.CurrentEpochParticipation = make([]byte, numValidators)
		s.InactivityScores = make([]uint64, numValidators)

		for i := 0; i < numValidators; i++ {

			pubKey := make([]byte, PublicKeyLength)
			pubKey[0] = byte(i >> 8)
			pubKey[1] = byte(i)

			s.Validators[i] = &zondpb.Validator{
				PublicKey:         pubKey,
				EffectiveBalance:  maxBalance,
				Slashed:           false,
				ActivationEpoch:   0,
				ExitEpoch:         params.BeaconConfig().FarFutureEpoch,
				WithdrawableEpoch: params.BeaconConfig().FarFutureEpoch,
			}
			s.Balances[i] = maxBalance
			s.PreviousEpochParticipation[i] = FlagsFullParticipation
			s.CurrentEpochParticipation[i] = FlagsNoParticipation
		}

		committeeKeys := make([][]byte, cfg.SyncCommitteeSize)
		for i := uint64(0); i < cfg.SyncCommitteeSize; i++ {
			valIdx := int(i)
			if valIdx < numValidators {
				committeeKeys[i] = s.Validators[valIdx].PublicKey
			} else {
				committeeKeys[i] = make([]byte, PublicKeyLength)
			}
		}

		s.CurrentSyncCommittee = &zondpb.SyncCommittee{
			Pubkeys: committeeKeys,
		}
		s.NextSyncCommittee = &zondpb.SyncCommittee{
			Pubkeys: committeeKeys,
		}

		return nil
	}
}

func getTotalRawBalance(s state.BeaconState) uint64 {
	total := uint64(0)
	balances := s.Balances()
	for _, b := range balances {
		total += b
	}
	return total
}

func setParticipationFlags(state state.BeaconState, numValidators int, value byte) error {
	flags := make([]byte, numValidators)
	if value != 0 {
		for i := range flags {
			flags[i] = value
		}
	}
	if err := state.SetCurrentParticipationBits(flags); err != nil {
		return err
	}
	if err := state.SetPreviousParticipationBits(flags); err != nil {
		return err
	}

	return nil
}

func printDetailedEpochReport(
	epoch int,
	currentSupply uint64,
	deltaSupply int64,
	deltaIssued uint64,
	deltaBurned uint64,
	propReward uint64,
	attesterReward uint64,
	syncReward uint64,
	totalNetChange int64,
	m *SimulationMetrics,
	state state.BeaconState,
) {
	toZond := func(val uint64) float64 { return float64(val) / 1e9 }
	toZondSigned := func(val int64) float64 { return float64(val) / 1e9 }

	statusTag := "[ OK ]"

	if deltaSupply < 0 {
		statusTag = "[BURN]"
	} else if deltaBurned > deltaIssued {
		statusTag = "[WARN]"
	}

	fmt.Printf("%s Epoch %d Report:\n", statusTag, epoch)

	fmt.Printf("   [THIS EPOCH] Net: %18s Gwei (%+10.4f Zond) | Issued: %18s Gwei (+%8.4f) | Burned: %18s Gwei (-%8.4f)\n",
		formatWithCommas(deltaSupply), toZondSigned(deltaSupply),
		formatWithCommas(int64(deltaIssued)), toZond(deltaIssued),
		strings.Replace(formatWithCommas(int64(deltaBurned)), "+", "-", 1), toZond(deltaBurned))

	fmt.Printf("      ├─ Proposers: %18s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(propReward)), toZond(propReward), percentage(propReward, deltaIssued))
	fmt.Printf("      ├─ Attesters: %18s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(attesterReward)), toZond(attesterReward), percentage(attesterReward, deltaIssued))
	fmt.Printf("      └─ Sync Comm: %18s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(syncReward)), toZond(syncReward), percentage(syncReward, deltaIssued))

	fmt.Printf("   [CUMULATIVE] Net: %18s Gwei (%+10.4f Zond) | Issued: %18s Gwei (+%8.4f) | Supply: %24s Gwei (%.4f Zond)\n",
		formatWithCommas(totalNetChange), toZondSigned(totalNetChange),
		formatWithCommas(int64(m.TotalIssued())), toZond(m.TotalIssued()),
		formatWithCommasUint(currentSupply), toZond(currentSupply))

	finalized := state.FinalizedCheckpoint().Epoch
	justified := state.CurrentJustifiedCheckpoint().Epoch
	fmt.Printf(".  [STATUS] Finalized Epoch: %d | Justified Epoch: %d\n", finalized, justified)

	fmt.Println("   ----------------------------------------------------------------")
}

func processSyncCommitteeSimulated(ctx context.Context, state state.BeaconState, missProb float64) (state.BeaconState, uint64, uint64, uint64, error) {
	cfg := params.BeaconConfig()
	committeeSize := cfg.SyncCommitteeSize

	syncBits := bitfield.NewBitvector16()
	for i := range committeeSize {
		if i >= 16 {
			break
		}
		if rand.Float64() > missProb {
			syncBits.SetBitAt(i, true)
		}
	}

	activeBalance, err := helpers.TotalActiveBalance(state)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	proposerReward, participantReward, err := altair.SyncRewards(activeBalance)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	proposerIndex, err := helpers.BeaconProposerIndex(ctx, state)
	if err != nil {
		return nil, 0, 0, 0, err
	}

	totalParticipantReward := uint64(0)
	earnedProposerReward := uint64(0)
	totalPenalized := uint64(0)

	for i := range committeeSize {
		if i >= 16 {
			break
		}

		vIdx := primitives.ValidatorIndex(i)

		if syncBits.BitAt(i) {
			if err := helpers.IncreaseBalance(state, vIdx, participantReward); err != nil {
				return nil, 0, 0, 0, err
			}
			totalParticipantReward += participantReward
			earnedProposerReward += proposerReward
		} else {
			if err := helpers.DecreaseBalance(state, vIdx, participantReward); err != nil {
				return nil, 0, 0, 0, err
			}
			totalPenalized += participantReward
		}
	}

	if earnedProposerReward > 0 {
		if err := helpers.IncreaseBalance(state, proposerIndex, earnedProposerReward); err != nil {
			return nil, 0, 0, 0, err
		}
	}

	return state, totalParticipantReward, earnedProposerReward, totalPenalized, nil
}

func createAttestation(ctx context.Context, state state.BeaconState, slot primitives.Slot, commIndex primitives.CommitteeIndex, missProb float64) (*zondpb.Attestation, error) {
	committee, err := helpers.BeaconCommitteeFromState(ctx, state, slot, commIndex)
	if err != nil {
		return nil, fmt.Errorf("failed to get committee: %w", err)
	}

	committeeSize := uint64(len(committee))
	aggregationBits := bitfield.NewBitlist(committeeSize)
	activeParticipants := 0
	for i := range committeeSize {
		if rand.Float64() > missProb {
			aggregationBits.SetBitAt(i, true)
			activeParticipants++
		}
	}

	mockSignatures := make([][]byte, activeParticipants)
	for i := range mockSignatures {
		mockSignatures[i] = make([]byte, SignatureLength)
	}

	attestationEpoch := slots.ToEpoch(slot)
	var sourceCheckpoint *zondpb.Checkpoint

	if attestationEpoch == slots.ToEpoch(state.Slot()) {
		sourceCheckpoint = state.CurrentJustifiedCheckpoint()
	} else {
		sourceCheckpoint = state.PreviousJustifiedCheckpoint()
	}

	blockRoot, err := state.BlockRootAtIndex(uint64(slot % params.BeaconConfig().SlotsPerHistoricalRoot))
	if err != nil {
		blockRoot = make([]byte, RootLength)
	}

	return &zondpb.Attestation{
		AggregationBits: aggregationBits,
		Data: &zondpb.AttestationData{
			Slot:            slot,
			CommitteeIndex:  commIndex,
			BeaconBlockRoot: blockRoot,
			Source:          sourceCheckpoint,
			Target: &zondpb.Checkpoint{
				Epoch: attestationEpoch,
				Root:  make([]byte, RootLength),
			},
		},
		Signatures: mockSignatures,
	}, nil
}

func processSingleAttestation(
	ctx context.Context,
	state state.BeaconState,
	slot primitives.Slot,
	commIndex primitives.CommitteeIndex,
	missProb float64,
) (state.BeaconState, uint64, uint64, error) {
	attestation, err := createAttestation(ctx, state, slot, commIndex, missProb)
	if err != nil {
		return nil, 0, 0, err
	}

	proposerIndex, err := helpers.BeaconProposerIndex(ctx, state)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("could not get proposer index: %w", err)
	}

	balanceBefore := state.Balances()[proposerIndex]
	totalActiveBalance, err := helpers.TotalActiveBalance(state)
	if err != nil {
		return nil, 0, 0, err
	}

	newState, err := altair.ProcessAttestationNoVerifySignature(ctx, state, attestation, totalActiveBalance)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("ProcessAttestation failed: %w", err)
	}

	balanceAfter := newState.Balances()[proposerIndex]
	var reward, penalty uint64
	if balanceAfter > balanceBefore {
		reward = balanceAfter - balanceBefore
	} else if balanceAfter < balanceBefore {
		penalty = balanceBefore - balanceAfter
	}

	return newState, reward, penalty, nil
}

func simulateWholeEpochProposers(ctx context.Context, state state.BeaconState, missProb float64, pendingRewardFromPrevEpoch uint64) (state.BeaconState, uint64, uint64, uint64, uint64, uint64, error) {
	cfg := params.BeaconConfig()
	slotsPerEpoch := uint64(cfg.SlotsPerEpoch)
	startSlot := state.Slot()
	epoch := slots.ToEpoch(startSlot)

	totalPropReward := uint64(0)
	totalPropPenalty := uint64(0)
	totalSyncReward := uint64(0)
	totalSyncPenalty := uint64(0)

	activeIndices, err := helpers.ActiveValidatorIndices(ctx, state, epoch)
	if err != nil {
		return nil, 0, 0, 0, 0, 0, fmt.Errorf("could not get active indices: %w", err)
	}
	activeValCount := uint64(len(activeIndices))
	committeeCount := helpers.SlotCommitteeCount(activeValCount)

	for i := range slotsPerEpoch {
		currentSlot := startSlot + primitives.Slot(i)

		var syncParticipantReward, syncProposerReward, syncBurned uint64
		if currentSlot > 0 {
			state, syncParticipantReward, syncProposerReward, syncBurned, err = processSyncCommitteeSimulated(ctx, state, missProb)
			if err != nil {
				return nil, 0, 0, 0, 0, 0, fmt.Errorf("sync committee failed: %v", err)
			}
		}

		totalSyncReward += syncParticipantReward
		totalPropReward += syncProposerReward
		totalSyncPenalty += syncBurned

		attestationSlot := currentSlot - 1

		if i == 0 {
			totalPropReward += pendingRewardFromPrevEpoch
		} else {
			for c := range committeeCount {
				commIndex := primitives.CommitteeIndex(c)

				var r, p uint64
				state, r, p, err = processSingleAttestation(ctx, state, attestationSlot, commIndex, missProb)
				if err != nil {
					return nil, 0, 0, 0, 0, 0, fmt.Errorf("error at slot %d comm %d: %w", attestationSlot, commIndex, err)
				}

				totalPropReward += r
				totalPropPenalty += p
			}
		}

		state.SetSlot(currentSlot + 1)
	}

	pendingRewardForNext := uint64(0)
	lastEpochSlot := startSlot + primitives.Slot(slotsPerEpoch) - 1
	for c := range committeeCount {
		commIndex := primitives.CommitteeIndex(c)
		var err error
		var r uint64
		state, r, _, err = processSingleAttestation(ctx, state, lastEpochSlot, commIndex, missProb)
		if err != nil {
			return nil, 0, 0, 0, 0, 0, fmt.Errorf("failed processing final slot attestations: %w", err)
		}
		pendingRewardForNext += r
	}

	return state, totalPropReward, totalPropPenalty, totalSyncReward, totalSyncPenalty, pendingRewardForNext, nil
}

func processEpoch(ctx context.Context, state state.BeaconState) (state.BeaconState, uint64, uint64, error) {
	currentFlags, _ := state.CurrentEpochParticipation()
	previousFlags, _ := state.PreviousEpochParticipation()

	mergedFlags := make([]byte, len(currentFlags))
	for i := range mergedFlags {
		mergedFlags[i] = currentFlags[i] | previousFlags[i]
	}

	if err := state.SetPreviousParticipationBits(mergedFlags); err != nil {
		return nil, 0, 0, err
	}
	if err := state.SetCurrentParticipationBits(make([]byte, len(currentFlags))); err != nil {
		return nil, 0, 0, err
	}

	validators, balance, err := altair.InitializePrecomputeValidators(ctx, state)
	if err != nil {
		return nil, 0, 0, err
	}

	validators, balance, err = altair.ProcessEpochParticipation(ctx, state, balance, validators)
	if err != nil {
		return nil, 0, 0, err
	}

	state, err = precompute.ProcessJustificationAndFinalizationPreCompute(state, balance)
	if err != nil {
		return nil, 0, 0, err
	}

	state, validators, err = altair.ProcessInactivityScores(ctx, state, validators)
	if err != nil {
		return nil, 0, 0, err
	}

	prevBalances := make([]uint64, len(state.(*state_native.BeaconState).Balances()))
	copy(prevBalances, state.Balances())

	state, err = altair.ProcessRewardsAndPenaltiesPrecompute(state, balance, validators)
	if err != nil {
		return nil, 0, 0, err
	}

	state, err = altair.ProcessParticipationFlagUpdates(state)
	if err != nil {
		return nil, 0, 0, err
	}

	state, err = e.ProcessRandaoMixesReset(state)
	if err != nil {
		return nil, 0, 0, err
	}

	attesterReward := uint64(0)
	attesterPenalty := uint64(0)
	for idx, newBal := range state.Balances() {
		oldBal := prevBalances[idx]
		if newBal > oldBal {
			attesterReward += (newBal - oldBal)
		} else if newBal < oldBal {
			attesterPenalty += (oldBal - newBal)
		}
	}

	return state, attesterReward, attesterPenalty, nil
}

func rewardAndPenaltySimulation(numValidators int, epochsToRun int, missStrategy MissStrategy) error {
	cfg := params.BeaconConfig()

	fmt.Println("------------------------------------------------")
	fmt.Printf("CONFIGURATION:\n")
	fmt.Printf("Seconds Per Slot: %d\n", cfg.SecondsPerSlot)
	fmt.Printf("Slots Per Epoch:  %d\n", cfg.SlotsPerEpoch)
	fmt.Printf("Sync Committee Size: %d\n", cfg.SyncCommitteeSize)

	state, err := util.NewBeaconStateCapella(WithMockValidators(numValidators))
	if err != nil {
		return fmt.Errorf("Failed to create new beacon state: %v", err)
	}

	initialSupply := getTotalRawBalance(state)

	ctx := context.Background()

	metrics := &SimulationMetrics{}

	prevIssued := uint64(0)
	prevBurned := uint64(0)
	prevSupply := initialSupply

	pendingReward := uint64(0)

	for epoch := range epochsToRun {
		if err := setParticipationFlags(state, numValidators, FlagsNoParticipation); err != nil {
			return fmt.Errorf("Failed to clear flags: %v", err)
		}

		missProb := missStrategy(epoch)

		var totalPropReward, totalPropPenalty, totalSyncReward, totalSyncPenalty uint64
		state, totalPropReward, totalPropPenalty, totalSyncReward, totalSyncPenalty, pendingReward, err = simulateWholeEpochProposers(ctx, state, missProb, pendingReward)
		if err != nil {
			return fmt.Errorf("Proposer simulation failed: %v", err)
		}

		var attesterReward, attesterPenalty uint64
		if epoch > 0 {
			state, attesterReward, attesterPenalty, err = processEpoch(ctx, state)
			if err != nil {
				return fmt.Errorf("Epoch boundary processing failed: %v", err)
			}
		}
		metrics.AddEpochData(attesterReward, attesterPenalty, totalPropReward, totalPropPenalty, totalSyncReward, totalSyncPenalty)

		currentIssued := metrics.TotalIssued()
		currentBurned := metrics.TotalBurned()
		currentSupply := getTotalRawBalance(state)

		deltaIssued := currentIssued - prevIssued
		deltaBurned := currentBurned - prevBurned
		deltaSupply := int64(currentSupply) - int64(prevSupply)

		totalNetChange := int64(currentSupply) - int64(initialSupply)

		printDetailedEpochReport(
			epoch,
			currentSupply,
			deltaSupply,
			deltaIssued,
			deltaBurned,
			totalPropReward,
			attesterReward,
			totalSyncReward,
			totalNetChange,
			metrics,
			state)

		prevIssued = currentIssued
		prevBurned = currentBurned
		prevSupply = currentSupply
	}

	metrics.printFinalDetailedReport()

	return nil
}

func TestRewardAndPenaltySimulation(t *testing.T) {
	numValidators := 3000
	epochsToRun := 250

	missStrategy := func(epoch int) float64 {
		return 0.02
	}

	err := rewardAndPenaltySimulation(numValidators, epochsToRun, missStrategy)
	require.NoError(t, err)
}

func TestRewardAndPenaltyInactivityModeSimulation(t *testing.T) {
	numValidators := 3000
	epochsToRun := 10

	missStrategy := func(epoch int) float64 {
		return 0.40
	}

	err := rewardAndPenaltySimulation(numValidators, epochsToRun, missStrategy)
	require.NoError(t, err)
}

func TestRewardAndPenaltyDynamicScenario(t *testing.T) {
	numValidators := 3000
	epochsToRun := 300

	dynamicStrategy := func(epoch int) float64 {
		if epoch < 50 {
			return 0.02
		}
		if epoch >= 50 && epoch <= 120 {
			return 0.45
		}
		return 0.02
	}

	err := rewardAndPenaltySimulation(numValidators, epochsToRun, dynamicStrategy)
	require.NoError(t, err)
}

func TestVerifyAgainstPrivateTestnet(t *testing.T) {
	numValidators := 512
	epochsToRun := 10

	missStrategy := func(epoch int) float64 { return 0.0 }

	err := rewardAndPenaltySimulation(numValidators, epochsToRun, missStrategy)
	require.NoError(t, err)
}
