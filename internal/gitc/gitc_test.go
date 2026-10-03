package gitc

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
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

// Authors asks git once for many files. The author of a file is the person
// whose commit first added it, so a later edit by someone else never changes
// who wrote it. A file moved inside the folder counts as added by whoever
// moved it, which is what a one-file log said too.
func TestAuthorsAsksOnceForManyFiles(t *testing.T) {
	repo := firstCommitRepo(t, "Ana")
	// A checkout that spots renames must still answer the same way, so the
	// flag the test would need is on for every case below.
	git(t, repo, "config", "diff.renames", "true")
	writeFile(t, filepath.Join(repo, "old.md"), "old\n")
	git(t, repo, "add", "old.md")
	git(t, repo, "commit", "-qm", "ana adds old")
	t.Setenv("GIT_AUTHOR_NAME", "Budi")
	t.Setenv("GIT_COMMITTER_NAME", "Budi")
	writeFile(t, filepath.Join(repo, "a.md"), "a again\n")
	writeFile(t, filepath.Join(repo, "catatan-é.md"), "c\n")
	git(t, repo, "mv", "old.md", "moved.md")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-qm", "budi edits, adds and moves")
	// A name that only ever shows up through a move: git must answer for it as
	// a file the mover added, not read the move as nothing at all.
	t.Setenv("GIT_AUTHOR_NAME", "Citra")
	t.Setenv("GIT_COMMITTER_NAME", "Citra")
	git(t, repo, "mv", "moved.md", "pindah.md")
	git(t, repo, "commit", "-qm", "citra moves it to a name nobody used")
	// A file dropped and written again later still answers for whoever put it
	// there the first time, so the newest name on the file never wins.
	git(t, repo, "rm", "-q", "catatan-é.md")
	git(t, repo, "commit", "-qm", "citra drops a file")
	writeFile(t, filepath.Join(repo, "catatan-é.md"), "written again\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-qm", "citra writes it again")

	at := func(name string) string { return filepath.Join(repo, name) }
	got := Authors(repo, []string{at("a.md"), at("catatan-é.md"), at("pindah.md"), at("new.md")})
	want := map[string]string{
		at("a.md"):         "Ana",
		at("catatan-é.md"): "Budi",
		at("pindah.md"):    "Citra",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Authors = %v, want %v (new.md was never committed, so it has no entry)", got, want)
	}

	// The board asks about a folder, not the checkout root, so the keys have
	// to come back the way they went in.
	folder := filepath.Join(repo, "plans")
	if err := os.MkdirAll(folder, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("GIT_AUTHOR_NAME", "Budi")
	t.Setenv("GIT_COMMITTER_NAME", "Budi")
	inside := filepath.Join(folder, "rencana.md")
	writeFile(t, inside, "p\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-qm", "budi writes a plan")
	if got := Authors(folder, []string{inside}); !reflect.DeepEqual(got, map[string]string{inside: "Budi"}) {
		t.Fatalf("Authors on a folder inside the checkout = %v, want {%s: Budi}", got, inside)
	}
}

// Some asks have nothing to find. None of them may answer with names.
func TestAuthorsWithNothingToFind(t *testing.T) {
	repo := firstCommitRepo(t, "Ana")
	// No paths at all must not turn into a log of the whole repo.
	if got := Authors(repo, nil); len(got) != 0 {
		t.Errorf("no paths gave %v, want an empty map", got)
	}
	// A repo with a .git but no commit has no author for anything.
	empty := t.TempDir()
	git(t, empty, "init", "-q", "-b", "main")
	if got := Authors(empty, []string{filepath.Join(empty, "a.md")}); len(got) != 0 {
		t.Errorf("a repo with no commit gave %v, want an empty map", got)
	}
	// A folder outside any checkout answers at once, without running git: a
	// fake git that writes down every call must never be reached.
	spy := t.TempDir()
	marks := filepath.Join(spy, "called")
	fake := "#!/bin/sh\necho called >> " + marks + "\nexit 1\n"
	if err := os.WriteFile(filepath.Join(spy, "git"), []byte(fake), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", spy)
	if got := Authors(spy, []string{filepath.Join(repo, "a.md")}); len(got) != 0 {
		t.Errorf("a folder with no repo gave %v, want an empty map", got)
	}
	if _, err := os.Stat(marks); !os.IsNotExist(err) {
		t.Errorf("git ran for a folder outside any checkout")
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

// repoAt is setupRepo with its first commit made at when, so a test can say
// how old a commit is.
func repoAt(t *testing.T, when string) string {
	t.Helper()
	t.Setenv("GIT_AUTHOR_DATE", when)
	t.Setenv("GIT_COMMITTER_DATE", when)
	return setupRepo(t)
}

// commitAllAt commits every change in the repo with both git dates at when.
func commitAllAt(t *testing.T, repo, when, msg string) {
	t.Helper()
	t.Setenv("GIT_AUTHOR_DATE", when)
	t.Setenv("GIT_COMMITTER_DATE", when)
	git(t, repo, "add", "-A")
	git(t, repo, "commit", "-q", "--allow-empty", "-m", msg)
}

// LastChange is the commit time of the newest commit that touched one of the
// paths. A path is a plain name: no glob and no git path magic, since a page
// lists names and the wiki matches them by prefix.
func TestLastChange(t *testing.T) {
	repo := repoAt(t, "2020-01-01T00:00:00Z") // a.md and b.md
	if err := os.MkdirAll(filepath.Join(repo, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "dir", "c.md"), "c\n")
	commitAllAt(t, repo, "2021-01-01T00:00:00Z", "add dir/c.md")
	writeFile(t, filepath.Join(repo, "a.md"), "a again\n")
	commitAllAt(t, repo, "2022-01-01T00:00:00Z", "edit a.md")
	writeFile(t, filepath.Join(repo, "dir", "c.md"), "c again\n")
	// Seven hours ahead of UTC: the same moment as midnight UTC.
	commitAllAt(t, repo, "2023-01-01T07:00:00+07:00", "edit dir/c.md")

	year := func(y int) time.Time { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC) }
	tests := []struct {
		name  string
		paths []string
		want  time.Time
	}{
		{"a file only the first commit touched", []string{"b.md"}, year(2020)},
		{"a file edited later", []string{"a.md"}, year(2022)},
		{"the newest of several paths", []string{"b.md", "a.md"}, year(2022)},
		{"a folder is the files in it", []string{"dir/"}, year(2023)},
		{"a path no commit touched", []string{"nothing.md"}, time.Time{}},
		{"no paths is not the whole repo", nil, time.Time{}},
		{"a glob is a name, not a pattern", []string{"*.md"}, time.Time{}},
		{"git path magic is a name, not magic", []string{":(top)a.md"}, time.Time{}},
	}
	for _, tt := range tests {
		got, err := LastChange(repo, tt.paths)
		if err != nil || !got.Equal(tt.want) {
			t.Errorf("%s: LastChange(%v) = %v, %v, want %v", tt.name, tt.paths, got, err, tt.want)
		}
	}
	if got, err := LastChange(t.TempDir(), []string{"a.md"}); err == nil {
		t.Errorf("a folder that is no checkout gave %v and no error", got)
	}
}

// A clean merge brings no change of its own. Landing a branch must not turn
// every page that covers its files into a stale one, so the answer is the commit
// that made the change and not the merge that carried it.
func TestLastChangeSeesPastACleanMerge(t *testing.T) {
	repo := repoAt(t, "2020-01-01T00:00:00Z") // a.md and b.md
	git(t, repo, "checkout", "-q", "-b", "side")
	writeFile(t, filepath.Join(repo, "a.md"), "side edits a\n")
	commitAllAt(t, repo, "2021-01-01T00:00:00Z", "side edits a.md")
	git(t, repo, "checkout", "-q", "main")
	writeFile(t, filepath.Join(repo, "b.md"), "main edits b\n")
	commitAllAt(t, repo, "2022-01-01T00:00:00Z", "main edits b.md")
	t.Setenv("GIT_AUTHOR_DATE", "2023-01-01T00:00:00Z")
	t.Setenv("GIT_COMMITTER_DATE", "2023-01-01T00:00:00Z")
	git(t, repo, "merge", "-q", "--no-ff", "-m", "merge side", "side")

	year := func(y int) time.Time { return time.Date(y, 1, 1, 0, 0, 0, 0, time.UTC) }
	for path, want := range map[string]time.Time{"a.md": year(2021), "b.md": year(2022)} {
		if got, err := LastChange(repo, []string{path}); err != nil || !got.Equal(want) {
			t.Errorf("LastChange(%s) = %v, %v, want %v, the commit before the merge", path, got, err, want)
		}
	}
}

// Changed lists the files that differ across a range. A move counts as a file
// that left and a file that came, so a folder a file moved out of is changed.
func TestChanged(t *testing.T) {
	repo := repoAt(t, "2020-01-01T00:00:00Z") // a.md and b.md
	first := git(t, repo, "rev-parse", "HEAD")
	if err := os.MkdirAll(filepath.Join(repo, "dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(repo, "a.md"), "a again\n")
	writeFile(t, filepath.Join(repo, "dir", "c.md"), "c\n")
	writeFile(t, filepath.Join(repo, "café é.md"), "a name with spaces and a letter past ASCII\n")
	commitAllAt(t, repo, "2021-01-01T00:00:00Z", "edit one, add two")
	second := git(t, repo, "rev-parse", "HEAD")
	git(t, repo, "mv", "b.md", "dir/b.md")
	commitAllAt(t, repo, "2022-01-01T00:00:00Z", "move b.md")
	third := git(t, repo, "rev-parse", "HEAD")
	git(t, repo, "rm", "-q", "a.md")
	commitAllAt(t, repo, "2023-01-01T00:00:00Z", "delete a.md")
	fourth := git(t, repo, "rev-parse", "HEAD")

	tests := []struct {
		name string
		rng  string
		want []string
	}{
		{"an edit and two adds", first + ".." + second, []string{"a.md", "café é.md", "dir/c.md"}},
		{"a move", second + ".." + third, []string{"b.md", "dir/b.md"}},
		{"a delete", third + ".." + fourth, []string{"a.md"}},
		{"three dots", first + "..." + second, []string{"a.md", "café é.md", "dir/c.md"}},
		{"a range with no change", second + ".." + second, nil},
	}
	for _, tt := range tests {
		got, err := Changed(repo, tt.rng)
		slices.Sort(got)
		if err != nil || !slices.Equal(got, tt.want) {
			t.Errorf("%s: Changed(%s) = %q, %v, want %q", tt.name, tt.rng, got, err, tt.want)
		}
	}

	// A range that is not one must not be read as something else: a file name
	// is no revision, and a dash would start a git option. --output writes a
	// file, which a read-only check must never do.
	out := filepath.Join(t.TempDir(), "out.txt")
	for _, rng := range []string{"", "no-such-ref..HEAD", "dir/c.md", "--output=" + out} {
		if got, err := Changed(repo, rng); err == nil {
			t.Errorf("Changed(%q) = %q and no error, want an error", rng, got)
		}
	}
	if _, err := os.Stat(out); err == nil {
		t.Error("a range that starts with a dash made git write a file")
	}
}

// A branch is checked against main while main keeps moving. What main changed
// after the branch left it is not what the branch changed.
func TestChangedStartsWhereTheSidesSplit(t *testing.T) {
	repo := repoAt(t, "2020-01-01T00:00:00Z") // a.md and b.md
	git(t, repo, "checkout", "-q", "-b", "side")
	writeFile(t, filepath.Join(repo, "side.md"), "side\n")
	commitAllAt(t, repo, "2021-01-01T00:00:00Z", "side adds a file")
	git(t, repo, "checkout", "-q", "main")
	writeFile(t, filepath.Join(repo, "a.md"), "main moved on\n")
	commitAllAt(t, repo, "2022-01-01T00:00:00Z", "main edits a file")
	git(t, repo, "checkout", "-q", "side")

	for _, rng := range []string{"main..HEAD", "main...HEAD", "main..", "..main"} {
		want := []string{"side.md"}
		if rng == "..main" { // the other way round: what main did, seen from side
			want = []string{"a.md"}
		}
		if got, err := Changed(repo, rng); err != nil || !slices.Equal(got, want) {
			t.Errorf("Changed(%q) = %q, %v, want %q", rng, got, err, want)
		}
	}
}

// Touched reads the commits, not the two ends, so a file a commit wrote and a
// later commit put back still comes back. That is what tells a file the branch
// worked on from one git only copied into the worktree.
func TestTouched(t *testing.T) {
	repo := setupRepo(t)
	git(t, repo, "checkout", "-q", "-b", "side")
	writeFile(t, filepath.Join(repo, "b.md"), "b on the side\n")
	commitAllAt(t, repo, "2021-01-01T00:00:00Z", "side writes b.md")
	// Put b.md back the way main holds it, in its own commit.
	git(t, repo, "checkout", "-q", "main", "--", "b.md")
	commitAllAt(t, repo, "2022-01-01T00:00:00Z", "side puts b.md back")

	got, err := Touched(repo, "main..side")
	if err != nil {
		t.Fatal(err)
	}
	// Changed reads the two ends and would not see b.md here.
	if want := []string{"b.md"}; !slices.Equal(got, want) {
		t.Errorf("Touched(main..side) = %q, want %q even though the bytes are back", got, want)
	}
	if diff, err := Changed(repo, "main..side"); err != nil || len(diff) != 0 {
		t.Errorf("Changed(main..side) = %q, %v, want nothing: the bytes are back", diff, err)
	}

	// A range that is not a revision must not be read as one. A dash would
	// start a git option, and a file name is not a commit range.
	for _, rng := range []string{"--help", "b.md..HEAD"} {
		if got, err := Touched(repo, rng); err == nil {
			t.Errorf("Touched(%q) = %q and no error, want an error", rng, got)
		}
	}
}

// A branch that committed nothing has no file of its own, so Touched gives
// nothing rather than every file the branch holds.
func TestTouchedOnABranchWithNoCommits(t *testing.T) {
	repo := setupRepo(t)
	got, err := Touched(repo, "main..side")
	if err == nil && len(got) != 0 {
		t.Errorf("Touched on a branch with no commits = %q, want nothing", got)
	}
}

// The newest commit on a branch is read by its short hash and subject, and a
// prefix picks the newest commit that says something with it. A prefix no
// commit says gives nothing, so a caller can tell "none" from a branch that
// has no commits at all.
func TestNewestCommit(t *testing.T) {
	repo := setupRepo(t)
	git(t, repo, "checkout", "-q", "-b", "side")
	commitAllAt(t, repo, "2021-01-01T00:00:00Z", "acta: tick fix round 1")
	commitAllAt(t, repo, "2021-02-01T00:00:00Z", "state: write Next")
	commitAllAt(t, repo, "2021-03-01T00:00:00Z", "acta: tick fix round 2")

	hash, subject, err := NewestCommit(repo, "side", "")
	if err != nil {
		t.Fatal(err)
	}
	if subject != "acta: tick fix round 2" || len(hash) != 7 {
		t.Errorf("newest commit = %q %q, want a 7 char hash and the round 2 subject", hash, subject)
	}
	if _, subject, err := NewestCommit(repo, "side", "acta: tick fix round "); err != nil {
		t.Fatal(err)
	} else if subject != "acta: tick fix round 2" {
		t.Errorf("prefix gave %q, want the round 2 subject", subject)
	}
	if _, subject, err := NewestCommit(repo, "side", "acta: tick fix round 7"); err != nil {
		t.Fatalf("a prefix no commit says must not fail: %v", err)
	} else if subject != "" {
		t.Errorf("a prefix no commit says gave %q, want nothing", subject)
	}
}
