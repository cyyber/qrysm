package kv

import (
	"context"
	"fmt"
	"os"
	"path"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/consensus-types/blocks"
	"github.com/theQRL/qrysm/io/file"
	bolt "go.etcd.io/bbolt"
	"go.opencensus.io/trace"
)

const backupsDirectoryName = "backups"

// Backup the database to the datadir backup directory.
// Example for backup at slot 345: $DATADIR/backups/qrysm_beacondb_at_slot_0000345.backup
//
// The backup is written to a temporary file, synced to disk and only then
// renamed to its final name, so a file with the final name is always complete.
func (s *Store) Backup(ctx context.Context, outputDir string, permissionOverride bool) error {
	ctx, span := trace.StartSpan(ctx, "BeaconDB.Backup")
	defer span.End()

	var backupsDir string
	var err error
	if outputDir != "" {
		backupsDir, err = file.ExpandPath(outputDir)
		if err != nil {
			return err
		}
	} else {
		backupsDir = path.Join(s.databasePath, backupsDirectoryName)
	}
	head, err := s.HeadBlock(ctx)
	if err != nil {
		return err
	}
	if err := blocks.BeaconBlockIsNil(head); err != nil {
		return err
	}
	// Ensure the backups directory exists.
	if err := file.HandleBackupDir(backupsDir, permissionOverride); err != nil {
		return err
	}
	// State summaries are buffered in memory and only written to the database
	// periodically; flush them so the backup does not lack summaries for blocks
	// it contains.
	if err := s.saveCachedStateSummariesDB(ctx); err != nil {
		return errors.Wrap(err, "could not flush state summaries before backup")
	}
	backupPath := path.Join(backupsDir, fmt.Sprintf("qrysm_beacondb_at_slot_%07d.backup", head.Block().Slot()))
	log.WithField("backup", backupPath).Info("Writing backup database.")

	// Each backup writes its own temporary file. Two backups for the same head
	// slot can run at the same time (the webhook can be triggered twice), and a
	// shared temporary name would let one of them rename the other's unfinished
	// copy into place.
	tmpFile, err := os.CreateTemp(backupsDir, path.Base(backupPath)+".*.tmp")
	if err != nil {
		return errors.Wrap(err, "could not create temporary backup file")
	}
	tmpPath := tmpFile.Name()
	if err := tmpFile.Close(); err != nil {
		return errors.Wrap(err, "could not close temporary backup file")
	}
	copyDB, err := bolt.Open(
		tmpPath,
		params.BeaconIoConfig().ReadWritePermissions,
		&bolt.Options{NoSync: true, Timeout: params.BeaconIoConfig().BoltTimeout, FreelistType: bolt.FreelistMapType},
	)
	if err != nil {
		return err
	}
	copyDB.AllocSize = boltAllocSize

	if err := s.copyBucketsTo(ctx, copyDB); err != nil {
		if closeErr := copyDB.Close(); closeErr != nil {
			log.WithError(closeErr).Error("Failed to close backup database")
		}
		if rmErr := os.Remove(tmpPath); rmErr != nil {
			log.WithError(rmErr).Error("Failed to remove incomplete backup database")
		}
		return err
	}
	// Writes were buffered without fsync; make the copy durable before it is
	// given its final name.
	if err := copyDB.Sync(); err != nil {
		if closeErr := copyDB.Close(); closeErr != nil {
			log.WithError(closeErr).Error("Failed to close backup database")
		}
		return errors.Wrap(err, "could not sync backup database")
	}
	if err := copyDB.Close(); err != nil {
		return errors.Wrap(err, "could not close backup database")
	}
	if err := os.Rename(tmpPath, backupPath); err != nil {
		return errors.Wrap(err, "could not move backup database into place")
	}
	return nil
}

// copyBucketsTo copies every bucket of the store into copyDB, one key per
// transaction pair, so that no long-running read transaction is held open.
func (s *Store) copyBucketsTo(ctx context.Context, copyDB *bolt.DB) error {
	// Prefetch all keys of buckets, and inner keys in a
	// bucket to use less memory usage when backing up.
	var bucketKeys [][]byte
	bucketMap := make(map[string][][]byte)
	err := s.db.View(func(tx *bolt.Tx) error {
		return tx.ForEach(func(name []byte, b *bolt.Bucket) error {
			newName := make([]byte, len(name))
			copy(newName, name)
			bucketKeys = append(bucketKeys, newName)
			var innerKeys [][]byte
			err := b.ForEach(func(k, v []byte) error {
				if k == nil {
					return nil
				}
				nKey := make([]byte, len(k))
				copy(nKey, k)
				innerKeys = append(innerKeys, nKey)
				return nil
			})
			if err != nil {
				return err
			}
			bucketMap[string(newName)] = innerKeys
			return nil
		})
	})
	if err != nil {
		return err
	}
	// Utilize much smaller writes, compared to
	// writing for a whole bucket in a single transaction. Also
	// prevent long-running read transactions, as Bolt doesn't
	// handle those well.
	for _, k := range bucketKeys {
		log.Debugf("Copying bucket %s\n", k)
		innerKeys := bucketMap[string(k)]
		for _, ik := range innerKeys {
			if err := ctx.Err(); err != nil {
				return err
			}
			err = s.db.View(func(tx *bolt.Tx) error {
				bkt := tx.Bucket(k)
				if bkt == nil {
					return nil
				}
				// The key may have been deleted since the prefetch (hot state
				// cleanup, invalid block pruning). Copying it with a nil value
				// would store an empty entry that cannot be decoded on restore.
				v := bkt.Get(ik)
				if v == nil {
					return nil
				}
				return copyDB.Update(func(tx2 *bolt.Tx) error {
					b2, err := tx2.CreateBucketIfNotExists(k)
					if err != nil {
						return err
					}
					return b2.Put(ik, v)
				})
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
