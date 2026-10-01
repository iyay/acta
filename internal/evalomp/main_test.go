package evalomp

import (
	"os"
	"testing"

	"github.com/iyay/acta/internal/testguard"
)

// TestMain starts the guard before anything else runs, so a test run nobody
// waits for any more stops itself instead of burning the CPU.
func TestMain(m *testing.M) {
	testguard.Watch()
	os.Exit(m.Run())
}
