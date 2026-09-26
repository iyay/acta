package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestModulePath pins the module path so the rename to acta sticks:
// it walks up from this test file until go.mod is found, then checks
// the module line is exactly the new name.
func TestModulePath(t *testing.T) {
	dir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("go.mod not found walking up from test dir")
		}
		dir = parent
	}
	raw, err := os.ReadFile(filepath.Join(dir, "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	for line := range strings.Lines(string(raw)) {
		if mod, ok := strings.CutPrefix(strings.TrimSpace(line), "module "); ok {
			if mod != "github.com/iyay/acta" {
				t.Fatalf("module = %q, want %q", mod, "github.com/iyay/acta")
			}
			return
		}
	}
	t.Fatal("no module line in go.mod")
}
