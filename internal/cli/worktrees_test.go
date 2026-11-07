package cli

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// strayRepo makes a repo with one linked worktree beside it, idle for two
// days, so the worktree rules read it as a clean stray.
func strayRepo(t *testing.T) (repo, stray string) {
	t.Helper()
	repo = debtRepo(t)
	stray = filepath.Join(filepath.Dir(repo), filepath.Base(repo)+"-stray")
	git := func(dir string, args ...string) string {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
		return strings.TrimSpace(string(out))
	}
	git(repo, "worktree", "add", "-q", "-b", "stray", stray)
	admin := git(stray, "rev-parse", "--absolute-git-dir")
	old := time.Now().Add(-48 * time.Hour)
	for _, f := range []string{"HEAD", "index", filepath.Join("logs", "HEAD")} {
		if err := os.Chtimes(filepath.Join(admin, f), old, old); err != nil {
			t.Fatal(err)
		}
	}
	return repo, stray
}

func TestWorktreesListsThenCleans(t *testing.T) {
	repo, stray := strayRepo(t)
	inDir(t, repo, func() {
		out := mustRun(t, "worktrees", "--json")
		var e struct{ Path, Branch, State string }
		if err := json.Unmarshal([]byte(strings.TrimSpace(out)), &e); err != nil {
			t.Fatalf("json %q: %v", out, err)
		}
		if e.Branch != "stray" || e.State != "clean" {
			t.Fatalf("listed %+v", e)
		}
		if _, err := os.Stat(stray); err != nil {
			t.Fatalf("a listing removed the worktree: %v", err)
		}
		out = mustRun(t, "worktrees", "--clean")
		if !strings.Contains(out, "removed") || !strings.Contains(out, "stray") {
			t.Fatalf("clean output %q", out)
		}
	})
	if _, err := os.Stat(stray); !os.IsNotExist(err) {
		t.Fatalf("stray still on disk: %v", err)
	}
}

func TestWorktreesOutsideGitFails(t *testing.T) {
	var stdout, stderr strings.Builder
	inDir(t, t.TempDir(), func() {
		if code := Run([]string{"worktrees"}, strings.NewReader(""), false, &stdout, &stderr); code == exitOK {
			t.Fatal("want a non-zero exit outside git")
		}
	})
	if stderr.Len() == 0 {
		t.Fatal("want the error on stderr")
	}
}

func TestWorktreesBadFlagPrintsUsage(t *testing.T) {
	var stdout, stderr strings.Builder
	if code := Run([]string{"worktrees", "--nope"}, strings.NewReader(""), false, &stdout, &stderr); code != exitBadInput {
		t.Fatalf("exit %d", code)
	}
	if !strings.Contains(stderr.String(), "acta worktrees") {
		t.Fatalf("stderr %q", stderr.String())
	}
	if !strings.Contains(mustRun(t, "help"), "worktrees") {
		t.Fatal("help does not list worktrees")
	}
}
