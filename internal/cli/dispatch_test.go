package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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
	plan := filepath.Join(dir, ".acta", "plans", "2026-09-29-p.md")
	if err := os.MkdirAll(filepath.Dir(plan), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(plan, []byte(dispatchPlan), 0o644); err != nil {
		t.Fatal(err)
	}
	run("init", "-q", "-b", "main")
	// Name the author inside the repo, not in the env, so tests that make
	// commits can still run side by side.
	run("config", "user.name", "test")
	run("config", "user.email", "test@example.com")
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

// fakeHerdr writes a stand-in herdr into a fresh temp folder and points PATH
// at it, so a test can read back exactly what the command wanted to send.
func fakeHerdr(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	script := "#!/bin/sh\n" +
		"printf '%s\\n' \"$@\" >> \"$HERDR_LOG\"\n" +
		"if [ \"${HERDR_EXIT:-0}\" != 0 ]; then echo \"herdr says no\" >&2; fi\n" +
		"exit ${HERDR_EXIT:-0}\n"
	if err := os.WriteFile(filepath.Join(dir, "herdr"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", dir+string(os.PathListSeparator)+os.Getenv("PATH"))
	t.Setenv("HERDR_LOG", filepath.Join(dir, "herdr.log"))
	t.Setenv("HERDR_EXIT", "0")
	t.Setenv("HERDR_PANE_ID", "wM:p9")
	return filepath.Join(dir, "herdr.log")
}

func runReplyBack(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"reply-back"}, args...), nil, false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// initRecord runs dispatch init, so reply-back starts from a record the
// orchestrator really wrote.
func initRecord(t *testing.T, dir string) storedRecord {
	t.Helper()
	code, _, stderr := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md")
	if code != exitOK {
		t.Fatalf("dispatch init exit %d, stderr %q", code, stderr)
	}
	return readRecord(t, dir)
}

// tickAllTasks ticks both boxes in the plan file and commits, so the board
// reads the plan as done and head moves past base. It returns the new head.
func tickAllTasks(t *testing.T, dir string) string {
	t.Helper()
	plan := filepath.Join(dir, ".acta", "plans", "2026-09-29-p.md")
	if err := os.WriteFile(plan, []byte(strings.ReplaceAll(dispatchPlan, "- [ ]", "- [x]")), 0o644); err != nil {
		t.Fatal(err)
	}
	for _, args := range [][]string{{"add", "-A"}, {"commit", "-q", "-m", "done"}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	return dispatchGitOut(t, dir, "rev-parse", "HEAD")
}

// wantNoSend fails when the herdr log exists, because every refusing path must
// leave the orchestrator pane alone.
func wantNoSend(t *testing.T, log string) {
	t.Helper()
	if _, err := os.Stat(log); !os.IsNotExist(err) {
		t.Fatalf("herdr was called anyway: %v, log %q", err, read(t, log))
	}
}

func herdrLog(t *testing.T, log string) []string {
	t.Helper()
	return strings.Split(strings.TrimRight(read(t, log), "\n"), "\n")
}

func TestReplyBackRefusesOpenTasks(t *testing.T) {
	dir := dispatchRepo(t)
	log := fakeHerdr(t)
	initRecord(t, dir)
	code, _, stderr := runReplyBack(t, dir)
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d", code, exitBadInput)
	}
	for _, want := range []string{"plans/2026-09-29-p#task-1", "#task-2"} {
		if !strings.Contains(stderr, want) {
			t.Errorf("stderr %q does not name %s", stderr, want)
		}
	}
	wantNoSend(t, log)
}

func TestReplyBackSendsReviewWhenDone(t *testing.T) {
	dir := dispatchRepo(t)
	log := fakeHerdr(t)
	r := initRecord(t, dir)
	head := tickAllTasks(t, dir)
	code, stdout, stderr := runReplyBack(t, dir)
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	want := []string{"agent", "prompt", "wM:pH", fmt.Sprintf(
		"/acta:review %s..%s - plan .acta/plans/2026-09-29-p.md, round main, pane wM:p9", r.Base, head)}
	if got := herdrLog(t, log); !slices.Equal(got, want) {
		t.Fatalf("herdr got %q, want %q", got, want)
	}
	if !strings.Contains(stdout, "sent to wM:pH") {
		t.Errorf("stdout %q does not say what was sent", stdout)
	}
}

func TestReplyBackBlockedSendsProse(t *testing.T) {
	dir := dispatchRepo(t)
	log := fakeHerdr(t)
	initRecord(t, dir)
	code, _, stderr := runReplyBack(t, dir, "--blocked", "tests need a db")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	got := herdrLog(t, log)
	if len(got) != 4 || got[3] != "main blocked: tests need a db - pane wM:p9" {
		t.Fatalf("herdr got %q, want the blocked prose", got)
	}
}

// replyRecordJSON builds a record body by hand, so a test can hand reply-back
// a record the orchestrator would never write.
func replyRecordJSON(pane, base, plan string) string {
	return fmt.Sprintf(`{"pane":%q,"base":%q,"plan":%q,"round":"main"}`, pane, base, plan)
}

func TestReplyBackRejectsBadRecords(t *testing.T) {
	full := strings.Repeat("a", 40)
	cases := []struct {
		name string
		gone bool
		body string
	}{
		{"no record", true, ""},
		{"broken json", false, "{"},
		{"pane with shell text", false, replyRecordJSON("wM:pH; rm", full, ".acta/plans/2026-09-29-p.md")},
		{"base too short", false, replyRecordJSON("wM:pH", "1234567", ".acta/plans/2026-09-29-p.md")},
		{"plan outside the root", false, replyRecordJSON("wM:pH", full, "../x.md")},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := dispatchRepo(t)
			log := fakeHerdr(t)
			initRecord(t, dir)
			path := filepath.Join(dir, ".acta", ".dispatch.json")
			if c.gone {
				if err := os.Remove(path); err != nil {
					t.Fatal(err)
				}
			} else if err := os.WriteFile(path, []byte(c.body), 0o644); err != nil {
				t.Fatal(err)
			}
			code, _, stderr := runReplyBack(t, dir)
			if code != exitBadInput {
				t.Fatalf("exit %d, want %d", code, exitBadInput)
			}
			if strings.TrimSpace(stderr) == "" {
				t.Error("no line on stderr")
			}
			wantNoSend(t, log)
		})
	}
}

