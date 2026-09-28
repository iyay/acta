package hook

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestEnsureGitignoreAddsTheLineOnce(t *testing.T) {
	root := gitRoot(t)
	for range 3 {
		if err := EnsureGitignore(root, ".agents.json"); err != nil {
			t.Fatal(err)
		}
	}
	got := readFile(t, filepath.Join(root, ".gitignore"))
	if got != ".agents.json\n" {
		t.Fatalf("gitignore = %q, want %q", got, ".agents.json\n")
	}
}

func TestEnsureGitignoreKeepsExistingLines(t *testing.T) {
	root := gitRoot(t)
	if err := os.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.tmp\n# notes\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	got := readFile(t, filepath.Join(root, ".gitignore"))
	for _, want := range []string{"*.tmp", "# notes", ".agents.json"} {
		if !strings.Contains(got, want) {
			t.Errorf("gitignore %q lost %q", got, want)
		}
	}
	if n := strings.Count(got, ".agents.json"); n != 1 {
		t.Errorf("gitignore holds the line %d times: %q", n, got)
	}
}

func TestEnsureGitignoreWithoutARootFolderWritesNothing(t *testing.T) {
	root := filepath.Join(gitRoot(t), "gone")
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf("a missing root folder must stay missing: %v", err)
	}
}

func TestEnsureGitignoreOutsideGitWritesNothing(t *testing.T) {
	// A folder with no .git anywhere above it, so nothing gets written.
	root := filepath.Join(t.TempDir(), "planning")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if inGit, _ := repoTop(root); inGit {
		t.Skipf("%s sits inside a git repo", root)
	}
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(root, ".gitignore")); !os.IsNotExist(err) {
		t.Errorf("outside git nothing may be written: %v", err)
	}
}

// gitRoot makes a git repo with a planning root folder in it.
func gitRoot(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if out, err := exec.Command("git", "-C", dir, "init", "-q", "-b", "main").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	root := filepath.Join(dir, ".pm")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	return root
}

func repoTop(dir string) (bool, string) {
	out, err := exec.Command("git", "-C", dir, "rev-parse", "--show-toplevel").Output()
	return err == nil, strings.TrimSpace(string(out))
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func TestEnsureGitignoreRefusesALinkedGitignore(t *testing.T) {
	cases := map[string]func(t *testing.T, root, outside string) string{
		"dangling target outside": func(t *testing.T, root, outside string) string {
			return filepath.Join(outside, "new.conf")
		},
		"existing target outside": func(t *testing.T, root, outside string) string {
			p := filepath.Join(outside, "victim.conf")
			if err := os.WriteFile(p, []byte("precious=1\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return p
		},
		"target inside the repo": func(t *testing.T, root, outside string) string {
			p := filepath.Join(root, "notes.txt")
			if err := os.WriteFile(p, []byte("keep me\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			return p
		},
	}
	for name, target := range cases {
		t.Run(name, func(t *testing.T) {
			root := gitRoot(t)
			outside := t.TempDir()
			dst := target(t, root, outside)
			before, beforeErr := os.ReadFile(dst)
			if err := os.Symlink(dst, filepath.Join(root, ".gitignore")); err != nil {
				t.Fatal(err)
			}
			err := EnsureGitignore(root, ".agents.json")
			if err == nil {
				t.Fatal("want an error for a linked .gitignore, got nil")
			}
			// The message is the only thing the user ever sees here, so it has
			// to name the file and say it is a link.
			if !strings.Contains(err.Error(), "is a link") || !strings.Contains(err.Error(), filepath.Join(root, ".gitignore")) {
				t.Fatalf("error does not name the linked path: %v", err)
			}
			after, afterErr := os.ReadFile(dst)
			if os.IsNotExist(beforeErr) != os.IsNotExist(afterErr) || string(before) != string(after) {
				t.Fatalf("link target changed: before %q (%v), after %q (%v)", before, beforeErr, after, afterErr)
			}
		})
	}
}

func TestEnsureGitignoreRefusesARootLinkedOutTheRepo(t *testing.T) {
	cases := map[string]func(t *testing.T) string{
		"plain folder outside":   func(t *testing.T) string { return t.TempDir() },
		"folder in another repo": func(t *testing.T) string { return gitRoot(t) },
	}
	for name, outside := range cases {
		t.Run(name, func(t *testing.T) {
			repo := gitRoot(t)
			out := outside(t)
			root := filepath.Join(repo, ".acta")
			if err := os.Symlink(out, root); err != nil {
				t.Fatal(err)
			}
			// Both checks report, so one run shows the two ways the write
			// escapes: no error, and a file created outside the repo.
			if err := EnsureGitignore(root, ".agents.json"); err == nil {
				t.Error("want an error for a root linked out of the repo, got nil")
			}
			if _, err := os.Stat(filepath.Join(out, ".gitignore")); !os.IsNotExist(err) {
				t.Errorf("wrote a .gitignore outside the repo (stat err %v)", err)
			}
		})
	}
}

func TestEnsureGitignoreFollowsARootLinkedInsideTheRepo(t *testing.T) {
	repo := gitRoot(t)
	real := filepath.Join(repo, "docs", "planning")
	if err := os.MkdirAll(real, 0o755); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(repo, ".acta")
	if err := os.Symlink(real, root); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(real, ".gitignore")); got != ".agents.json\n" {
		t.Fatalf("gitignore = %q, want %q", got, ".agents.json\n")
	}
}

func TestEnsureGitignoreWorksThroughALinkedParentFolder(t *testing.T) {
	folder := gitRoot(t)
	// gitRoot hands back the root folder, but the link has to point at the
	// folder that holds .git, or the walk above it finds no repo at all.
	inGit, top := repoTop(folder)
	if !inGit {
		t.Fatalf("no git repo above %s", folder)
	}
	link := filepath.Join(t.TempDir(), "via")
	if err := os.Symlink(top, link); err != nil {
		t.Fatal(err)
	}
	root := filepath.Join(link, ".pm")
	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := EnsureGitignore(root, ".agents.json"); err != nil {
		t.Fatal(err)
	}
	if got := readFile(t, filepath.Join(folder, ".gitignore")); got != ".agents.json\n" {
		t.Fatalf("gitignore = %q, want %q", got, ".agents.json\n")
	}
}
