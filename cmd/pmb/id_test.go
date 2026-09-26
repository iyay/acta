package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func gitCommit(t *testing.T, dir, msg string) {
	t.Helper()
	for _, args := range [][]string{{"add", "."}, {"commit", "-q", "-m", msg}} {
		if out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
}

func TestIDGivesAndIsIdempotent(t *testing.T) {
	dir := fixtureRepo(t)
	before := commitCount(t, dir)
	out, errOut, code := pmb(t, dir, "", "id")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	for _, want := range []string{"BUG-1", "PLAN-1", "SPEC-1"} {
		if !strings.Contains(out, want) {
			t.Fatalf("change lines lack %s: %q", want, out)
		}
	}
	n, _ := strconv.Atoi(strings.TrimSpace(before))
	if got := commitCount(t, dir); got != strconv.Itoa(n+1) {
		t.Fatalf("commits %s, want %d", got, n+1)
	}
	again, errOut, code := pmb(t, dir, "", "id")
	if code != 0 {
		t.Fatalf("second exit %d: %s", code, errOut)
	}
	if again != "" {
		t.Fatalf("second run printed %q, want silent", again)
	}
	if got := commitCount(t, dir); got != strconv.Itoa(n+1) {
		t.Fatalf("commits %s after second run, want %d", got, n+1)
	}
}

func TestIDResolvesInShowTickSet(t *testing.T) {
	dir := fixtureRepo(t)
	if _, errOut, code := pmb(t, dir, "", "id"); code != 0 {
		t.Fatalf("id exit %d: %s", code, errOut)
	}
	out, errOut, code := pmb(t, dir, "", "show", "PLAN-3")
	if code != 0 {
		t.Fatalf("show exit %d: %s", code, errOut)
	}
	if !strings.HasPrefix(out, "PLAN-3 ") || !strings.Contains(out, "plans/2026-09-23-lonely") {
		t.Fatalf("show does not lead with short id: %q", out)
	}
	for _, args := range [][]string{
		{"show", "plan-3"},
		{"tick", "PLAN-2.1", "--step", "1"},
		{"set", "BUG-2", "status", "fixing"},
	} {
		if _, errOut, code := pmb(t, dir, "", args...); code != 0 {
			t.Errorf("%v: exit %d: %s", args, code, errOut)
		}
	}
}

func TestIDUnknown(t *testing.T) {
	dir := fixtureRepo(t)
	if _, errOut, code := pmb(t, dir, "", "id"); code != 0 {
		t.Fatalf("id exit %d: %s", code, errOut)
	}
	_, errOut, code := pmb(t, dir, "", "show", "PLAN-999")
	if code != 1 || !strings.Contains(errOut, "unknown id PLAN-999") {
		t.Fatalf("exit %d stderr %q, want 1 with unknown id PLAN-999", code, errOut)
	}
}

// With auto_commit off the ids are written and left in the working tree, and
// the command still says what it did and succeeds.
func TestIDAutoCommitOffPrintsLinesAndExitsZero(t *testing.T) {
	dir := fixtureRepo(t)
	if err := os.WriteFile(filepath.Join(dir, ".pm.yaml"), []byte("auto_commit: false\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommit(t, dir, "auto_commit off")
	before := commitCount(t, dir)
	out, errOut, code := pmb(t, dir, "", "id")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "SPEC-") {
		t.Fatalf("no change lines: %q", out)
	}
	if got := commitCount(t, dir); got != before {
		t.Fatalf("commits %s, want %s with auto_commit off", got, before)
	}
}
func TestIDFixDuplicates(t *testing.T) {
	dir := fixtureRepo(t)
	if _, errOut, code := pmb(t, dir, "", "id"); code != 0 {
		t.Fatalf("id exit %d: %s", code, errOut)
	}
	src, err := os.ReadFile(filepath.Join(dir, ".pm/bugs/2026-09-26-open.md"))
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, ".pm/bugs/2026-09-27-dup.md"), src, 0o644); err != nil {
		t.Fatal(err)
	}
	gitCommit(t, dir, "dup")
	out, errOut, code := pmb(t, dir, "", "id", "--fix-duplicates")
	if code != 0 {
		t.Fatalf("exit %d: %s", code, errOut)
	}
	if !strings.Contains(out, "-> BUG-") {
		t.Fatalf("change line lacks -> BUG-: %q", out)
	}
	again, _, code := pmb(t, dir, "", "id", "--fix-duplicates")
	if code != 0 {
		t.Fatalf("second exit %d", code)
	}
	if again != "" {
		t.Fatalf("second run printed %q, want silent", again)
	}
}

func TestListShowsShortIDFirst(t *testing.T) {
	dir := fixtureRepo(t)
	if _, errOut, code := pmb(t, dir, "", "id"); code != 0 {
		t.Fatalf("id exit %d: %s", code, errOut)
	}
	out, _, code := pmb(t, dir, "", "list", "--type", "story")
	if code != 0 {
		t.Fatalf("list exit %d", code)
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) == 0 {
		t.Fatalf("empty list: %q", out)
	}
	withoutID := 0
	for _, line := range lines {
		i := strings.Index(line, "SPEC-")
		if i < 0 {
			i = strings.Index(line, "PLAN-")
		}
		j := strings.Index(line, "specs/")
		if j < 0 {
			j = strings.Index(line, "plans/")
		}
		if i < 0 && strings.Contains(line, "2026-09-17-broken") {
			// The one file whose frontmatter will not parse keeps no id.
			withoutID++
			continue
		}
		if i < 0 || j < 0 || i > j {
			t.Fatalf("short id not first: %q", line)
		}
	}
	if withoutID != 1 {
		t.Fatalf("%d lines without a short id, want only the broken one", withoutID)
	}
	out, _, code = pmb(t, dir, "", "list", "--json")
	if code != 0 {
		t.Fatalf("list --json exit %d", code)
	}
	var items []struct {
		ID      string `json:"id"`
		ShortID string `json:"short_id"`
		Hash    string `json:"hash"`
	}
	if err := json.Unmarshal([]byte(out), &items); err != nil {
		t.Fatalf("bad json %v: %s", err, out)
	}
	if len(items) == 0 {
		t.Fatal("no items")
	}
	// Only the file whose frontmatter will not parse has no id, so it is the
	// one that cannot be listed with a short id first.
	noID := 0
	for _, it := range items {
		if it.ShortID == "" || it.Hash == "" {
			if strings.Contains(it.ID, "2026-09-17-broken") {
				noID++
				continue
			}
			t.Fatalf("item lacks short_id/hash: %+v", it)
		}
		if !strings.Contains(it.ID, "/") {
			t.Fatalf("id is not the path: %+v", it)
		}
	}
	if noID != 1 {
		t.Fatalf("%d items without ids, want only the broken one", noID)
	}
}
