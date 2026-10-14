package gitc

import (
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"slices"
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

// One path goes through CommitOrFold, so it folds. More paths make a new commit.
func TestCommitPathsOnePathFolds(t *testing.T) {
	const msg = "chore(scratch): add to item"
	t.Run("one path folds", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		before := commitCount(t, repo)
		writeFile(t, a, "one\ntwo\n")
		r := CommitPaths(repo, []string{a}, "chore(spec): fix wording")
		if r != (Result{Committed: true, Folded: true}) {
			t.Fatalf("got %+v", r)
		}
		if got := commitCount(t, repo); got != before {
			t.Fatalf("commit count %s, want %s", got, before)
		}
		if got, want := git(t, repo, "log", "-1", "--format=%B"), msg+"\n\nchore(spec): fix wording"; got != want {
			t.Fatalf("message %q, want %q", got, want)
		}
	})
	t.Run("two paths make a new commit", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		before := commitCount(t, repo)
		writeFile(t, a, "one\ntwo\n")
		writeFile(t, filepath.Join(repo, "b.md"), "b changed\n")
		r := CommitPaths(repo, []string{a, filepath.Join(repo, "b.md")}, "chore(spec): fix wording")
		if r != (Result{Committed: true}) {
			t.Fatalf("got %+v", r)
		}
		if got, want := commitCount(t, repo), before; got == want {
			t.Fatalf("commit count stayed %s, want a new commit", got)
		}
		if got := git(t, repo, "log", "-1", "--format=%s"); got != "chore(spec): fix wording" {
			t.Fatalf("subject %q", got)
		}
	})
	t.Run("one path keeps the dirty-file-outside-list guard", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		writeFile(t, a, "one\ntwo\n")
		writeFile(t, filepath.Join(repo, "b.md"), "b changed\n")
		if r := CommitPaths(repo, []string{a}, "chore(spec): fix wording"); !r.Committed {
			t.Fatalf("got %+v", r)
		}
		if got := git(t, repo, "status", "--porcelain"); got != "M b.md" {
			t.Fatalf("other dirty file not left alone: %q", got)
		}
	})
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

