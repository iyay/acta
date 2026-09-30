package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/iyay/acta/internal/write"
)

// tickRepo makes a bare .acta root (no git needed: cmdTick never shells out
// to git) holding one plan with one task and one debt file with two lines.
func tickRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	plans := filepath.Join(dir, ".acta", "plans")
	debt := filepath.Join(dir, ".acta", "debt")
	if err := os.MkdirAll(plans, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(debt, 0o755); err != nil {
		t.Fatal(err)
	}
	planBody := "---\nid: PLAN-1\nhash: aaaa\n---\n# P\n\n### Task 1: One\n- [ ] a\n"
	if err := os.WriteFile(filepath.Join(plans, "2026-09-27-n.md"), []byte(planBody), 0o644); err != nil {
		t.Fatal(err)
	}
	debtBody := "---\nid: DEBT-1\nhash: bbbb\n---\n# Review NOTEs\n\n- [ ] a\n- [ ] b\n"
	if err := os.WriteFile(filepath.Join(debt, "2026-09-27-d.md"), []byte(debtBody), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

func runTick(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	// A temp cache folder, so a cli test never writes a lock into the
	// real cache folder of whoever runs the tests.
	cache := t.TempDir()
	t.Setenv("HOME", cache)
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"tick"}, args...), nil, false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

func debtFile(dir string) string {
	return filepath.Join(dir, ".acta", "debt", "2026-09-27-d.md")
}

func read(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func planN(dir string) string {
	return filepath.Join(dir, ".acta", "plans", "2026-09-27-n.md")
}

// undoRepo holds one plan whose single task carries three checkboxes, under a
// spec, so an undo can be watched on a task that was only partly ticked.
func undoRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"specs", "plans"} {
		if err := os.MkdirAll(filepath.Join(dir, ".acta", d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"specs/2026-09-20-s-design.md": "---\nid: SPEC-1\nhash: ssss\n---\n# S\n",
		"plans/2026-09-21-a.md":        "---\nid: PLAN-1\nhash: aaaa\n---\n# A\n\n**Spec:** `.acta/specs/2026-09-20-s-design.md`\n\n### Task 1: One\n- [ ] a\n- [ ] b\n- [ ] c\n",
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, ".acta", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

func TestCmdTickDebtWontfix(t *testing.T) {
	dir := tickRepo(t)
	code, _, stderr := runTick(t, dir, "DEBT-1.2", "--wontfix")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	got := read(t, debtFile(dir))
	if !strings.Contains(got, "- [-] b") {
		t.Fatalf("b was not marked wontfix: %q", got)
	}
	if !strings.Contains(got, "- [ ] a") {
		t.Fatalf("a changed: %q", got)
	}
}

func TestCmdTickDebtAllLowerCase(t *testing.T) {
	dir := tickRepo(t)
	code, _, stderr := runTick(t, dir, "debt-1.1", "--all")
	if code != exitOK {
		t.Fatalf("exit %d stderr %q", code, stderr)
	}
	got := read(t, debtFile(dir))
	if !strings.Contains(got, "- [x] a") {
		t.Fatalf("a was not ticked: %q", got)
	}
}

func TestCmdTickDebtRejectsStep(t *testing.T) {
	dir := tickRepo(t)
	code, _, _ := runTick(t, dir, "DEBT-1.1", "--step", "1")
	if code != exitSkipped {
		t.Fatalf("exit %d, want %d", code, exitSkipped)
	}
	got := read(t, debtFile(dir))
	if got != "---\nid: DEBT-1\nhash: bbbb\n---\n# Review NOTEs\n\n- [ ] a\n- [ ] b\n" {
		t.Fatalf("file changed: %q", got)
	}
}

func TestCmdTickDebtLinePastEnd(t *testing.T) {
	dir := tickRepo(t)
	code, _, _ := runTick(t, dir, "DEBT-1.9", "--all")
	if code != exitBadInput {
		t.Fatalf("exit %d, want %d", code, exitBadInput)
	}
}

func TestCmdTickWontfixRejectsPlanTask(t *testing.T) {
	dir := tickRepo(t)
	code, _, _ := runTick(t, dir, "PLAN-1.1", "--wontfix")
	if code != exitSkipped {
		t.Fatalf("exit %d, want %d", code, exitSkipped)
	}
	got := read(t, filepath.Join(dir, ".acta", "plans", "2026-09-27-n.md"))
	if got != "---\nid: PLAN-1\nhash: aaaa\n---\n# P\n\n### Task 1: One\n- [ ] a\n" {
		t.Fatalf("plan file changed: %q", got)
	}
}

// datesRepo holds one spec and the plans that carry its tasks, so a tick can
// be followed up to the spec. With two, the second plan stays open.
func datesRepo(t *testing.T, plans int) string {
	t.Helper()
	dir := t.TempDir()
	for _, d := range []string{"specs", "plans"} {
		if err := os.MkdirAll(filepath.Join(dir, ".acta", d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	files := map[string]string{
		"specs/2026-09-20-s-design.md": "---\nid: SPEC-1\nhash: ssss\n---\n# S\n",
		"plans/2026-09-21-a.md":        "---\nid: PLAN-1\nhash: aaaa\n---\n# A\n\n**Spec:** `.acta/specs/2026-09-20-s-design.md`\n\n### Task 1: One\n- [ ] a\n\n### Task 2: Two\n- [ ] b\n",
	}
	if plans == 2 {
		files["plans/2026-09-22-b.md"] = "---\nid: PLAN-2\nhash: bbbb\n---\n# B\n\n**Spec:** `.acta/specs/2026-09-20-s-design.md`\n\n### Task 1: One\n- [ ] c\n"
	}
	for name, body := range files {
		if err := os.WriteFile(filepath.Join(dir, ".acta", name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return dir
}

// onDay moves the clock the date writers read, so a test can put a tick on a
// known day, and puts it back after the test.
func onDay(t *testing.T, day int) {
	t.Helper()
	old := write.Now
	write.Now = func() time.Time { return time.Date(2026, 9, day, 10, 0, 0, 0, time.UTC) }
	t.Cleanup(func() { write.Now = old })
}

func specOf(dir string) string { return filepath.Join(dir, ".acta", "specs", "2026-09-20-s-design.md") }
func planA(dir string) string  { return filepath.Join(dir, ".acta", "plans", "2026-09-21-a.md") }

// wantsStarted says the file carries the started day and time a tick wrote.
// The date is written as text, so the file holds it in quotes.
func wantsStarted(t *testing.T, path, day string) {
	t.Helper()
	if s := read(t, path); !strings.Contains(s, `started: "`+day+`"`) {
		t.Errorf("%s lacks started %s: %q", path, day, s)
	}
}

func TestCmdTickWritesStartedOnPlanAndSpec(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	if code, _, errs := runTick(t, dir, "plans/2026-09-21-a#task-1", "--start"); code != exitOK {
		t.Fatalf("exit %d: %s", code, errs)
	}
	wantsStarted(t, planA(dir), "2026-09-26 10:00:00")
	wantsStarted(t, specOf(dir), "2026-09-26 10:00:00")
	// The day work began is written once, so a later tick on another day
	// must not move it.
	onDay(t, 27)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	wantsStarted(t, planA(dir), "2026-09-26 10:00:00")
	wantsStarted(t, specOf(dir), "2026-09-26 10:00:00")
	for _, p := range []string{planA(dir), specOf(dir)} {
		if s := read(t, p); strings.Contains(s, "2026-09-27") {
			t.Errorf("a second tick moved a date in %s: %q", p, s)
		}
	}
}

func TestCmdTickStepStartsThePlanToo(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--step", "1")
	wantsStarted(t, planA(dir), "2026-09-26 10:00:00")
	wantsStarted(t, specOf(dir), "2026-09-26 10:00:00")
}

func TestCmdTickFinishesPlanOnlyWhenAllTasksDone(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	wantsStarted(t, planA(dir), "2026-09-26 10:00:00")
	if strings.Contains(read(t, planA(dir)), "finished:") {
		t.Fatalf("plan finished with a task still open: %q", read(t, planA(dir)))
	}
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--all")
	for _, p := range []string{planA(dir), specOf(dir)} {
		if s := read(t, p); !strings.Contains(s, `finished: "2026-09-26 10:00:00"`) {
			t.Errorf("%s lacks finished: %q", p, s)
		}
	}
}

// The day a plan and its spec closed is the day people read, so a later tick
// on another day must not move it.
func TestCmdTickKeepsTheCloseDay(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--all")
	for _, p := range []string{planA(dir), specOf(dir)} {
		if s := read(t, p); !strings.Contains(s, `finished: "2026-09-26 10:00:00"`) {
			t.Fatalf("%s lacks finished on day 26: %q", p, s)
		}
	}
	onDay(t, 27)
	if code, _, errs := runTick(t, dir, "plans/2026-09-21-a#task-1", "--start"); code != exitOK {
		t.Fatalf("exit %d: %s", code, errs)
	}
	for _, p := range []string{planA(dir), specOf(dir)} {
		if s := read(t, p); !strings.Contains(s, `finished: "2026-09-26 10:00:00"`) {
			t.Errorf("%s lost the day it closed: %q", p, s)
		}
	}
}

func TestCmdTickSpecWaitsForEveryPlan(t *testing.T) {
	dir := datesRepo(t, 2)
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--all")
	if s := read(t, planA(dir)); !strings.Contains(s, `finished: "2026-09-26 10:00:00"`) {
		t.Errorf("plan A lacks finished: %q", s)
	}
	if s := read(t, specOf(dir)); strings.Contains(s, "finished:") {
		t.Errorf("spec finished while plan B is open: %q", s)
	}
}

func TestCmdTickDebtLineWritesNoDate(t *testing.T) {
	dir := tickRepo(t)
	onDay(t, 26)
	runTick(t, dir, "DEBT-1.1", "--all")
	runTick(t, dir, "DEBT-1.2", "--wontfix")
	if s := read(t, debtFile(dir)); strings.Contains(s, "started:") || strings.Contains(s, "finished:") {
		t.Errorf("debt tick wrote a date: %q", s)
	}
}

// gitInit makes dir a git repo holding one commit of everything in it, so a
// test can count the commits a command leaves behind.
func gitInit(t *testing.T, dir string) {
	t.Helper()
	gitOut(t, dir, "init", "-q")
	gitOut(t, dir, "config", "user.email", "test@example.com")
	gitOut(t, dir, "config", "user.name", "Test")
	commitAll(t, dir, "start", ".")
}

func TestCmdTickMakesNoCommit(t *testing.T) {
	dir := datesRepo(t, 1)
	gitInit(t, dir)
	before := gitOut(t, dir, "rev-list", "--count", "HEAD")
	onDay(t, 26)
	runTick(t, dir, "plans/2026-09-21-a#task-1", "--all")
	runTick(t, dir, "plans/2026-09-21-a#task-2", "--all")
	if after := gitOut(t, dir, "rev-list", "--count", "HEAD"); after != before {
		t.Errorf("tick made a commit: %s -> %s", before, after)
	}
}

func TestTickUndoPutsATaskBackToOpen(t *testing.T) {
	dir := tickRepo(t)
	if code, _, errOut := runTick(t, dir, "--all", "plans/2026-09-27-n#task-1"); code != exitOK {
		t.Fatalf("tick --all: %d %s", code, errOut)
	}
	code, out, errOut := runTick(t, dir, "--undo", "plans/2026-09-27-n#task-1")
	if code != exitOK || !strings.Contains(out, "0/1") {
		t.Fatalf("undo: %d %q %q", code, out, errOut)
	}
	if got := read(t, planN(dir)); !strings.Contains(got, "- [ ] a") {
		t.Fatalf("task not open: %q", got)
	}
}

func TestTickUndoPutsADebtLineBackToOpen(t *testing.T) {
	dir := tickRepo(t)
	if code, _, errOut := runTick(t, dir, "--wontfix", "DEBT-1.1"); code != exitOK {
		t.Fatalf("wontfix: %d %s", code, errOut)
	}
	if code, _, errOut := runTick(t, dir, "--undo", "DEBT-1.1"); code != exitOK {
		t.Fatalf("undo: %d %s", code, errOut)
	}
	if got := read(t, debtFile(dir)); !strings.Contains(got, "- [ ] a\n- [ ] b") {
		t.Fatalf("debt line not open: %q", got)
	}
}

func TestTickUndoMixedWithAnotherActionIsRefused(t *testing.T) {
	dir := tickRepo(t)
	for _, id := range []string{"plans/2026-09-27-n#task-1", "DEBT-1.1"} {
		for _, other := range []string{"--all", "--start", "--wontfix", "--step=1"} {
			before := map[string]string{planN(dir): read(t, planN(dir)), debtFile(dir): read(t, debtFile(dir))}
			if code, _, _ := runTick(t, dir, "--undo", other, id); code == exitOK {
				t.Errorf("--undo %s %s: exit 0", other, id)
			}
			for path, want := range before {
				if got := read(t, path); got != want {
					t.Errorf("--undo %s %s changed %s: %q", other, id, path, got)
				}
			}
		}
	}
}

// A task ticked halfway has one box done and two open, and the day work began
// is already written. Undo opens every box and leaves that day alone.
func TestTickUndoOpensEveryBoxOfAPartlyTickedTask(t *testing.T) {
	dir := undoRepo(t)
	onDay(t, 26)
	if code, _, errOut := runTick(t, dir, "plans/2026-09-21-a#task-1", "--step", "1"); code != exitOK {
		t.Fatalf("step 1: %d %s", code, errOut)
	}
	if code, out, errOut := runTick(t, dir, "plans/2026-09-21-a#task-1", "--undo"); code != exitOK || !strings.Contains(out, "0/3") {
		t.Fatalf("undo: %d %q %q", code, out, errOut)
	}
	want := "---\nid: PLAN-1\nhash: aaaa\nstarted: \"2026-09-26 10:00:00\"\n---\n# A\n\n**Spec:** `.acta/specs/2026-09-20-s-design.md`\n\n### Task 1: One\n- [ ] a\n- [ ] b\n- [ ] c\n"
	if got := read(t, planA(dir)); got != want {
		t.Fatalf("plan =\n%q\nwant\n%q", got, want)
	}
	if s := read(t, specOf(dir)); strings.Contains(s, "finished:") {
		t.Errorf("undo closed the spec: %q", s)
	}
}

// A task only started was never ticked, so an undo has nothing to open and
// must leave the plan and the spec byte for byte as they were.
func TestTickUndoAfterStartWritesNothing(t *testing.T) {
	dir := datesRepo(t, 1)
	onDay(t, 26)
	if code, _, errOut := runTick(t, dir, "plans/2026-09-21-a#task-1", "--start"); code != exitOK {
		t.Fatalf("start: %d %s", code, errOut)
	}
	before := map[string]string{planA(dir): read(t, planA(dir)), specOf(dir): read(t, specOf(dir))}
	if code, _, errOut := runTick(t, dir, "plans/2026-09-21-a#task-1", "--undo"); code != exitOK {
		t.Fatalf("undo: %d %s", code, errOut)
	}
	for path, want := range before {
		if got := read(t, path); got != want {
			t.Errorf("undo changed %s: %q", path, got)
		}
	}
}

func TestTickUndoOnAnOpenTaskChangesNothing(t *testing.T) {
	dir := tickRepo(t)
	before := read(t, planN(dir))
	if code, out, errOut := runTick(t, dir, "plans/2026-09-27-n#task-1", "--undo"); code != exitOK || out == "" {
		t.Fatalf("undo: %d %q %q", code, out, errOut)
	}
	if got := read(t, planN(dir)); got != before {
		t.Fatalf("plan changed: %q", got)
	}
}

func TestTickUndoOnAnOpenDebtLineChangesNothing(t *testing.T) {
	dir := tickRepo(t)
	before := read(t, debtFile(dir))
	if code, out, errOut := runTick(t, dir, "DEBT-1.2", "--undo"); code != exitOK || out == "" {
		t.Fatalf("undo: %d %q %q", code, out, errOut)
	}
	if got := read(t, debtFile(dir)); got != before {
		t.Fatalf("debt file changed: %q", got)
	}
}

// Undo is a slip of the finger, not work, so nobody is recorded as working on
// the item: the file stays missing, and a record an earlier tick wrote is
// left exactly as it was.
func TestTickUndoWritesNoAgentRecord(t *testing.T) {
	t.Run("no record file", func(t *testing.T) {
		dir := tickRepo(t)
		if code, _, errOut := runTick(t, dir, "plans/2026-09-27-n#task-1", "--undo", "--agent", "omp"); code != exitOK {
			t.Fatalf("undo: %d %s", code, errOut)
		}
		if _, err := os.Stat(filepath.Join(dir, ".acta", ".agents.json")); !os.IsNotExist(err) {
			t.Fatalf("undo wrote an agent record: %v", err)
		}
	})
	t.Run("existing record untouched", func(t *testing.T) {
		dir := tickRepo(t)
		if code, _, errOut := runTick(t, dir, "plans/2026-09-27-n#task-1", "--all", "--agent", "omp"); code != exitOK {
			t.Fatalf("tick: %d %s", code, errOut)
		}
		rec := filepath.Join(dir, ".acta", ".agents.json")
		before := read(t, rec)
		if code, _, errOut := runTick(t, dir, "plans/2026-09-27-n#task-1", "--undo", "--agent", "omp"); code != exitOK {
			t.Fatalf("undo: %d %s", code, errOut)
		}
		if got := read(t, rec); got != before {
			t.Fatalf("record changed: %q", got)
		}
	})
}

// Only a task and a debt line carry boxes, so anything else is refused the
// same way a tick refuses it.
func TestTickUndoRefusesAnItemThatIsNotATask(t *testing.T) {
	dir := datesRepo(t, 1)
	for _, id := range []string{"SPEC-1", "PLAN-1"} {
		before := read(t, planA(dir))
		if code, _, _ := runTick(t, dir, id, "--undo"); code != exitBadInput {
			t.Errorf("--undo %s: exit %d, want %d", id, code, exitBadInput)
		}
		if got := read(t, planA(dir)); got != before {
			t.Errorf("--undo %s changed the plan: %q", id, got)
		}
	}
}

func TestTickUsageNamesUndo(t *testing.T) {
	dir := tickRepo(t)
	code, _, errOut := runTick(t, dir, "--help")
	if code != exitOK {
		t.Fatalf("help: %d", code)
	}
	if !strings.Contains(errOut, "--undo") {
		t.Fatalf("help hides --undo: %q", errOut)
	}
}
