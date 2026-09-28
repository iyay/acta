package cli

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

const dispatchPlan = "---\nid: PLAN-1\nhash: aaaa\n---\n# P\n\n### Task 1: A\n- [ ] a\n\n### Task 2: B\n- [ ] b\n"

// dispatchRepo makes a git repo with a planning root holding one plan with two
// tasks, so dispatch init has a base, a branch and a plan on the board.
func dispatchRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	t.Setenv("GIT_AUTHOR_NAME", "test")
	t.Setenv("GIT_COMMITTER_NAME", "test")
	t.Setenv("GIT_AUTHOR_EMAIL", "test@example.com")
	t.Setenv("GIT_COMMITTER_EMAIL", "test@example.com")
	plan := filepath.Join(dir, ".acta", "plans", "2026-09-29-p.md")
	if err := os.MkdirAll(filepath.Dir(plan), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan, []byte(dispatchPlan), 0o644); err != nil {
		t.Fatal(err)
	}
	run("init", "-q", "-b", "main")
	run("add", ".")
	run("commit", "-q", "-m", "init")
	return dir
}

// storedRecord mirrors the record file on disk. It is declared here so the
// first test run fails on the missing command, not on a missing type.
type storedRecord struct {
	Pane  string `json:"pane"`
	Base  string `json:"base"`
	Plan  string `json:"plan"`
	Round string `json:"round"`
}

func runDispatchInit(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"dispatch", "init"}, args...), nil, false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func dispatchGitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return strings.TrimSpace(string(out))
}

// TestCheckRecordCoversEveryPlanPath calls the checks straight, because a
// record is untrusted input that later commands read from disk.
func TestCheckRecordCoversEveryPlanPath(t *testing.T) {
	dir := dispatchRepo(t)
	t.Chdir(dir)
	cfg, b, code := loadBoard("", io.Discard)
	if code != exitOK {
		t.Fatalf("load exit %d", code)
	}
	good := dispatchRecord{Pane: "wM:pH", Base: dispatchGitOut(t, dir, "rev-parse", "HEAD"),
		Plan: ".acta/plans/2026-09-29-p.md", Round: "main"}
	it, err := checkRecord(cfg, b, good)
	if err != nil || it == nil || it.ID != "plans/2026-09-29-p" {
		t.Fatalf("good record: item %v err %v", it, err)
	}
	outside := []string{
		"../outside.md",
		filepath.Join(dir, ".acta", "plans", "2026-09-29-p.md"),
		".acta/debt/2026-09-29-d.md",
		"",
	}
	for _, plan := range outside {
		bad := good
		bad.Plan = plan
		if it, err := checkRecord(cfg, b, bad); err == nil {
			t.Fatalf("plan %q passed the checks, item %v", plan, it)
		}
	}
	for _, bad := range []dispatchRecord{
		{Pane: "", Base: good.Base, Plan: good.Plan, Round: good.Round},
		{Pane: "wM:pH; rm", Base: good.Base, Plan: good.Plan, Round: good.Round},
		{Pane: good.Pane, Base: "abc1234", Plan: good.Plan, Round: good.Round},
		{Pane: good.Pane, Base: strings.ToUpper(good.Base), Plan: good.Plan, Round: good.Round},
		{Pane: good.Pane, Base: good.Base, Plan: good.Plan, Round: "Bad Round"},
		{Pane: good.Pane, Base: good.Base, Plan: good.Plan, Round: ""},
	} {
		if it, err := checkRecord(cfg, b, bad); err == nil {
			t.Fatalf("record %+v passed the checks, item %v", bad, it)
		}
	}
}

func readRecord(t *testing.T, dir string) storedRecord {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, ".acta", ".dispatch.json"))
	if err != nil {
		t.Fatal(err)
	}
	var r storedRecord
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatal(err)
	}
	return r
}

