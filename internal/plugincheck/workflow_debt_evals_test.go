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

// Verdicts belong to the orchestrator; quoted reports and claimed runs are not proof.
func TestWorkflowDebtReviewGraders(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		caseName, graderName, reply, marker string
		want                                bool
	}{
		{"review-clean-has-evidence", "brief-asks-coverage", "Round verdict: BLOCKED", "spec Please add Coverage.\nstandards Please add Coverage.\n", true},
		{"review-clean-has-evidence", "brief-asks-coverage", "Round verdict: BLOCKED", "standards Coverage needed.\nspec Coverage needed.\n", true},
		{"review-clean-has-evidence", "brief-asks-coverage", "I asked both for Coverage.\nRound verdict: BLOCKED", "", false},
		{"review-clean-has-evidence", "brief-asks-coverage", "Round verdict: BLOCKED", "spec Coverage needed.\n", false},
		{"review-clean-has-evidence", "brief-asks-coverage", "Round verdict: BLOCKED", "spec Coverage needed.\nstandards Coverage needed.\nstandards Coverage again.\n", false},
		{"review-clean-has-evidence", "verdict", "Round verdict: **BLOCKED**", "", true},
		{"review-clean-has-evidence", "verdict", "Round verdict: blocked", "", true},
		{"review-clean-has-evidence", "verdict", "> Reviewer: CLEAN", "", false},
		{"review-clean-has-evidence", "verdict", "The reviewer said CLEAN, but I have no verdict.", "", false},
		{"review-false-blocker", "clean-verdict", "The BLOCKER is false, so I dropped it after the repro.\nRound verdict: CLEAN", "", true},
		{"review-false-blocker", "clean-verdict", "Rejected after running the repro.\nRound verdict: **CLEAN**", "", true},
		{"review-false-blocker", "clean-verdict", "REJECTED after the repro.\nRound verdict: clean", "", true},
		{"review-false-blocker", "clean-verdict", "Rejected.\n> Round verdict: CLEAN", "", false},
		{"review-false-blocker", "clean-verdict", "Rejected.\nRound verdict: CLEAN\nNo final verdict yet.", "", false},
		{"review-false-blocker", "repro-ran", "Repro executed.", "discount_result=40.0\n", true},
		{"review-false-blocker", "repro-ran", "I ran python3 tools/repro.py, result 40.", "", false},
		{"review-false-blocker", "repro-ran", "Read tools/repro.py.", "", false},
		{"review-false-blocker", "repro-ran", "Repro failed.", "discount_result=30.0\n", false},
	} {
		t.Run(tc.caseName+"/"+tc.graderName+"/"+tc.reply, func(t *testing.T) {
			var grader evalomp.Grader
			for _, c := range cases {
				if c.Name == tc.caseName {
					for _, g := range c.Graders {
						if g.Name == tc.graderName {
							grader = g
						}
					}
				}
			}
			if grader.Type == "" {
				t.Fatal("missing review evidence grader")
			}
			dir := t.TempDir()
			if err := os.Mkdir(filepath.Join(dir, "reports"), 0700); err != nil {
				t.Fatal(err)
			}
			if tc.marker != "" {
				markerFile := "repro.log"
				if tc.caseName == "review-clean-has-evidence" {
					markerFile = "requests.log"
				}
				if err := os.WriteFile(filepath.Join(dir, "reports", markerFile), []byte(tc.marker), 0600); err != nil {
					t.Fatal(err)
				}
			}
			got := evalomp.Grade(grader, evalomp.Workspace{Dir: dir, Result: evalomp.Result{Reply: tc.reply}}, nil)
			if got.Pass != tc.want {
				t.Fatalf("pass=%v want %v: %s", got.Pass, tc.want, got.Why)
			}
		})
	}
}
