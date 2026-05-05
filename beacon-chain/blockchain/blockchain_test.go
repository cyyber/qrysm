package blockchain

import (
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

func TestMain(m *testing.M) {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.SetOutput(io.Discard)

	if stopTests {
		logrus.Warn("Skipping beacon-chain/blockchain tests: re-run with `-tags develop` to execute")
		os.Exit(0)
	}
	m.Run()
}
