package cli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// commitsRepo makes a temp repo with one plan (short id PLN-0127, hash
// broaksz, tasks 1 to 3) and a run of commits that name tasks. Each commit gets
// its own author date, so the order checks do not depend on the clock.
func commitsRepo(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, ".acta", "plans"), 0o755); err != nil {
		t.Fatal(err)
	}
	plan := "---\nid: PLN-0127\nhash: broaksz\n---\n# Viewer\n\n### Task 01: One\n- [ ] a\n\n### Task 02: Two\n- [ ] a\n\n### Task 03: Three\n- [ ] a\n"
	if err := os.WriteFile(filepath.Join(dir, ".acta", "plans", "2026-10-09-viewer.md"), []byte(plan), 0o644); err != nil {
		t.Fatal(err)
	}
	gitOut(t, dir, "init", "-q", "-b", "main")
	gitOut(t, dir, "config", "user.name", "test")
	gitOut(t, dir, "config", "user.email", "test@example.com")
	gitOut(t, dir, "add", ".acta")
	commitAt(t, dir, "2026-10-01T10:00:00+00:00", "chore(plan): add plan")
	return dir
}

// commitAt makes an empty commit with a fixed author date.
func commitAt(t *testing.T, dir, date, msg string) {
	t.Helper()
	t.Setenv("GIT_AUTHOR_DATE", date)
	t.Setenv("GIT_COMMITTER_DATE", date)
	gitOut(t, dir, "commit", "-q", "--allow-empty", "-m", msg)
}

// linkedRepo adds commits in an order that differs from the task order, so
// sorting by date is told apart from sorting by task.
func linkedRepo(t *testing.T) string {
	dir := commitsRepo(t)
	commitAt(t, dir, "2026-10-02T10:00:00+00:00", "feat(a): second task first\n\nTask: PLN-broaksz#2")
	commitAt(t, dir, "2026-10-03T10:00:00+00:00", "feat(b): first task later\n\nTask: PLN-broaksz#1")
	commitAt(t, dir, "2026-10-04T10:00:00+00:00", "chore(c): chore on task one\n\nTask: PLN-broaksz#1")
	commitAt(t, dir, "2026-10-05T10:00:00+00:00", "feat(d): two tasks\n\nTask: PLN-broaksz#2\nTask: PLN-broaksz#3")
	commitAt(t, dir, "2026-10-06T10:00:00+00:00", "feat(e): another plan\n\nTask: PLN-zzzzzzz#1")
	return dir
}

func runCommits(t *testing.T, dir string, args ...string) (int, string, string) {
	t.Helper()
	cache := t.TempDir()
	t.Setenv("HOME", cache)
	t.Setenv("XDG_CACHE_HOME", cache)
	t.Chdir(dir)
	var stdout, stderr strings.Builder
	code := Run(append([]string{"commits"}, args...), strings.NewReader(""), false, &stdout, &stderr)
	return code, stdout.String(), stderr.String()
}

// shaOf gives the full sha of the commit whose subject starts with subject.
func shaOf(t *testing.T, dir, subject string) string {
	t.Helper()
	return gitOut(t, dir, "log", "-1", "--format=%H", "--grep=^"+subject)
}

func TestCommitsPlanByShortIDAndHash(t *testing.T) {
	dir := linkedRepo(t)
	want := []string{
		shaOf(t, dir, "feat(a)")[:7] + "  2026-10-02  #2  feat(a): second task first  (main)",
		shaOf(t, dir, "feat(b)")[:7] + "  2026-10-03  #1  feat(b): first task later  (main)",
		shaOf(t, dir, "feat(d)")[:7] + "  2026-10-05  #2,#3  feat(d): two tasks  (main)",
	}
	for _, id := range []string{"PLN-0127", "PLN-broaksz"} {
		code, stdout, stderr := runCommits(t, dir, id)
		if code != exitOK {
			t.Fatalf("%s: exit %d, stderr %q", id, code, stderr)
		}
		if got := strings.TrimRight(stdout, "\n"); got != strings.Join(want, "\n") {
			t.Errorf("%s: stdout =\n%s\nwant\n%s", id, got, strings.Join(want, "\n"))
		}
	}
}

