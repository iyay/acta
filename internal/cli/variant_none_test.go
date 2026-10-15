//go:build notui

package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestVariantCLI(t *testing.T) {
	if Variant != "cli" {
		t.Errorf("Variant = %q, want cli", Variant)
	}
}

// Bare acta in the small build has no board to open. It shows the help and
// says how to get the board, and it does not fail.
func TestNotuiBareActaPrintsHelp(t *testing.T) {
	t.Chdir(commitRepo(t))
	var out, errs bytes.Buffer
	if code := Run(nil, strings.NewReader(""), false, &out, &errs); code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, errs.String())
	}
	if !strings.Contains(out.String(), renderHelp()) {
		t.Errorf("bare acta did not print the help: %q", out.String())
	}
	if !strings.Contains(out.String(), "ACTA_TUI=1") {
		t.Errorf("bare acta did not say how to get the TUI: %q", out.String())
	}
}