// Authors names folders to git, never files, so the command line stays short
// however many files the board asks about. A file in an asked folder that was
// not asked about must not come back.
func TestAuthorsAsksByFolder(t *testing.T) {
	repo := firstCommitRepo(t, "Ana")
	for _, d := range []string{"a", "b"} {
		if err := os.MkdirAll(filepath.Join(repo, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeFile(t, filepath.Join(repo, "a", "x.md"), "x\n")
	writeFile(t, filepath.Join(repo, "a", "z.md"), "z\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-qm", "ana writes a/x.md and a/z.md")
	t.Setenv("GIT_AUTHOR_NAME", "Budi")
	t.Setenv("GIT_COMMITTER_NAME", "Budi")
	writeFile(t, filepath.Join(repo, "b", "y.md"), "y\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-qm", "budi writes b/y.md")

	x, y := filepath.Join(repo, "a", "x.md"), filepath.Join(repo, "b", "y.md")
	got := Authors(repo, []string{x, y})
	want := map[string]string{x: "Ana", y: "Budi"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Authors = %v, want %v (a/z.md was not asked, so it has no entry)", got, want)
	}

	// Many files in two folders: the log gets each folder once and no file name.
	var many []string
	for i := 0; i < 300; i++ {
		many = append(many, filepath.Join(repo, "a", "file-"+strings.Repeat("n", i%7)+string(rune('a'+i%26))+".md"))
		many = append(many, filepath.Join(repo, "b", "file-"+strings.Repeat("m", i%5)+string(rune('a'+i%26))+".md"))
	}
	args := authorArgs(repo, many)
	at := slices.Index(args, "--")
	if at < 0 {
		t.Fatalf("args %v have no --", args)
	}
	if got, want := args[at+1:], []string{"a", "b"}; !reflect.DeepEqual(got, want) {
		t.Errorf("pathspecs = %v, want %v", got, want)
	}
	if line := strings.Join(args, " "); strings.Contains(line, ".md") {
		t.Errorf("the git log line names a file: %s", line)
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

// IsAncestor says whether one commit came before another, so the round that
// came later wins whatever subject it uses. A missing commit is an error,
// not a no, so a typo never looks like an order.
func TestIsAncestor(t *testing.T) {
	repo := setupRepo(t)
	parent := git(t, repo, "rev-parse", "HEAD")
	writeFile(t, filepath.Join(repo, "c.md"), "c\n")
	git(t, repo, "add", ".")
	git(t, repo, "commit", "-q", "-m", "second")
	child := git(t, repo, "rev-parse", "HEAD")
	if ok, err := IsAncestor(repo, parent, child); err != nil || !ok {
		t.Errorf("IsAncestor(parent, child) = %v, %v, want true, no error", ok, err)
	}
	if ok, err := IsAncestor(repo, child, parent); err != nil || ok {
		t.Errorf("IsAncestor(child, parent) = %v, %v, want false, no error", ok, err)
	}
	if _, err := IsAncestor(repo, parent, "no-such-commit"); err == nil {
		t.Errorf("IsAncestor with a missing commit gave no error, want one")
	}
}

// foldRepo is a checkout whose newest commit is the first CommitOrFold of
// a.md under msg, written by the name git config gives. The next call with
// the same msg is the one that may fold.
func foldRepo(t *testing.T, msg string) (repo, a string) {
	t.Helper()
	repo = setupRepo(t)
	git(t, repo, "config", "user.name", "test")
	a = filepath.Join(repo, "a.md")
	writeFile(t, a, "one\n")
	r := CommitOrFold(repo, a, msg, false)
	if !r.Committed || r.Folded {
		t.Fatalf("first call = %+v, want a plain commit", r)
	}
	return repo, a
}

func commitCount(t *testing.T, repo string) string {
	t.Helper()
	return git(t, repo, "rev-list", "--count", "HEAD")
}

func TestCommitOrFold(t *testing.T) {
	const msg = "chore(scratch): add to item"
	t.Run("folds", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		before := commitCount(t, repo)
		writeFile(t, a, "one\ntwo\n")
		r := CommitOrFold(repo, a, msg, false)
		if r != (Result{Committed: true, Folded: true}) {
			t.Fatalf("got %+v", r)
		}
		if got := commitCount(t, repo); got != before {
			t.Fatalf("commit count %s, want %s", got, before)
		}
		if got := git(t, repo, "show", "HEAD:a.md"); got != "one\ntwo" {
			t.Fatalf("HEAD holds %q, want both lines", got)
		}
		if got := git(t, repo, "status", "--porcelain"); got != "" {
			t.Fatalf("tree not clean: %q", got)
		}
	})

	// Another planning subject on the same file still folds, and git keeps
	// every subject so the history says what each write was.
	t.Run("a different chore( subject folds and keeps both", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		before := commitCount(t, repo)
		writeFile(t, a, "one\ntwo\n")
		r := CommitOrFold(repo, a, "chore(spec): fix wording", false)
		if r != (Result{Committed: true, Folded: true}) {
			t.Fatalf("got %+v", r)
		}
		if got := commitCount(t, repo); got != before {
			t.Fatalf("commit count %s, want %s", got, before)
		}
		if got, want := git(t, repo, "log", "-1", "--format=%s"), msg; got != want {
			t.Fatalf("subject %q, want %q", got, want)
		}
		if got, want := git(t, repo, "log", "-1", "--format=%B"), msg+"\n\nchore(spec): fix wording"; got != want {
			t.Fatalf("message %q, want %q", got, want)
		}
		if got := git(t, repo, "show", "HEAD:a.md"); got != "one\ntwo" {
			t.Fatalf("HEAD holds %q", got)
		}

		// A third subject is one more body line, and a repeat adds nothing.
		writeFile(t, a, "one\ntwo\nthree\n")
		if r := CommitOrFold(repo, a, "chore(scratch): third", false); !r.Folded {
			t.Fatalf("third got %+v", r)
		}
		writeFile(t, a, "one\ntwo\nthree\nfour\n")
		if r := CommitOrFold(repo, a, "chore(spec): fix wording", false); !r.Folded {
			t.Fatalf("repeat got %+v", r)
		}
		want := msg + "\n\nchore(spec): fix wording\nchore(scratch): third"
		if got := git(t, repo, "log", "-1", "--format=%B"); got != want {
			t.Fatalf("message %q, want %q", got, want)
		}
		if got := commitCount(t, repo); got != before {
			t.Fatalf("commit count %s, want %s", got, before)
		}
	})
	t.Run("the same subject twice adds no body", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		writeFile(t, a, "one\ntwo\n")
		if r := CommitOrFold(repo, a, msg, false); !r.Folded {
			t.Fatalf("got %+v", r)
		}
		if got := git(t, repo, "log", "-1", "--format=%B"); got != msg {
			t.Fatalf("message %q, want only %q", got, msg)
		}
	})

	// HEAD may hold a hand-written body and a trailer block. The fold must keep
	// the body as written and keep the trailers readable as trailers.
	t.Run("a fold keeps the body and the trailer block", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		head := msg + "\n\nIntro paragraph.\n  indented line\n\nCo-Authored-By: X <x@x>\n"
		msgFile := filepath.Join(t.TempDir(), "msg.txt")
		writeFile(t, msgFile, head)
		git(t, repo, "commit", "-q", "--amend", "-F", msgFile)
		writeFile(t, a, "one\ntwo\n")
		if r := CommitOrFold(repo, a, "chore(spec): fix wording", false); !r.Folded {
			t.Fatalf("got %+v", r)
		}
		want := msg + "\n\nIntro paragraph.\n  indented line\nchore(spec): fix wording\n\nCo-Authored-By: X <x@x>"
		if got := git(t, repo, "log", "-1", "--format=%B"); got != want {
			t.Fatalf("message %q, want %q", got, want)
		}
		if got := git(t, repo, "log", "-1", "--format=%(trailers)"); got != "Co-Authored-By: X <x@x>" {
			t.Fatalf("trailers %q, want the Co-Authored-By line", got)
		}
	})

	// Each case breaks one fold rule. The call must make a new commit.
	rules := []struct {
		name  string
		msg   string // message of the second call, msg when empty
		setup func(t *testing.T, repo string)
	}{
		{name: "head is not a chore( commit", setup: func(t *testing.T, repo string) {
			git(t, repo, "commit", "-q", "--amend", "-m", "feat: x")
		}},
		{name: "head also changes another file", setup: func(t *testing.T, repo string) {
			writeFile(t, filepath.Join(repo, "b.md"), "b2\n")
			git(t, repo, "commit", "-q", "--amend", "--no-edit", "-a")
		}},
		{name: "head is a merge", setup: func(t *testing.T, repo string) {
			git(t, repo, "checkout", "-q", "-b", "side", "HEAD~1")
			writeFile(t, filepath.Join(repo, "c.md"), "c\n")
			git(t, repo, "add", "c.md")
			git(t, repo, "commit", "-q", "-m", "side")
			git(t, repo, "checkout", "-q", "main")
			git(t, repo, "merge", "-q", "--no-ff", "-m", msg, "side")
			git(t, repo, "branch", "-q", "-D", "side")
		}},
		{name: "a remote branch holds head", setup: func(t *testing.T, repo string) {
			git(t, repo, "update-ref", "refs/remotes/origin/main", "HEAD")
		}},
		{name: "another local branch holds head", setup: func(t *testing.T, repo string) {
			git(t, repo, "branch", "other")
		}},
		{name: "a linked worktree branch holds head", setup: func(t *testing.T, repo string) {
			git(t, repo, "worktree", "add", "-q", "-b", "wt", filepath.Join(t.TempDir(), "wt"))
		}},
		{name: "a detached worktree sits on head", setup: func(t *testing.T, repo string) {
			git(t, repo, "worktree", "add", "-q", "--detach", filepath.Join(t.TempDir(), "wt"))
		}},
		{name: "a detached worktree sits past head", setup: func(t *testing.T, repo string) {
			wt := filepath.Join(t.TempDir(), "wt")
			git(t, repo, "worktree", "add", "-q", "--detach", wt)
			writeFile(t, filepath.Join(wt, "c.md"), "c\n")
			git(t, wt, "add", "c.md")
			git(t, wt, "commit", "-q", "-m", "ahead")
		}},
		{name: "head author is someone else", setup: func(t *testing.T, repo string) {
			git(t, repo, "commit", "-q", "--amend", "--no-edit", "--author=Other <o@example.com>")
		}},
		{name: "git has no user name", setup: func(t *testing.T, repo string) {
			t.Setenv("GIT_CONFIG_GLOBAL", os.DevNull)
			t.Setenv("GIT_CONFIG_SYSTEM", os.DevNull)
			git(t, repo, "config", "--unset", "user.name")
		}},
	}
	for _, c := range rules {
		t.Run(c.name, func(t *testing.T) {
			repo, a := foldRepo(t, msg)
			if c.setup != nil {
				c.setup(t, repo)
			}
			before := commitCount(t, repo)
			second := c.msg
			if second == "" {
				second = msg
			}
			writeFile(t, a, "one\ntwo\n")
			r := CommitOrFold(repo, a, second, false)
			if r != (Result{Committed: true}) {
				t.Fatalf("got %+v", r)
			}
			if got, want := commitCount(t, repo), before; got == want {
				t.Fatalf("commit count stayed %s, want a new commit", got)
			}
			if got := git(t, repo, "show", "HEAD:a.md"); got != "one\ntwo" {
				t.Fatalf("HEAD holds %q", got)
			}
		})
	}

	// A tag marks a release. The tagged commit must keep its hash, so the
	// edit goes into a new commit and the tag stays where it was.
	for _, c := range []struct {
		name string
		tag  []string
	}{
		{"a lightweight tag on head", []string{"tag", "v1"}},
		{"an annotated tag on head", []string{"tag", "-a", "v1", "-m", "release"}},
	} {
		t.Run(c.name+" blocks the fold", func(t *testing.T) {
			repo, a := foldRepo(t, msg)
			git(t, repo, c.tag...)
			tagged := git(t, repo, "rev-parse", "HEAD")
			before := commitCount(t, repo)
			writeFile(t, a, "one\ntwo\n")
			r := CommitOrFold(repo, a, msg, false)
			if r != (Result{Committed: true}) {
				t.Fatalf("got %+v", r)
			}
			if got := commitCount(t, repo); got == before {
				t.Fatalf("commit count stayed %s, want a new commit", got)
			}
			if got := git(t, repo, "rev-parse", "v1^{commit}"); got != tagged {
				t.Fatalf("tag points at %s, want %s", got, tagged)
			}
			if got := git(t, repo, "rev-parse", "HEAD~1"); got != tagged {
				t.Fatalf("parent is %s, want the tagged commit %s", got, tagged)
			}
			if got := git(t, repo, "show", "HEAD:a.md"); got != "one\ntwo" {
				t.Fatalf("HEAD holds %q", got)
			}
		})
	}

	// These paths never reach the fold check. They must match Commit.
	t.Run("file was dirty", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		before := commitCount(t, repo)
		r := CommitOrFold(repo, a, msg, true)
		if r.Committed || r.Folded || !strings.Contains(r.Reason, "other uncommitted changes") {
			t.Fatalf("got %+v", r)
		}
		if got := commitCount(t, repo); got != before {
			t.Fatalf("commit count %s, want %s", got, before)
		}
	})
	t.Run("mid-merge", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		writeFile(t, filepath.Join(repo, ".git", "MERGE_HEAD"), git(t, repo, "rev-parse", "HEAD")+"\n")
		writeFile(t, a, "two\n")
		r := CommitOrFold(repo, a, msg, false)
		if r.Committed || r.Folded || !strings.Contains(r.Reason, "merge") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("detached HEAD", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		git(t, repo, "checkout", "-q", "--detach")
		before := commitCount(t, repo)
		writeFile(t, a, "two\n")
		r := CommitOrFold(repo, a, msg, false)
		if r.Committed || r.Folded || !strings.Contains(r.Reason, "detached") {
			t.Fatalf("got %+v", r)
		}
		if got := commitCount(t, repo); got != before {
			t.Fatalf("commit count %s, want %s", got, before)
		}
	})
	t.Run("not a repo", func(t *testing.T) {
		dir := t.TempDir()
		p := filepath.Join(dir, "a.md")
		writeFile(t, p, "x\n")
		r := CommitOrFold(dir, p, msg, false)
		if r.Committed || r.Folded || !strings.Contains(r.Reason, "not a git repo") {
			t.Fatalf("got %+v", r)
		}
	})
	t.Run("nothing changed", func(t *testing.T) {
		// Commit refuses a file with no change. A fold must not rewrite HEAD for it.
		repo, a := foldRepo(t, msg)
		head := git(t, repo, "rev-parse", "HEAD")
		r := CommitOrFold(repo, a, msg, false)
		if r.Committed || r.Folded || !strings.Contains(r.Reason, "commit failed") {
			t.Fatalf("got %+v", r)
		}
		if got := git(t, repo, "rev-parse", "HEAD"); got != head {
			t.Fatalf("HEAD moved to %s, want %s", got, head)
		}
	})
	t.Run("amend fails, plain commit follows", func(t *testing.T) {
		repo, a := foldRepo(t, msg)
		before := commitCount(t, repo)
		// The hook fails once, so the amend fails and the plain commit passes.
		hook := filepath.Join(repo, ".git", "hooks", "pre-commit")
		writeFile(t, hook, "#!/bin/sh\n[ -e \"$0.seen\" ] && exit 0\ntouch \"$0.seen\"\necho no >&2\nexit 1\n")
		if err := os.Chmod(hook, 0o755); err != nil {
			t.Fatal(err)
		}
		writeFile(t, a, "one\ntwo\n")
		r := CommitOrFold(repo, a, msg, false)
		if r != (Result{Committed: true}) {
			t.Fatalf("got %+v", r)
		}
		if got, want := commitCount(t, repo), before; got == want {
			t.Fatalf("commit count stayed %s, want a new commit", got)
		}
		if got := git(t, repo, "show", "HEAD:a.md"); got != "one\ntwo" {
			t.Fatalf("HEAD holds %q", got)
		}
	})
}

