package board

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
)

// runRepo makes a git repo with a worktree on a branch called live. The main
// files are committed on main, the worktree files on live, so a plan that only
// the branch works on is written there. It gives the config of the main tree
// and the path of the worktree as git writes it.
func runRepo(t *testing.T, main, live map[string]string) (config.Config, string) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "repo")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	gitRun(t, dir, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(dir, "README.md"), "a repo\n")
	for rel, body := range main {
		writeFile(t, filepath.Join(dir, ".acta", rel), body)
	}
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-q", "-m", "init")
	wt := filepath.Join(t.TempDir(), "live")
	gitRun(t, dir, "worktree", "add", "-q", "-b", "live", wt)
	if len(live) > 0 {
		commitFiles(t, wt, "live work", live)
	}
	real, err := filepath.EvalSymlinks(wt)
	if err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load(dir, "")
	if err != nil {
		t.Fatal(err)
	}
	return cfg, real
}

// commitFiles writes the given planning files in dir and commits them, the way
// an agent writes in its own worktree.
func commitFiles(t *testing.T, dir, msg string, files map[string]string) {
	t.Helper()
	for rel, body := range files {
		writeFile(t, filepath.Join(dir, ".acta", rel), body)
	}
	gitRun(t, dir, "add", "-A")
	gitRun(t, dir, "commit", "-q", "-m", msg)
}

// runIds gives the ids of the running plans, so a test can ask who is running.
func runIds(runs []RunState) []string {
	var out []string
	for _, r := range runs {
		out = append(out, r.Plan.ID)
	}
	return out
}

const (
	startedPlan = "---\nid: PLN-0082\nstarted: \"2026-10-05 08:32:15\"\n---\n" +
		"# Live work state\n\n### Task 1: acta state set\n\n- [x] write it\n\n" +
		"### Task 2: acta state view\n\n- [ ] read it\n\n### Task 3: session start\n\n- [ ] hook it\n\n" +
		"## State\n\n### Next\n\nrun the state test\n\n### Findings\n\nthe hook is cheap\n\n" +
		"### Open rulings\n\nnone yet\n"
	finishedPlan = "---\nid: PLN-0083\nstarted: \"2026-10-05 08:32:15\"\nfinished: \"2026-10-05 09:00:00\"\n---\n" +
		"# Landed plan\n\n### Task 1: acta state set\n\n- [x] write it\n"
	untouchedPlan = "---\nid: PLN-0084\n---\n# Fresh plan\n\n### Task 1: acta state set\n\n- [ ] write it\n"
	mainTreePlan  = "---\nid: PLN-0090\nstarted: \"2026-10-05 07:00:00\"\n---\n" +
		"# Main tree plan\n\n### Task 1: acta state set\n\n- [x] write it\n\n### Task 2: read it\n\n- [ ] read it\n"
)

