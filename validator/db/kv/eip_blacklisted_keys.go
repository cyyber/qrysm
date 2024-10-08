package kv

import (
	"context"

	field_params "github.com/theQRL/qrysm/config/fieldparams"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

// EIPImportBlacklistedPublicKeys returns keys that were marked as blacklisted during EIP-3076 slashing
// protection imports, ensuring that we can prevent these keys from having duties at runtime.
func (s *Store) EIPImportBlacklistedPublicKeys(ctx context.Context) ([][field_params.DilithiumPubkeyLength]byte, error) {
	_, span := trace.StartSpan(ctx, "Validator.EIPImportBlacklistedPublicKeys")
	defer span.End()
	var err error
	publicKeys := make([][field_params.DilithiumPubkeyLength]byte, 0)
	err = s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(slashablePublicKeysBucket)
		return bucket.ForEach(func(key []byte, _ []byte) error {
			if key != nil {
				var pubKeyBytes [field_params.DilithiumPubkeyLength]byte
				copy(pubKeyBytes[:], key)
				publicKeys = append(publicKeys, pubKeyBytes)
			}
			return nil
		})
	})
	return publicKeys, err
}

// SaveEIPImportBlacklistedPublicKeys stores a list of blacklisted public keys that
// were determined during EIP-3076 slashing protection imports.
func (s *Store) SaveEIPImportBlacklistedPublicKeys(ctx context.Context, publicKeys [][field_params.DilithiumPubkeyLength]byte) error {
	_, span := trace.StartSpan(ctx, "Validator.SaveEIPImportBlacklistedPublicKeys")
	defer span.End()
	return s.db.Update(func(tx *bolt.Tx) error {
		bkt := tx.Bucket(slashablePublicKeysBucket)
		for _, pubKey := range publicKeys {
			// We write the public key to disk in the bucket. The value written for the key does not
			// matter as we'll only be looking at the keys in the bucket when fetching from disk.
			if err := bkt.Put(pubKey[:], []byte{1}); err != nil {
				return err
			}
		}
		return nil
	})
}
