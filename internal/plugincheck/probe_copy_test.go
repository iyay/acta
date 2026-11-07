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
		name          string
		broken        []string // commands stubbed to fail
		tmpInRepo     bool     // TMPDIR points inside the repo
		gitFile       bool     // .git is a pointer file, as in a linked worktree
		noGit         bool     // a plain folder with no .git at all
		realGit       bool     // the repo is made with git init
		fromSub       bool     // the script runs from a subfolder
		tmpInOther    bool     // TMPDIR sits inside some other git repo
		plantFallback bool     // a folder already waits at the fallback name
		plantSymlink  bool     // the thing waiting there is a symlink instead
		gitInitFails  bool     // git works, but git init in the copy fails
		gitDirEnv     bool     // GIT_DIR is exported, as inside a git hook
		wantOK        bool
	}{
		{name: "copy works", wantOK: true},
		{name: "linked worktree", gitFile: true, wantOK: true},
		{name: "no .git", noGit: true, wantOK: true},
		{name: "mktemp fails, mkdir fallback", broken: []string{"mktemp"}, wantOK: true},
		{name: "mktemp and mkdir fail", broken: []string{"mktemp", "mkdir"}},
		{name: "cp fails", broken: []string{"cp"}},
		{name: "fallback lands in the repo", broken: []string{"mktemp"}, tmpInRepo: true},
		{name: "fallback folder already there", broken: []string{"mktemp"}, plantFallback: true},
		{name: "fallback symlink already there", broken: []string{"mktemp"}, plantFallback: true, plantSymlink: true},
		{name: "git init fails", realGit: true, gitInitFails: true},
		{name: "GIT_DIR exported", realGit: true, gitDirEnv: true, wantOK: true},
		{name: "subfolder with TMPDIR in the repo", realGit: true, fromSub: true, tmpInRepo: true},
		{name: "run from a subfolder", realGit: true, fromSub: true, wantOK: true},
		{name: "TMPDIR inside another repo", realGit: true, tmpInOther: true, wantOK: true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			repo := t.TempDir()
			writeFile(t, filepath.Join(repo, "tracked.txt"), "edited, not committed")
			writeFile(t, filepath.Join(repo, "untracked.txt"), "new")
			writeFile(t, filepath.Join(repo, ".env.example"), "KEY=")
			switch {
			case c.noGit:
			case c.realGit:
				gitInit(t, repo)
			case c.gitFile:
				writeFile(t, filepath.Join(repo, ".git"), "gitdir: /somewhere/.git/worktrees/x")
			default:
				writeFile(t, filepath.Join(repo, ".git", "HEAD"), "ref: refs/heads/main")
			}
			dir := repo
			if c.fromSub {
				dir = filepath.Join(repo, "sub")
				writeFile(t, filepath.Join(dir, "inner.txt"), "deep")
			}
			tmp := t.TempDir()
			switch {
			case c.tmpInRepo:
				tmp = filepath.Join(repo, "tmp")
			case c.tmpInOther:
				other := t.TempDir()
				gitInit(t, other)
				tmp = filepath.Join(other, "tmp")
			}
			if err := os.MkdirAll(tmp, 0o755); err != nil {
				t.Fatal(err)
			}
			stubs := t.TempDir()
			for _, name := range c.broken {
				writeFile(t, filepath.Join(stubs, name), "#!/bin/sh\nexit 1\n")
				if err := os.Chmod(filepath.Join(stubs, name), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			if c.gitInitFails {
				realGit, err := exec.LookPath("git")
				if err != nil {
					t.Fatal(err)
				}
				// Only git init fails, so the copy is made and the last step breaks.
				stub := "#!/bin/sh\nif [ \"$1\" = init ]; then exit 1; fi\nexec " + realGit + " \"$@\"\n"
				writeFile(t, filepath.Join(stubs, "git"), stub)
				if err := os.Chmod(filepath.Join(stubs, "git"), 0o755); err != nil {
					t.Fatal(err)
				}
			}
			// The symlink points at a folder outside the temp dir, so a copy
			// that followed it would land there.
			target := t.TempDir()
			cmd := exec.Command("bash", script)
			if c.plantFallback {
				// exec keeps the process id, so the thing made here has the
				// very name the script picks for its fallback.
				plant := `mkdir "$TMPDIR/acta-probe-$$"; exec bash "$0"`
				if c.plantSymlink {
					plant = `ln -s "$1" "$TMPDIR/acta-probe-$$"; exec bash "$0"`
				}
				cmd = exec.Command("bash", "-c", plant, script, target)
			}
			cmd.Dir = dir
			cmd.Env = append(os.Environ(), "TMPDIR="+tmp, "PATH="+stubs+string(os.PathListSeparator)+os.Getenv("PATH"),
				"GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
			if c.gitDirEnv {
				cmd.Env = append(cmd.Env, "GIT_DIR="+filepath.Join(repo, ".git"))
			}
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
				if c.plantFallback {
					planted, _ := filepath.Glob(filepath.Join(tmp, "acta-probe-*"))
					for _, p := range planted {
						if left, _ := os.ReadDir(p); len(left) > 0 {
							t.Fatalf("script copied into the planted folder %s", p)
						}
					}
					if left, _ := os.ReadDir(target); len(left) > 0 {
						t.Fatalf("script copied through the planted symlink")
					}
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
			want := []string{"tracked.txt", "untracked.txt", ".env.example"}
			if c.fromSub {
				want = append(want, filepath.Join("sub", "inner.txt"))
			}
			for _, f := range want {
				if _, err := os.Stat(filepath.Join(got, f)); err != nil {
					t.Errorf("copy is missing %s: %v", f, err)
				}
			}
			checkCopyGit(t, got, c.realGit)
		})
	}
}

// checkCopyGit makes sure git in the copy can only reach the copy. A copied
// .git file would point at the real repo, and no .git at all would let git
// walk up into whatever repo holds the temp folder.
func checkCopyGit(t *testing.T, copyDir string, fromGit bool) {
	t.Helper()
	st, err := os.Lstat(filepath.Join(copyDir, ".git"))
	if !fromGit {
		if err == nil {
			t.Errorf("a copy of a plain folder must not hold .git")
		}
		return
	}
	if err != nil || !st.IsDir() {
		t.Fatalf("copy of a git repo needs its own .git folder, got err=%v", err)
	}
	cmd := exec.Command("git", "rev-parse", "--show-toplevel")
	cmd.Dir = copyDir
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git in the copy failed: %v", err)
	}
	wantTop, _ := filepath.EvalSymlinks(copyDir)
	gotTop, _ := filepath.EvalSymlinks(strings.TrimSpace(string(out)))
	if gotTop != wantTop {
		t.Errorf("git in the copy reaches %q, want the copy %q", gotTop, wantTop)
	}
}

func gitInit(t *testing.T, dir string) {
	t.Helper()
	cmd := exec.Command("git", "init", "-q", dir)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_GLOBAL=/dev/null", "GIT_CONFIG_NOSYSTEM=1")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
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
