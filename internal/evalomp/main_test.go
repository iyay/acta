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
	// Every test in this package spawns shells or omp, and shells read HOME
	// even when asked only for TMPDIR. Point HOME at a throwaway folder so
	// no path can reach the user's real config.
	home, err := os.MkdirTemp("", "acta-eval-omp-test-home-")
	if err != nil {
		panic(err)
	}
	if err := os.Setenv("HOME", home); err != nil {
		panic(err)
	}
	code := m.Run()
	os.RemoveAll(home)
	os.Exit(code)
}