// commitFileAt writes one file, commits it with both git dates at when, and
// returns the full sha. An empty when keeps the dates git picks itself.
func commitFileAt(t *testing.T, repo, name, body, when, msg string) string {
	t.Helper()
	if when != "" {
		t.Setenv("GIT_AUTHOR_DATE", when)
		t.Setenv("GIT_COMMITTER_DATE", when)
	}
	writeFile(t, filepath.Join(repo, name), body)
	git(t, repo, "add", name)
	git(t, repo, "commit", "-q", "-m", msg)
	return git(t, repo, "rev-parse", "HEAD")
}

// emptyCommitAt makes a commit with no file change and returns its full sha.
func emptyCommitAt(t *testing.T, repo, when, msg string) string {
	t.Helper()
	if when != "" {
		t.Setenv("GIT_AUTHOR_DATE", when)
		t.Setenv("GIT_COMMITTER_DATE", when)
	}
	git(t, repo, "commit", "-q", "--allow-empty", "-m", msg)
	return git(t, repo, "rev-parse", "HEAD")
}

// mergeSide merges the branch into the current one with a merge commit of its
// own, so both sides stay visible in the history.
func mergeSide(t *testing.T, repo, branch string) {
	t.Helper()
	git(t, repo, "merge", "-q", "--no-ff", "-m", "merge "+branch, branch)
}

