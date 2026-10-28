package usage

import (
	"os"
	"testing"

	"github.com/iyay/acta/internal/testguard"
)

// TestMain stops a test run nobody waits for, like the other packages do.
func TestMain(m *testing.M) {
	testguard.Watch()
	os.Exit(m.Run())
}
