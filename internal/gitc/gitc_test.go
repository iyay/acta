package gitc

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func setupRepo(t *testing.T) string {
	t.Helper()
	for _, k := range []string{"GIT_AUTHOR_NAME", "GIT_COMMITTER_NAME"} {
		t.Setenv(k, "test")
	}
	for _, k := range []string{"GIT_AUTHOR_EMAIL", "GIT_COMMITTER_EMAIL"} {
		t.Setenv(k, "test@example.com")
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(dir, "a.md"), "a\n")
	writeFile(t, filepath.Join(dir, "b.md"), "b\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "init")
	return dir
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestCommitOnlyThatPath(t *testing.T) {
	repo := setupRepo(t)
	writeFile(t, filepath.Join(repo, "b.md"), "b staged\n")
	git(t, repo, "add", "b.md")
	a := filepath.Join(repo, "a.md")
	writeFile(t, a, "a changed\n")

	r := Commit(repo, a, "pm: a status done", false)
	if !r.Committed {
		t.Fatalf("not committed: %s", r.Reason)
	}
	if got := git(t, repo, "log", "-1", "--format=%s"); got != "pm: a status done" {
		t.Fatalf("last commit = %q", got)
	}
	if got := git(t, repo, "show", "--name-only", "--format=", "HEAD"); got != "a.md" {
		t.Fatalf("commit touched %q, want only a.md", got)
	}
	if got := git(t, repo, "diff", "--cached", "--name-only"); got != "b.md" {
		t.Fatalf("staged after commit = %q, want b.md still staged", got)
	}
}

func TestCommitNewFile(t *testing.T) {
	repo := setupRepo(t)
	p := filepath.Join(repo, "new.md")
	writeFile(t, p, "new\n")
	if r := Commit(repo, p, "pm: new bug", false); !r.Committed {
		t.Fatalf("not committed: %s", r.Reason)
	}
}

func TestCommitSkips(t *testing.T) {
	t.Run("file was dirty", func(t *testing.T) {
		repo := setupRepo(t)
		r := Commit(repo, filepath.Join(repo, "a.md"), "m", true)
		if r.Committed || !strings.Contains(r.Reason, "other uncommitted changes") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("mid-merge", func(t *testing.T) {
		repo := setupRepo(t)
		head := git(t, repo, "rev-parse", "HEAD")
		writeFile(t, filepath.Join(repo, ".git", "MERGE_HEAD"), head+"\n")
		writeFile(t, filepath.Join(repo, "a.md"), "x\n")
		r := Commit(repo, filepath.Join(repo, "a.md"), "m", false)
		if r.Committed || !strings.Contains(r.Reason, "merge") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("detached HEAD", func(t *testing.T) {
		repo := setupRepo(t)
		git(t, repo, "checkout", "-q", "--detach")
		writeFile(t, filepath.Join(repo, "a.md"), "x\n")
		r := Commit(repo, filepath.Join(repo, "a.md"), "m", false)
		if r.Committed || !strings.Contains(r.Reason, "detached") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("not a repo", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "a.md")
		writeFile(t, p, "x\n")
		r := Commit(dir, p, "m", false)
		if r.Committed || !strings.Contains(r.Reason, "not a git repo") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("hook fails, change stays", func(t *testing.T) {
		repo := setupRepo(t)
		hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
		writeFile(t, hook, "#!/bin/sh\necho no >&2\nexit 1\n")
		if err := os.Chmod(hook, 0o755); err != nil {
			t.Fatal(err)
		}
		a := filepath.Join(repo, "a.md")
		writeFile(t, a, "x\n")
		r := Commit(repo, a, "m", false)
		if r.Committed || !strings.Contains(r.Reason, "commit failed") {
			t.Fatalf("got %+v", r)
		}
		if b, _ := os.ReadFile(a); string(b) != "x\n" {
			t.Fatalf("change lost: %q", b)
		}
	})
}

func TestIsDirty(t *testing.T) {
	repo := setupRepo(t)
	a := filepath.Join(repo, "a.md")
	if d, err := IsDirty(repo, a); err != nil || d {
		t.Fatalf("clean file: dirty=%v err=%v", d, err)
	}
	writeFile(t, a, "changed\n")
	if d, _ := IsDirty(repo, a); !d {
		t.Fatal("changed file not dirty")
	}
	if d, _ := IsDirty(repo, filepath.Join(repo, "missing.md")); d {
		t.Fatal("missing file reads as dirty")
	}
}

func TestCommitPaths(t *testing.T) {
	dir := setupRepo(t)
	writeFile(t, filepath.Join(dir, "a.md"), "a changed\n")
	writeFile(t, filepath.Join(dir, "b.md"), "b changed\n")
	writeFile(t, filepath.Join(dir, "c.md"), "c new\n")
	res := CommitPaths(dir, []string{"a.md", "b.md"}, "two files")
	if !res.Committed {
		t.Fatalf("not committed: %s", res.Reason)
	}
	if names := git(t, dir, "show", "--name-only", "--format=", "HEAD"); names != "a.md\nb.md" {
		t.Fatalf("commit holds %q", names)
	}
	if out := git(t, dir, "status", "--porcelain"); out != "?? c.md" {
		t.Fatalf("third file not left out: %q", out)
	}
}

// FirstSeen answers with the place of a file in the branch history, so two
// files keep a stable order even when their commits share one second.
func TestFirstSeen(t *testing.T) {
	dir := setupRepo(t)
	first, err := FirstSeen(dir, "a.md")
	if err != nil || first != 0 {
		t.Fatalf("first commit index = %d, %v", first, err)
	}
	writeFile(t, filepath.Join(dir, "a.md"), "a again\n")
	git(t, dir, "add", "a.md")
	git(t, dir, "commit", "-q", "-m", "second")
	again, err := FirstSeen(dir, "a.md")
	if err != nil || again != first {
		t.Fatalf("after second commit = %d, want first %d, err %v", again, first, err)
	}
	t.Setenv("GIT_AUTHOR_DATE", "2020-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2020-01-01T00:00:00Z")
	writeFile(t, filepath.Join(dir, "c.md"), "c one second too\n")
	git(t, dir, "add", "c.md")
	git(t, dir, "commit", "-q", "-m", "same second as the commit before")
	same, err := FirstSeen(dir, "c.md")
	if err != nil || same <= again {
		t.Fatalf("later file index = %d, want past %d, err %v", same, again, err)
	}
	if untracked, err := FirstSeen(dir, "untracked.md"); err != nil || untracked <= same {
		t.Fatalf("untracked = %d, want past %d, err %v", untracked, same, err)
	}
}

// A file that only arrives through a merge is seen at the merge, so the file
// that was on main keeps the lower place.
func TestFirstSeenCountsTheMerge(t *testing.T) {
	dir := setupRepo(t)
	git(t, dir, "checkout", "-q", "-b", "side")
	writeFile(t, filepath.Join(dir, "side.md"), "side\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "side adds its file")
	git(t, dir, "checkout", "-q", "main")
	writeFile(t, filepath.Join(dir, "main.md"), "main\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "main adds its file")
	git(t, dir, "merge", "-q", "--no-ff", "-m", "merge side", "side")
	main, err := FirstSeen(dir, "main.md")
	if err != nil {
		t.Fatal(err)
	}
	side, err := FirstSeen(dir, "side.md")
	if err != nil {
		t.Fatal(err)
	}
	if main >= side {
		t.Fatalf("main file at %d, side file at %d, want main first", main, side)
	}
}

// The author of a file is the person whose commit first added it, so a later
// edit by someone else never changes who wrote the file.
func TestAuthorIsTheFirstCommit(t *testing.T) {
	repo := firstCommitRepo(t, "Ana")
	t.Setenv("GIT_AUTHOR_NAME", "Budi")
	t.Setenv("GIT_COMMITTER_NAME", "Budi")
	writeFile(t, filepath.Join(repo, "a.md"), "a again\n")
	git(t, repo, "commit", "-qam", "second")

	if name, ok := Author(repo, filepath.Join(repo, "a.md")); !ok || name != "Ana" {
		t.Fatalf("Author = %q, %v, want Ana, true", name, ok)
	}
	if name, ok := Author(repo, filepath.Join(repo, "new.md")); ok || name != "" {
		t.Errorf("a file git never saw gave %q, %v, want empty, false", name, ok)
	}
	// A repo with a .git but no commit has no author for anything either.
	empty := t.TempDir()
	git(t, empty, "init", "-q", "-b", "main")
	if name, ok := Author(empty, filepath.Join(empty, "a.md")); ok || name != "" {
		t.Errorf("a repo with no commit gave %q, %v, want empty, false", name, ok)
	}
	// A folder outside any checkout answers at once, without running git.
	if name, ok := Author(t.TempDir(), filepath.Join(repo, "a.md")); ok || name != "" {
		t.Errorf("a folder with no repo gave %q, %v, want empty, false", name, ok)
	}
}

// UserName is the name this checkout commits under, for a file git has no
// commit of yet. It says empty when git has no name to give.
func TestUserName(t *testing.T) {
	repo := firstCommitRepo(t, "Ana")
	git(t, repo, "config", "user.name", "Sari")
	if got := UserName(repo); got != "Sari" {
		t.Errorf("UserName = %q, want Sari", got)
	}
	if got := UserName(t.TempDir()); got != "" {
		t.Errorf("a folder with no repo named %q, want empty", got)
	}
	// Whatever the machine keeps in its own config, a checkout that names
	// nobody itself answers empty.
	t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
	t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
	bare := t.TempDir()
	git(t, bare, "init", "-q", "-b", "main")
	if got := UserName(bare); got != "" {
		t.Errorf("a checkout with no user.name gave %q, want empty", got)
	}
}

// firstCommitRepo is a checkout whose one commit was written by name.
func firstCommitRepo(t *testing.T, name string) string {
	t.Helper()
	for k, v := range map[string]string{
		"GIT_AUTHOR_NAME": name, "GIT_COMMITTER_NAME": name,
		"GIT_AUTHOR_EMAIL": name + "@example.com", "GIT_COMMITTER_EMAIL": name + "@example.com",
	} {
		t.Setenv(k, v)
	}
	dir, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	git(t, dir, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(dir, "a.md"), "a\n")
	git(t, dir, "add", ".")
	git(t, dir, "commit", "-q", "-m", "first")
	return dir
}
