package db

import (
	"io"
	"os"
	"path"
	"strings"
	"time"

	"github.com/pkg/errors"
	"github.com/theQRL/qrysm/beacon-chain/db/kv"
	"github.com/theQRL/qrysm/cmd"
	"github.com/theQRL/qrysm/config/params"
	"github.com/theQRL/qrysm/io/file"
	"github.com/theQRL/qrysm/io/prompt"
	"github.com/urfave/cli/v2"
	bolt "go.etcd.io/bbolt"
)

const dbExistsYesNoPrompt = "A database file already exists in the target directory. " +
	"Are you sure that you want to overwrite it? [y/n]"

// Restore a beacon chain database.
//
// The source must be a readable database file that is not the target itself.
// The copy is written to a temporary file, synced and then renamed over the
// target, so an interrupted restore never leaves a truncated database under
// the database file name. A target that is open by a running beacon node is
// refused.
func Restore(cliCtx *cli.Context) error {
	sourceFile := cliCtx.String(cmd.RestoreSourceFileFlag.Name)
	targetDir := cliCtx.String(cmd.RestoreTargetDirFlag.Name)

	restoreDir := path.Join(targetDir, kv.BeaconNodeDbDirName)
	targetFile := path.Join(restoreDir, kv.DatabaseFileName)
	if file.FileExists(targetFile) {
		same, err := sameFile(sourceFile, targetFile)
		if err != nil {
			return err
		}
		if same {
			// Copying a file onto itself truncates it before a byte is read.
			return errors.New("the restore source is the database file in the target directory")
		}
		resp, err := prompt.ValidatePrompt(
			os.Stdin, dbExistsYesNoPrompt, prompt.ValidateYesOrNo,
		)
		if err != nil {
			return errors.Wrap(err, "could not validate choice")
		}
		if strings.EqualFold(resp, "n") {
			log.Info("Restore aborted")
			return nil
		}
		if err := ensureDatabaseNotInUse(targetFile); err != nil {
			return err
		}
	}
	if err := verifyDatabaseFile(sourceFile); err != nil {
		return errors.Wrap(err, "the restore source is not a readable database file")
	}
	if err := file.MkdirAll(restoreDir); err != nil {
		return err
	}
	// A fresh temporary file: a fixed name could be the source itself (a
	// leftover of an interrupted restore), which the copy would truncate
	// before reading it.
	tmp, err := os.CreateTemp(restoreDir, kv.DatabaseFileName+".restore.*.tmp")
	if err != nil {
		return errors.Wrap(err, "could not create a temporary restore file")
	}
	tmpFile := tmp.Name()
	if err := tmp.Close(); err != nil {
		return errors.Wrap(err, "could not close the temporary restore file")
	}
	if err := copyFileSync(sourceFile, tmpFile); err != nil {
		if rmErr := os.Remove(tmpFile); rmErr != nil && !os.IsNotExist(rmErr) {
			log.WithError(rmErr).Error("Could not remove incomplete restore file")
		}
		return err
	}
	if err := os.Rename(tmpFile, targetFile); err != nil {
		return errors.Wrap(err, "could not move the restored database into place")
	}

	log.Info("Restore completed successfully")
	return nil
}

// sameFile reports whether the two paths name the same file.
func sameFile(a, b string) (bool, error) {
	ai, err := os.Stat(a)
	if err != nil {
		return false, errors.Wrap(err, "could not stat the restore source")
	}
	bi, err := os.Stat(b)
	if err != nil {
		return false, errors.Wrap(err, "could not stat the restore target")
	}
	return os.SameFile(ai, bi), nil
}

// ensureDatabaseNotInUse refuses to overwrite a database file that another
// process (a running beacon node) holds locked. A file that cannot be opened for
// another reason is about to be replaced anyway.
func ensureDatabaseNotInUse(dbPath string) error {
	db, err := bolt.Open(dbPath, params.BeaconIoConfig().ReadWritePermissions, &bolt.Options{Timeout: time.Second, ReadOnly: true})
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return errors.New("the target database is in use by another process; stop the beacon node before restoring")
		}
		return nil
	}
	return db.Close()
}

// verifyDatabaseFile checks that the path opens as a bolt database.
func verifyDatabaseFile(dbPath string) error {
	if !file.FileExists(dbPath) {
		return errors.New("source file does not exist at provided path")
	}
	db, err := bolt.Open(dbPath, params.BeaconIoConfig().ReadWritePermissions, &bolt.Options{Timeout: time.Second, ReadOnly: true})
	if err != nil {
		if errors.Is(err, bolt.ErrTimeout) {
			return errors.New("the source database is in use by another process")
		}
		return err
	}
	return db.Close()
}

// copyFileSync copies src to dst and syncs dst to disk before returning.
func copyFileSync(src, dst string) (err error) {
	in, err := os.Open(src) // #nosec G304
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := in.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, params.BeaconIoConfig().ReadWritePermissions) // #nosec G304
	if err != nil {
		return err
	}
	defer func() {
		if closeErr := out.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
	}()
	if _, err = io.Copy(out, in); err != nil {
		return err
	}
	return out.Sync()
}