func TestCommitsTaskNumberWithOrWithoutZero(t *testing.T) {
	dir := linkedRepo(t)
	want := []string{
		shaOf(t, dir, "feat(a)")[:7] + "  2026-10-02  #2  feat(a): second task first  (main)",
		shaOf(t, dir, "feat(d)")[:7] + "  2026-10-05  #2  feat(d): two tasks  (main)",
	}
	for _, task := range []string{"2", "02"} {
		code, stdout, stderr := runCommits(t, dir, "PLN-0127", task)
		if code != exitOK {
			t.Fatalf("task %s: exit %d, stderr %q", task, code, stderr)
		}
		if got := strings.TrimRight(stdout, "\n"); got != strings.Join(want, "\n") {
			t.Errorf("task %s: stdout =\n%s\nwant\n%s", task, got, strings.Join(want, "\n"))
		}
	}
}

func TestCommitsAllShowsChores(t *testing.T) {
	dir := linkedRepo(t)
	code, stdout, stderr := runCommits(t, dir, "PLN-0127", "1", "--all")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	want := []string{
		shaOf(t, dir, "feat(b)")[:7] + "  2026-10-03  #1  feat(b): first task later  (main)",
		shaOf(t, dir, "chore(c)")[:7] + "  2026-10-04  #1  chore(c): chore on task one  (main)",
	}
	if got := strings.TrimRight(stdout, "\n"); got != strings.Join(want, "\n") {
		t.Errorf("stdout =\n%s\nwant\n%s", got, strings.Join(want, "\n"))
	}
	_, hidden, _ := runCommits(t, dir, "PLN-0127", "1")
	if strings.Contains(hidden, "chore(c)") {
		t.Errorf("chore commit shown without --all: %q", hidden)
	}
}

func TestCommitsJSON(t *testing.T) {
	dir := linkedRepo(t)
	code, stdout, stderr := runCommits(t, dir, "PLN-broaksz", "--json", "--all")
	if code != exitOK {
		t.Fatalf("exit %d, stderr %q", code, stderr)
	}
	var got []struct {
		Sha     string `json:"sha"`
		Date    string `json:"date"`
		Tasks   []int  `json:"tasks"`
		Subject string `json:"subject"`
		Branch  string `json:"branch"`
		Chore   bool   `json:"chore"`
	}
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("bad json %q: %v", stdout, err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d commits, want 4: %s", len(got), stdout)
	}
	if got[0].Sha != shaOf(t, dir, "feat(a)") || got[0].Date != "2026-10-02T10:00:00Z" ||
		got[0].Subject != "feat(a): second task first" || got[0].Branch != "main" || got[0].Chore {
		t.Errorf("first entry wrong: %+v", got[0])
	}
	if !got[2].Chore || got[2].Tasks[0] != 1 {
		t.Errorf("chore entry wrong: %+v", got[2])
	}
	if len(got[3].Tasks) != 2 || got[3].Tasks[0] != 2 || got[3].Tasks[1] != 3 {
		t.Errorf("two-task entry wrong: %+v", got[3])
	}
	// Without --all the chore is left out of the JSON too.
	_, filtered, _ := runCommits(t, dir, "PLN-broaksz", "--json")
	if strings.Contains(filtered, "chore(c)") {
		t.Errorf("chore commit in json without --all: %s", filtered)
	}
}

func TestCommitsNothingLinked(t *testing.T) {
	dir := commitsRepo(t)
	code, stdout, stderr := runCommits(t, dir, "PLN-0127")
	if code != exitOK || strings.TrimSpace(stdout) != "no linked commits" {
		t.Errorf("exit %d, stdout %q, stderr %q", code, stdout, stderr)
	}
	code, stdout, _ = runCommits(t, dir, "PLN-0127", "--json")
	if code != exitOK || strings.TrimSpace(stdout) != "[]" {
		t.Errorf("json: exit %d, stdout %q", code, stdout)
	}
	// Task 3 has no commit of its own in the linked repo, only a shared one.
	dir = linkedRepo(t)
	code, stdout, _ = runCommits(t, dir, "PLN-0127", "3", "--json")
	if code != exitOK || !strings.Contains(stdout, "feat(d)") {
		t.Errorf("task 3: exit %d, stdout %q", code, stdout)
	}
}

func TestCommitsBadInput(t *testing.T) {
	dir := linkedRepo(t)
	cases := map[string][]string{
		"no args":         {},
		"unknown plan":    {"PLN-9999"},
		"unknown hash":    {"PLN-nothere"},
		"task not in it":  {"PLN-0127", "9"},
		"task not number": {"PLN-0127", "abc"},
		"too many args":   {"PLN-0127", "1", "2"},
		"not a plan":      {"plans/2026-10-09-viewer#task-01"},
	}
	for name, args := range cases {
		code, stdout, stderr := runCommits(t, dir, args...)
		if code != exitBadInput {
			t.Errorf("%s: exit %d, want %d (stdout %q, stderr %q)", name, code, exitBadInput, stdout, stderr)
		}
	}
}
