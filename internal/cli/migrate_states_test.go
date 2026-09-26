package cli

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/iyay/acta/internal/board"
	"github.com/iyay/acta/internal/config"
)

// gitOut runs git in dir and returns its trimmed output, failing the test
// when git fails.
func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimRight(string(out), "\n")
}

// gitRaw returns the bytes of a file in a commit, so a test can compare the
// text of a moved yaml exactly as git stored it.
func gitRaw(t *testing.T, dir, rev, path string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "show", rev+":"+path).CombinedOutput()
	if err != nil {
		t.Fatalf("git show %s:%s: %v %s", rev, path, err, out)
	}
	return string(out)
}

// repoSnapshot records everything a refusal must leave alone: the commit sha,
// the staged diff, the short status and the bytes of every file outside .git.
func repoSnapshot(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	fmt.Fprintf(&b, "HEAD %s\n", gitOut(t, dir, "rev-parse", "HEAD"))
	fmt.Fprintf(&b, "STATUS %q\n", gitOut(t, dir, "status", "--porcelain"))
	fmt.Fprintf(&b, "CACHED %q\n", gitOut(t, dir, "diff", "--cached"))
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			return rerr
		}
		if d.IsDir() {
			if rel == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		fmt.Fprintf(&b, "FILE %s %q\n", filepath.ToSlash(rel), string(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// writeRepoFile writes a file in the temp repo and returns its path.
func writeRepoFile(t *testing.T, dir, rel, body string) string {
	t.Helper()
	full := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(full, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return full
}

// commitAll stages the given paths and commits them, so a test starts from a
// known clean state.
func commitAll(t *testing.T, dir, msg string, paths ...string) {
	t.Helper()
	gitOut(t, dir, append([]string{"add", "--"}, paths...)...)
	gitOut(t, dir, "commit", "-q", "-m", msg)
}

// planRepo makes the fixture repo hold one plan file, so a test can ask the
// board whether the plans are still there after the move.
func planRepo(t *testing.T) string {
	t.Helper()
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm/plans/2026-01-01-demo.md", "# Demo\n\n- [ ] one thing\n")
	commitAll(t, dir, "add plan", ".pm")
	return dir
}

// assertRefused runs migrate-root and wants exit 1 with the repo untouched.
func assertRefused(t *testing.T, dir string) {
	t.Helper()
	before := repoSnapshot(t, dir)
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1 (stderr %q)", code, stderr.String())
	}
	if stderr.String() == "" {
		t.Fatal("a refusal must say what to do on stderr")
	}
	if after := repoSnapshot(t, dir); after != before {
		t.Fatalf("a refusal changed the repo\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// commitPaths lists the paths the move commit touches.
func commitPaths(t *testing.T, dir string) []string {
	t.Helper()
	out := gitOut(t, dir, "show", "--pretty=", "--name-status", "-M", "HEAD")
	var paths []string
	for _, l := range strings.Split(out, "\n") {
		if l == "" {
			continue
		}
		fields := strings.Fields(l)
		paths = append(paths, fields[len(fields)-1])
	}
	return paths
}

func TestMigrateRootCleanRepoCommitsOnlyTheMove(t *testing.T) {
	dir := migrateRepo(t)
	before := gitOut(t, dir, "rev-parse", "HEAD")

	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d stderr %q", code, stderr.String())
	}
	if got := gitOut(t, dir, "rev-list", "--count", before+"..HEAD"); got != "1" {
		t.Fatalf("want exactly one new commit, got %s", got)
	}
	if s := gitOut(t, dir, "status", "--porcelain"); s != "" {
		t.Fatalf("want a clean tree after the move, got %q", s)
	}
	want := map[string]bool{".pm/note.md": true, ".acta/note.md": true}
	for _, p := range commitPaths(t, dir) {
		if !want[p] {
			t.Fatalf("commit touches %q, want only the .pm to .acta renames", p)
		}
	}
}

func TestMigrateRootRenamesSettingsYaml(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "dirs:\n  plans: plans\n  specs: specs\n")
	commitAll(t, dir, "add yaml", ".pm.yaml")

	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d stderr %q", code, stderr.String())
	}
	got := gitRaw(t, dir, "HEAD", ".acta.yaml")
	if want := "dirs:\n  plans: plans\n  specs: specs\n"; got != want {
		t.Fatalf(".acta.yaml is %q, want %q", got, want)
	}
	if _, err := os.Stat(filepath.Join(dir, ".pm.yaml")); !os.IsNotExist(err) {
		t.Fatal(".pm.yaml still exists")
	}
	for _, p := range commitPaths(t, dir) {
		if p != ".pm/note.md" && p != ".acta/note.md" && p != ".pm.yaml" && p != ".acta.yaml" {
			t.Fatalf("commit touches %q, want only the renames", p)
		}
	}
}

func TestMigrateRootDropsDefaultRootLine(t *testing.T) {
	for _, rootLine := range []string{"root: .pm", "root: ./.pm", "root: .pm/", "root: ./.pm/"} {
		t.Run(rootLine, func(t *testing.T) {
			dir := planRepo(t)
			writeRepoFile(t, dir, ".pm.yaml", rootLine+"\n\n# the board folder\ndirs:\n  plans: plans\n")
			commitAll(t, dir, "add yaml", ".pm.yaml")

			var stdout, stderr strings.Builder
			if code := migrateRoot(dir, &stdout, &stderr); code != 0 {
				t.Fatalf("exit %d stderr %q", code, stderr.String())
			}
			got := gitRaw(t, dir, "HEAD", ".acta.yaml")
			if want := "\n# the board folder\ndirs:\n  plans: plans\n"; got != want {
				t.Fatalf(".acta.yaml is %q, want %q with no root line", got, want)
			}
			// The rewritten file must not point the board at the folder that
			// is gone, so the plans are still listed after the move.
			t.Setenv("ACTA_ROOT", "")
			t.Setenv("PM_ROOT", "")
			cfg, err := config.Load(dir, "")
			if err != nil {
				t.Fatal(err)
			}
			if want := filepath.Join(dir, ".acta"); cfg.Root != want {
				t.Fatalf("root is %q, want %q", cfg.Root, want)
			}
			b, err := board.Load(cfg)
			if err != nil {
				t.Fatal(err)
			}
			if n := len(b.List(board.KindPlan, true)); n != 1 {
				t.Fatalf("board lists %d plans, want 1", n)
			}
		})
	}
}

func TestMigrateRootRefusesForeignRoot(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "root: docs/plans\n")
	commitAll(t, dir, "add yaml", ".pm.yaml")
	assertRefused(t, dir)
}

func TestMigrateRootRefusesUntrackedYaml(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "root: .pm\n")
	assertRefused(t, dir)
}

// TestMigrateRootRefusesIgnoredUntrackedYaml covers a .pm.yaml that is
// untracked and also listed in .gitignore. A plain git status cannot see an
// ignored file, so the command has to ask git whether the file is tracked at
// all and refuse before the first move. Without that the move happens and
// git mv fails late with a raw git error and exit 3.
func TestMigrateRootRefusesIgnoredUntrackedYaml(t *testing.T) {
	dir := planRepo(t)
	writeRepoFile(t, dir, ".gitignore", ".pm.yaml\n")
	commitAll(t, dir, "ignore yaml", ".gitignore")
	writeRepoFile(t, dir, ".pm.yaml", "root: .pm\n")
	if s := gitOut(t, dir, "status", "--porcelain"); strings.Contains(s, ".pm.yaml") {
		t.Fatalf("fixture broken, .pm.yaml shows in status %q", s)
	}

	before := repoSnapshot(t, dir)
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1 (stderr %q)", code, stderr.String())
	}
	if s := stderr.String(); !strings.Contains(s, ".pm.yaml") || strings.Count(s, "\n") != 1 {
		t.Fatalf("stderr %q must be one line naming .pm.yaml", s)
	}
	if after := repoSnapshot(t, dir); after != before {
		t.Fatalf("a refusal changed the repo\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// TestMigrateRootMovesYamlThatIsTrackedButIgnored pins the other side of that
// check: a .pm.yaml that is committed and only added to .gitignore later is
// still tracked, so git can rename it and the move goes ahead.
func TestMigrateRootMovesYamlThatIsTrackedButIgnored(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "dirs:\n  plans: plans\n")
	commitAll(t, dir, "add yaml", ".pm.yaml")
	writeRepoFile(t, dir, ".gitignore", ".pm.yaml\n")
	commitAll(t, dir, "ignore yaml", ".gitignore")

	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d stderr %q", code, stderr.String())
	}
	if body := gitRaw(t, dir, "HEAD", ".acta.yaml"); body != "dirs:\n  plans: plans\n" {
		t.Fatalf(".acta.yaml in HEAD is %q", body)
	}
}

func TestMigrateRootRefusesModifiedYaml(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "dirs:\n  plans: plans\n")
	commitAll(t, dir, "add yaml", ".pm.yaml")
	writeRepoFile(t, dir, ".pm.yaml", "dirs:\n  plans: plans\n  bugs: bugs\n")
	assertRefused(t, dir)
}

