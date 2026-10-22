package cli

import (
	"strings"
	"testing"
)

const cliAddPlan = "---\nid: PLN-0082\nhash: k3f2\n---\n# Live work state\n\n### Task 1: acta state set\n\n- [ ] test\n"

// runPlan runs acta plan with the lines an agent would pipe in.
func runPlan(t *testing.T, dir, body string, args ...string) (int, string, string) {
	t.Helper()
	cache := t.TempDir()
	t.Setenv("HOME", cache)
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"plan"}, args...), strings.NewReader(body), false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// The command prints the id of the new task, so the agent can tick it without
// reading the plan, and the plan id for review notes.
func TestCmdPlanAddPrintsWhatItAdded(t *testing.T) {
	dir := stateRepo(t, cliAddPlan)
	code, stdout, stderr := runPlan(t, dir, "fix it\n", "add", "PLN-0082", "fix-round", "--title", "Fix", "--verify", "it works")
	if code != exitOK || stdout != "plans/2026-10-05-live-work-state#task-2 added\n" {
		t.Fatalf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	code, stdout, stderr = runPlan(t, dir, "round 1 CLEAN\n", "add", "PLN-0082", "review-notes")
	if code != exitOK || stdout != "PLN-0082\n" {
		t.Fatalf("exit %d stdout %q stderr %q", code, stdout, stderr)
	}
	got := read(t, stateFile(dir))
	for _, want := range []string{"## Fix round 1\n", "### Task 2: Fix\n", "**verify:** it works\n", "- [ ] fix it\n", "## Review notes\n\n- round 1 CLEAN\n"} {
		if !strings.Contains(got, want) {
			t.Errorf("plan lacks %q:\n%s", want, got)
		}
	}
}

// A bad request exits with the bad-input code, says why, and changes nothing.
func TestCmdPlanAddRefusesAndChangesNothing(t *testing.T) {
	dir := stateRepo(t, cliAddPlan)
	before := read(t, stateFile(dir))
	head := gitOut(t, dir, "rev-parse", "HEAD")
	for _, c := range []struct {
		name, body string
		args       []string
		want       string
	}{
		{"unknown section", "x\n", []string{"add", "PLN-0082", "notes"}, "fix-round, polish or review-notes"},
		{"empty stdin", "", []string{"add", "PLN-0082", "polish"}, "nothing on stdin"},
		{"no title", "x\n", []string{"add", "PLN-0082", "fix-round", "--verify", "V"}, "--title"},
		{"missing section", "x\n", []string{"add", "PLN-0082"}, "usage: acta plan add"},
		{"no sub command", "x\n", []string{"PLN-0082"}, "usage: acta plan add"},
	} {
		code, _, stderr := runPlan(t, dir, c.body, c.args...)
		if code != exitBadInput || !strings.Contains(stderr, c.want) {
			t.Errorf("%s: exit %d stderr %q, want bad input naming %q", c.name, code, stderr, c.want)
		}
	}
	if got := read(t, stateFile(dir)); got != before {
		t.Fatalf("file changed: %q", got)
	}
	if gitOut(t, dir, "rev-parse", "HEAD") != head {
		t.Fatal("a refused call made a commit")
	}
}
