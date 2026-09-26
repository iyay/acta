package trees

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"pm-board/internal/config"
)

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
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "test")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "test@example.com")
	}
	t.Setenv("PM_ROOT", "")
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo = filepath.Join(base, "repo")
	run(t, base, "init", "-q", "-b", "main", repo)
	write(t, filepath.Join(repo, ".pm/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [ ] a\n- [ ] b\n")
	run(t, repo, "add", ".")
	run(t, repo, "commit", "-q", "-m", "init")
	wt = filepath.Join(base, "repo-feat")
	run(t, repo, "worktree", "add", "-q", wt, "-b", "feat")
	write(t, filepath.Join(wt, ".pm/plans/2026-09-21-a.md"), "# Plan A\n\n### Task 1: One\n- [x] a\n- [ ] b\n")
	return repo, wt
}

func TestOthersAndLoad(t *testing.T) {
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
	if it := b.Get("plans/2026-09-21-a#task-1"); it.Status != "doing" || it.Worktree != "feat" {
		t.Fatalf("task = %+v", it)
	}

	// Run from inside the worktree: the main checkout becomes the other tree.
	wcfg, _ := config.Load(wt, "")
	if o := Others(wcfg); len(o) != 1 || o[0].Branch != "main" {
		t.Fatalf("from the worktree: %+v", o)
	}
}

func TestOthersReadsUnmergedBranches(t *testing.T) {
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

func TestOthersOutsideGit(t *testing.T) {
	cfg, _ := config.Load(t.TempDir(), "")
	if Others(cfg) != nil {
		t.Fatal("no trees outside git")
	}
}

func TestWatchDirs(t *testing.T) {
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
