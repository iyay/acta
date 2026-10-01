package trees

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/config"
	"github.com/iyay/acta/internal/testguard"
)

// TestMain clears an inherited PM_ROOT once, so no helper has to set env
// and block parallel tests.
func TestMain(m *testing.M) {
	testguard.Watch()
	os.Unsetenv("PM_ROOT")
	os.Exit(m.Run())
}

func run(t *testing.T, dir string, args ...string) {
	t.Helper()
	if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func write(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func setup(t *testing.T) (repo, wt string) {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo = filepath.Join(base, "repo")
	run(t, base, "init", "-q", "-b", "main", repo)
	// Name the author inside the repo, not in the env, so tests that make
	// commits can still run side by side.
	run(t, repo, "config", "user.name", "test")
	run(t, repo, "config", "user.email", "test@example.com")
	write(t, filepath.Join(repo, ".pm/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [ ] a\n- [ ] b\n")
	run(t, repo, "add", ".")
	run(t, repo, "commit", "-q", "-m", "init")
	wt = filepath.Join(base, "repo-feat")
	run(t, repo, "worktree", "add", "-q", wt, "-b", "feat")
	write(t, filepath.Join(wt, ".pm/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [x] a\n- [ ] b\n")
	return repo, wt
}

func TestOthersAndLoad(t *testing.T) {
	t.Parallel()

	repo, wt := setup(t)
	cfg, err := config.Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	others := Others(cfg)
	if len(others) != 1 || others[0].Branch != "feat" || others[0].Cfg.RepoRoot != wt {
		t.Fatalf("others = %+v", others)
	}
	b, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if it := b.Get("plans/2026-09-21-a#task-1"); it.Status != "in-progress" || it.Worktree != "feat" {
		t.Fatalf("task = %+v", it)
	}

	// Run from inside the worktree: the main checkout becomes the other tree.
	wcfg, _ := config.Load(wt, "")
	if o := Others(wcfg); len(o) != 1 || o[0].Branch != "main" {
		t.Fatalf("from the worktree: %+v", o)
	}
}

func TestOthersReadsUnmergedBranches(t *testing.T) {
	t.Parallel()

	repo, _ := setup(t)
	run(t, repo, "checkout", "-q", "-b", "feat-x")
	write(t, filepath.Join(repo, ".pm/specs/2026-09-26-x.md"), "# Branch story\n")
	run(t, repo, "add", ".")
	run(t, repo, "commit", "-q", "-m", "spec on feat-x")
	run(t, repo, "checkout", "-q", "main")

	cfg, _ := config.Load(repo, "")
	b, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	it := b.Get("specs/2026-09-26-x")
	if it == nil || it.Worktree != "feat-x" || it.OnDisk {
		t.Fatalf("branch story = %+v", it)
	}
	var names []string
	for _, tr := range Others(cfg) {
		names = append(names, tr.Branch)
	}
	if strings.Join(names, ",") != "feat,feat-x" {
		t.Fatalf("trees = %v (feat is a worktree, read from disk once; main is HEAD)", names)
	}

	write(t, filepath.Join(repo, ".pm.yaml"), "branches: []\n")
	off, _ := config.Load(repo, "")
	if b2, _ := Load(off); b2.Get("specs/2026-09-26-x") != nil {
		t.Fatal("branches: [] must turn branch reading off")
	}
	write(t, filepath.Join(repo, ".pm.yaml"), "branches: [\"fix-*\"]\n")
	filtered, _ := config.Load(repo, "")
	if b3, _ := Load(filtered); b3.Get("specs/2026-09-26-x") != nil {
		t.Fatal("a pattern list must keep only matching branches")
	}
}

func TestOthersReadsWorktreeKeptOnOldRoot(t *testing.T) {
	t.Parallel()

	repo, _ := setup(t)
	// The main tree moved to .acta, but the worktree on disk still keeps .pm.
	write(t, filepath.Join(repo, ".acta/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [ ] a\n- [ ] b\n")
	cfg, err := config.Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if it := b.Get("plans/2026-09-21-a#task-1"); it.Status != "in-progress" || it.Worktree != "feat" {
		t.Fatalf("task = %+v, want the checked box from the .pm worktree", it)
	}
}

func TestOthersReadsWorktreeAlreadyOnNewRoot(t *testing.T) {
	t.Parallel()

	repo, wt := setup(t)
	// The main tree still uses .pm, but the worktree already moved to .acta.
	if err := os.RemoveAll(filepath.Join(wt, ".pm")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(wt, ".acta/plans/2026-09-21-b.md"), "# Plan B\n\n### Task 1: One\n- [x] a\n- [ ] b\n")
	cfg, err := config.Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if it := b.Get("plans/2026-09-21-b#task-1"); it == nil || it.Worktree != "feat" {
		t.Fatalf("task = %+v, want the plan from the .acta worktree", it)
	}
}

func TestOthersKeepsMatchingRootName(t *testing.T) {
	t.Parallel()

	repo, wt := setup(t)
	// Both trees moved to .acta: the worktree must be read from .acta itself.
	write(t, filepath.Join(repo, ".acta/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [ ] a\n- [ ] b\n")
	if err := os.RemoveAll(filepath.Join(wt, ".pm")); err != nil {
		t.Fatal(err)
	}
	write(t, filepath.Join(wt, ".acta/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [x] a\n- [ ] b\n")
	cfg, err := config.Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	others := Others(cfg)
	if len(others) != 1 || others[0].Branch != "feat" || others[0].Cfg.Root != filepath.Join(wt, ".acta") {
		t.Fatalf("others = %+v, want feat read from its own .acta", others)
	}
}

func TestOthersReadsUnmergedBranchOnOldRoot(t *testing.T) {
	t.Parallel()

	repo, _ := setup(t)
	run(t, repo, "checkout", "-q", "-b", "feat-x")
	write(t, filepath.Join(repo, ".pm/specs/2026-09-26-x.md"), "# Branch story\n")
	run(t, repo, "add", ".")
	run(t, repo, "commit", "-q", "-m", "spec on feat-x")
	run(t, repo, "checkout", "-q", "main")
	// The main tree moved to .acta, but feat-x only ever held .pm files
	// and is not checked out anywhere.
	write(t, filepath.Join(repo, ".acta/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [ ] a\n- [ ] b\n")
	cfg, err := config.Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	b, err := Load(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if it := b.Get("specs/2026-09-26-x"); it == nil || it.Worktree != "feat-x" || it.OnDisk {
		t.Fatalf("branch story = %+v, want feat-x read from git", it)
	}
}

func TestOthersListsCheckedOutBranchOnce(t *testing.T) {
	t.Parallel()

	repo, _ := setup(t)
	write(t, filepath.Join(repo, ".acta/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [ ] a\n- [ ] b\n")
	cfg, err := config.Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, tr := range Others(cfg) {
		if tr.Branch == "feat" {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("feat listed %d times, want once", n)
	}
}

func TestOthersOutsideGit(t *testing.T) {
	t.Parallel()

	cfg, _ := config.Load(t.TempDir(), "")
	if Others(cfg) != nil {
		t.Fatal("no trees outside git")
	}
}

func TestWatchDirs(t *testing.T) {
	t.Parallel()

	repo, wt := setup(t)
	cfg, _ := config.Load(repo, "")
	dirs := WatchDirs(cfg, func(c config.Config) []string { return []string{c.Root} })
	joined := strings.Join(dirs, "\n")
	if !strings.Contains(joined, filepath.Join(wt, ".pm")) || !strings.Contains(joined, filepath.Join(repo, ".git", "worktrees")) || !strings.Contains(joined, filepath.Join(repo, ".git", "refs", "heads")) {
		t.Fatalf("dirs = %v", dirs)
	}
	if strings.Contains(joined, filepath.Join(repo, ".pm")+"\n") {
		t.Fatal("the main tree's folders belong to the caller, not to WatchDirs")
	}
}
