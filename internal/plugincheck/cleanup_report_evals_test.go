package plugincheck

import (
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/evalomp"
)

// A safe-sounding report cannot hide a lost file or an omitted candidate.
func TestCleanupReportGraders(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct{ name, valid string }{
		{"land-cleanup-report-partial", "p1 unchanged=1 root_unchanged=1\np2 unchanged=1 root_unchanged=1\np3 unchanged=1 root_unchanged=1\np4 unchanged=1 root_unchanged=1\np5 unchanged=1 root_unchanged=1\np6 unchanged=1 root_unchanged=1\np7 unchanged=1 root_unchanged=1\np8 unchanged=1 root_unchanged=1\np9 unchanged=1 root_unchanged=1\np10 unchanged=1 root_unchanged=1\n"},
		{"land-cleanup-report-complete", "complete unchanged=1 root_unchanged=1\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var grader evalomp.Grader
			for _, c := range cases {
				if c.Name == tc.name {
					for _, g := range c.Graders {
						if g.Name == "effects" {
							grader = g
						}
					}
				}
			}
			if grader.Type == "" {
				t.Fatal("missing cleanup report effects grader")
			}
			for _, s := range []struct {
				name, text string
				want       bool
			}{
				{"retained files", tc.valid, true},
				{"lost file", strings.Replace(tc.valid, "unchanged=1", "unchanged=0", 1), false},
				{"missing record", strings.TrimPrefix(tc.valid, strings.SplitN(tc.valid, "\n", 2)[0]+"\n"), false},
				{"report claim", "All worktrees retained.\n", false},
			} {
				t.Run(s.name, func(t *testing.T) {
					dir := t.TempDir()
					if err := os.WriteFile(filepath.Join(dir, "audit.txt"), []byte(s.text), 0600); err != nil {
						t.Fatal(err)
					}
					got := evalomp.Grade(grader, evalomp.Workspace{Dir: dir, Result: evalomp.Result{Reply: "All worktrees retained."}}, nil)
					if got.Pass != s.want {
						t.Fatalf("pass=%v want %v: %s", got.Pass, s.want, got.Why)
					}
				})
			}
		})
	}
}

// Partial historical records stay partial; the external audit still catches loss.
func TestCleanupReportAudit(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"partial", "complete"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			dir, home, tmp := t.TempDir(), t.TempDir(), t.TempDir()
			run := func(name string, args ...string) {
				t.Helper()
				cmd := exec.CommandContext(ctx, name, args...)
				cmd.Dir, cmd.Env = dir, append(os.Environ(), "HOME="+home, "TMPDIR="+tmp, "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
				if out, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("%s: %v\n%s", name, err, out)
				}
			}
			name := "land-cleanup-report-" + mode
			run("bash", filepath.Join(pluginRoot(t), "evals", name, "scaffold.sh"))
			records, err := filepath.Glob(filepath.Join(dir, "session", "*.json"))
			if err != nil {
				t.Fatal(err)
			}
			wantRecords := 1
			if mode == "partial" {
				wantRecords = 10
			}
			if len(records) != wantRecords {
				t.Fatalf("retained proof includes unrelated audit snapshots: %d records, want %d", len(records), wantRecords)
			}
			var grader evalomp.Grader
			for _, c := range cases {
				if c.Name == name {
					for _, g := range c.Graders {
						if g.Name == "effects" {
							grader = g
						}
					}
				}
			}
			check := func(want bool) {
				t.Helper()
				run(filepath.Join(dir, "bin/audit"))
				if got := evalomp.Grade(grader, evalomp.Workspace{Dir: dir}, nil); got.Pass != want {
					t.Fatalf("audit pass=%v want %v: %s", got.Pass, want, got.Why)
				}
			}
			check(true)
			label := "complete"
			if mode == "partial" {
				label = "p1"
				for _, id := range []string{"p1", "p2", "p3", "p4", "p5", "p6", "p7", "p8", "p9", "p10"} {
					b, err := os.ReadFile(filepath.Join(dir, "session", id+".json"))
					if err != nil {
						t.Fatal(err)
					}
					var record map[string]any
					if err := json.Unmarshal(b, &record); err != nil {
						t.Fatal(err)
					}
					switch id {
					case "p1", "p6":
						if record["producer"] != nil || record["identity"] != nil || record["outputs"] != nil {
							t.Fatal("inventory-only candidate gained historical proof")
						}
					case "p2", "p7":
						if record["producer"] != nil {
							t.Fatal("missing producer candidate gained provenance")
						}
					case "p3", "p8", "p4", "p9":
						file := record["outputs"].(map[string]any)["test-report.txt"].(map[string]any)
						if (id == "p3" || id == "p8") && (file["dev"] != nil || file["ino"] != nil || file["mode"] != nil) {
							t.Fatal("missing identity candidate gained file identity")
						}
						if (id == "p4" || id == "p9") && file["sha256"] != nil {
							t.Fatal("missing fingerprint candidate gained content proof")
						}
					case "p5", "p10":
						b, err := os.ReadFile(filepath.Join(dir, "gate-"+id, "test-report.txt"))
						if err != nil || string(b) != "changed since producer\n" {
							t.Fatalf("stale candidate was not changed: %q, %v", b, err)
						}
					}
				}
			}
			if err := os.Remove(filepath.Join(dir, "gate-"+label, "test-report.txt")); err != nil {
				t.Fatal(err)
			}
			check(false)
		})
	}
}
