package update

import (
	"os"
	"testing"

	"github.com/iyay/acta/internal/testguard"
)

// TestMain joins the test guard. It stops a stuck test run from leaving a
// test binary running after the run is killed.
func TestMain(m *testing.M) {
	testguard.Watch()
	os.Exit(m.Run())
}