// A plan counts as running only when it has started, has not finished, and has
// a worktree. Each of the three missing on its own keeps it out of the list,
// and a repo with nothing started has no running plans at all.
func TestRunningNeedsStartNoFinishAndAWorktree(t *testing.T) {
	t.Parallel()

	cfg, _ := runRepo(t, nil, map[string]string{
		"plans/2026-10-05-live.md":   startedPlan,
		"plans/2026-10-06-landed.md": finishedPlan,
		"plans/2026-10-07-fresh.md":  untouchedPlan,
	})
	runs := Running(cfg)
	if got := runIds(runs); len(got) != 1 || got[0] != "plans/2026-10-05-live" {
		t.Fatalf("running = %v, want only the started plan in the worktree", got)
	}

	bare := filepath.Join(t.TempDir(), "solo")
	if err := os.MkdirAll(filepath.Join(bare, ".acta", "plans"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(bare, ".acta", "plans", "2026-10-05-live.md"), []byte(startedPlan), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, bare, "init", "-q", "-b", "main")
	gitRun(t, bare, "add", ".")
	gitRun(t, bare, "commit", "-q", "-m", "init")
	sole, err := config.Load(bare, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := Running(sole); len(got) != 0 {
		t.Errorf("a started plan with no worktree is running: %v", runIds(got))
	}
}

// A plan whose every box is ticked is still running: the State section is the
// last record of the work, so it stays on show after the last task lands.
func TestRunningKeepsAPlanWithEveryBoxTicked(t *testing.T) {
	t.Parallel()

	ticked := strings.ReplaceAll(startedPlan, "- [ ]", "- [x]")
	cfg, _ := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": ticked})
	runs := Running(cfg)
	if got := runIds(runs); len(got) != 1 {
		t.Fatalf("running = %v, want the fully ticked plan to stay running", got)
	}
	if runs[0].Task != nil {
		t.Errorf("current task = %q, want none when every box is ticked", runs[0].Task.ID)
	}
}

// The current task is the first one with an open box, so the view points at the
// work a fresh session takes over, not at the last task that was left.
func TestRunningTaskIsTheFirstOneStillOpen(t *testing.T) {
	t.Parallel()

	cfg, _ := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	if runs[0].Task == nil {
		t.Fatal("the running plan has no current task")
	}
	if got := runs[0].Task.ID; got != "plans/2026-10-05-live#task-2" {
		t.Errorf("current task = %q, want task-2, the first task with an open box", got)
	}
}

// The last commit comes from the branch, not from the plan file. A plan whose
// body names a different commit does not change what the view prints.
func TestRunningCommitIsTheBranchTipNotTheFileText(t *testing.T) {
	t.Parallel()

	cfg, wt := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	lie := strings.Replace(startedPlan, "### Next\n", "### Next\n\nlast commit: 9999999 the file lies\n", 1)
	if err := os.WriteFile(filepath.Join(wt, ".acta", "plans", "2026-10-05-live.md"), []byte(lie), 0o644); err != nil {
		t.Fatal(err)
	}
	gitRun(t, wt, "add", ".")
	gitRun(t, wt, "commit", "-q", "-m", "state: plant a lie in the plan")

	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	if !strings.Contains(runs[0].Commit, "plant a lie in the plan") {
		t.Errorf("last commit = %q, want the newest commit on the branch", runs[0].Commit)
	}
	if strings.Contains(runs[0].Commit, "the file lies") {
		t.Errorf("last commit came from the file text: %q", runs[0].Commit)
	}
}

// The review round is the newest commit that opens one, so a branch with two
// rounds reports the later one.
func TestRunningRoundIsTheNewestFixRound(t *testing.T) {
	t.Parallel()

	cfg, wt := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	for _, msg := range []string{"acta: tick fix round 1", "acta: tick fix round 3"} {
		gitRun(t, wt, "commit", "-q", "--allow-empty", "-m", msg)
	}
	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	if runs[0].Round != "3" {
		t.Errorf("round = %q, want 3, the newest fix round on the branch", runs[0].Round)
	}
}

// The round is found from the new subject and from the old one, so history
// keeps working after the subject changes.
func TestRunningRoundReadsBothSubjects(t *testing.T) {
	t.Parallel()

	cfg, wt := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	gitRun(t, wt, "commit", "-q", "--allow-empty", "-m", "acta: tick fix round 2")
	gitRun(t, wt, "commit", "-q", "--allow-empty", "-m", "chore(plan): tick fix round 5")
	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	if runs[0].Round != "5" {
		t.Errorf("round = %q, want 5, the newest fix round in either subject", runs[0].Round)
	}
}

// A branch with no fix round commit has no round, and the view says so instead
// of leaving the line out or printing an empty number.
func TestRunningRoundIsNoneWithoutAFixRound(t *testing.T) {
	t.Parallel()

	cfg, _ := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	if runs[0].Round != "" {
		t.Errorf("round = %q, want none", runs[0].Round)
	}
	if !strings.Contains(runs[0].Full(), "round: none") {
		t.Errorf("the full view does not say round: none\n%s", runs[0].Full())
	}
}

// The short view is what a session start hook prints, so it never takes more
// than three lines however long the plan id, the worktree path and the commit
// subject are.
func TestShortHoldsThreeLines(t *testing.T) {
	t.Parallel()

	cfg, _ := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	if got := len(runs[0].Short()); got > 3 {
		t.Errorf("Short() gave %d lines, want at most 3\n%q", got, runs[0].Short())
	}
	for _, want := range []string{"plans/2026-10-05-live", "#task-2", "run the state test"} {
		if !strings.Contains(strings.Join(runs[0].Short(), "\n"), want) {
			t.Errorf("Short() does not name %q\n%q", want, runs[0].Short())
		}
	}
}

// The full view prints the four facts and then the three subsections as the
// file holds them, so a session reads back what the last one wrote.
func TestFullShowsTheFactsAndTheThreeSubsections(t *testing.T) {
	t.Parallel()

	cfg, wt := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	full := runs[0].Full()
	for _, want := range []string{
		"task: plans/2026-10-05-live#task-2",
		"worktree: " + wt,
		"round: none",
		"### Next\n\nrun the state test",
		"### Findings\n\nthe hook is cheap",
		"### Open rulings\n\nnone yet",
	} {
		if !strings.Contains(full, want) {
			t.Errorf("the full view does not hold %q\n%s", want, full)
		}
	}
	if !strings.Contains(full, "last commit: ") {
		t.Errorf("the full view does not name the last commit\n%s", full)
	}
}

// A plan whose State section holds two of the three subsections shows those
// two, so a part nobody wrote is not printed as an empty heading.
func TestFullShowsOnlyTheSubsectionsTheFileHolds(t *testing.T) {
	t.Parallel()

	two := strings.Replace(startedPlan, "### Open rulings\n\nnone yet\n", "", 1)
	cfg, _ := runRepo(t, nil, map[string]string{"plans/2026-10-05-live.md": two})
	runs := Running(cfg)
	if len(runs) != 1 {
		t.Fatalf("running = %v, want one plan", runIds(runs))
	}
	full := runs[0].Full()
	if !strings.Contains(full, "### Next") || !strings.Contains(full, "### Findings") {
		t.Errorf("the full view lost a subsection the file holds\n%s", full)
	}
	if strings.Contains(full, "### Open rulings") {
		t.Errorf("the full view printed a subsection the file does not hold\n%s", full)
	}
}

// The main tree and a linked worktree each hold a started plan of their own. Git
// lists the main checkout first, so a running plan is every worktree but that
// first one. The answer is the same from both sides: a session started inside
// the worktree finds the plan it was building, the main tree still sees that
// plan, and neither sees the plan that only lives in the main tree.
func TestRunningCountsEveryWorktreeButTheMainOne(t *testing.T) {
	t.Parallel()

	main, wt := runRepo(t, map[string]string{"plans/2026-10-05-in-main.md": mainTreePlan},
		map[string]string{"plans/2026-10-05-live.md": startedPlan})
	// The worktree also holds a copy of the main tree's plan. Nobody worked on
	// that copy there, so it is not running in the worktree.
	inWt, err := config.Load(wt, "")
	if err != nil {
		t.Fatal(err)
	}

	if got := runIds(Running(inWt)); len(got) != 1 || got[0] != "plans/2026-10-05-live" {
		t.Errorf("from the worktree: running = %v, want only the plan of that worktree", got)
	}
	if got := runIds(Running(main)); len(got) != 1 || got[0] != "plans/2026-10-05-live" {
		t.Errorf("from the main tree: running = %v, want only the plan of the worktree", got)
	}
	for _, runs := range [][]RunState{Running(inWt), Running(main)} {
		for _, r := range runs {
			if sameDir(r.Worktree, main.RepoRoot) {
				t.Errorf("the main checkout was counted as a worktree: %s", r.Worktree)
			}
		}
	}
}

// A started plan that lives only in the main tree is never running, from either
// folder. The main checkout is not a worktree of itself.
func TestRunningNeverCountsAMainTreePlan(t *testing.T) {
	t.Parallel()

	main, wt := runRepo(t, map[string]string{"plans/2026-10-05-in-main.md": mainTreePlan}, nil)
	// The worktree holds the same bytes, so it is not running there either.
	inWt, err := config.Load(wt, "")
	if err != nil {
		t.Fatal(err)
	}

	if got := runIds(Running(main)); len(got) != 0 {
		t.Errorf("from the main tree: running = %v, want none", got)
	}
	if got := runIds(Running(inWt)); len(got) != 0 {
		t.Errorf("from the worktree: running = %v, want none", got)
	}
}

// Git copies every tracked file into a new worktree, so a started plan that
// only lives in the main tree shows up in the worktree byte for byte. That
// copy is not work in progress: nobody built it there, so it is not running.
// The plan the branch really works on stays running from both folders.
func TestRunningSkipsAPlanTheBranchNeverTouched(t *testing.T) {
	t.Parallel()

	// The main tree holds a plan that has started. The branch never writes it,
	// so its copy in the worktree holds the same bytes.
	main, wt := runRepo(t, map[string]string{
		"plans/2026-10-05-mainonly.md": mainTreePlan,
		"plans/2026-10-05-live.md":     untouchedPlan,
	}, map[string]string{"plans/2026-10-05-live.md": startedPlan})
	inWt, err := config.Load(wt, "")
	if err != nil {
		t.Fatal(err)
	}
	// The two copies are the same file, which is why the file alone cannot
	// answer whether the branch worked on it.
	a, err := os.ReadFile(filepath.Join(main.RepoRoot, ".acta", "plans", "2026-10-05-mainonly.md"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(filepath.Join(wt, ".acta", "plans", "2026-10-05-mainonly.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("the fixture is wrong: the two copies differ")
	}

	for _, c := range []config.Config{main, inWt} {
		if got := runIds(Running(c)); len(got) != 1 || got[0] != "plans/2026-10-05-live" {
			t.Errorf("running = %v, want only the plan the branch works on", got)
		}
	}
}

// A branch that works on a plan only works on it in its own worktree. A
// second linked worktree that never wrote the plan keeps the same bytes, so
// only the worktree that built it lists the plan as running.
func TestRunningOnlyCountsTheWorktreeThatBuiltThePlan(t *testing.T) {
	t.Parallel()

	main, wt := runRepo(t, map[string]string{"plans/2026-10-05-shared.md": mainTreePlan},
		map[string]string{"plans/2026-10-05-live.md": startedPlan})
	// A second worktree on its own branch. It holds the shared plan as a copy
	// and never touches it.
	other := filepath.Join(t.TempDir(), "other")
	gitRun(t, main.RepoRoot, "worktree", "add", "-q", "-b", "other", other)

	runs := Running(main)
	if got := runIds(runs); len(got) != 1 || got[0] != "plans/2026-10-05-live" {
		t.Fatalf("running = %v, want only the plan the live branch works on", got)
	}
	if !sameDir(runs[0].Worktree, wt) {
		t.Errorf("the plan was counted in %s, want the worktree that built it %s", runs[0].Worktree, wt)
	}
}

// A branch that has no commits past the scaffold has done nothing yet, so it
// has no running plan even when the plan file in it says it started.
func TestRunningSkipsAWorktreeWithNoBranchWork(t *testing.T) {
	t.Parallel()

	main, wt := runRepo(t, map[string]string{"plans/2026-10-05-shared.md": mainTreePlan}, nil)
	inWt, err := config.Load(wt, "")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []config.Config{main, inWt} {
		if got := runIds(Running(c)); len(got) != 0 {
			t.Errorf("running = %v, want none: the branch wrote nothing", got)
		}
	}
}

// The main tree is the yardstick for the plan files, so a main checkout that
// cannot be read must not crash and must not turn its plans into running ones.
func TestRunningSurvivesAMainCheckoutItCannotRead(t *testing.T) {
	t.Parallel()
	if os.Geteuid() == 0 {
		t.Skip("root reads every file, so an unreadable folder is not one here")
	}

	main, wt := runRepo(t, map[string]string{"plans/2026-10-05-mainonly.md": mainTreePlan},
		map[string]string{"plans/2026-10-05-live.md": startedPlan})
	// Hide the main checkout's plan folder. The worktree still holds a copy of
	// the plan, and the main checkout no longer answers for it.
	if err := os.Chmod(filepath.Join(main.RepoRoot, ".acta", "plans"), 0o000); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = os.Chmod(filepath.Join(main.RepoRoot, ".acta", "plans"), 0o755)
	})

	if got := runIds(Running(main)); len(got) != 1 || got[0] != "plans/2026-10-05-live" {
		t.Errorf("running = %v, want only the plan the branch works on", got)
	}

	inWt, err := config.Load(wt, "")
	if err != nil {
		t.Fatal(err)
	}
	if got := runIds(Running(inWt)); len(got) != 1 || got[0] != "plans/2026-10-05-live" {
		t.Errorf("from the worktree: running = %v, want only the plan the branch works on", got)
	}
}

// A branch that wrote the plan and then put the old bytes back leaves the plan
// looking untouched. The commits are still there, so the work is still running
// and the view has to keep showing it.
func TestRunningKeepsAPlanTheBranchWroteAndReverted(t *testing.T) {
	t.Parallel()

	main, wt := runRepo(t, map[string]string{"plans/2026-10-05-shared.md": mainTreePlan}, nil)
	commitFiles(t, wt, "tick the plan", map[string]string{"plans/2026-10-05-shared.md": startedPlan})
	// Put the plan back the way the main tree holds it.
	commitFiles(t, wt, "revert the plan", map[string]string{"plans/2026-10-05-shared.md": mainTreePlan})
	inWt, err := config.Load(wt, "")
	if err != nil {
		t.Fatal(err)
	}

	for _, c := range []config.Config{main, inWt} {
		if got := runIds(Running(c)); len(got) != 1 || got[0] != "plans/2026-10-05-shared" {
			t.Errorf("running = %v, want the plan the branch wrote", got)
		}
	}
}