func TestMigrateRootRefusesStagedYaml(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "dirs:\n  plans: plans\n")
	commitAll(t, dir, "add yaml", ".pm.yaml")
	writeRepoFile(t, dir, ".pm.yaml", "dirs:\n  plans: plans\n  bugs: bugs\n")
	gitOut(t, dir, "add", "--", ".pm.yaml")
	assertRefused(t, dir)
}

func TestMigrateRootLeavesStagedFileAlone(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, "other.txt", "mine\n")
	gitOut(t, dir, "add", "--", "other.txt")

	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d stderr %q", code, stderr.String())
	}
	for _, p := range commitPaths(t, dir) {
		if p == "other.txt" {
			t.Fatal("the move commit swallowed the file the user had staged")
		}
	}
	if s := gitOut(t, dir, "status", "--porcelain", "--", "other.txt"); s != "A  other.txt" {
		t.Fatalf("other.txt status is %q, want it still staged", s)
	}
	if gitOut(t, dir, "log", "-1", "--pretty=%s", "HEAD", "--", "other.txt") != "" {
		t.Fatal("other.txt is already committed")
	}
}

func TestMigrateRootLeavesUnstagedFileAlone(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, "other.txt", "one\n")
	commitAll(t, dir, "add other", "other.txt")
	writeRepoFile(t, dir, "other.txt", "two\n")

	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d stderr %q", code, stderr.String())
	}
	if s := gitOut(t, dir, "status", "--porcelain", "--", "other.txt"); s != " M other.txt" {
		t.Fatalf("other.txt status is %q, want it still modified and uncommitted", s)
	}
	if body := gitRaw(t, dir, "HEAD", "other.txt"); body != "one\n" {
		t.Fatalf("other.txt in HEAD is %q, want the old text", body)
	}
}

