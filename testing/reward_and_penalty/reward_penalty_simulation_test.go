package reward_and_penalty

import (
	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"testing"

	"github.com/theQRL/go-bitfield"
	"github.com/theQRL/qrysm/beacon-chain/core/altair"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/beacon-chain/state"
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

type EpochStats struct {
	ProposerReward  uint64
	ProposerPenalty uint64
	AttesterReward  uint64
	AttesterPenalty uint64
	SyncReward      uint64
	SyncPenalty     uint64
}

type SimulationMetrics struct {
	Stats EpochStats
}

func (m *SimulationMetrics) Add(s EpochStats) {
	m.Stats.ProposerReward += s.ProposerReward
	m.Stats.ProposerPenalty += s.ProposerPenalty
	m.Stats.AttesterReward += s.AttesterReward
	m.Stats.AttesterPenalty += s.AttesterPenalty
	m.Stats.SyncReward += s.SyncReward
	m.Stats.SyncPenalty += s.SyncPenalty
}

func (m *SimulationMetrics) TotalIssued() uint64 {
	return m.Stats.ProposerReward + m.Stats.AttesterReward + m.Stats.SyncReward
}

func (m *SimulationMetrics) TotalBurned() uint64 {
	return m.Stats.ProposerPenalty + m.Stats.AttesterPenalty + m.Stats.SyncPenalty
}

func (m *SimulationMetrics) printFinalDetailedReport() {
	toZond := func(val uint64) float64 { return float64(val) / 1e9 }

	fmt.Printf("\nBREAKDOWN:\n")

	totalIssued := m.TotalIssued()
	fmt.Printf("1. TOTAL ISSUED:   %20s Gwei (%.4f Zond)\n", formatWithCommas(int64(totalIssued)), toZond(totalIssued))
	fmt.Printf("   ├─ Proposers:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(m.Stats.ProposerReward)), toZond(m.Stats.ProposerReward), percentage(m.Stats.ProposerReward, totalIssued))
	fmt.Printf("   ├─ Attesters:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(m.Stats.AttesterReward)), toZond(m.Stats.AttesterReward), percentage(m.Stats.AttesterReward, totalIssued))
	fmt.Printf("   └─ Sync Comm:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(m.Stats.SyncReward)), toZond(m.Stats.SyncReward), percentage(m.Stats.SyncReward, totalIssued))

	totalBurned := m.TotalBurned()
	burnedStr := strings.Replace(formatWithCommas(int64(totalBurned)), "+", "-", 1)

	fmt.Printf("2. TOTAL BURNED:   %20s Gwei (%.4f Zond)\n", burnedStr, toZond(totalBurned))
	fmt.Printf("   ├─ Proposers:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		strings.Replace(formatWithCommas(int64(m.Stats.ProposerPenalty)), "+", "-", 1), toZond(m.Stats.ProposerPenalty), percentage(m.Stats.ProposerPenalty, totalBurned))
	fmt.Printf("   ├─ Attesters:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		strings.Replace(formatWithCommas(int64(m.Stats.AttesterPenalty)), "+", "-", 1), toZond(m.Stats.AttesterPenalty), percentage(m.Stats.AttesterPenalty, totalBurned))
	fmt.Printf("   └─ Sync Comm:   %20s Gwei (%.4f Zond) [%.2f%%]\n",
		strings.Replace(formatWithCommas(int64(m.Stats.SyncPenalty)), "+", "-", 1), toZond(m.Stats.SyncPenalty), percentage(m.Stats.SyncPenalty, totalBurned))

	fmt.Printf("========================================\n")
}

type Simulator struct {
	ctx           context.Context
	state         state.BeaconState
	numValidators int
	metrics       *SimulationMetrics
}

func NewSimulator(numValidators int) (*Simulator, error) {
	s, err := util.NewBeaconStateCapella(WithMockValidators(numValidators))
	if err != nil {
		return nil, err
	}
	return &Simulator{
		ctx:           context.Background(),
		state:         s,
		numValidators: numValidators,
		metrics:       &SimulationMetrics{},
	}, nil
}

func (s *Simulator) applyAndMeasure(action func() error) (reward uint64, penalty uint64, err error) {
	balances := s.state.Balances()
	prevBalances := make([]uint64, len(balances))
	copy(prevBalances, balances)

	if err := action(); err != nil {
		return 0, 0, err
	}

	newBalances := s.state.Balances()
	for i, newBal := range newBalances {
		oldBal := prevBalances[i]
		if newBal > oldBal {
			reward += (newBal - oldBal)
		} else if newBal < oldBal {
			penalty += (oldBal - newBal)
		}
	}
	return reward, penalty, nil
}

func (s *Simulator) processAttestation(slot primitives.Slot, commIndex primitives.CommitteeIndex, missProb float64) error {
	attestation, err := createAttestation(s.ctx, s.state, slot, commIndex, missProb)
	if err != nil {
		return err
	}

	activeBal, err := helpers.TotalActiveBalance(s.state)
	if err != nil {
		return err
	}

	newState, err := altair.ProcessAttestationNoVerifySignature(s.ctx, s.state, attestation, activeBal)
	if err != nil {
		return err
	}

	s.state = newState
	return nil
}

func (s *Simulator) processSyncCommittee(missProb float64) (uint64, error) {
	cfg := params.BeaconConfig()
	committeeSize := cfg.SyncCommitteeSize

	syncBits := bitfield.NewBitvector16()
	limit := uint64(16)
	if committeeSize < 16 {
		limit = committeeSize
	}

	for i := uint64(0); i < limit; i++ {
		if rand.Float64() > missProb {
			syncBits.SetBitAt(i, true)
		}
	}

	activeBalance, err := helpers.TotalActiveBalance(s.state)
	if err != nil {
		return 0, err
	}

	proposerReward, participantReward, err := altair.SyncRewards(activeBalance)
	if err != nil {
		return 0, err
	}

	proposerIndex, err := helpers.BeaconProposerIndex(s.ctx, s.state)
	if err != nil {
		return 0, err
	}

	earnedProposerReward := uint64(0)

	for i := uint64(0); i < limit; i++ {
		vIdx := primitives.ValidatorIndex(i)

		if syncBits.BitAt(i) {
			if err := helpers.IncreaseBalance(s.state, vIdx, participantReward); err != nil {
				return 0, err
			}
			earnedProposerReward += proposerReward
		} else {
			if err := helpers.DecreaseBalance(s.state, vIdx, participantReward); err != nil {
				return 0, err
			}
		}
	}

	if earnedProposerReward > 0 {
		if err := helpers.IncreaseBalance(s.state, proposerIndex, earnedProposerReward); err != nil {
			return 0, err
		}
	}

	return earnedProposerReward, nil
}

func (s *Simulator) RunEpoch(epoch int, missProb float64, pendingReward *uint64) (EpochStats, error) {
	cfg := params.BeaconConfig()
	slotsPerEpoch := uint64(cfg.SlotsPerEpoch)
	startSlot := s.state.Slot()

	if err := setParticipationFlags(s.state, s.numValidators, FlagsNoParticipation); err != nil {
		return EpochStats{}, err
	}

	stats := EpochStats{}
	activeIndices, err := helpers.ActiveValidatorIndices(s.ctx, s.state, slots.ToEpoch(startSlot))
	if err != nil {
		return stats, err
	}
	committeeCount := helpers.SlotCommitteeCount(uint64(len(activeIndices)))

	for i := range slotsPerEpoch {
		currentSlot := startSlot + primitives.Slot(i)

		if currentSlot == 0 {
			s.state.SetSlot(currentSlot + 1)
			continue
		}

		var proposerCut uint64
		r, p, err := s.applyAndMeasure(func() error {
			var errInner error
			proposerCut, errInner = s.processSyncCommittee(missProb)
			return errInner
		})
		if err != nil {
			return stats, err
		}

		if r >= proposerCut {
			stats.SyncReward += (r - proposerCut)
			stats.ProposerReward += proposerCut
		} else {
			stats.SyncReward += r
		}
		stats.SyncPenalty += p

		if i == 0 {
			stats.ProposerReward += *pendingReward
		} else {
			attestationSlot := currentSlot - 1
			for c := range committeeCount {
				commIndex := primitives.CommitteeIndex(c)
				r, p, err := s.applyAndMeasure(func() error {
					return s.processAttestation(attestationSlot, commIndex, missProb)
				})
				if err != nil {
					return stats, err
				}
				stats.ProposerReward += r
				stats.ProposerPenalty += p
			}
		}

		s.state.SetSlot(currentSlot + 1)
	}

	*pendingReward = 0
	lastEpochSlot := startSlot + primitives.Slot(slotsPerEpoch) - 1

	if lastEpochSlot > 0 {
		for c := range committeeCount {
			commIndex := primitives.CommitteeIndex(c)
			r, _, err := s.applyAndMeasure(func() error {
				return s.processAttestation(lastEpochSlot, commIndex, missProb)
			})
			if err != nil {
				return stats, err
			}
			*pendingReward += r
		}
	}

	if epoch == 0 {
		_ = s.state.SetPreviousParticipationBits(make([]byte, s.numValidators))
	} else {
		currentFlags, _ := s.state.CurrentEpochParticipation()
		previousFlags, _ := s.state.PreviousEpochParticipation()
		mergedFlags := make([]byte, len(currentFlags))
		for i := range mergedFlags {
			mergedFlags[i] = currentFlags[i] | previousFlags[i]
		}
		_ = s.state.SetPreviousParticipationBits(mergedFlags)
	}

	_ = s.state.SetCurrentParticipationBits(make([]byte, s.numValidators))

	balancesBefore := make([]uint64, len(s.state.Balances()))
	copy(balancesBefore, s.state.Balances())

	r, p, err := s.applyAndMeasure(func() error {
		var err error
		s.state, err = altair.ProcessEpoch(s.ctx, s.state)
		return err
	})
	if err != nil {

		return stats, err
	}

	if epoch == 0 {
		for i, bal := range balancesBefore {
			s.state.Balances()[i] = bal
		}
		stats.AttesterReward = 0
		stats.AttesterPenalty = 0
	} else {
		stats.AttesterReward = r
		stats.AttesterPenalty = p
	}

	return stats, nil
}

func rewardAndPenaltySimulation(numValidators int, epochsToRun int, missStrategy MissStrategy) error {
	sim, err := NewSimulator(numValidators)
	if err != nil {
		return err
	}

	initialSupply := getTotalRawBalance(sim.state)
	pendingReward := uint64(0)

	printConfig()

	for epoch := range epochsToRun {
		missProb := missStrategy(epoch)

		stats, err := sim.RunEpoch(epoch, missProb, &pendingReward)
		if err != nil {
			return err
		}

		sim.metrics.Add(stats)
		currentSupply := getTotalRawBalance(sim.state)

		printDetailedEpochReport(epoch, currentSupply, initialSupply, stats, sim.metrics, sim.state)
	}

	sim.metrics.printFinalDetailedReport()
	return nil
}

func printDetailedEpochReport(epoch int, currentSupply uint64, initialSupply uint64, stats EpochStats, cumMetrics *SimulationMetrics, state state.BeaconState) {
	toZond := func(val uint64) float64 { return float64(val) / 1e9 }
	toZondSigned := func(val int64) float64 { return float64(val) / 1e9 }

	thisEpochIssued := stats.ProposerReward + stats.AttesterReward + stats.SyncReward
	thisEpochBurned := stats.ProposerPenalty + stats.AttesterPenalty + stats.SyncPenalty
	thisEpochNet := int64(thisEpochIssued) - int64(thisEpochBurned)

	cumulativeNet := int64(currentSupply) - int64(initialSupply)
	cumulativeIssued := cumMetrics.TotalIssued()

	statusTag := "[ OK ]"
	if thisEpochNet < 0 {
		statusTag = "[BURN]"
	}

	fmt.Printf("%s Epoch %d Report:\n", statusTag, epoch)
	fmt.Printf("   [THIS EPOCH] Net: %18s Gwei (%+10.4f Zond) | Issued: %18s Gwei (+%8.4f) | Burned: %18s Gwei (-%8.4f)\n",
		formatWithCommas(thisEpochNet), toZondSigned(thisEpochNet),
		formatWithCommas(int64(thisEpochIssued)), toZond(thisEpochIssued),
		strings.Replace(formatWithCommas(int64(thisEpochBurned)), "+", "-", 1), toZond(thisEpochBurned))

	fmt.Printf("      ├─ Proposers: %18s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(stats.ProposerReward)), toZond(stats.ProposerReward), percentage(stats.ProposerReward, thisEpochIssued))
	fmt.Printf("      ├─ Attesters: %18s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(stats.AttesterReward)), toZond(stats.AttesterReward), percentage(stats.AttesterReward, thisEpochIssued))
	fmt.Printf("      └─ Sync Comm: %18s Gwei (%.4f Zond) [%.2f%%]\n",
		formatWithCommas(int64(stats.SyncReward)), toZond(stats.SyncReward), percentage(stats.SyncReward, thisEpochIssued))

	fmt.Printf("   [CUMULATIVE] Net: %18s Gwei (%+10.4f Zond) | Issued: %18s Gwei (+%8.4f) | Supply: %24s Gwei (%.4f Zond)\n",
		formatWithCommas(cumulativeNet), toZondSigned(cumulativeNet),
		formatWithCommas(int64(cumulativeIssued)), toZond(cumulativeIssued),
		formatWithCommas(int64(currentSupply)), toZond(currentSupply))

	finalized := state.FinalizedCheckpoint().Epoch
	justified := state.CurrentJustifiedCheckpoint().Epoch
	fmt.Printf(".  [STATUS] Finalized Epoch: %d | Justified Epoch: %d\n", finalized, justified)
	fmt.Println("   ----------------------------------------------------------------")
}

func formatWithCommas(n int64) string {
	in := []rune(fmt.Sprintf("%d", n))
	sign := ""
	if n < 0 {
		sign = "-"
		in = in[1:]
	} else if n > 0 {
		sign = "+"
	}

	out := make([]rune, 0, len(in)+(len(in)-1)/3)
	for i, r := range in {
		if i > 0 && (len(in)-i)%3 == 0 {
			out = append(out, ',')
		}
		out = append(out, r)
	}
	return sign + string(out)
}

func percentage(part, total uint64) float64 {
	if total == 0 {
		return 0
	}
	return (float64(part) / float64(total)) * 100
}

func printConfig() {
	cfg := params.BeaconConfig()
	fmt.Println("------------------------------------------------")
	fmt.Printf("CONFIGURATION:\n")
	fmt.Printf("Seconds Per Slot: %d\n", cfg.SecondsPerSlot)
	fmt.Printf("Slots Per Epoch:  %d\n", cfg.SlotsPerEpoch)
	fmt.Printf("Sync Committee Size: %d\n", cfg.SyncCommitteeSize)
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

func createAttestation(ctx context.Context, state state.BeaconState, slot primitives.Slot, commIndex primitives.CommitteeIndex, missProb float64) (*zondpb.Attestation, error) {
	committee, err := helpers.BeaconCommitteeFromState(ctx, state, slot, commIndex)
	if err != nil {
		return nil, err
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
			s.PreviousEpochParticipation[i] = FlagsNoParticipation
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

		s.CurrentSyncCommittee = &zondpb.SyncCommittee{Pubkeys: committeeKeys}
		s.NextSyncCommittee = &zondpb.SyncCommittee{Pubkeys: committeeKeys}

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

func TestRewardAndPenaltySimulation(t *testing.T) {
	numValidators := 3000
	epochsToRun := 250
	missStrategy := func(epoch int) float64 { return 0.02 }

	err := rewardAndPenaltySimulation(numValidators, epochsToRun, missStrategy)
	require.NoError(t, err)
}

func TestRewardAndPenaltyInactivityModeSimulation(t *testing.T) {
	numValidators := 3000
	epochsToRun := 10
	missStrategy := func(epoch int) float64 { return 0.40 }

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
