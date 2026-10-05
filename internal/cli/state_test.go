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

// startedStatePlan is a plan with work going on: started, some boxes ticked,
// some open, and a State section with all three subsections.
const startedStatePlan = "---\nid: PLN-0082\nstarted: \"2026-10-05 08:32:15\"\n---\n" +
	"# Live work state\n\n### Task 1: acta state set\n\n- [x] write it\n\n" +
	"### Task 2: acta state view\n\n- [ ] read it\n\n### Task 3: session start\n\n- [ ] hook it\n\n" +
	"## State\n\n### Next\n\nrun the state test\n\n### Findings\n\nthe hook is cheap\n\n" +
	"### Open rulings\n\nnone yet\n"

// stateFreshPlan is the plan before any work: no stamp and every box open.
const stateFreshPlan = "---\nid: PLN-0082\n---\n# Live work state\n\n" +
	"### Task 1: acta state set\n\n- [ ] write it\n\n" +
	"### Task 2: acta state view\n\n- [ ] read it\n\n### Task 3: session start\n\n- [ ] hook it\n"

// stateFinishedPlan is that plan after the last build: started, finished,
// every box ticked, and a Findings line. Land runs at that moment, so this is
// the shape of the plan the land gate has to read.
const stateFinishedPlan = "---\nid: PLN-0082\nstarted: \"2026-10-05 08:32:15\"\nfinished: \"2026-10-05 09:00:00\"\n---\n" +
	"# Live work state\n\n### Task 1: acta state set\n\n- [x] write it\n\n" +
	"### Task 2: acta state view\n\n- [x] read it\n\n### Task 3: session start\n\n- [x] hook it\n\n" +
	"## State\n\n### Findings\n\nThe row is comma separated, so a price cannot hold a comma.\n"

// stateWorktreePlan makes a git repo holding the given plan and a worktree on a
// branch that commits inWt as the work. The main file is what git copied into
// the worktree; the branch commit is what says the work was done there.
func stateWorktreePlan(t *testing.T, plan, inWt string) (dir, wt string) {
	t.Helper()
	dir = stateRepo(t, plan)
	wt = filepath.Join(t.TempDir(), "live")
	gitOut(t, dir, "worktree", "add", "-q", "-b", "live", wt)
	if err := os.WriteFile(stateFile(wt), []byte(inWt), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, wt, "add", "-A")
	gitOut(t, wt, "commit", "-q", "-m", "state: the agent works on the plan")
	return dir, wt
}

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
	if got := gitOut(t, dir, "log", "-1", "--format=%s"); got != "chore(plan): state plans/2026-10-05-live-work-state" {
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

// stateWorktreeRepo makes a git repo holding the given plan and a worktree on
// a branch, so the plan counts as running. The branch commits the plan file,
// because that is the work: a plan git only copied into the worktree is not
// running there.
func stateWorktreeRepo(t *testing.T, plan string) (dir, wt string) {
	t.Helper()
	return stateWorktreePlan(t, plan, strings.Replace(plan, "- [ ] hook it", "- [x] hook it", 1))
}

// A started plan with a worktree shows what the work is standing on: the
// first task with an open box, the branch's last commit, the worktree path,
// the review round, and then the three subsections as the file holds them.
func TestCmdStatePrintsTheRunningPlan(t *testing.T) {
	dir, wt := stateWorktreeRepo(t, startedStatePlan)
	gitOut(t, wt, "commit", "-q", "--allow-empty", "-m", "state: show the view")
	code, out, stderr := runState(t, dir, "", "plans/2026-10-05-live-work-state")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	for _, want := range []string{
		"plans/2026-10-05-live-work-state#task-2",
		"state: show the view",
		wt,
		"round: none",
		"### Next\n\nrun the state test",
		"### Findings\n\nthe hook is cheap",
		"### Open rulings\n\nnone yet",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not hold %q\n%s", want, out)
		}
	}
	if strings.Contains(out, "#task-1") || strings.Contains(out, "#task-3") {
		t.Errorf("stdout names a task that is not the first open one\n%s", out)
	}
}