func TestMigrateRootRefusesUntrackedPm(t *testing.T) {
	dir := migrateRepo(t)
	gitOut(t, dir, "rm", "-q", "-r", ".pm")
	gitOut(t, dir, "commit", "-q", "-m", "drop pm")
	writeRepoFile(t, dir, ".pm/note.md", "hi\n")
	assertRefused(t, dir)
}

func TestMigrateRootRefusesIgnoredPm(t *testing.T) {
	dir := migrateRepo(t)
	gitOut(t, dir, "rm", "-q", "-r", ".pm")
	writeRepoFile(t, dir, ".gitignore", ".pm/\n")
	gitOut(t, dir, "add", "--", ".gitignore")
	gitOut(t, dir, "commit", "-q", "-m", "ignore pm")
	writeRepoFile(t, dir, ".pm/note.md", "hi\n")
	if s := gitOut(t, dir, "status", "--porcelain"); strings.Contains(s, ".pm") {
		t.Fatalf("fixture broken, .pm shows in status %q", s)
	}
	assertRefused(t, dir)
}

func TestMigrateRootUndoesWhenCommitFails(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "root: .pm\n")
	commitAll(t, dir, "add yaml", ".pm.yaml")
	hook := writeRepoFile(t, dir, ".git/hooks/pre-commit", "#!/bin/sh\nexit 1\n")
	if err := os.Chmod(hook, 0o755); err != nil {
		t.Fatal(err)
	}
	before := gitOut(t, dir, "rev-parse", "HEAD")

	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code == 0 {
		t.Fatal("want a non-zero exit when the commit is refused")
	}
	if got := gitOut(t, dir, "rev-parse", "HEAD"); got != before {
		t.Fatalf("HEAD moved to %s, want %s", got, before)
	}
	if s := gitOut(t, dir, "status", "--porcelain"); s != "" {
		t.Fatalf("want a clean tree after the undo, got %q", s)
	}
	if _, err := os.Stat(filepath.Join(dir, ".pm", "note.md")); err != nil {
		t.Fatalf(".pm/note.md is gone: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta")); !os.IsNotExist(err) {
		t.Fatal(".acta/ was left behind")
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); !os.IsNotExist(err) {
		t.Fatal(".acta.yaml was left behind")
	}
	if body := gitRaw(t, dir, "HEAD", ".pm.yaml"); body != "root: .pm\n" {
		t.Fatalf(".pm.yaml in HEAD is %q", body)
	}
	if stderr.String() == "" {
		t.Fatal("want a message on stderr")
	}
}

