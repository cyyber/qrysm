package kv

import (
	"sort"
	"testing"

	"github.com/theQRL/go-bitfield"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"github.com/theQRL/qrysm/testing/util"
)

func TestKV_BlockAttestation_CanSaveRetrieve(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&qrysmpb.Attestation{Data: &qrysmpb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b1101}})
	att2 := util.HydrateAttestation(&qrysmpb.Attestation{Data: &qrysmpb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b1101}})
	att3 := util.HydrateAttestation(&qrysmpb.Attestation{Data: &qrysmpb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b1101}})
	atts := []*qrysmpb.Attestation{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.SaveBlockAttestation(att))
	}
	// Diff bit length should not panic.
	att4 := util.HydrateAttestation(&qrysmpb.Attestation{Data: &qrysmpb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b11011}})
	if err := cache.SaveBlockAttestation(att4); err != bitfield.ErrBitlistDifferentLength {
		t.Errorf("Unexpected error: wanted %v, got %v", bitfield.ErrBitlistDifferentLength, err)
	}

	returned := cache.BlockAttestations()

	sort.Slice(returned, func(i, j int) bool {
		return returned[i].Data.Slot < returned[j].Data.Slot
	})

	assert.DeepEqual(t, atts, returned)
}

func TestKV_BlockAttestation_CanDelete(t *testing.T) {
	cache := NewAttCaches()

	att1 := util.HydrateAttestation(&qrysmpb.Attestation{Data: &qrysmpb.AttestationData{Slot: 1}, AggregationBits: bitfield.Bitlist{0b1101}})
	att2 := util.HydrateAttestation(&qrysmpb.Attestation{Data: &qrysmpb.AttestationData{Slot: 2}, AggregationBits: bitfield.Bitlist{0b1101}})
	att3 := util.HydrateAttestation(&qrysmpb.Attestation{Data: &qrysmpb.AttestationData{Slot: 3}, AggregationBits: bitfield.Bitlist{0b1101}})
	atts := []*qrysmpb.Attestation{att1, att2, att3}

	for _, att := range atts {
		require.NoError(t, cache.SaveBlockAttestation(att))
	}

	require.NoError(t, cache.DeleteBlockAttestation(att1))
	require.NoError(t, cache.DeleteBlockAttestation(att3))

	returned := cache.BlockAttestations()
	wanted := []*qrysmpb.Attestation{att2}
	assert.DeepEqual(t, wanted, returned)
}

func TestKV_BlockAttestation_DeletePreservesPendingParticipants(t *testing.T) {
	cache := NewAttCaches()
	processed := recoveryAttestation(0b10001)
	pending := recoveryAttestation(0b10010)
	require.NoError(t, cache.SaveBlockAttestation(processed))
	require.NoError(t, cache.SaveBlockAttestation(pending))
	for range 2 {
		require.NoError(t, cache.DeleteBlockAttestation(processed))
		require.DeepSSZEqual(t, []*qrysmpb.Attestation{pending}, cache.BlockAttestations(), "a later retry must still see unprocessed participants with the same data")
		seen, err := cache.hasSeenAggregatedBit(pending)
		require.NoError(t, err)
		require.Equal(t, false, seen, "deletion must only mark the processed participants seen")
	}
	require.NoError(t, cache.DeleteBlockAttestation(pending))
	require.Equal(t, 0, len(cache.BlockAttestations()))
}
