package stategen

import (
	"context"
	"encoding/hex"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/theQRL/qrysm/beacon-chain/state"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/encoding/bytesutil"
	"go.opencensus.io/trace"
)

// MigrateToCold advances the finalized info in between the cold and hot state sections.
// It moves the recent finalized states from the hot section to the cold section and
// only preserves the ones that are on archived point.
func (s *State) MigrateToCold(ctx context.Context, fRoot [32]byte) error {
	ctx, span := trace.StartSpan(ctx, "stateGen.MigrateToCold")
	defer span.End()

	// When migrating states we choose to acquire the migration lock before
	// proceeding. This is to prevent multiple migration routines from overwriting each
	// other.
	s.migrationLock.Lock()
	defer s.migrationLock.Unlock()

	s.finalizedInfo.lock.RLock()
	oldFSlot := s.finalizedInfo.slot
	s.finalizedInfo.lock.RUnlock()

	fBlock, err := s.beaconDB.Block(ctx, fRoot)
	if err != nil {
		return err
	}
	fSlot := fBlock.Block().Slot()
	if oldFSlot > fSlot {
		return nil
	}

	// Start at previous finalized slot, stop at current finalized slot (it will be handled in the next migration).
	// If the slot is on archived point, save the state of that slot to the DB.
	for slot := oldFSlot; slot < fSlot; slot++ {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if slot%s.slotsPerArchivedPoint == 0 && slot != 0 {
			cached, exists, err := s.epochBoundaryStateCache.getBySlot(slot)
			if err != nil {
				return fmt.Errorf("could not get epoch boundary state for slot %d", slot)
			}
			// The cache is populated for every epoch boundary block that gets processed and is
			// only ever evicted by size or when a block turns out to be invalid, never when a
			// block loses fork choice. Its slot index is also first-write-wins, so a reorged
			// sibling processed ahead of the canonical block at the same slot keeps the slot
			// key. Either way the cached root can be non-canonical, so it has to be checked
			// before its state is archived.
			if exists && !s.beaconDB.IsFinalizedBlock(ctx, cached.root) {
				log.WithFields(logrus.Fields{
					"slot": slot,
					"root": fmt.Sprintf("%#x", cached.root),
				}).Debug("Ignoring non-canonical epoch boundary state while migrating to cold")
				exists = false
			}

			var aRoot [32]byte
			var aState state.BeaconState

			// When the epoch boundary state is not in cache due to skip slot scenario,
			// we have to regenerate the state which will represent epoch boundary.
			// By finding the highest available canonical block below epoch boundary slot,
			// we generate the state for that block root.
			if exists {
				aRoot = cached.root
				aState = cached.state
			} else {
				aRoot, err = s.canonicalRootAtOrBelow(ctx, slot-1, oldFSlot)
				if err != nil {
					return err
				}
				// There's no need to generate the state if the state already exists in the DB.
				// We can skip saving the state.
				if !s.beaconDB.HasState(ctx, aRoot) {
					aState, err = s.StateByRoot(ctx, aRoot)
					if err != nil {
						return err
					}
				}
			}

			if s.beaconDB.HasState(ctx, aRoot) {
				// If you are migrating a state and its already part of the hot state cache saved to the db,
				// you can just remove it from the hot state cache as it becomes redundant.
				s.saveHotStateDB.lock.Lock()
				roots := s.saveHotStateDB.blockRootsOfSavedStates
				for i := range roots {
					if aRoot == roots[i] {
						s.saveHotStateDB.blockRootsOfSavedStates = append(roots[:i], roots[i+1:]...)
						// There shouldn't be duplicated roots in `blockRootsOfSavedStates`.
						// Break here is ok.
						break
					}
				}
				s.saveHotStateDB.lock.Unlock()
				continue
			}

			if err := s.beaconDB.SaveState(ctx, aState, aRoot); err != nil {
				return err
			}
			log.WithFields(
				logrus.Fields{
					"slot": aState.Slot(),
					"root": hex.EncodeToString(bytesutil.Trunc(aRoot[:])),
				}).Info("Saved state in DB")
		}
	}

	// Update finalized info in memory.
	fInfo, ok, err := s.epochBoundaryStateCache.getByBlockRoot(fRoot)
	if err != nil {
		return err
	}
	if ok {
		s.SaveFinalizedState(fSlot, fRoot, fInfo.state)
	}

	return nil
}

// canonicalRootAtOrBelow returns the root of the highest canonical block at or below the given slot,
// skipping slots that hold only blocks which lost fork choice. Those are never deleted from the slot index.
// Canonicality comes from the finalized index, which is decisive below the finalized checkpoint slot.
// floor is a slot already known to be canonical; resolving past it is an error rather than a walk to genesis.
func (s *State) canonicalRootAtOrBelow(ctx context.Context, slot, floor primitives.Slot) ([32]byte, error) {
	// HighestRootsBelowSlot reports a strictly lower slot, so next decreases every round.
	for next := slot + 1; ; {
		high, roots, err := s.beaconDB.HighestRootsBelowSlot(ctx, next)
		if err != nil {
			return [32]byte{}, err
		}
		if high < floor {
			return [32]byte{}, errUnknownBlock
		}
		canonical := make([][32]byte, 0, 1)
		for _, r := range roots {
			if s.beaconDB.IsFinalizedBlock(ctx, r) {
				canonical = append(canonical, r)
			}
		}
		switch len(canonical) {
		case 1:
			return canonical[0], nil
		case 0:
			if high == 0 {
				return [32]byte{}, errUnknownBlock
			}
			next = high
		default:
			return [32]byte{}, errUnknownBlock
		}
	}
}