func TestCheckCommits(t *testing.T) {
	const page, checked = "page.md", ".checked"
	// newRepo has a first commit that does not touch the page.
	newRepo := func(t *testing.T) string { return setupRepo(t) }

	t.Run("a commit that changed the page file", func(t *testing.T) {
		repo := newRepo(t)
		first := commitFileAt(t, repo, page, "one\n", "", "add page")
		second := commitFileAt(t, repo, page, "two\n", "", "edit page")
		commitFileAt(t, repo, "b.md", "other\n", "", "edit other file")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil {
			t.Fatal(err)
		}
		if want := []string{second, first}; !reflect.DeepEqual(got, want) {
			t.Errorf("got %v, want %v", got, want)
		}
	})
	t.Run("nothing at all is an empty list", func(t *testing.T) {
		repo := newRepo(t)
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || len(got) != 0 {
			t.Errorf("got %v, %v, want no commits", got, err)
		}
	})
	t.Run("a commit that adds the id line to the checked file", func(t *testing.T) {
		repo := newRepo(t)
		c := commitFileAt(t, repo, checked, "wiki/p 2026-01-01\n", "", "chore(wiki): re-check wiki/p")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || !reflect.DeepEqual(got, []string{c}) {
			t.Errorf("got %v, %v, want [%s]", got, err, c)
		}
	})
	t.Run("a new date on the line is a check again", func(t *testing.T) {
		repo := newRepo(t)
		first := commitFileAt(t, repo, checked, "wiki/p 2026-01-01\n", "", "check one")
		second := commitFileAt(t, repo, checked, "wiki/p 2026-02-01\n", "", "check two")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || !slices.Contains(got, first) || !slices.Contains(got, second) {
			t.Errorf("got %v, %v, want %s and %s", got, err, first, second)
		}
	})
	t.Run("one commit with two lines answers to both ids", func(t *testing.T) {
		repo := newRepo(t)
		c := commitFileAt(t, repo, checked, "wiki/p 2026-01-01\nwiki/q 2026-01-01\n", "", "check both")
		for _, id := range []string{"wiki/p", "wiki/q"} {
			got, err := CheckCommits(repo, page, checked, id)
			if err != nil || !reflect.DeepEqual(got, []string{c}) {
				t.Errorf("%s: got %v, %v, want [%s]", id, got, err, c)
			}
		}
	})
	t.Run("a commit that only removes the line is not a check", func(t *testing.T) {
		repo := newRepo(t)
		add := commitFileAt(t, repo, checked, "wiki/p 2026-01-01\nwiki/q 2026-01-01\n", "", "check both")
		commitFileAt(t, repo, checked, "wiki/q 2026-01-01\n", "", "drop p")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || !reflect.DeepEqual(got, []string{add}) {
			t.Errorf("got %v, %v, want [%s]", got, err, add)
		}
	})
	t.Run("a line for another id or a longer id is not a check", func(t *testing.T) {
		repo := newRepo(t)
		commitFileAt(t, repo, checked, "wiki/other 2026-01-01\nwiki/p-two 2026-01-01\n", "", "check others")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || len(got) != 0 {
			t.Errorf("got %v, %v, want no commits", got, err)
		}
	})
	t.Run("a commit that does both is listed once", func(t *testing.T) {
		repo := newRepo(t)
		writeFile(t, filepath.Join(repo, checked), "wiki/p 2026-01-01\n")
		git(t, repo, "add", checked)
		c := commitFileAt(t, repo, page, "one\n", "", "edit page and check it")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || !reflect.DeepEqual(got, []string{c}) {
			t.Errorf("got %v, %v, want [%s]", got, err, c)
		}
	})
	t.Run("a commit outside HEAD is not listed", func(t *testing.T) {
		repo := newRepo(t)
		git(t, repo, "checkout", "-q", "-b", "side")
		commitFileAt(t, repo, checked, "wiki/p 2026-01-01\n", "", "check on side")
		commitFileAt(t, repo, page, "side\n", "", "edit page on side")
		git(t, repo, "checkout", "-q", "main")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || len(got) != 0 {
			t.Errorf("got %v, %v, want no commits", got, err)
		}
	})
	t.Run("a page change on both sides of a merge is listed on both", func(t *testing.T) {
		repo := newRepo(t)
		git(t, repo, "checkout", "-q", "-b", "side")
		onSide := commitFileAt(t, repo, page, "same\n", "", "edit page on side")
		git(t, repo, "checkout", "-q", "main")
		onMain := commitFileAt(t, repo, page, "same\n", "", "edit page on main")
		mergeSide(t, repo, "side")
		got, err := CheckCommits(repo, page, checked, "wiki/p")
		if err != nil || !slices.Contains(got, onSide) || !slices.Contains(got, onMain) {
			t.Errorf("got %v, %v, want both %s and %s", got, err, onSide, onMain)
		}
	})
	t.Run("a page file with odd characters is a plain name", func(t *testing.T) {
		repo := newRepo(t)
		commitFileAt(t, repo, page, "one\n", "", "add page")
		got, err := CheckCommits(repo, "*.md", checked, "wiki/p")
		if err != nil || len(got) != 0 {
			t.Errorf("a glob matched: got %v, %v", got, err)
		}
	})
	t.Run("a folder that is no checkout is an error", func(t *testing.T) {
		if got, err := CheckCommits(t.TempDir(), page, checked, "wiki/p"); err == nil {
			t.Errorf("got %v and no error", got)
		}
	})
}

