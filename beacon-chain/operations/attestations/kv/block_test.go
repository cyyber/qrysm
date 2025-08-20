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
