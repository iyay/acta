package cli

import (
	"os/exec"
	"strings"
	"testing"
)

// The small build is only small while nothing pulls the TUI back in. One
// stray import would quietly undo it, so this test fails on it.
func TestNotuiBuildLeavesTUIOut(t *testing.T) {
	out, err := exec.Command("go", "list", "-deps", "-tags", "notui", "github.com/iyay/acta/cmd/acta").CombinedOutput()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, out)
	}
	for _, dep := range strings.Fields(string(out)) {
		if dep == "github.com/iyay/acta/internal/tui" || strings.HasPrefix(dep, "github.com/charmbracelet/glamour") {
			t.Errorf("the notui build still links %s", dep)
		}
	}
}
