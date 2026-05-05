//go:build !develop

package params_test

import (
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

// TestMain refuses to run tests in this package unless the `develop` build tag
// is set, while allowing `go test ./...` (which doesn't pass build tags) to
// short-circuit cleanly instead of aborting the whole package with log.Fatal.
func TestMain(m *testing.M) {
	log.Warn("Skipping config/params tests: re-run with `-tags develop` to execute")
	os.Exit(0)
}
