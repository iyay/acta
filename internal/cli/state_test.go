package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// stateRepo makes a git repo holding one plan with no State section yet.
func stateRepo(t *testing.T, plan string) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"specs", "plans"} {
		if err := os.MkdirAll(filepath.Join(dir, ".acta", d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, ".acta", "plans", "2026-10-05-live-work-state.md"), []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "init", "-q", "-b", "main")
	gitOut(t, dir, "config", "user.name", "test")
	gitOut(t, dir, "config", "user.email", "test@example.com")
	gitOut(t, dir, "add", ".")
	gitOut(t, dir, "commit", "-q", "-m", "init")
	return dir
}

// runState runs acta state with the body an agent would pipe in.
func runState(t *testing.T, dir, body string, args ...string) (int, string, string) {
	t.Helper()
	// A temp cache folder, so a cli test never writes a lock into the real
	// cache folder of whoever runs the tests.
	cache := t.TempDir()
	t.Setenv("HOME", cache)
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"state"}, args...), strings.NewReader(body), false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func stateFile(dir string) string {
	return filepath.Join(dir, ".acta", "plans", "2026-10-05-live-work-state.md")
}

const cliStatePlan = "---\nstatus: in-progress\n---\n# Live work state\n\n### Task 1: acta state set\n- [ ] test\n"

// The body comes from stdin and the plan file gets the subsection under it.
func TestCmdStateSetReadsStdin(t *testing.T) {
	dir := stateRepo(t, cliStatePlan)
	code, _, stderr := runState(t, dir, "run the gate\n", "set", "plans/2026-10-05-live-work-state", "next")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	want := cliStatePlan + "\n## State\n\n### Next\n\nrun the gate\n"
	if got := read(t, stateFile(dir)); got != want {
		t.Fatalf("file = %q\nwant %q", got, want)
	}
	if got := gitOut(t, dir, "log", "-1", "--format=%s"); got != "acta: state plans/2026-10-05-live-work-state" {
		t.Fatalf("commit subject %q", got)
	}
}

// Each part writes under its own heading, and the plan id accepts the short
// form the board gives it.
func TestCmdStateSetTakesTheShortIdAndEachPart(t *testing.T) {
	dir := stateRepo(t, "---\nid: PLN-0082\nhash: k1a2b3\n---\n# Live work state\n\n### Task 1: acta state set\n- [ ] test\n")
	for _, c := range []struct{ part, heading string }{
		{"next", "### Next"},
		{"findings", "### Findings"},
		{"rulings", "### Open rulings"},
	} {
		code, _, stderr := runState(t, dir, "one line\n", "set", "PLN-0082", c.part)
		if code != exitOK {
			t.Fatalf("%s: exit %d stderr %q", c.part, code, stderr)
		}
		if got := read(t, stateFile(dir)); !strings.Contains(got, c.heading+"\n\none line\n") {
			t.Fatalf("%s: file = %q, want the body under %q", c.part, got, c.heading)
		}
	}
}

// A body over the cap, a part name outside the three, and an id the board does
// not know each exit bad input with the reason named, and the file on disk is
// left as it was.
func TestCmdStateSetRefusesAndChangesNothing(t *testing.T) {
	dir := stateRepo(t, cliStatePlan)
	before := read(t, stateFile(dir))
	head := gitOut(t, dir, "rev-parse", "HEAD")
	for _, c := range []struct{ name, body, want string }{
		{"over cap", strings.Repeat("line\n", 11), "at most 10 lines"},
		{"unknown part", "x\n", "next, findings or rulings"},
		{"unknown plan", "x\n", "unknown id plans/2026-10-06-nope"},
	} {
		id, part := "plans/2026-10-05-live-work-state", "next"
		if c.name == "unknown part" {
			part = "ruling"
		}
		if c.name == "unknown plan" {
			id = "plans/2026-10-06-nope"
		}
		code, _, stderr := runState(t, dir, c.body, "set", id, part)
		if code != exitBadInput {
			t.Fatalf("%s: exit %d, stderr %q, want bad input", c.name, code, stderr)
		}
		if !strings.Contains(stderr, c.want) {
			t.Errorf("%s: stderr %q does not name %q", c.name, stderr, c.want)
		}
	}
	if got := read(t, stateFile(dir)); got != before {
		t.Fatalf("file changed: %q", got)
	}
	if gitOut(t, dir, "rev-parse", "HEAD") != head {
		t.Fatal("a refused call made a commit")
	}
}

// The bare view is not built yet, so it says so and exits bad input instead of
// printing nothing a reader could mistake for an answer.
func TestCmdStateBareSaysItIsNotBuiltYet(t *testing.T) {
	dir := stateRepo(t, cliStatePlan)
	for _, args := range [][]string{nil, {"plans/2026-10-05-live-work-state"}, {"set"}} {
		code, out, stderr := runState(t, dir, "", args...)
		if code != exitBadInput {
			t.Fatalf("%v: exit %d, want bad input", args, code)
		}
		if out != "" {
			t.Errorf("%v: stdout %q, want nothing", args, out)
		}
		if !strings.Contains(stderr, "acta state set") {
			t.Errorf("%v: stderr %q does not point at acta state set", args, stderr)
		}
	}
}

// Too many or too few words after "set" is a usage error, not a crash.
func TestCmdStateSetUsage(t *testing.T) {
	dir := stateRepo(t, cliStatePlan)
	for _, args := range [][]string{{"set", "plans/2026-10-05-live-work-state"}, {"set", "a", "b", "c"}} {
		code, _, stderr := runState(t, dir, "x\n", args...)
		if code != exitBadInput || !strings.Contains(stderr, "usage: acta state set") {
			t.Fatalf("%v: exit %d stderr %q", args, code, stderr)
		}
	}
}
