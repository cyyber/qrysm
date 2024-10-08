package kv

import (
	"context"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	field_params "github.com/theQRL/qrysm/config/fieldparams"
	"github.com/theQRL/qrysm/testing/require"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(io.Discard)

	m.Run()
}

// setupDB instantiates and returns a DB instance for the validator client.
func setupDB(t testing.TB, pubkeys [][field_params.DilithiumPubkeyLength]byte) *Store {
	db, err := NewKVStore(context.Background(), t.TempDir(), &Config{
		PubKeys: pubkeys,
	})
	require.NoError(t, err, "Failed to instantiate DB")
	err = db.UpdatePublicKeysBuckets(pubkeys)
	require.NoError(t, err, "Failed to create old buckets for public keys")
	t.Cleanup(func() {
		require.NoError(t, db.Close(), "Failed to close database")
		require.NoError(t, db.ClearDB(), "Failed to clear database")
	})
	return db
}