func TestReplyBackEmptyBlockedFails(t *testing.T) {
	dir := dispatchRepo(t)
	log := fakeHerdr(t)
	initRecord(t, dir)
	tickAllTasks(t, dir)
	code, _, stderr := runReplyBack(t, dir, "--blocked", "")
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d", code, exitBadInput)
	}
	if strings.TrimSpace(stderr) == "" {
		t.Error("no line on stderr")
	}
	wantNoSend(t, log)
}

func TestReplyBackHerdrFailureExits3(t *testing.T) {
	t.Run("herdr exits non-zero", func(t *testing.T) {
		dir := dispatchRepo(t)
		fakeHerdr(t)
		initRecord(t, dir)
		tickAllTasks(t, dir)
		t.Setenv("HERDR_EXIT", "2")
		code, _, stderr := runReplyBack(t, dir)
		if code != exitOther {
			t.Fatalf("exit %d, want %d", code, exitOther)
		}
		if !strings.Contains(stderr, "herdr says no") {
			t.Errorf("stderr %q does not show what herdr said", stderr)
		}
	})
	t.Run("herdr missing from PATH", func(t *testing.T) {
		dir := dispatchRepo(t)
		fakeHerdr(t)
		initRecord(t, dir)
		tickAllTasks(t, dir)
		t.Setenv("PATH", t.TempDir())
		if code, _, _ := runReplyBack(t, dir); code != exitOther {
			t.Fatalf("exit %d, want %d", code, exitOther)
		}
	})
}

// TestReplyBackOwnPaneFallsBackToUnknown covers the two pane values that are
// not a pane at all: a missing one and one carrying shell text.
func TestReplyBackOwnPaneFallsBackToUnknown(t *testing.T) {
	cases := []struct {
		name string
		pane string
		none bool
	}{
		{"pane not set", "", true},
		{"pane with shell text", "wM:pH; rm", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := dispatchRepo(t)
			log := fakeHerdr(t)
			initRecord(t, dir)
			tickAllTasks(t, dir)
			// t.Setenv first, so the test cleanup still restores the real value.
			t.Setenv("HERDR_PANE_ID", c.pane)
			if c.none {
				if err := os.Unsetenv("HERDR_PANE_ID"); err != nil {
					t.Fatal(err)
				}
			}
			if code, _, stderr := runReplyBack(t, dir); code != exitOK {
				t.Fatalf("exit %d, stderr %q", code, stderr)
			}
			got := herdrLog(t, log)
			if len(got) != 4 || !strings.HasSuffix(got[3], "pane unknown") {
				t.Fatalf("herdr got %q, want the last line to end in pane unknown", got)
			}
		})
	}
}

// TestRoundFromBranch covers the slug rule straight, because git refuses
// many of these names as real branches.
func TestRoundFromBranch(t *testing.T) {
	t.Parallel()

	long := strings.Repeat("a", 63) + "-bcd"
	for _, c := range []struct {
		in, want string
		ok       bool
	}{
		{"main", "main", true},
		{"feat/X", "feat-x", true},
		{"--a--b--", "a-b", true},
		{"HEAD", "head", true},
		{strings.Repeat("a", 70), strings.Repeat("a", 64), true},
		{long, strings.Repeat("a", 63), true},
		{"///", "", false},
		{"", "", false},
	} {
		got, ok := roundFromBranch(c.in)
		if got != c.want || ok != c.ok {
			t.Errorf("roundFromBranch(%q) = %q, %v; want %q, %v", c.in, got, ok, c.want, c.ok)
		}
	}
}

func TestDispatchInitSlugsTheBranch(t *testing.T) {
	dir := dispatchRepo(t)
	dispatchGitOut(t, dir, "checkout", "-q", "-b", "feat/X")
	code, _, stderr := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if r := readRecord(t, dir); r.Round != "feat-x" {
		t.Fatalf("round %q, want feat-x", r.Round)
	}
}

func TestDispatchInitDetachedHeadIsHead(t *testing.T) {
	dir := dispatchRepo(t)
	dispatchGitOut(t, dir, "checkout", "-q", "--detach")
	code, _, stderr := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	if r := readRecord(t, dir); r.Round != "head" {
		t.Fatalf("round %q, want head", r.Round)
	}
}

func TestDispatchInitExplicitRoundIsNotRewritten(t *testing.T) {
	dir := dispatchRepo(t)
	code, _, _ := runDispatchInit(t, dir, "--pane", "wM:pH", "--plan", ".acta/plans/2026-09-29-p.md", "--round", "Feat-X")
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d for an upper-case --round", code, exitBadInput)
	}
	if _, err := os.Stat(filepath.Join(dir, ".acta", ".dispatch.json")); !os.IsNotExist(err) {
		t.Fatalf("record written for a bad --round: %v", err)
	}
}
