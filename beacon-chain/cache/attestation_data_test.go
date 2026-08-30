package cache_test

import (
	"context"
	"testing"

	"github.com/theQRL/qrysm/beacon-chain/cache"
	qrysmpb "github.com/theQRL/qrysm/proto/qrysm/v1alpha1"
	"github.com/theQRL/qrysm/testing/assert"
	"github.com/theQRL/qrysm/testing/require"
	"google.golang.org/protobuf/proto"
)

func TestAttestationCache_RoundTrip(t *testing.T) {
	ctx := context.Background()
	c := cache.NewAttestationCache()

	req := &qrysmpb.AttestationDataRequest{
		CommitteeIndex: 0,
		Slot:           1,
	}

	response, err := c.Get(ctx, req)
	assert.NoError(t, err)
	assert.Equal(t, (*qrysmpb.AttestationData)(nil), response)

	assert.NoError(t, c.MarkInProgress(req))

	res := &qrysmpb.AttestationData{
		Target: &qrysmpb.Checkpoint{Epoch: 5, Root: make([]byte, 32)},
	}

	assert.NoError(t, c.Put(ctx, req, res))
	assert.NoError(t, c.MarkNotInProgress(req))

	response, err = c.Get(ctx, req)
	assert.NoError(t, err)

	if !proto.Equal(response, res) {
		t.Error("Expected equal protos to return from cache")
	}
}

func TestAttestationCache_Clear(t *testing.T) {
	ctx := context.Background()
	c := cache.NewAttestationCache()
	req := &qrysmpb.AttestationDataRequest{Slot: 5, CommitteeIndex: 0}
	require.NoError(t, c.Put(ctx, req, &qrysmpb.AttestationData{Slot: 5, BeaconBlockRoot: make([]byte, 32)}))
	res, err := c.Get(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)

	c.Clear()

	res, err = c.Get(ctx, req)
	require.NoError(t, err)
	require.Equal(t, (*qrysmpb.AttestationData)(nil), res)

	// The cache keeps working after being cleared.
	require.NoError(t, c.Put(ctx, req, &qrysmpb.AttestationData{Slot: 5, BeaconBlockRoot: make([]byte, 32)}))
	res, err = c.Get(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
}
