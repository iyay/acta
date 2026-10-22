package cli

import (
	"strings"
	"testing"
)

// guardPlan is a plan with a number id, so every write command can name it
// the short way.
const guardPlan = "---\nid: PLN-0082\n---\n# Live work state\n\n### Task 1: acta state set\n\n- [ ] write it\n"

// runIn runs one acta command in dir with body on stdin.
func runIn(t *testing.T, dir, body string, args ...string) (int, string, string) {
	t.Helper()
	cache := t.TempDir()
	t.Setenv("HOME", cache)
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(args, strings.NewReader(body), false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// guardWrites are the three write commands the guard covers, each in a form
// that would change the plan.
var guardWrites = []struct {
	body string
	args []string
}{
	{"", []string{"set", "PLN-0082", "title", "New title"}},
	{"run the gate\n", []string{"state", "set", "PLN-0082", "next"}},
	{"round 1 CLEAN\n", []string{"plan", "add", "PLN-0082", "review-notes"}},
}

// While a linked worktree builds the plan, no write from the main checkout
// changes it, and each refusal names the worktree to use instead.
func TestGuardRefusesMainWhileAWorktreeBuilds(t *testing.T) {
	dir, _ := stateWorktreePlan(t, guardPlan, strings.Replace(guardPlan, "- [ ] write it", "- [x] write it", 1))
	before := read(t, stateFile(dir))
	head := gitOut(t, dir, "rev-parse", "HEAD")
	for _, c := range guardWrites {
		code, _, stderr := runIn(t, dir, c.body, c.args...)
		if code != exitBadInput || !strings.Contains(stderr, "PLN-0082 is building in ") ||
			!strings.Contains(stderr, "live; change it there") {
			t.Errorf("%v: exit %d stderr %q, want the building refusal", c.args, code, stderr)
		}
	}
	if got := read(t, stateFile(dir)); got != before {
		t.Fatalf("file changed: %q", got)
	}
	if gitOut(t, dir, "rev-parse", "HEAD") != head {
		t.Fatal("a refused write made a commit")
	}
}

// The same commands still write inside the worktree that builds the plan.
func TestGuardLetsTheWorktreeWrite(t *testing.T) {
	_, wt := stateWorktreePlan(t, guardPlan, strings.Replace(guardPlan, "- [ ] write it", "- [x] write it", 1))
	for _, c := range guardWrites {
		if code, _, stderr := runIn(t, wt, c.body, c.args...); code != exitOK {
			t.Errorf("%v in the worktree: exit %d stderr %q", c.args, code, stderr)
		}
	}
}

// Once the worktree is gone, as after land, the main checkout writes again.
func TestGuardLetsMainWriteOnceTheWorktreeIsGone(t *testing.T) {
	dir, wt := stateWorktreePlan(t, guardPlan, strings.Replace(guardPlan, "- [ ] write it", "- [x] write it", 1))
	gitOut(t, dir, "worktree", "remove", "--force", wt)
	for _, c := range guardWrites {
		if code, _, stderr := runIn(t, dir, c.body, c.args...); code != exitOK {
			t.Errorf("%v after the worktree is gone: exit %d stderr %q", c.args, code, stderr)
		}
	}
}
