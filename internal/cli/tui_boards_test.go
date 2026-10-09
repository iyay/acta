package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/iyay/acta/internal/config"
)

func runGit(t *testing.T, dir string, args ...string) {
	t.Helper()
	if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
}

func writeFixture(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The first frame must not wait for git, so its board holds only the main
// tree and no author. Every later load must see the other worktrees and the
// authors, or the board would stay half empty.
func TestTUIBoardsFirstFrameIsMainOnly(t *testing.T) {
	// An inherited root would point the config at another repo.
	t.Setenv("ACTA_ROOT", "")
	t.Setenv("PM_ROOT", "")

	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo := filepath.Join(base, "repo")
	runGit(t, base, "init", "-q", "-b", "main", repo)
	// Set Ana in the env too: a name set in the caller's env beats the repo
	// config, so the config alone would break under such an env.
	t.Setenv("GIT_AUTHOR_NAME", "Ana")
	t.Setenv("GIT_COMMITTER_NAME", "Ana")
	t.Setenv("GIT_AUTHOR_EMAIL", "ana@example.com")
	t.Setenv("GIT_COMMITTER_EMAIL", "ana@example.com")
	// The repo config stays: it names the author of files never committed.
	runGit(t, repo, "config", "user.name", "Ana")
	runGit(t, repo, "config", "user.email", "ana@example.com")
	writeFixture(t, filepath.Join(repo, ".acta/plans/2026-09-21-main.md"), "# Main plan\n\n### Task 1: One\n- [ ] a\n")
	runGit(t, repo, "add", ".")
	runGit(t, repo, "commit", "-q", "-m", "init")
	wt := filepath.Join(base, "repo-feat")
	runGit(t, repo, "worktree", "add", "-q", wt, "-b", "feat")
	writeFixture(t, filepath.Join(wt, ".acta/plans/2026-09-22-feat.md"), "# Feat plan\n\n### Task 1: One\n- [ ] a\n")
	runGit(t, wt, "add", ".")
	runGit(t, wt, "commit", "-q", "-m", "feat plan")

	cfg, err := config.Load(repo, "")
	if err != nil {
		t.Fatal(err)
	}
	first, load, err := tuiBoards(cfg)
	if err != nil {
		t.Fatal(err)
	}

	const mainID, featID = "plans/2026-09-21-main", "plans/2026-09-22-feat"
	if first.Get(mainID) == nil {
		t.Fatal("first board lacks the main plan")
	}
	if first.Get(featID) != nil {
		t.Errorf("first board holds the worktree plan %s", featID)
	}
	if a := first.Get(mainID).Author; a != "" {
		t.Errorf("first board author = %q, want empty", a)
	}

	later, err := load()
	if err != nil {
		t.Fatal(err)
	}
	if later.Get(featID) == nil {
		t.Errorf("later board lacks the worktree plan %s", featID)
	}
	if it := later.Get(mainID); it == nil || it.Author != "Ana" {
		t.Errorf("later board main plan = %+v, want author Ana", it)
	}
}
