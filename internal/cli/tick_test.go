package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
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
