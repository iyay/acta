package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// migrateRepo makes a git repo with .pm/ holding one tracked file.
func migrateRepo(t *testing.T) string {
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
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	run("init", "-q", "-b", "main")
	if err := os.MkdirAll(filepath.Join(dir, ".pm"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".pm", "note.md"), []byte("hi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", ".")
	run("commit", "-q", "-m", "init")
	return dir
}

func gitLog(t *testing.T, dir string) []string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "log", "--format=%s").CombinedOutput()
	if err != nil {
		t.Fatalf("git log: %v %s", err, out)
	}
	var msgs []string
	for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		if l != "" {
			msgs = append(msgs, l)
		}
	}
	return msgs
}

func TestMigrateRootHappyPath(t *testing.T) {
	dir := migrateRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".pm.yaml"), []byte("root: .pm\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "add", ".").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v %s", err, out)
	}
	if out, err := exec.Command("git", "-C", dir, "commit", "-q", "-m", "add yaml").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
	before := len(gitLog(t, dir))

	var stdout, stderr strings.Builder
	code := migrateRoot(dir, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("exit %d stderr %q", code, stderr.String())
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta", "note.md")); err != nil {
		t.Fatalf(".acta/note.md missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".pm")); !os.IsNotExist(err) {
		t.Fatalf(".pm still exists")
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta.yaml")); err != nil {
		t.Fatalf(".acta.yaml missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, ".pm.yaml")); !os.IsNotExist(err) {
		t.Fatalf(".pm.yaml still exists")
	}
	// The move commit must hold the renames and nothing else.
	want := map[string]bool{".pm/note.md": true, ".acta/note.md": true, ".pm.yaml": true, ".acta.yaml": true}
	for _, p := range commitPaths(t, dir) {
		if !want[p] {
			t.Fatalf("commit touches %q, want only the move", p)
		}
	}
	if got := gitRaw(t, dir, "HEAD", ".acta.yaml"); strings.Contains(got, "root:") {
		t.Fatalf(".acta.yaml keeps a root line: %q", got)
	}
	msgs := gitLog(t, dir)
	if len(msgs) != before+1 {
		t.Fatalf("want one new commit, log is %v", msgs)
	}
	if msgs[0] != "acta: move root folder .pm to .acta" {
		t.Fatalf("commit message %q", msgs[0])
	}
	if out := stdout.String(); !strings.Contains(out, ".pm") || !strings.Contains(out, ".acta") {
		t.Fatalf("output %q must name what moved", out)
	}
}

func TestMigrateRootRefusesWhenActaExists(t *testing.T) {
	dir := migrateRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, ".acta"), 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	} else if !strings.Contains(stderr.String(), ".acta") {
		t.Fatalf("stderr %q must name .acta/", stderr.String())
	}
}

func TestMigrateRootRefusesWhenPmMissing(t *testing.T) {
	dir := migrateRepo(t)
	if out, err := exec.Command("git", "-C", dir, "rm", "-q", "-r", ".pm").CombinedOutput(); err != nil {
		t.Fatalf("git rm: %v %s", err, out)
	}
	if out, err := exec.Command("git", "-C", dir, "commit", "-q", "-m", "drop pm").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", err, out)
	}
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	} else if !strings.Contains(stderr.String(), ".pm") {
		t.Fatalf("stderr %q must name .pm/", stderr.String())
	}
}

func TestMigrateRootRefusesDirtyTracked(t *testing.T) {
	dir := migrateRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".pm", "note.md"), []byte("dirty\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
}

func TestMigrateRootRefusesUntracked(t *testing.T) {
	dir := migrateRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".pm", "new.md"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	if code := migrateRoot(dir, &stdout, &stderr); code != 1 {
		t.Fatalf("exit %d, want 1", code)
	}
}

func TestMigrateRootOutsideGit(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".pm"), 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr strings.Builder
	// A missing git repo must be a message, never a panic.
	func() {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("panicked: %v", r)
			}
		}()
		if code := migrateRoot(dir, &stdout, &stderr); code == 0 {
			t.Fatal("want non-zero exit outside git")
		}
	}()
	if stderr.String() == "" {
		t.Fatal("want a message on stderr")
	}
}
