package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestTickCommand(t *testing.T) {
	dir := fixtureRepo(t)
	plan := filepath.Join(dir, ".pm/plans/2026-09-21-alpha.md")
	before, _ := os.ReadFile(plan)

	out, errOut, code := pmb(t, dir, "", "tick", "plans/2026-09-21-alpha#task-2", "--step", "2")
	if code != 0 || strings.TrimSpace(out) != "plans/2026-09-21-alpha#task-2 2/2" {
		t.Fatalf("exit %d out %q err %q", code, out, errOut)
	}
	after, _ := os.ReadFile(plan)
	if strings.Count(string(after), "- [x]") != strings.Count(string(before), "- [x]")+1 {
		t.Fatal("tick did not add exactly one ticked box")
	}
	if out, _, _ := pmb(t, dir, "", "list", "--type", "task", "--json"); strings.Contains(out, "task-2") {
		t.Fatal("task 2 should now be done and leave the active list")
	}
	if st, _, _ := pmb(t, dir, "", "tick", "plans/2026-09-21-alpha#task-2", "--all"); st == "" {
		t.Fatal("--all on a done task should still print progress")
	}

	for _, args := range [][]string{
		{"tick"},
		{"tick", "plans/nope#task-1"},
		{"tick", "specs/2026-09-20-alpha"},
		{"tick", "docs/superpowers/plans/2026-01-02-old#task-1"},
		{"tick", "plans/2026-09-21-alpha#task-1", "--step", "9"},
		{"tick", "plans/2026-09-21-alpha#task-1", "--step", "1", "--all"},
		{"tick", "plans/2026-09-21-alpha#task-1"},
	} {
		if _, _, code := pmb(t, dir, "", args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
	}
	if n := commitCount(t, dir); n != "1" {
		t.Fatalf("tick must never commit: %s commits, want 1", n)
	}
}
func TestTickHelp(t *testing.T) {
	dir := fixtureRepo(t)
	for _, args := range [][]string{
		{"tick", "--help"},
		{"tick", "-help"},
		{"tick", "-h"},
		{"tick", "plans/2026-09-21-alpha#task-1", "-h"},
	} {
		out, errOut, code := pmb(t, dir, "", args...)
		if code != 0 {
			t.Errorf("%v: exit %d, want 0", args, code)
		}
		if got := out + errOut; !strings.Contains(got, "plans/<stem>#task-N") {
			t.Errorf("%v: output %q names no plans/<stem>#task-N id", args, got)
		}
		for _, flag := range []string{"-step", "-all", "tick every checkbox of the task"} {
			if !strings.Contains(out+errOut, flag) {
				t.Errorf("%v: output %q does not list %q", args, out+errOut, flag)
			}
		}
	}
}

func TestTickMixedLineEndingsFromBoardLine(t *testing.T) {
	dir := fixtureRepo(t)
	// Plan's BLOCKER repro: a CRLF plan with no frontmatter. pmb set
	// prepends an LF-only frontmatter block, so the file ends up mixed.
	body := "# Plan\r\n\r\n### Task 1: a\r\n- [ ] a1\r\n\r\n### Task 2: b\r\n- [ ] b1\r\n"
	rel := filepath.Join(".pm", "plans", "2026-09-29-crlf.md")
	if err := os.WriteFile(filepath.Join(dir, rel), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "add", rel).CombinedOutput(); err != nil {
		t.Fatalf("git add: %v %s", out, err)
	}
	if out, err := exec.Command("git", "-C", dir, "commit", "-q", "-m", "crlf plan").CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v %s", out, err)
	}
	id := "plans/2026-09-29-crlf#task-1"
	if _, errOut, code := pmb(t, dir, "", "set", "plans/2026-09-29-crlf", "status", "in-progress"); code != 0 {
		t.Fatalf("set: exit %d: %s", code, errOut)
	}
	pre, _ := os.ReadFile(filepath.Join(dir, rel))
	out, errOut, code := pmb(t, dir, "", "tick", id, "--step", "1")
	if code != 0 || strings.TrimSpace(out) != id+" 1/1" {
		t.Fatalf("tick: exit %d out %q err %q", code, out, errOut)
	}
	after, _ := os.ReadFile(filepath.Join(dir, rel))
	// Exactly the Nth box as the board counts it, no other byte changes.
	if want := strings.Replace(string(pre), "- [ ] a1", "- [x] a1", 1); string(after) != want {
		t.Fatalf("got %q want %q", after, want)
	}
	show, errOut, code := pmb(t, dir, "", "show", id, "--json")
	if code != 0 {
		t.Fatalf("show: exit %d: %s", code, errOut)
	}
	var it jsonItem
	decode(t, show, &it)
	if it.Progress.Done != 1 || it.Progress.Total != 1 {
		t.Fatalf("show progress = %+v, tick printed 1/1", it.Progress)
	}
}

func commitCount(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-list", "--count", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}
