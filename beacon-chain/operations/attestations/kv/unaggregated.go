package kv

import (
	"context"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/beacon-chain/core/helpers"
	"github.com/theQRL/qrysm/consensus-types/primitives"
	zondpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"go.opencensus.io/trace"
)

// SaveUnaggregatedAttestation saves an unaggregated attestation in cache.
func (c *AttCaches) SaveUnaggregatedAttestation(att *zondpb.Attestation) error {
	if att == nil {
		return nil
	}
	if helpers.IsAggregated(att) {
		return errors.New("attestation is aggregated")
	}

	seen, err := c.hasSeenBit(att)
	if err != nil {
		return err
	}
	if seen {
		return nil
	}

	r, err := hashFn(att)
	if err != nil {
		return errors.Wrap(err, "could not tree hash attestation")
	}
	att = zondpb.CopyAttestation(att) // Copied.
	c.unAggregateAttLock.Lock()
	defer c.unAggregateAttLock.Unlock()
	c.unAggregatedAtt[r] = att

	return nil
}

// SaveUnaggregatedAttestations saves a list of unaggregated attestations in cache.
func (c *AttCaches) SaveUnaggregatedAttestations(atts []*zondpb.Attestation) error {
	for _, att := range atts {
		if err := c.SaveUnaggregatedAttestation(att); err != nil {
			return err
		}
	}

	return nil
}

// UnaggregatedAttestations returns all the unaggregated attestations in cache.
func (c *AttCaches) UnaggregatedAttestations() ([]*zondpb.Attestation, error) {
	c.unAggregateAttLock.RLock()
	defer c.unAggregateAttLock.RUnlock()
	unAggregatedAtts := c.unAggregatedAtt
	atts := make([]*zondpb.Attestation, 0, len(unAggregatedAtts))
	for _, att := range unAggregatedAtts {
		seen, err := c.hasSeenBit(att)
		if err != nil {
			return nil, err
		}
		if !seen {
			atts = append(atts, zondpb.CopyAttestation(att) /* Copied */)
		}
	}
	return atts, nil
}

// UnaggregatedAttestationsBySlotIndex returns the unaggregated attestations in cache,
// filtered by committee index and slot.
func (c *AttCaches) UnaggregatedAttestationsBySlotIndex(ctx context.Context, slot primitives.Slot, committeeIndex primitives.CommitteeIndex) []*zondpb.Attestation {
	ctx, span := trace.StartSpan(ctx, "operations.attestations.kv.UnaggregatedAttestationsBySlotIndex")
	defer span.End()

	atts := make([]*zondpb.Attestation, 0)

	c.unAggregateAttLock.RLock()
	defer c.unAggregateAttLock.RUnlock()

	unAggregatedAtts := c.unAggregatedAtt
	for _, a := range unAggregatedAtts {
		if slot == a.Data.Slot && committeeIndex == a.Data.CommitteeIndex {
			atts = append(atts, a)
		}
	}

	return atts
}

// DeleteUnaggregatedAttestation deletes the unaggregated attestations in cache.
func (c *AttCaches) DeleteUnaggregatedAttestation(att *zondpb.Attestation) error {
	if att == nil {
		return nil
	}
	if helpers.IsAggregated(att) {
		return errors.New("attestation is aggregated")
	}

	if err := c.insertSeenBit(att); err != nil {
		return err
	}

	r, err := hashFn(att)
	if err != nil {
		return errors.Wrap(err, "could not tree hash attestation")
	}

	c.unAggregateAttLock.Lock()
	defer c.unAggregateAttLock.Unlock()
	delete(c.unAggregatedAtt, r)

	return nil
}

// DeleteSeenUnaggregatedAttestations deletes the unaggregated attestations in cache
// that have been already processed once. Returns number of attestations deleted.
func (c *AttCaches) DeleteSeenUnaggregatedAttestations() (int, error) {
	c.unAggregateAttLock.Lock()
	defer c.unAggregateAttLock.Unlock()

	count := 0
	for _, att := range c.unAggregatedAtt {
		if att == nil || helpers.IsAggregated(att) {
			continue
		}
		if seen, err := c.hasSeenBit(att); err == nil && seen {
			r, err := hashFn(att)
			if err != nil {
				return count, errors.Wrap(err, "could not tree hash attestation")
			}
			delete(c.unAggregatedAtt, r)
			count++
		}
	}
	return count, nil
}

// UnaggregatedAttestationCount returns the number of unaggregated attestations key in the pool.
func (c *AttCaches) UnaggregatedAttestationCount() int {
	c.unAggregateAttLock.RLock()
	defer c.unAggregateAttLock.RUnlock()
	return len(c.unAggregatedAtt)
}
