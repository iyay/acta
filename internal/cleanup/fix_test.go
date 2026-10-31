package cleanup

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

func (r testRepo) load() (config.Config, *board.Board) {
	r.t.Helper()
	cfg, err := config.Load(r.dir, "")
	if err != nil {
		r.t.Fatal(err)
	}
	b, err := board.Load(cfg)
	if err != nil {
		r.t.Fatal(err)
	}
	return cfg, b
}

func (r testRepo) read(p string) string {
	r.t.Helper()
	body, err := os.ReadFile(filepath.Join(r.dir, p))
	if err != nil {
		r.t.Fatal(err)
	}
	return string(body)
}

func (r testRepo) out(args ...string) string {
	r.t.Helper()
	return strings.TrimSpace(gitOutput(r.t, r.dir, args...))
}

var fixFiles = map[string]string{
	// The spec has a plan, so the board ignores its written draft.
	".acta/specs/2026-01-01-alpha.md": "---\nid: SPC-0001\nstatus: draft\n---\n# Alpha\n",
	".acta/plans/2026-01-01-alpha.md": "---\nparent: specs/2026-01-01-alpha\n---\n# Alpha plan\n\n### Task 1: A\n- [ ] a\n",
	// Every box is ticked, but the written draft keeps the plan open.
	".acta/plans/2026-01-02-ticked.md": "---\nstatus: draft\n---\n# Ticked\n\n### Task 1: A\n- [x] a\n",
	// No tasks at all: nothing proves this plan is finished.
	".acta/plans/2026-01-03-empty.md": "---\nstatus: draft\n---\n# Empty\n",
	// Half done: still open work.
	".acta/plans/2026-01-04-half.md": "---\nstatus: draft\n---\n# Half\n\n### Task 1: A\n- [x] a\n\n### Task 2: B\n- [ ] b\n",
}

func TestFixWritesEveryNoJudgmentChangeInOneCommit(t *testing.T) {
	t.Parallel()
	r := newRepo(t, 0, fixFiles)
	before := r.out("rev-list", "--count", "HEAD")
	cfg, b := r.load()
	o, changes, err := Fix(cfg, b)
	if err != nil || !o.Committed {
		t.Fatalf("outcome %+v err %v", o, err)
	}
	if len(changes) != 2 {
		t.Fatalf("changes = %+v, want the spec and the ticked plan", changes)
	}
	if got := r.read(".acta/specs/2026-01-01-alpha.md"); got != "---\nid: SPC-0001\n---\n# Alpha\n" {
		t.Errorf("spec keeps its ignored status: %q", got)
	}
	if got := r.read(".acta/plans/2026-01-02-ticked.md"); !strings.Contains(got, "status: done\n") {
		t.Errorf("ticked plan not done: %q", got)
	}
	for _, p := range []string{".acta/plans/2026-01-03-empty.md", ".acta/plans/2026-01-04-half.md"} {
		if got := r.read(p); got != fixFiles[p] {
			t.Errorf("%s changed: %q", p, got)
		}
	}
	if after := r.out("rev-list", "--count", "HEAD"); after != "2" || before != "1" {
		t.Fatalf("commits before %s after %s, want one new commit", before, after)
	}
	if got := r.out("status", "--porcelain"); got != "" {
		t.Fatalf("left changes: %s", got)
	}
	_, b = r.load()
	if again, err := Plan(cfg, b); err != nil || len(again) != 0 {
		t.Fatalf("a second run still finds %+v (err %v)", again, err)
	}
}

func TestFixWithNothingToDoMakesNoCommit(t *testing.T) {
	t.Parallel()
	r := newRepo(t, 0, map[string]string{".acta/plans/2026-01-04-half.md": fixFiles[".acta/plans/2026-01-04-half.md"]})
	cfg, b := r.load()
	o, changes, err := Fix(cfg, b)
	if err != nil || o.Committed || len(changes) != 0 {
		t.Fatalf("outcome %+v changes %+v err %v", o, changes, err)
	}
	if got := r.out("rev-list", "--count", "HEAD"); got != "1" {
		t.Fatalf("commits = %s", got)
	}
}

// A file the user is still editing would be swept into the commit, so the
// whole fix stops before it writes anything.
func TestFixRefusesWhenATargetIsDirty(t *testing.T) {
	t.Parallel()
	r := newRepo(t, 0, fixFiles)
	r.write(map[string]string{".acta/plans/2026-01-02-ticked.md": fixFiles[".acta/plans/2026-01-02-ticked.md"] + "note\n"})
	cfg, b := r.load()
	if _, _, err := Fix(cfg, b); err == nil {
		t.Fatal("want an error for the dirty plan")
	}
	if got := r.read(".acta/specs/2026-01-01-alpha.md"); got != fixFiles[".acta/specs/2026-01-01-alpha.md"] {
		t.Fatalf("the spec was written before the check failed: %q", got)
	}
	if got := r.out("rev-list", "--count", "HEAD"); got != "1" {
		t.Fatalf("commits = %s", got)
	}
}

// A plan a linked worktree is building keeps its ticks there, so main never
// closes it.
func TestFixLeavesABuildingPlanAlone(t *testing.T) {
	t.Parallel()
	plan := ".acta/plans/2026-01-02-ticked.md"
	r := newRepo(t, 0, map[string]string{plan: fixFiles[plan]})
	wt := filepath.Join(t.TempDir(), "wt")
	r.git(0, "worktree", "add", "-q", wt, "-b", "building")
	testRepo{t, wt}.commit(0, "chore(plan): tick", map[string]string{plan: fixFiles[plan] + "\n"})
	cfg, b := r.load()
	changes, err := Plan(cfg, b)
	if err != nil || len(changes) != 0 {
		t.Fatalf("changes %+v err %v, want none", changes, err)
	}
}

func gitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return string(out)
}