func TestUnseen(t *testing.T) {
	paths := []string{"src/"}
	srcFile := func(t *testing.T, repo, body, when, msg string) string {
		t.Helper()
		if err := os.MkdirAll(filepath.Join(repo, "src"), 0o755); err != nil {
			t.Fatal(err)
		}
		return commitFileAt(t, repo, "src/x.go", body, when, msg)
	}
	short := func(t *testing.T, repo, sha string) string { return git(t, repo, "rev-parse", "--short", sha) }
	// wantNone and wantCommit check one answer from Unseen.
	wantNone := func(t *testing.T, repo string, checks []string) {
		t.Helper()
		sha, subject, err := Unseen(repo, checks, paths)
		if err != nil || sha != "" || subject != "" {
			t.Errorf("got %q %q %v, want nothing unseen", sha, subject, err)
		}
	}
	wantCommit := func(t *testing.T, repo string, checks []string, full, subject string) {
		t.Helper()
		sha, gotSubject, err := Unseen(repo, checks, paths)
		if err != nil || sha != short(t, repo, full) || gotSubject != subject {
			t.Errorf("got %q %q %v, want %q %q", sha, gotSubject, err, short(t, repo, full), subject)
		}
	}

	t.Run("no check commits and a path change", func(t *testing.T) {
		repo := setupRepo(t)
		p := srcFile(t, repo, "one\n", "", "change src")
		wantCommit(t, repo, nil, p, "change src")
	})
	t.Run("no check commits and no path change", func(t *testing.T) {
		repo := setupRepo(t)
		commitFileAt(t, repo, "b.md", "other\n", "", "change something else")
		wantNone(t, repo, nil)
	})
	t.Run("the check commit is the path commit itself", func(t *testing.T) {
		repo := setupRepo(t)
		p := srcFile(t, repo, "one\n", "", "change src")
		wantNone(t, repo, []string{p})
	})
	t.Run("a check after the path commit", func(t *testing.T) {
		repo := setupRepo(t)
		srcFile(t, repo, "one\n", "", "change src")
		c := emptyCommitAt(t, repo, "", "check")
		wantNone(t, repo, []string{c})
	})
	t.Run("a path commit after the check", func(t *testing.T) {
		repo := setupRepo(t)
		c := emptyCommitAt(t, repo, "", "check")
		p := srcFile(t, repo, "one\n", "", "change src")
		wantCommit(t, repo, []string{c}, p, "change src")
	})
	t.Run("the newest of several unseen commits", func(t *testing.T) {
		repo := setupRepo(t)
		srcFile(t, repo, "one\n", "", "first change")
		p := srcFile(t, repo, "two\n", "", "second change")
		wantCommit(t, repo, nil, p, "second change")
	})
	t.Run("only the path commits before the check count as seen", func(t *testing.T) {
		repo := setupRepo(t)
		srcFile(t, repo, "one\n", "", "first change")
		c := emptyCommitAt(t, repo, "", "check")
		p := srcFile(t, repo, "two\n", "", "second change")
		wantCommit(t, repo, []string{c}, p, "second change")
	})
	t.Run("a check on one side of a merge and the path change on the other", func(t *testing.T) {
		repo := setupRepo(t)
		git(t, repo, "checkout", "-q", "-b", "side")
		p := srcFile(t, repo, "side\n", "", "change src on side")
		git(t, repo, "checkout", "-q", "main")
		c := emptyCommitAt(t, repo, "", "check on main")
		mergeSide(t, repo, "side")
		wantCommit(t, repo, []string{c}, p, "change src on side")
	})
	t.Run("a check on both sides of a merge", func(t *testing.T) {
		repo := setupRepo(t)
		git(t, repo, "checkout", "-q", "-b", "side")
		srcFile(t, repo, "side\n", "", "change src on side")
		onSide := emptyCommitAt(t, repo, "", "check on side")
		git(t, repo, "checkout", "-q", "main")
		onMain := emptyCommitAt(t, repo, "", "check on main")
		mergeSide(t, repo, "side")
		wantNone(t, repo, []string{onMain, onSide})
	})
	t.Run("checks on both sides cover a change on each side", func(t *testing.T) {
		repo := setupRepo(t)
		git(t, repo, "checkout", "-q", "-b", "side")
		srcFile(t, repo, "side\n", "", "change src on side")
		onSide := emptyCommitAt(t, repo, "", "check on side")
		git(t, repo, "checkout", "-q", "main")
		if err := os.MkdirAll(filepath.Join(repo, "src"), 0o755); err != nil {
			t.Fatal(err)
		}
		commitFileAt(t, repo, "src/y.go", "main\n", "", "change src on main")
		onMain := emptyCommitAt(t, repo, "", "check on main")
		mergeSide(t, repo, "side")
		wantNone(t, repo, []string{onMain, onSide})
	})
	t.Run("a path commit dated before the check but not its ancestor", func(t *testing.T) {
		repo := setupRepo(t)
		git(t, repo, "checkout", "-q", "-b", "side")
		p := srcFile(t, repo, "side\n", "2020-01-01T00:00:00Z", "change src on side")
		git(t, repo, "checkout", "-q", "main")
		c := emptyCommitAt(t, repo, "2025-01-01T00:00:00Z", "check on main")
		mergeSide(t, repo, "side")
		wantCommit(t, repo, []string{c}, p, "change src on side")
	})
	t.Run("a path commit dated after the check but its ancestor", func(t *testing.T) {
		repo := setupRepo(t)
		srcFile(t, repo, "one\n", "2030-01-01T00:00:00Z", "change src")
		c := emptyCommitAt(t, repo, "2025-01-01T00:00:00Z", "check")
		wantNone(t, repo, []string{c})
	})
	t.Run("the same change on both sides still counts on the unchecked side", func(t *testing.T) {
		repo := setupRepo(t)
		git(t, repo, "checkout", "-q", "-b", "side")
		p := srcFile(t, repo, "same\n", "", "change src on side")
		git(t, repo, "checkout", "-q", "main")
		srcFile(t, repo, "same\n", "", "change src on main")
		c := emptyCommitAt(t, repo, "", "check on main")
		mergeSide(t, repo, "side")
		wantCommit(t, repo, []string{c}, p, "change src on side")
	})
	t.Run("no paths is not the whole repo", func(t *testing.T) {
		repo := setupRepo(t)
		srcFile(t, repo, "one\n", "", "change src")
		sha, subject, err := Unseen(repo, nil, nil)
		if err != nil || sha != "" || subject != "" {
			t.Errorf("got %q %q %v, want nothing", sha, subject, err)
		}
	})
	t.Run("a glob is a name, not a pattern", func(t *testing.T) {
		repo := setupRepo(t)
		srcFile(t, repo, "one\n", "", "change src")
		sha, _, err := Unseen(repo, nil, []string{"src/*.go"})
		if err != nil || sha != "" {
			t.Errorf("got %q %v, want nothing", sha, err)
		}
	})
	t.Run("a folder that is no checkout is an error", func(t *testing.T) {
		if sha, _, err := Unseen(t.TempDir(), nil, paths); err == nil {
			t.Errorf("got %q and no error", sha)
		}
	})
}

