package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// flakeRun runs acta flake in dir and gives back the exit code and both
// streams, so each case can check all three.
func flakeRun(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	var code int
	inDir(t, dir, func() {
		code = Run(append([]string{"flake"}, args...), strings.NewReader(""), false, &stdout, &stderr)
	})
	return code, stdout.String(), stderr.String()
}

// A first record files the bug and prints its id, title and flake count on
// one line.
func TestFlakeRecordPrintsBugLine(t *testing.T) {
	fixDay(t, "2026-10-31")
	dir := debtRepo(t)
	code, out, errOut := flakeRun(t, dir, "record", "internal/tui.TestClick", "--plan", "PLAN-3")
	if code != exitOK || errOut != "" {
		t.Fatalf("exit %d stderr %q", code, errOut)
	}
	want := regexp.MustCompile(`^BUG-\d{4} Flaky test internal/tui\.TestClick \(1 flakes\)\n$`)
	if !want.MatchString(out) {
		t.Fatalf("stdout %q does not match %s", out, want)
	}
}

// A record onto a bug that is already open prints that bug's own title, not
// the title a new flake bug would get.
func TestFlakeRecordKeepsOldBugTitle(t *testing.T) {
	fixDay(t, "2026-10-31")
	dir := debtRepo(t)
	bug := filepath.Join(dir, ".pm", "bugs", "2026-10-01-click.md")
	if err := os.MkdirAll(filepath.Dir(bug), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bug, []byte("---\nid: BUG-0007\nflaky_tests: [internal/tui.TestClick]\n---\n# Click test is shaky\n\n## Symptom\nIt flakes.\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "."}, {"commit", "-q", "-m", "bug"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	code, out, errOut := flakeRun(t, dir, "record", "internal/tui.TestClick")
	if code != exitOK || errOut != "" {
		t.Fatalf("exit %d stderr %q", code, errOut)
	}
	if want := "BUG-0007 Click test is shaky (1 flakes)\n"; out != want {
		t.Fatalf("stdout %q want %q", out, want)
	}
}

// A refusal from the write step reaches stderr with the bad-input exit and
// prints nothing on stdout.
func TestFlakeRecordRefusalGoesToStderr(t *testing.T) {
	dir := debtRepo(t)
	code, out, errOut := flakeRun(t, dir, "record", "internal/tui.TestClick", "--plan", "PLAN-404")
	if code != exitBadInput || out != "" || !strings.Contains(errOut, "unknown id PLAN-404") {
		t.Fatalf("exit %d stdout %q stderr %q", code, out, errOut)
	}
}

// A call with no name, or with a subcommand acta does not know, prints the
// usage line.
func TestFlakeBadCallPrintsUsage(t *testing.T) {
	dir := debtRepo(t)
	for _, args := range [][]string{{}, {"record"}, {"bogus", "x"}} {
		code, out, errOut := flakeRun(t, dir, args...)
		if code != exitBadInput || out != "" || errOut != usageLine("flake")+"\n" {
			t.Errorf("%v: exit %d stdout %q stderr %q", args, code, out, errOut)
		}
	}
}

// acta help flake shows the entry, and the entry sits in the Write group.
func TestHelpFlakeShowsEntry(t *testing.T) {
	code, out, errOut := helpRun("help", "flake")
	if code != exitOK || errOut != "" || !strings.Contains(out, "acta flake record <name>") {
		t.Fatalf("exit %d stdout %q stderr %q", code, out, errOut)
	}
	if c, _ := findHelpCmd("flake"); c.group != "Write" {
		t.Fatalf("flake group %q want Write", c.group)
	}
}
