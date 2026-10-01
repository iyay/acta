package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The hook reads HERDR_ENV, so the session text can name a herdr tab as the
// extra way to run a second brainstorm. Without it, the text stays silent.
// The phrase belongs to the herdr block only; the skill index has its own line
// about the dispatch skill.
const herdrLine = "this session runs in a herdr tab"

func TestHookSessionStartNamesHerdrOnlyInsideHerdr(t *testing.T) {
	t.Chdir(hookRepo(t))
	t.Setenv("PM_VOICE_FILE", "")

	// Unset, not empty: an unset variable must not count as herdr either.
	saved, had := os.LookupEnv("HERDR_ENV")
	if err := os.Unsetenv("HERDR_ENV"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if had {
			_ = os.Setenv("HERDR_ENV", saved)
		}
	})

	if _, out, _ := runHook(t, "session-start", ""); strings.Contains(out, herdrLine) {
		t.Errorf("session start outside herdr offers a herdr tab:\n%s", out)
	}
	t.Setenv("HERDR_ENV", "1")
	if _, out, _ := runHook(t, "session-start", ""); !strings.Contains(out, herdrLine) {
		t.Errorf("herdr session start missing the herdr tab sentence:\n%s", out)
	}
	t.Setenv("HERDR_ENV", "0")
	if _, out, _ := runHook(t, "session-start", ""); strings.Contains(out, herdrLine) {
		t.Errorf("HERDR_ENV=0 counts as herdr:\n%s", out)
	}
}

// TestHookPreToolBlocksBareGoTest runs the real command line in a repo that
// keeps its tests in scripts/test, which is the one place a bare go test is
// stopped. A repo without that script is left alone, and a broken hook must
// never stand in the way of any Bash call.
func TestHookPreToolBlocksBareGoTest(t *testing.T) {
	withScript := goTestRepoCLI(t)
	code, out, errb := runGoTestHook(t, withScript, "go test ./... -count=1")
	if code != exitBlock {
		t.Fatalf("a bare go test: exit %d, want %d", code, exitBlock)
	}
	if out != "" {
		t.Fatalf("the block message must go to stderr only, stdout was %q", out)
	}
	want := "acta: run tests with scripts/test, not go test. It adds -short and the machine-wide lock. Use: scripts/test ./..."
	if !strings.Contains(errb, want) {
		t.Fatalf("stderr %q, want it to end in %q", errb, want)
	}

	// Without the script there is no lock to join, so the command runs.
	noScript := hookRepo(t)
	if code, out, errb := runGoTestHook(t, noScript, "go test ./..."); code != exitOK || out != "" || errb != "" {
		t.Fatalf("a repo with no scripts/test: exit %d, stdout %q, stderr %q, want exit 0 and nothing printed", code, out, errb)
	}

	// A repo with a scripts/test but no planning root is still a repo the
	// block applies to: the script is what matters, not the .acta folder.
	noPlanning := t.TempDir()
	if err := os.MkdirAll(filepath.Join(noPlanning, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(noPlanning, "scripts", "test"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", noPlanning, "init", "-q").CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if code, _, errb := runGoTestHook(t, noPlanning, "go test ./..."); code != exitBlock {
		t.Fatalf("a bare go test with no planning root: exit %d, stderr %q, want %d", code, errb, exitBlock)
	}
}

// goTestRepoCLI is a git repo whose git top-level holds scripts/test.
func goTestRepoCLI(t *testing.T) string {
	t.Helper()
	dir := hookRepo(t)
	if err := os.MkdirAll(filepath.Join(dir, "scripts"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "scripts", "test"), []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runGoTestHook(t *testing.T, dir, command string) (int, string, string) {
	t.Helper()
	t.Chdir(dir)
	return runHook(t, "pre-tool", hookEvent("s1", command))
}
