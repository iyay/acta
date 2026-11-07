package plugincheck

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestProbeCopy runs the debug skill's probe-copy.sh on a fake repo. A probe
// writes only after `D=$(bash probe-copy.sh) && cd "$D"`, so the script must
// print a path outside the repo when it works, and print nothing and fail
// when it cannot make the copy. Otherwise a probe could edit the repo itself.
func TestProbeCopy(t *testing.T) {
	script := filepath.Join(pluginRoot(t), "skills", "debug", "probe-copy.sh")
	cases := []struct {
		name      string
		broken    []string // commands stubbed to fail
		tmpInRepo bool     // TMPDIR points inside the repo
		gitFile   bool     // .git is a pointer file, as in a linked worktree
		noGit     bool     // a plain folder with no .git at all
		wantOK    bool
	}{
		{name: "copy works", wantOK: true},
		{name: "linked worktree", gitFile: true, wantOK: true},
		{name: "no .git", noGit: true, wantOK: true},
		{name: "mktemp fails, mkdir fallback", broken: []string{"mktemp"}, wantOK: true},
		{name: "mktemp and mkdir fail", broken: []string{"mktemp", "mkdir"}},
		{name: "cp fails", broken: []string{"cp"}},
		{name: "fallback lands in the repo", broken: []string{"mktemp"}, tmpInRepo: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := t.TempDir()
			writeFile(t, filepath.Join(repo, "tracked.txt"), "edited, not committed")
			writeFile(t, filepath.Join(repo, "untracked.txt"), "new")
			writeFile(t, filepath.Join(repo, ".env.example"), "KEY=")
			switch {
			case c.noGit:
			case c.gitFile:
				writeFile(t, filepath.Join(repo, ".git"), "gitdir: /somewhere/.git/worktrees/x")
			default:
				writeFile(t, filepath.Join(repo, ".git", "HEAD"), "ref: refs/heads/main")
			}
			tmp := t.TempDir()
			if c.tmpInRepo {
				tmp = filepath.Join(repo, "tmp")
				if err := os.MkdirAll(tmp, 0o755); err != nil {
					t.Fatal(err)
				}
			}
			stubs := t.TempDir()
			for _, name := range c.broken {
				writeFile(t, filepath.Join(stubs, name), "#!/bin/sh\nexit 1\n")
				if err := os.Chmod(filepath.Join(stubs, name), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			cmd := exec.Command("bash", script)
			cmd.Dir = repo
			cmd.Env = append(os.Environ(), "TMPDIR="+tmp, "PATH="+stubs+string(os.PathListSeparator)+os.Getenv("PATH"))
			out, err := cmd.Output()
			got := strings.TrimSpace(string(out))
			if !c.wantOK {
				if err == nil || got != "" {
					t.Fatalf("want a failed exit and no output, got err=%v out=%q", err, got)
				}
				// Failing is not enough: the script must not have made a
				// folder inside the repo on its way out.
				if left, _ := os.ReadDir(tmp); c.tmpInRepo && len(left) > 0 {
					t.Fatalf("script left %d entries inside the repo", len(left))
				}
				return
			}
			if err != nil {
				t.Fatalf("script failed: %v", err)
			}
			realRepo, _ := filepath.EvalSymlinks(repo)
			if got == "" || strings.HasPrefix(got+"/", realRepo+"/") {
				t.Fatalf("copy path %q must be outside the repo %q", got, realRepo)
			}
			for _, f := range []string{"tracked.txt", "untracked.txt", ".env.example"} {
				if _, err := os.Stat(filepath.Join(got, f)); err != nil {
					t.Errorf("copy is missing %s: %v", f, err)
				}
			}
			// A copied .git would let git in the copy change the real repo:
			// a worktree's .git file points at the real repo's git folder.
			if _, err := os.Lstat(filepath.Join(got, ".git")); err == nil {
				t.Errorf("copy must not hold .git")
			}
		})
	}
}

func writeFile(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}