// The bare command prints one line per running plan, and nothing at all in a
// repo where no plan is running, so an empty answer means an empty board.
func TestCmdStateBarePrintsOneLinePerRunningPlan(t *testing.T) {
	dir, _ := stateWorktreeRepo(t, startedStatePlan)
	code, out, stderr := runState(t, dir, "")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	if lines := strings.Split(strings.TrimRight(out, "\n"), "\n"); len(lines) != 1 ||
		!strings.Contains(lines[0], "plans/2026-10-05-live-work-state") {
		t.Fatalf("bare stdout = %q, want one line naming the running plan", out)
	}

	bare := stateRepo(t, startedStatePlan)
	code, out, stderr = runState(t, bare, "")
	if code != exitOK {
		t.Fatalf("repo with no worktree: exit %d stderr %q", code, stderr)
	}
	if out != "" {
		t.Errorf("repo with no worktree printed %q, want nothing", out)
	}
}

// A plan that is not running, and an id the board never knew, both fail with
// the reason on stderr and print nothing, so a caller cannot mistake silence
// for an answer.
func TestCmdStateRefusesAPlanThatIsNotRunning(t *testing.T) {
	dir := stateRepo(t, startedStatePlan)
	for _, id := range []string{"plans/2026-10-05-live-work-state", "plans/2026-10-06-nope"} {
		code, out, stderr := runState(t, dir, "", id)
		if code != exitBadInput {
			t.Errorf("%s: exit %d, want bad input", id, code)
		}
		if out != "" {
			t.Errorf("%s: stdout %q, want nothing", id, out)
		}
		if !strings.Contains(stderr, "no running plan") {
			t.Errorf("%s: stderr %q does not say the plan is not running", id, stderr)
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

// Land runs after the last box is ticked, so the plan it lands carries a
// finished stamp. Asking that plan for its view still has to print, because
// the land gate reads it there and then. The bare list and the session start
// hook leave it out, because work on it is over.
func TestCmdStateShowsAFinishedPlanInAWorktree(t *testing.T) {
	_, wt := stateWorktreePlan(t, stateFreshPlan, stateFinishedPlan)
	gitOut(t, wt, "commit", "-q", "--allow-empty", "-m", "state: the last task landed")

	code, out, stderr := runState(t, wt, "", "plans/2026-10-05-live-work-state")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q, want the finished plan to print", code, stderr)
	}
	for _, want := range []string{
		"plans/2026-10-05-live-work-state  Live work state",
		"task: none, every box is ticked",
		"state: the last task landed",
		wt,
		"round: none",
		"### Findings\n\nThe row is comma separated, so a price cannot hold a comma.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("stdout does not hold %q\n%s", want, out)
		}
	}

	if code, bare, stderr := runState(t, wt, ""); code != exitOK || bare != "" {
		t.Errorf("bare listing = %q (exit %d, stderr %q), want nothing: the plan is finished", bare, code, stderr)
	}
	if hook := sessionStart(t, wt); strings.Contains(hook, "plans/2026-10-05-live-work-state") {
		t.Errorf("the session start names a finished plan:\n%s", hook)
	}

	code, out, stderr = runState(t, wt, "", "plans/2026-10-06-nope")
	if code != exitBadInput || out != "" || !strings.Contains(stderr, "no running plan") {
		t.Errorf("an id the board never knew: exit %d stdout %q stderr %q, want bad input and nothing on stdout", code, out, stderr)
	}
}

// A plan in a worktree that has no State section yet still has a view: the
// four facts print and the three subsections stay out, so a missing section
// is not a failed command.
func TestCmdStateShowsAWorktreePlanWithNoStateSection(t *testing.T) {
	plan := "---\nid: PLN-0082\nstarted: \"2026-10-05 08:32:15\"\n---\n# Live work state\n\n" +
		"### Task 1: acta state set\n\n- [x] write it\n\n### Task 2: acta state view\n\n- [ ] read it\n"
	_, wt := stateWorktreePlan(t, plan, strings.Replace(plan, "- [ ] read it", "- [x] read it", 1))

	code, out, stderr := runState(t, wt, "", "plans/2026-10-05-live-work-state")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	if !strings.Contains(out, "task: none, every box is ticked") {
		t.Errorf("stdout does not hold the facts\n%s", out)
	}
	for _, gone := range []string{"### Next", "### Findings", "### Open rulings"} {
		if strings.Contains(out, gone) {
			t.Errorf("a plan with no State section printed %q\n%s", gone, out)
		}
	}
}
