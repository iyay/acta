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

func commitCount(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-list", "--count", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}
