package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/evalomp"
)

// Missing forcing-path calls and extra unrelated runs must never earn a pass.
func TestWorkflowDebtLand(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	var grader evalomp.Grader
	for _, c := range cases {
		if c.Name == "land-parent-gate-triggers" {
			grader = c.Graders[0]
		}
	}
	if grader.Type == "" {
		t.Fatal("missing land parent gate grader")
	}
	valid := "start\nfull-A full\nfull-A slow --area=unit\nfull-B full\nfull-B slow --all\nfull-C full\nfull-D full\nfull-D slow --all\nfull-E full\nfull-E slow --all\nfull-F full\nfull-F slow --all\ntidy-A full\ntidy-A slow --area=unit\ntidy-B full\ntidy-B slow --all\ntidy-C full\ntidy-D full\ntidy-D slow --all\ntidy-E full\ntidy-E slow --all\ntidy-F full\ntidy-F slow --all\n"
	for _, c := range []struct {
		name, log string
		want      bool
	}{
		{"all selected gates", valid, true},
		{"forcing path skipped", strings.ReplaceAll(valid, "full-B slow --all\n", ""), false},
		{"no run paths skipped", strings.ReplaceAll(valid, "tidy-D slow --all\n", ""), false},
		{"missing metadata skipped", strings.ReplaceAll(valid, "full-E slow --all\n", ""), false},
		{"unclear paths skipped", strings.ReplaceAll(valid, "tidy-F slow --all\n", ""), false},
		{"wrong narrow form", strings.ReplaceAll(valid, "full-B slow --all", "full-B slow --area=unit"), false},
		{"unaffected gate repeated", valid + "full-C slow --all\n", false},
		{"unneeded whole run", valid + "tidy-A slow --all\n", false},
		{"full suite skipped", strings.ReplaceAll(valid, "tidy-C full\n", ""), false},
		{"commands quoted only", "read scripts/slow --all\n", false},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "calls.log"), []byte(c.log), 0600); err != nil {
				t.Fatal(err)
			}
			got := evalomp.Grade(grader, evalomp.Workspace{Dir: dir}, nil)
			if got.Pass != c.want {
				t.Fatalf("pass=%v, want %v: %s", got.Pass, c.want, got.Why)
			}
		})
	}
}
