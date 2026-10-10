package kv

import (
	"context"
	"testing"

	"github.com/theQRL/qrysm/consensus-types/primitives"
	"github.com/theQRL/qrysm/testing/require"
)

func TestStateMigrationCursor_SaveRetrieve(t *testing.T) {
	ctx := context.Background()
	db := setupDB(t)

	_, found, err := db.StateMigrationCursor(ctx)
	require.NoError(t, err)
	require.Equal(t, false, found)

	require.NoError(t, db.SaveStateMigrationCursor(ctx, 20224))
	slot, found, err := db.StateMigrationCursor(ctx)
	require.NoError(t, err)
	require.Equal(t, true, found)
	require.Equal(t, primitives.Slot(20224), slot)

	require.NoError(t, db.SaveStateMigrationCursor(ctx, 30336))
	slot, _, err = db.StateMigrationCursor(ctx)
	require.NoError(t, err)
	require.Equal(t, primitives.Slot(30336), slot)
}