func TestCommitEmpty(t *testing.T) {
	repo := setupRepo(t)
	parentTree := git(t, repo, "rev-parse", "HEAD^{tree}")
	// A staged file must stay staged and stay out of the empty commit.
	writeFile(t, filepath.Join(repo, "staged.md"), "staged\n")
	git(t, repo, "add", "staged.md")

	sha, err := CommitEmpty(repo, "chore(wiki): re-check wiki/p")
	if err != nil {
		t.Fatal(err)
	}
	if want := git(t, repo, "rev-parse", "--short", "HEAD"); sha != want {
		t.Errorf("sha = %q, want HEAD's short sha %q", sha, want)
	}
	if got := git(t, repo, "rev-parse", "HEAD^{tree}"); got != parentTree {
		t.Errorf("the commit's tree is %s, want the parent's tree %s", got, parentTree)
	}
	if got := git(t, repo, "diff", "--cached", "--name-only"); got != "staged.md" {
		t.Errorf("staged files = %q, want only staged.md", got)
	}
	if got := git(t, repo, "log", "-1", "--format=%s"); got != "chore(wiki): re-check wiki/p" {
		t.Errorf("subject = %q", got)
	}
}

// stopOnConflict leaves the repo in the middle of op ("cherry-pick" or
// "rebase"), stopped on a conflict in a.md.
func stopOnConflict(t *testing.T, repo, op string) {
	t.Helper()
	git(t, repo, "checkout", "-q", "-b", "side")
	writeFile(t, filepath.Join(repo, "a.md"), "side\n")
	git(t, repo, "commit", "-q", "-am", "side edit")
	git(t, repo, "checkout", "-q", "main")
	writeFile(t, filepath.Join(repo, "a.md"), "main\n")
	git(t, repo, "commit", "-q", "-am", "main edit")
	args := []string{"-C", repo, "cherry-pick", "side"}
	if op == "rebase" {
		git(t, repo, "checkout", "-q", "side")
		args = []string{"-C", repo, "rebase", "main"}
	}
	// Both branches changed a.md, so git stops on a conflict.
	if out, err := exec.Command("git", args...).CombinedOutput(); err == nil {
		t.Fatalf("%s should conflict: %s", op, out)
	}
}

func TestCommitEmptyRefusesWhileGitIsStopped(t *testing.T) {
	for _, op := range []string{"cherry-pick", "rebase"} {
		t.Run(op, func(t *testing.T) {
			repo := setupRepo(t)
			stopOnConflict(t, repo, op)
			before := git(t, repo, "rev-parse", "HEAD")
			if sha, err := CommitEmpty(repo, "chore(wiki): re-check p"); err == nil {
				t.Fatalf("CommitEmpty made %s in the middle of a %s", sha, op)
			}
			if after := git(t, repo, "rev-parse", "HEAD"); after != before {
				t.Fatalf("HEAD moved from %s to %s", before, after)
			}
		})
	}
}