func TestMigrateRootKeepsEmptyYaml(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "")
	commitAll(t, dir, "add yaml", ".pm.yaml")

	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 0 {
		t.Fatalf("exit %d stderr %q", code, stderr.String())
	}
	if got := gitRaw(t, dir, "HEAD", ".acta.yaml"); got != "" {
		t.Fatalf(".acta.yaml is %q, want it empty", got)
	}
}

func TestMigrateRootRefusalNamesTheProblem(t *testing.T) {
	dir := migrateRepo(t)
	writeRepoFile(t, dir, ".pm.yaml", "root: .pm\n")
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
	if s := stderr.String(); !strings.Contains(s, ".pm.yaml") {
		t.Fatalf("stderr %q must name .pm.yaml", s)
	}
}

// repoState records everything a refusal must leave alone, and it looks at
// every entry with os.Lstat instead of reading it. That way a .acta.yaml that
// is a folder, a link, a broken link or a link pointing at itself is part of
// the picture instead of an error in the walk.
func repoState(t *testing.T, dir string) string {
	t.Helper()
	var b strings.Builder
	fmt.Fprintf(&b, "HEAD %s\n", gitOut(t, dir, "rev-parse", "HEAD"))
	fmt.Fprintf(&b, "STATUS %q\n", gitOut(t, dir, "status", "--porcelain"))
	fmt.Fprintf(&b, "CACHED %q\n", gitOut(t, dir, "diff", "--cached"))
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, rerr := filepath.Rel(dir, p)
		if rerr != nil {
			return rerr
		}
		if d.IsDir() {
			if rel == ".git" {
				return filepath.SkipDir
			}
			fmt.Fprintf(&b, "DIR %s\n", filepath.ToSlash(rel))
			return nil
		}
		info, lerr := os.Lstat(p)
		if lerr != nil {
			return lerr
		}
		if info.Mode()&fs.ModeSymlink != 0 {
			// Record where the link points, not what it points at, so a link
			// loop never turns into a walk that never ends.
			target, terr := os.Readlink(p)
			if terr != nil {
				return terr
			}
			fmt.Fprintf(&b, "LINK %s -> %s\n", filepath.ToSlash(rel), target)
			return nil
		}
		data, rerr := os.ReadFile(p)
		if rerr != nil {
			return rerr
		}
		fmt.Fprintf(&b, "FILE %s %q\n", filepath.ToSlash(rel), string(data))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return b.String()
}

