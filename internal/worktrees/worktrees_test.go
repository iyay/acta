package worktrees

import (
	"encoding/json"
	"errors"
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

func TestClean(t *testing.T) {
	// Not parallel: it swaps runGit to see every git call.
	base, repo := newRepo(t)
	wt := func(name string, args ...string) string {
		p := filepath.Join(base, name)
		git(t, repo, append([]string{"worktree", "add", "-q", p}, args...)...)
		return p
	}
	gone := wt("gone", "-b", "gone")
	refused := wt("refused", "-b", "refused")
	detached := wt("detached", "--detach")
	dirty := wt("dirty", "-b", "dirty")
	write(t, filepath.Join(dirty, "junk.txt"), "x\n")
	missing := wt("missing", "-b", "missing")
	stuck := wt("stuck", "-b", "stuck")
	protected := wt("protected", "-b", "protected")
	for _, p := range []string{gone, refused, detached, dirty, missing, stuck} {
		age(t, p)
	}
	if err := os.RemoveAll(missing); err != nil {
		t.Fatal(err)
	}

	var calls [][]string
	real := runGit
	t.Cleanup(func() { runGit = real })
	runGit = func(dir string, args ...string) (string, error) {
		calls = append(calls, args)
		// git branch -d refuses a branch it thinks is not merged. Fake that
		// for one branch, so the refused path runs.
		if len(args) == 4 && args[0] == "branch" && args[3] == "refused" {
			return "", errors.New("error: the branch 'refused' is not fully merged")
		}
		// Fake a remove that git refuses, as on a folder it cannot delete.
		if len(args) == 3 && args[0] == "worktree" && args[1] == "remove" && same(args[2], stuck) {
			return "", errors.New("fatal: cannot remove stuck")
		}
		return real(dir, args...)
	}

	// A listing right before the clean must not stop it.
	if _, err := List(repo, nil, time.Now()); err != nil {
		t.Fatal(err)
	}
	got, err := Clean(repo, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}

	cases := []struct {
		path, state, reason string
		kept                bool
	}{
		{gone, "removed", "", false},
		{refused, "removed", "not fully merged", true},
		{detached, "removed", "", false},
		{missing, "removed", "", false},
		{stuck, "kept", "cannot remove stuck", false},
		{dirty, "dirty", "", false},
		{protected, "protected", "24 hours", false},
	}
	for _, c := range cases {
		e := byPath(t, got, c.path)
		if e.State != c.state || e.BranchKept != c.kept || !strings.Contains(e.Reason, c.reason) {
			t.Errorf("%s: got %+v, want state %q kept %v reason %q", filepath.Base(c.path), e, c.state, c.kept, c.reason)
		}
	}
	for _, p := range []string{gone, refused, detached, missing} {
		if strings.Contains(git(t, repo, "worktree", "list"), p+" ") {
			t.Errorf("%s still listed", filepath.Base(p))
		}
	}
	for _, p := range []string{dirty, protected, stuck} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("%s was touched: %v", filepath.Base(p), err)
		}
	}
	branches := git(t, repo, "branch", "--list")
	if strings.Contains(branches, "gone") || !strings.Contains(branches, "refused") {
		t.Errorf("branches after clean:\n%s", branches)
	}
	for _, args := range calls {
		for _, a := range args {
			if a == "--force" || a == "-f" || a == "-D" || a == "-ff" {
				t.Errorf("git called with %s: %v", a, args)
			}
		}
		if len(args) > 0 && args[0] == "branch" && (len(args) != 4 || args[1] != "-d" || args[2] != "--") {
			t.Errorf("unexpected branch call %v", args)
		}
	}
	for _, args := range calls {
		if len(args) > 3 && args[0] == "branch" && args[3] == "" {
			t.Errorf("branch call for a detached worktree: %v", args)
		}
	}
}

// TestListUnknownWorkIsProtected covers the worktrees whose work or age the
// rules cannot be sure of. Each must stay protected.
func TestListUnknownWorkIsProtected(t *testing.T) {
	t.Parallel()
	base, repo := newRepo(t)

	// A detached worktree with a commit that no branch holds.
	own := filepath.Join(base, "own")
	git(t, repo, "worktree", "add", "-q", "--detach", own)
	write(t, filepath.Join(own, "d.txt"), "d\n")
	git(t, own, "add", ".")
	git(t, own, "commit", "-q", "-m", "only here")

	// A fresh worktree whose admin gitdir file holds a relative path.
	rel := filepath.Join(base, "rel")
	git(t, repo, "-c", "worktree.useRelativePaths=true", "worktree", "add", "-q", "-b", "rel", rel)

	// A locked worktree whose folder is gone, as on an unmounted drive.
	away := filepath.Join(base, "away")
	git(t, repo, "worktree", "add", "-q", "-b", "away", away)
	git(t, repo, "worktree", "lock", away)
	if err := os.RemoveAll(away); err != nil {
		t.Fatal(err)
	}

	// A branch ref deleted under its worktree, so rev-list cannot read it.
	gone := filepath.Join(base, "gone-ref")
	git(t, repo, "worktree", "add", "-q", "-b", "gone-ref", gone)
	age(t, gone)
	git(t, repo, "update-ref", "-d", "refs/heads/gone-ref")

	for _, p := range []string{own} {
		age(t, p)
	}
	got, err := List(repo, nil, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ path, reason string }{
		{own, "commits no branch"},
		{rel, "24 hours"},
		{away, "locked"},
		{gone, "cannot"},
	}
	for _, c := range cases {
		e := byPath(t, got, c.path)
		if e.State != StateProtected || !strings.Contains(e.Reason, c.reason) {
			t.Errorf("%s: got state %q reason %q, want protected with %q", filepath.Base(c.path), e.State, e.Reason, c.reason)
		}
	}
}

func TestEntryJSONHasEveryField(t *testing.T) {
	t.Parallel()
	raw, err := json.Marshal(Entry{Path: "/x", State: StateClean})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{"path", "branch", "state", "reason", "branch_kept", "status"} {
		if !strings.Contains(string(raw), `"`+f+`":`) {
			t.Errorf("%s missing from %s", f, raw)
		}
	}
}

// A worktree whose admin folder cannot be found has no activity time. The
// rules must keep it, since it may be in use right now.
func TestJudgeKeepsUnknownAge(t *testing.T) {
	t.Parallel()
	e := judge(t.TempDir(), porcelain{path: t.TempDir(), branch: "x"}, "", nil, time.Now())
	if e.State != StateProtected || !strings.Contains(e.Reason, "cannot read activity") {
		t.Fatalf("got %+v", e)
	}
}
