package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// commitRepo makes a temp repo whose spec was committed the way acta id does
// it: one chore commit holding only that file.
func commitRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".acta", "specs"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "README.md"), []byte("readme\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "init", "-q", "-b", "main")
	gitOut(t, dir, "config", "user.name", "test")
	gitOut(t, dir, "config", "user.email", "test@example.com")
	gitOut(t, dir, "add", "README.md")
	gitOut(t, dir, "commit", "-q", "-m", "init")
	writeCommitSpec(t, dir, "one\n")
	gitOut(t, dir, "add", ".acta/specs/x.md")
	gitOut(t, dir, "commit", "-q", "-m", "chore(spec): assign short ids")
	return dir
}

func writeCommitSpec(t *testing.T, dir, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, ".acta", "specs", "x.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// runCommit runs acta commit inside dir with a private cache folder.
func runCommit(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	cache := t.TempDir()
	t.Setenv("HOME", cache)
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"commit"}, args...), strings.NewReader(""), false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestCommitFoldsIntoThePlanningCommit(t *testing.T) {
	dir := commitRepo(t)
	before := commitCount(t, dir)
	writeCommitSpec(t, dir, "two\n")
	code, stdout, stderr := runCommit(t, dir, ".acta/specs/x.md", "-m", "chore(spec): fix wording")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if got := strings.TrimSpace(stdout); got != ".acta/specs/x.md" {
		t.Errorf("stdout = %q, want the path", got)
	}
	if got := commitCount(t, dir); got != before {
		t.Errorf("commit count = %d, want %d (folded)", got, before)
	}
	body := gitOut(t, dir, "log", "-1", "--format=%B")
	for _, want := range []string{"chore(spec): assign short ids", "chore(spec): fix wording"} {
		if !strings.Contains(body, want) {
			t.Errorf("commit message %q lacks %q", body, want)
		}
	}
	if st := gitOut(t, dir, "status", "--porcelain"); strings.TrimSpace(st) != "" {
		t.Errorf("tree not clean: %q", st)
	}
}

func TestCommitRefusesBadInput(t *testing.T) {
	dir := commitRepo(t)
	writeCommitSpec(t, dir, "two\n")
	outside := filepath.Join(dir, "README.md")
	if err := os.WriteFile(outside, []byte("changed\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		name string
		args []string
	}{
		{"outside the planning root", []string{"README.md", "-m", "chore(spec): x"}},
		{"outside by dot dot", []string{".acta/specs/../../README.md", "-m", "chore(spec): x"}},
		{"missing path", []string{".acta/specs/nope.md", "-m", "chore(spec): x"}},
		{"empty message", []string{".acta/specs/x.md", "-m", ""}},
		{"no message", []string{".acta/specs/x.md"}},
		{"no path", []string{"-m", "chore(spec): x"}},
		{"two paths", []string{".acta/specs/x.md", "README.md", "-m", "chore(spec): x"}},
	}
	before := commitCount(t, dir)
	for _, tc := range cases {
		code, _, stderr := runCommit(t, dir, tc.args...)
		if code != exitBadInput {
			t.Errorf("%s: exit %d, want %d", tc.name, code, exitBadInput)
		}
		if strings.TrimSpace(stderr) == "" {
			t.Errorf("%s: no error text", tc.name)
		}
		if got := commitCount(t, dir); got != before {
			t.Errorf("%s: commit count %d, want %d", tc.name, got, before)
		}
	}
}

func TestCommitSkippedWhenAutoCommitOff(t *testing.T) {
	dir := commitRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".acta.yaml"), []byte("auto_commit: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	writeCommitSpec(t, dir, "two\n")
	before := commitCount(t, dir)
	code, _, stderr := runCommit(t, dir, ".acta/specs/x.md", "-m", "chore(spec): fix wording")
	if code != exitSkipped {
		t.Fatalf("exit %d, want %d", code, exitSkipped)
	}
	if !strings.Contains(stderr, "auto_commit is off") {
		t.Errorf("stderr %q lacks the reason", stderr)
	}
	if got := commitCount(t, dir); got != before {
		t.Errorf("commit count = %d, want %d", got, before)
	}
}
