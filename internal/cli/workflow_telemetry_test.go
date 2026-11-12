package cli

import (
	"strings"
	"testing"

	"github.com/iyay/acta/internal/telemetry"
)

// workflowRun runs one acta command in dir and gives back today's workflow
// counts after it.
func workflowRun(t *testing.T, s *telemetry.Store, dir, body string, args ...string) (int, map[string]int) {
	t.Helper()
	var stdout, stderr strings.Builder
	var code int
	inDir(t, dir, func() {
		code = Run(args, strings.NewReader(body), false, &stdout, &stderr)
	})
	return code, todayCounts(t, s).Workflow
}

// The repo comes first: it sets its own HOME, and the telemetry home must win.
func workflowRepo(t *testing.T, files map[string]string) (string, *telemetry.Store) {
	t.Helper()
	dir := readyRepo(t, files)
	s, _ := runEnv(t, false)
	return dir, s
}

func TestIDCountsEachNewSpecAndPlanNumber(t *testing.T) {
	dir, s := workflowRepo(t, map[string]string{
		"specs/2026-11-01-a.md": "# A\n",
		"specs/2026-11-01-b.md": "# B\n",
		"plans/2026-11-01-p.md": "---\nparent: specs/2026-11-01-a\n---\n# P\n\n### Task 1: Do it\n- [ ] a\n",
		// It has its number already and gets only a hash, so it is no new spec.
		"specs/2026-11-01-old.md": "---\nid: SPC-0001\n---\n# Old\n",
	})
	code, got := workflowRun(t, s, dir, "", "id")
	if code != exitOK || got["spec_created"] != 2 || got["plan_created"] != 1 || len(got) != 2 {
		t.Fatalf("acta id = %d, workflow %v; want spec_created 2 and plan_created 1", code, got)
	}
	// Every item has its number now, so a second run gives none.
	if _, got = workflowRun(t, s, dir, "", "id"); got["spec_created"] != 2 || got["plan_created"] != 1 {
		t.Fatalf("second acta id counted again: %v", got)
	}
}

func TestFixDuplicatesCountsNothing(t *testing.T) {
	dir, s := workflowRepo(t, map[string]string{
		"specs/2026-11-01-a.md": "---\nid: SPC-0001\nhash: aaaaaaa\n---\n# A\n",
		"specs/2026-11-02-b.md": "---\nid: SPC-0001\nhash: bbbbbbb\n---\n# B\n",
	})
	code, got := workflowRun(t, s, dir, "", "id", "--fix-duplicates")
	if code != exitOK || len(got) != 0 {
		t.Fatalf("--fix-duplicates = %d, workflow %v; want nothing counted", code, got)
	}
}

func TestBugNewCountsOnlyAWrittenBug(t *testing.T) {
	dir, s := workflowRepo(t, map[string]string{"specs/2026-11-01-a.md": "---\nid: SPC-0001\n---\n# A\n"})
	if code, got := workflowRun(t, s, dir, "# Broken\n\nno symptom here\n", "bug", "new", "broken"); code == exitOK || len(got) != 0 {
		t.Fatalf("a refused bug = %d, workflow %v; want an error and nothing counted", code, got)
	}
	code, got := workflowRun(t, s, dir, "# Broken\n\n## Symptom\nIt breaks.\n", "bug", "new", "broken")
	if code != exitOK || got["bug_created"] != 1 || len(got) != 1 {
		t.Fatalf("bug new = %d, workflow %v; want bug_created 1", code, got)
	}
}

func TestSpecDatesCountsALandedPlan(t *testing.T) {
	dir, s := workflowRepo(t, map[string]string{
		"specs/2026-10-01-s.md": "---\nid: SPC-0009\n---\n# S\n",
		"plans/2026-10-01-a.md": "---\nid: PLN-0001\nparent: specs/2026-10-01-s\nstarted: \"2026-10-01 10:00:00\"\n---\n# Plan\n\n### Task 1: Do it\n- [x] a\n",
	})
	if code, got := workflowRun(t, s, dir, "", "spec-dates", "PLN-0404"); code == exitOK || len(got) != 0 {
		t.Fatalf("spec-dates of an unknown plan = %d, workflow %v; want an error and nothing counted", code, got)
	}
	code, got := workflowRun(t, s, dir, "", "spec-dates", "PLN-0001")
	if code != exitOK || got["plan_landed"] != 1 || len(got) != 1 {
		t.Fatalf("spec-dates = %d, workflow %v; want plan_landed 1", code, got)
	}
}
