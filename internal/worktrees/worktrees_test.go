package worktrees

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/testguard"
)

func TestMain(m *testing.M) {
	testguard.Watch()
	os.Exit(m.Run())
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
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

// newRepo makes a repo on main with one commit. The author lives in the repo
// config, so tests can run side by side.
func newRepo(t *testing.T) (base, repo string) {
	t.Helper()
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	repo = filepath.Join(base, "repo")
	git(t, base, "init", "-q", "-b", "main", repo)
	git(t, repo, "config", "user.name", "test")
	git(t, repo, "config", "user.email", "test@example.com")
	write(t, filepath.Join(repo, ".gitignore"), "node_modules/\n")
	write(t, filepath.Join(repo, "a.txt"), "a\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-q", "-m", "init")
	return base, repo
}

// age moves the times of a worktree's git files two days back, so it reads
// as idle.
func age(t *testing.T, wt string) {
	t.Helper()
	admin := git(t, wt, "rev-parse", "--absolute-git-dir")
	old := time.Now().Add(-48 * time.Hour)
	for _, f := range []string{"HEAD", "index", filepath.Join("logs", "HEAD")} {
		p := filepath.Join(admin, f)
		if _, err := os.Stat(p); err == nil {
			if err := os.Chtimes(p, old, old); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func byPath(t *testing.T, es []Entry, path string) Entry {
	t.Helper()
	for _, e := range es {
		if e.Path == path {
			return e
		}
	}
	t.Fatalf("no entry for %s in %+v", path, es)
	return Entry{}
}

func TestListStates(t *testing.T) {
	t.Parallel()
	base, repo := newRepo(t)
	wt := func(name string, args ...string) string {
		p := filepath.Join(base, name)
		git(t, repo, append([]string{"worktree", "add", "-q", p}, args...)...)
		return p
	}

	fresh := wt("fresh-plan", "-b", "fresh-plan")
	unmerged := wt("unmerged", "-b", "unmerged")
	write(t, filepath.Join(unmerged, "b.txt"), "b\n")
	git(t, unmerged, "add", ".")
	git(t, unmerged, "commit", "-q", "-m", "work")
	ticked := wt("ticked-plan", "-b", "ticked-plan")
	write(t, filepath.Join(ticked, "c.txt"), "c\n")
	git(t, ticked, "add", ".")
	git(t, ticked, "commit", "-q", "-m", "all boxes ticked")
	dropped := wt("dropped-plan", "-b", "dropped-plan")
	locked := wt("locked", "-b", "locked")
	git(t, repo, "worktree", "lock", locked)
	active := wt("active", "-b", "active")
	clean := wt("clean", "-b", "clean")
	dirty := wt("dirty", "-b", "dirty")
	for i := 0; i < 12; i++ {
		write(t, filepath.Join(dirty, "junk-"+string(rune('a'+i))+".txt"), "x\n")
	}
	ignored := wt("ignored", "-b", "ignored")
	write(t, filepath.Join(ignored, "node_modules", "x.js"), "x\n")
	detached := wt("detached", "--detach")
	harness := filepath.Join(repo, ".claude", "worktrees", "agent-1")
	git(t, repo, "worktree", "add", "-q", "-b", "agent-1", harness)
	missing := wt("missing", "-b", "missing")

	for _, p := range []string{fresh, unmerged, ticked, dropped, locked, clean, dirty, ignored, detached, harness, missing} {
		age(t, p)
	}
	if err := os.RemoveAll(missing); err != nil {
		t.Fatal(err)
	}

	// Keep the index times, to prove the listing does not touch them.
	before := map[string]time.Time{}
	for _, p := range []string{clean, dirty, ignored} {
		st, err := os.Stat(filepath.Join(git(t, p, "rev-parse", "--absolute-git-dir"), "index"))
		if err != nil {
			t.Fatal(err)
		}
		before[p] = st.ModTime()
	}

	plans := []Plan{
		{Slug: "fresh-plan", Status: "approved"},
		{Slug: "ticked-plan", Status: "done"},
		{Slug: "dropped-plan", Status: "dropped"},
	}
	got, err := List(repo, plans, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	for _, e := range got {
		if e.Path == repo {
			t.Fatalf("main checkout listed: %+v", e)
		}
	}
	cases := []struct {
		path, state, reason string
	}{
		{fresh, "protected", "open plan"},
		{unmerged, "protected", "unmerged"},
		{ticked, "protected", "unmerged"},
		{dropped, "clean", ""},
		{locked, "protected", "locked"},
		{active, "protected", "24 hours"},
		{clean, "clean", ""},
		{dirty, "dirty", ""},
		{ignored, "clean", ""},
		{detached, "clean", ""},
		{harness, "clean", ""},
		{missing, "missing", ""},
	}
	for _, c := range cases {
		e := byPath(t, got, c.path)
		if e.State != c.state || !strings.Contains(e.Reason, c.reason) {
			t.Errorf("%s: got state %q reason %q, want %q with %q", filepath.Base(c.path), e.State, e.Reason, c.state, c.reason)
		}
	}
	if e := byPath(t, got, detached); e.Branch != "" {
		t.Errorf("detached branch = %q, want empty", e.Branch)
	}
	if e := byPath(t, got, clean); e.Branch != "clean" {
		t.Errorf("clean branch = %q", e.Branch)
	}
	d := byPath(t, got, dirty)
	if len(d.Status) != 11 || d.Status[10] != "... and 2 more" {
		t.Errorf("dirty status = %q", d.Status)
	}

	for p, t0 := range before {
		st, err := os.Stat(filepath.Join(git(t, p, "rev-parse", "--absolute-git-dir"), "index"))
		if err != nil {
			t.Fatal(err)
		}
		if !st.ModTime().Equal(t0) {
			t.Errorf("%s: listing changed the index time", filepath.Base(p))
		}
	}
}

func TestListSkipsCallingWorktree(t *testing.T) {
	t.Parallel()
	base, repo := newRepo(t)
	here := filepath.Join(base, "here")
	git(t, repo, "worktree", "add", "-q", "-b", "here", here)
	age(t, here)
	got, err := List(here, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 0 {
		t.Fatalf("got %+v, want nothing", got)
	}
}

func TestListOutsideGit(t *testing.T) {
	t.Parallel()
	if _, err := List(t.TempDir(), nil, time.Now()); err == nil {
		t.Fatal("want an error outside a git repo")
	}
}
