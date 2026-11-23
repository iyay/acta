package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// tidyGit runs one git command in dir and fails the test when it fails.
func tidyGit(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// tidyRepo makes a repo with a base commit on main and a branch holding a
// code commit plus a chore commit on top of it.
func tidyRepo(t *testing.T) string {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Setenv("GIT_AUTHOR_NAME", "T")
	t.Setenv("GIT_AUTHOR_EMAIL", "t@example.com")
	t.Setenv("GIT_COMMITTER_NAME", "T")
	t.Setenv("GIT_COMMITTER_EMAIL", "t@example.com")
	dir := t.TempDir()
	tidyGit(t, dir, "init", "-q", "-b", "main")
	add := func(name, msg string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
		tidyGit(t, dir, "add", name)
		tidyGit(t, dir, "commit", "-q", "-m", msg)
	}
	add("base.txt", "chore: base")
	tidyGit(t, dir, "checkout", "-q", "-b", "work")
	add("a.txt", "feat: a")
	add("b.txt", "chore: b")
	return dir
}

func runTidy(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"tidy"}, args...), nil, false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func TestTidySuccessPrintsLine(t *testing.T) {
	dir := tidyRepo(t)
	code, out, errOut := runTidy(t, dir, "main", "work")
	if code != exitOK {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	want := "tidy: 2 commits -> 1, tree ok, parent " + tidyGit(t, dir, "rev-parse", "--short=7", "main") + ", folded 0\n"
	if out != want {
		t.Fatalf("stdout %q", out)
	}
	if tidyGit(t, dir, "rev-parse", "--verify", "refs/acta/tidy/work") == "" {
		t.Fatal("no tidy ref written")
	}
}

func TestTidyBadArgCount(t *testing.T) {
	dir := tidyRepo(t)
	for _, args := range [][]string{{}, {"main"}, {"main", "work", "extra"}} {
		code, out, errOut := runTidy(t, dir, args...)
		if code != exitBadInput || out != "" || !strings.Contains(errOut, "usage: acta tidy") {
			t.Fatalf("args %v: code %d, stdout %q, stderr %q", args, code, out, errOut)
		}
	}
}

func TestTidyUnknownRefLeavesRefsAlone(t *testing.T) {
	dir := tidyRepo(t)
	before := tidyGit(t, dir, "for-each-ref")
	code, out, errOut := runTidy(t, dir, "main", "nope")
	if code == exitOK || out != "" || errOut == "" {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errOut)
	}
	if after := tidyGit(t, dir, "for-each-ref"); after != before {
		t.Fatalf("refs changed:\n%s\n---\n%s", before, after)
	}
	// --plan reaches tidy: a hash no plan holds stops it the same way.
	code, out, errOut = runTidy(t, dir, "main", "work", "--plan", "zzzz999")
	if code == exitOK || out != "" || !strings.Contains(errOut, "zzzz999") {
		t.Fatalf("--plan: code %d, stdout %q, stderr %q", code, out, errOut)
	}
	if after := tidyGit(t, dir, "for-each-ref"); after != before {
		t.Fatalf("--plan: refs changed:\n%s\n---\n%s", before, after)
	}
}

func TestTidyNoCommitsOnTopFails(t *testing.T) {
	dir := tidyRepo(t)
	before := tidyGit(t, dir, "for-each-ref")
	code, out, errOut := runTidy(t, dir, "work", "work")
	if code == exitOK || out != "" || !strings.Contains(errOut, "no commits") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errOut)
	}
	if after := tidyGit(t, dir, "for-each-ref"); after != before {
		t.Fatal("refs changed")
	}
}

func TestTidyOntoFlagAccepted(t *testing.T) {
	dir := tidyRepo(t)
	code, out, errOut := runTidy(t, dir, "main", "work", "--onto", "main")
	if code != exitOK || !strings.HasPrefix(out, "tidy: ") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errOut)
	}
}

// Tidy mode lands a plan with tidy --plan and never runs land-notes, so a
// tidy with --plan that works counts the landed plan. A failed one counts
// nothing.
func TestTidyPlanCountsALandedPlan(t *testing.T) {
	dir := landNotesRepo(t, "BUG-0001, DBT-0001")
	s, _ := runEnv(t, false)
	// main^1 is main before the merge, so build still has its plan on top.
	if code, _, errOut := runTidy(t, dir, "main^1", "build", "--plan", "zzzz999"); code == exitOK {
		t.Fatalf("a plan no file holds passed: %q", errOut)
	}
	if got := todayCounts(t, s).Workflow; len(got) != 0 {
		t.Fatalf("a failed tidy counted %v", got)
	}
	if code, _, errOut := runTidy(t, dir, "main^1", "build", "--plan", "abcd123"); code != exitOK {
		t.Fatalf("code %d, stderr %q", code, errOut)
	}
	if got := todayCounts(t, s).Workflow; got["plan_landed"] != 1 || len(got) != 1 {
		t.Fatalf("workflow %v, want plan_landed 1", got)
	}
	// Land reruns tidy after a FLAKY line while the tidy ref still stands.
	if code, _, errOut := runTidy(t, dir, "main^1", "build", "--plan", "abcd123"); code != exitOK {
		t.Fatalf("rerun: code %d, stderr %q", code, errOut)
	}
	if got := todayCounts(t, s).Workflow; got["plan_landed"] != 1 {
		t.Fatalf("a rerun counted the land again: %v", got)
	}
	// Land deletes the ref at the end, so the next land counts again.
	tidyGit(t, dir, "update-ref", "-d", "refs/acta/tidy/build")
	if code, _, errOut := runTidy(t, dir, "main^1", "build", "--plan", "abcd123"); code != exitOK {
		t.Fatalf("new land: code %d, stderr %q", code, errOut)
	}
	if got := todayCounts(t, s).Workflow; got["plan_landed"] != 2 {
		t.Fatalf("workflow %v, want plan_landed 2 after a new land", got)
	}
}