func TestDispatchInitWritesHeadAndBranch(t *testing.T) {
	dir := dispatchRepo(t)
	code, stdout, stderr := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	want := dispatchGitOut(t, dir, "rev-parse", "HEAD")
	r := readRecord(t, dir)
	if r.Pane != "wM:pH" || r.Base != want || r.Plan != ".acta/plans/2026-09-29-p.md" || r.Round != "main" {
		t.Fatalf("record %+v, want base %s", r, want)
	}
	if !strings.Contains(stdout, ".dispatch.json") {
		t.Fatalf("stdout %q does not name the record", stdout)
	}
	if got := read(t, filepath.Join(dir, ".acta", ".gitignore")); !strings.Contains(got, ".dispatch.json") {
		t.Fatalf("gitignore %q has no .dispatch.json line", got)
	}
	// The command must not commit anything.
	if n := dispatchGitOut(t, dir, "rev-list", "--count", "HEAD"); n != "1" {
		t.Fatalf("commit count %s, want 1", n)
	}
}

func TestDispatchInitSecondRunMovesBase(t *testing.T) {
	dir := dispatchRepo(t)
	if code, _, stderr := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md"); code != exitOK {
		t.Fatalf("first run exit %d, stderr %q", code, stderr)
	}
	note := filepath.Join(dir, "note.md")
	if err := os.WriteFile(note, []byte("work\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "."}, {"commit", "-q", "-m", "work"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	if code, _, stderr := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md", "--round", "fix-1"); code != exitOK {
		t.Fatalf("second run exit %d, stderr %q", code, stderr)
	}
	r := readRecord(t, dir)
	if r.Base != dispatchGitOut(t, dir, "rev-parse", "HEAD") {
		t.Fatalf("base %s did not move to the new head", r.Base)
	}
	if r.Round != "fix-1" {
		t.Fatalf("round %q, want fix-1", r.Round)
	}
}

func TestDispatchInitRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		args []string
	}{
		{"empty pane", []string{"--pane", "", "--plan", ".acta/plans/2026-09-29-p.md"}},
		{"empty plan", []string{"--pane", "wM:pH", "--plan", ""}},
		{"pane with shell text", []string{"--pane", "wM:pH; rm", "--plan", ".acta/plans/2026-09-29-p.md"}},
		{"pane without colon", []string{"--pane", "wMpH", "--plan", ".acta/plans/2026-09-29-p.md"}},
		{"plan outside the root", []string{"--pane", "wM:pH", "--plan", "../outside.md"}},
		{"plan not on the board", []string{"--pane", "wM:pH", "--plan", ".acta/plans/missing.md"}},
		{"round with a space", []string{"--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md", "--round", "Bad Round"}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := dispatchRepo(t)
			code, _, stderr := runDispatchInit(t, dir, c.args...)
			if code != exitBadInput {
				t.Fatalf("exit %d, want %d", code, exitBadInput)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Fatal("no line on stderr")
			}
			if _, err := os.Stat(filepath.Join(dir, ".acta", ".dispatch.json")); !os.IsNotExist(err) {
				t.Fatalf("a record was written anyway: %v", err)
			}
		})
	}
}

// TestDispatchUnknownSubcommandFails covers the guard in Run, so a typo never
// runs init with no flags and writes an empty record.
func TestDispatchUnknownSubcommandFails(t *testing.T) {
	dir := dispatchRepo(t)
	t.Chdir(dir)
	for _, args := range [][]string{{"dispatch"}, {"dispatch", "start"}} {
		var stdout, stderr strings.Builder
		if code := Run(args, nil, false, &stdout, &stderr); code != exitBadInput {
			t.Fatalf("%v: exit %d, want %d", args, code, exitBadInput)
		}
		if !strings.Contains(stderr.String(), "usage: acta dispatch init") {
			t.Fatalf("%v: stderr %q has no usage line", args, stderr.String())
		}
		if _, err := os.Stat(filepath.Join(dir, ".acta", ".dispatch.json")); !os.IsNotExist(err) {
			t.Fatalf("%v: a record was written anyway: %v", args, err)
		}
	}
}

func TestDispatchInitOutsideGitFails(t *testing.T) {
	dir := t.TempDir()
	plan := filepath.Join(dir, ".acta", "plans", "2026-09-29-p.md")
	if err := os.MkdirAll(filepath.Dir(plan), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan, []byte(dispatchPlan), 0o644); err != nil {
		t.Fatal(err)
	}
	code, _, stderr := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md")
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d", code, exitBadInput)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Fatal("no line on stderr")
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta", ".dispatch.json")); !os.IsNotExist(err) {
		t.Fatalf("a record was written anyway: %v", err)
	}
}
