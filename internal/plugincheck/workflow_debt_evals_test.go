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

// Role checks must reject extra launches without counting helpers or resumes.
func TestWorkflowDebtReviewerRoles(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	both := `{"prompt":"ACTA_REVIEW_AXIS=BOTH\nReview Spec and Standards.","subagent_type":"general-purpose"}`
	spec := `{"prompt":"ACTA_REVIEW_AXIS=SPEC\nReview the spec.","subagent_type":"general-purpose"}`
	standards := `{"prompt":"ACTA_REVIEW_AXIS=STANDARDS\nReview standards.","subagent_type":"general-purpose"}`
	helper := `{"prompt":"Find files mentioning Spec and Standards; do not review."}`
	for _, tc := range []struct {
		name         string
		inputs       []string
		safe, polish bool
	}{
		{"empty", nil, false, false},
		{"helper", []string{helper}, false, false},
		{"quoted marker", []string{`{"prompt":"Discuss \"ACTA_REVIEW_AXIS=BOTH\" without reviewing."}`}, false, false},
		{"combined", []string{both}, true, false},
		{"combined with helper", []string{both, helper}, true, false},
		{"combined resumed", []string{both, `{"resume":"reviewer-1","prompt":"ACTA_REVIEW_AXIS=BOTH\nAdd Coverage."}`}, true, false},
		{"combined empty resume", []string{`{"resume":"","prompt":"ACTA_REVIEW_AXIS=BOTH\nReview both axes."}`}, true, false},
		{"combined null resume", []string{`{"prompt":"ACTA_REVIEW_AXIS=BOTH\nReview both axes.","resume":null}`}, true, false},
		{"two combined", []string{both, both}, false, false},
		{"spec only", []string{spec}, false, false},
		{"standards only", []string{standards}, false, false},
		{"separate axes", []string{spec, standards}, false, true},
		{"separate axes with helper", []string{helper, spec, standards}, false, true},
		{"separate axes resumed", []string{spec, standards, `{"prompt":"ACTA_REVIEW_AXIS=SPEC\nAdd Coverage.","resume":"spec-1"}`}, false, true},
		{"same identity other axis", []string{spec, `{"prompt":"ACTA_REVIEW_AXIS=STANDARDS\nReview standards.","resume":"spec-1"}`}, false, false},
		{"extra combined reviewer", []string{spec, standards, both}, false, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var calls []evalomp.Call
			for _, input := range tc.inputs {
				calls = append(calls, evalomp.Call{Tool: "Agent", Input: input})
			}
			// Messages return to an existing reviewer; they never launch one.
			calls = append(calls, evalomp.Call{Tool: "SendMessage", Input: `{"prompt":"ACTA_REVIEW_AXIS=BOTH\nAdd Coverage."}`})
			for _, target := range []struct {
				name    string
				graders []string
				want    bool
			}{
				{"review-auto-safe-one-reviewer", []string{"one-reviewer", "combined-role"}, tc.safe},
				{"polish-full-review", []string{"reviewer-dispatched", "standards-reviewer-dispatched", "reviewer-count"}, tc.polish},
			} {
				got := true
				for _, name := range target.graders {
					var grader evalomp.Grader
					for _, c := range cases {
						if c.Name == target.name {
							for _, g := range c.Graders {
								if g.Name == name {
									grader = g
								}
							}
						}
					}
					if grader.Type != "tool_used" {
						t.Errorf("missing fixed role grader %s/%s (type %q)", target.name, name, grader.Type)
						got = false
						continue
					}
					if !evalomp.Grade(grader, evalomp.Workspace{Result: evalomp.Result{Calls: calls}}, nil).Pass {
						got = false
					}
				}
				if got != target.want {
					t.Errorf("%s role checks pass=%v, want %v", target.name, got, target.want)
				}
			}
		})
	}
}

// Both durations need an explicit conflict; wording and order may vary.
func TestWorkflowDebtColdReadGraders(t *testing.T) {
	cases, err := evalomp.LoadCases(filepath.Join(pluginRoot(t), "evals"))
	if err != nil {
		t.Fatal(err)
	}
	var grader evalomp.Grader
	for _, c := range cases {
		if c.Name == "cold-read-planted-contradiction" {
			for _, g := range c.Graders {
				if g.Name == "names-both-lines" {
					grader = g
				}
			}
		}
	}
	if grader.Type == "" {
		t.Fatal("missing contradiction grader")
	}
	for _, tc := range []struct {
		reply string
		want  bool
	}{
		{"Storage: 30 days. Cleanup: 7 days.", false},
		{"No conflict: 30 days and 7 days.", false},
		{"No conflicts: 30 days and 7 days.", false},
		{"No contradictions found. Storage: 30 days. Cleanup: 7 days.", false},
		{"Storage keeps archives for 30 days, but Cleanup deletes after 7 days: these conflict.", true},
		{"The 7-day deletion contradicts the 30-day retention.", true},
		{"CONFLICT: 30-day retention and 7 days until deletion.", true},
		{"30 days conflicts with 7-day deletion.", true},
		{"30 days vs 7 days: two retention periods for one archive; the implementer cannot tell which wins.", true},
		{"Storage promises 30 days; the 7-day job breaks the promise.", true},
		{"Archives deleted at 7 days break the 30-day promise.", true},
		{"Storage: 30 days. Cleanup: 7 days. Both can't hold.", true},
		// Real replies from the 2026-11-16 runs that named the conflict but failed the old pattern.
		{"30 days vs 7 days. An implementer can't tell which wins.", true},
		{"30 days vs 7 days. An implementer can't tell which period is right. Either choice breaks the other line.", true},
		{"Retention is 30 days in one section and 7 days in the other. The job as written breaks the re-download promise.", true},
		{"Storage: 30 days. Cleanup: 7 days. I cannot tell which section is newer.", false},
		{"30 days conflicts with an unspecified cleanup policy.", false},
	} {
		got := evalomp.Grade(grader, evalomp.Workspace{Result: evalomp.Result{Reply: tc.reply}}, nil)
		if got.Pass != tc.want {
			t.Errorf("%q pass=%v want%v: %s", tc.reply, got.Pass, tc.want, got.Why)
		}
	}
	for _, name := range []string{"cold-read-planted-contradiction", "cold-read-false-claim"} {
		prompt := readFile(t, "evals", name, "prompt.md")
		if strings.Contains(prompt, "Give me every finding") {
			t.Errorf("%s still supplies the relay requirement", name)
		}
	}
	for _, file := range []string{
		"polish-full-review/graders/reviewer-dispatched.md", "polish-full-review/graders/standards-reviewer-dispatched.md",
		"review-auto-signal/graders/spec-reviewer-dispatched.md", "review-auto-signal/graders/standards-reviewer-dispatched.md",
		"review-depth-floor/graders/spec-reviewer-dispatched.md", "review-depth-floor/graders/standards-reviewer-dispatched.md",
		"review-clean-has-evidence/graders/brief-asks-coverage.md",
	} {
		if strings.Contains(readFile(t, "evals", file), "case-sensitive substring") {
			t.Errorf("%s describes the old matcher", file)
		}
	}
}