// assertRefusedActaYaml runs migrate-root and wants exit 1 with a message and
// a repo that is byte for byte what it was. wantName is the file the message
// has to name, or "" when another refusal wins and any clear line will do.
func assertRefusedActaYaml(t *testing.T, dir, wantName string) {
	t.Helper()
	before := repoState(t, dir)
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1 (stderr %q)", code, stderr.String())
	}
	if stderr.String() == "" {
		t.Fatal("a refusal must say what to do on stderr")
	}
	if wantName != "" && !strings.Contains(stderr.String(), wantName) {
		t.Fatalf("stderr %q must name %s", stderr.String(), wantName)
	}
	if after := repoState(t, dir); after != before {
		t.Fatalf("a refusal changed the repo\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

// writeTrackedActaYaml puts a .acta.yaml in the repo and commits it, so the
// file is known, tracked and clean.
func writeTrackedActaYaml(t *testing.T, dir, body string) {
	t.Helper()
	writeRepoFile(t, dir, ".acta.yaml", body)
	commitAll(t, dir, "add acta yaml", ".acta.yaml")
}

// symlink makes a link at rel in the repo pointing at target.
func symlink(t *testing.T, dir, target, rel string) {
	t.Helper()
	if err := os.Symlink(target, filepath.Join(dir, filepath.FromSlash(rel))); err != nil {
		t.Fatal(err)
	}
}

// actaYamlForms is every shape .acta.yaml can have in the repo root before the
// move. Each one has to make the command refuse with exit 1 and change
// nothing, because the move writes that exact name. wantName is the file the
// message has to name, or "" when the case also holds a state that refuses
// first, where any clear line will do.
var actaYamlForms = []struct {
	name     string
	wantName string
	setup    func(t *testing.T, dir string)
}{
	{"tracked file clean", ".acta.yaml", func(t *testing.T, dir string) {
		writeTrackedActaYaml(t, dir, "voice: id\n")
	}},
	{"untracked file", ".acta.yaml", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, ".acta.yaml", "root: .pm\n")
	}},
	{"ignored file", ".acta.yaml", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, ".gitignore", ".acta.yaml\n")
		commitAll(t, dir, "add ignore", ".gitignore")
		writeRepoFile(t, dir, ".acta.yaml", "root: .pm\n")
		if s := gitOut(t, dir, "status", "--porcelain"); strings.Contains(s, ".acta.yaml") {
			t.Fatalf("git status %q should not see the ignored .acta.yaml", s)
		}
	}},
	{"tracked file modified", ".acta.yaml", func(t *testing.T, dir string) {
		writeTrackedActaYaml(t, dir, "voice: id\n")
		writeRepoFile(t, dir, ".acta.yaml", "voice: other\n")
	}},
	{"tracked file staged", ".acta.yaml", func(t *testing.T, dir string) {
		writeTrackedActaYaml(t, dir, "voice: id\n")
		writeRepoFile(t, dir, ".acta.yaml", "voice: other\n")
		gitOut(t, dir, "add", "--", ".acta.yaml")
	}},
	{"empty file", ".acta.yaml", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, ".acta.yaml", "")
	}},
	{"empty folder", ".acta.yaml", func(t *testing.T, dir string) {
		if err := os.Mkdir(filepath.Join(dir, ".acta.yaml"), 0o755); err != nil {
			t.Fatal(err)
		}
	}},
	{"folder holding a file", ".acta.yaml", func(t *testing.T, dir string) {
		if err := os.Mkdir(filepath.Join(dir, ".acta.yaml"), 0o755); err != nil {
			t.Fatal(err)
		}
		writeRepoFile(t, dir, ".acta.yaml/inside.md", "hi\n")
	}},
	{"symlink to a file", ".acta.yaml", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, "other.yaml", "voice: id\n")
		symlink(t, dir, "other.yaml", ".acta.yaml")
	}},
	{"symlink to a folder", ".acta.yaml", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, "other/inside.md", "hi\n")
		symlink(t, dir, "other", ".acta.yaml")
	}},
	{"dangling symlink", ".acta.yaml", func(t *testing.T, dir string) {
		symlink(t, dir, "not-here.yaml", ".acta.yaml")
	}},
	{"symlink loop", ".acta.yaml", func(t *testing.T, dir string) {
		symlink(t, dir, ".acta.yaml", ".acta.yaml")
	}},
	{"with a tracked .pm.yaml", ".acta.yaml", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, ".pm.yaml", "root: .pm\n")
		commitAll(t, dir, "add yaml", ".pm.yaml")
		writeRepoFile(t, dir, ".acta.yaml", "root: .pm\n")
	}},
	{"with no .pm folder", "", func(t *testing.T, dir string) {
		// Removing the folder and committing it leaves the repo with no .pm
		// to move, so this form can be paired with that state as well.
		gitOut(t, dir, "rm", "-r", "-q", "--", ".pm")
		gitOut(t, dir, "commit", "-q", "-m", "drop pm")
		writeRepoFile(t, dir, ".acta.yaml", "root: .pm\n")
	}},
	{"with an .acta folder", "", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, ".acta/inside.md", "hi\n")
		writeRepoFile(t, dir, ".acta.yaml", "root: .pm\n")
	}},
	{"different letter case", ".acta.yaml", func(t *testing.T, dir string) {
		writeRepoFile(t, dir, ".ACTA.YAML", "root: .pm\n")
		// A case-insensitive filesystem already has this name taken. A
		// case-sensitive one does not, and there is nothing to refuse.
		if _, err := os.Lstat(filepath.Join(dir, ".acta.yaml")); err != nil {
			t.Skipf("this filesystem keeps letter case, so .ACTA.YAML is another name: %v", err)
		}
	}},
}

// TestMigrateRootRefusesEveryActaYaml walks every form .acta.yaml can take in
// the repo root. The move ends in a file with that name, so an existing one in
// any shape has to stop the command before the first git mv, whatever git
// thinks of it and whatever the filesystem follows links to.
func TestMigrateRootRefusesEveryActaYaml(t *testing.T) {
	for _, form := range actaYamlForms {
		t.Run(form.name, func(t *testing.T) {
			dir := migrateRepo(t)
			form.setup(t, dir)
			assertRefusedActaYaml(t, dir, form.wantName)
		})
	}
}
