package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestTickCommand(t *testing.T) {
	dir := fixtureRepo(t)
	plan := filepath.Join(dir, ".acta/plans/2026-09-21-alpha.md")
	before, _ := os.ReadFile(plan)

	out, errOut, code := acta(t, dir, "", "tick", "plans/2026-09-21-alpha#task-2", "--step", "2")
	if code != 0 || strings.TrimSpace(out) != "plans/2026-09-21-alpha#task-2 2/2" {
		t.Fatalf("exit %d out %q err %q", code, out, errOut)
	}
	after, _ := os.ReadFile(plan)
	if strings.Count(string(after), "- [x]") != strings.Count(string(before), "- [x]")+1 {
		t.Fatal("tick did not add exactly one ticked box")
	}
	if out, _, _ := acta(t, dir, "", "list", "--type", "task", "--json"); strings.Contains(out, "task-2") {
		t.Fatal("task 2 should now be done and leave the active list")
	}
	if st, _, _ := acta(t, dir, "", "tick", "plans/2026-09-21-alpha#task-2", "--all"); st == "" {
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
		if _, _, code := acta(t, dir, "", args...); code != 1 {
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
		out, errOut, code := acta(t, dir, "", args...)
		if code != 0 {
			t.Errorf("%v: exit %d, want 0", args, code)
		}
		if got := out + errOut; !strings.Contains(got, "acta tick <id>") {
			t.Errorf("%v: output %q names no acta tick <id> usage", args, got)
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
	rel := filepath.Join(".acta", "plans", "2026-09-29-crlf.md")
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
	if _, errOut, code := acta(t, dir, "", "set", "plans/2026-09-29-crlf", "status", "in-progress"); code != 0 {
		t.Fatalf("set: exit %d: %s", code, errOut)
	}
	pre, _ := os.ReadFile(filepath.Join(dir, rel))
	out, errOut, code := acta(t, dir, "", "tick", id, "--step", "1")
	if code != 0 || strings.TrimSpace(out) != id+" 1/1" {
		t.Fatalf("tick: exit %d out %q err %q", code, out, errOut)
	}
	after, _ := os.ReadFile(filepath.Join(dir, rel))
	// Exactly the Nth box as the board counts it, no other byte changes.
	if want := strings.Replace(string(pre), "- [ ] a1", "- [x] a1", 1); string(after) != want {
		t.Fatalf("got %q want %q", after, want)
	}
	show, errOut, code := acta(t, dir, "", "show", id, "--json")
	if code != 0 {
		t.Fatalf("show: exit %d: %s", code, errOut)
	}
	var it jsonItem
	decode(t, show, &it)
	if it.Progress.Done != 1 || it.Progress.Total != 1 {
		t.Fatalf("show progress = %+v, tick printed 1/1", it.Progress)
	}
}

// The board shows which agent works on a task, so a successful tick records
// the name in the worktree it ran in.
func TestTickRecordsTheAgent(t *testing.T) {
	cases := []struct {
		name    string
		args    []string
		env     string
		want    string // "" means no record at all
		wantErr bool
	}{
		{"flag wins", []string{"--agent", "omp"}, "claude-code_2-1-283_agent", "omp", false},
		{"claude code env", nil, "claude-code_2-1-283_agent", "claude", false},
		{"nothing set", nil, "", "", false},
		{"failed tick records nothing", []string{"--agent", "omp"}, "", "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			dir := fixtureRepo(t)
			t.Setenv("AI_AGENT", c.env)
			args := append([]string{"tick"}, c.args...)
			if c.wantErr {
				args = append(args, "plans/nope#task-1", "--step", "1")
			} else {
				args = append(args, "plans/2026-09-21-alpha#task-1", "--step", "1")
			}
			if _, errOut, code := acta(t, dir, "", args...); c.wantErr && code == 0 {
				t.Fatalf("exit 0, want a failure: %q", errOut)
			}
			rec := agentsFile(t, dir)
			recs := readRecords(t, rec)
			if c.want == "" {
				if len(recs) != 0 {
					t.Fatalf("%s holds %v, want no record", rec, recs)
				}
				return
			}
			got, ok := recs["plans/2026-09-21-alpha#task-1"]
			if !ok || got.Agent != c.want {
				t.Fatalf("record = %+v (present %t), want agent %q", got, ok, c.want)
			}
			if _, err := time.Parse(time.RFC3339, got.At); err != nil {
				t.Errorf("at %q is not RFC3339: %v", got.At, err)
			}
			if out := git(t, dir, "status", "--porcelain"); strings.Contains(out, ".agents.json") {
				t.Errorf("git status lists the record file: %q", out)
			}
			if n := commitCount(t, dir); n != "1" {
				t.Errorf("tick committed: %s commits, want 1", n)
			}
		})
	}
}

// agentsFile is where a tick records the agent in the default root folder.
func agentsFile(t *testing.T, dir string) string {
	t.Helper()
	return filepath.Join(dir, ".acta", ".agents.json")
}

func readRecords(t *testing.T, path string) map[string]struct {
	Agent   string `json:"agent"`
	At      string `json:"at"`
	Started bool   `json:"started,omitempty"`
} {
	t.Helper()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		t.Fatal(err)
	}
	var recs map[string]struct {
		Agent   string `json:"agent"`
		At      string `json:"at"`
		Started bool   `json:"started,omitempty"`
	}
	if err := json.Unmarshal(b, &recs); err != nil {
		t.Fatalf("%s: %v (%s)", path, err, b)
	}
	return recs
}

func git(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v %s", args, err, out)
	}
	return string(out)
}

func commitCount(t *testing.T, dir string) string {
	t.Helper()
	out, err := exec.Command("git", "-C", dir, "rev-list", "--count", "HEAD").Output()
	if err != nil {
		t.Fatal(err)
	}
	return strings.TrimSpace(string(out))
}

// --start marks the task started without touching a box, and it refuses to
// mix with the flags that tick boxes, writing nothing when refused.
func TestTickStartMarksStartedWithoutTicking(t *testing.T) {
	dir := fixtureRepo(t)
	plan := filepath.Join(dir, ".acta/plans/2026-09-21-alpha.md")
	id := "plans/2026-09-21-alpha#task-1"
	before, _ := os.ReadFile(plan)
	if _, errOut, code := acta(t, dir, "", "tick", id, "--start", "--agent", "omp"); code != 0 {
		t.Fatalf("--start: exit %d err %q, want 0", code, errOut)
	}
	after, _ := os.ReadFile(plan)
	// The plan gains the started date, but no box moves: count the ticked
	// boxes before and after, the same way the plain tick test does.
	if strings.Count(string(after), "- [x]") != strings.Count(string(before), "- [x]") {
		t.Fatal("--start must tick no box")
	}
	if !strings.Contains(string(after), `started: "`+time.Now().Format("2006-01-02")+`"`) {
		t.Fatalf("--start wrote no started date: %q", after)
	}
	recs := readRecords(t, agentsFile(t, dir))
	if !recs[id].Started || recs[id].Agent != "omp" {
		t.Fatalf("record = %+v, want started true by omp", recs[id])
	}
	record, _ := os.ReadFile(agentsFile(t, dir))
	for _, args := range [][]string{
		{"tick", id, "--start", "--all"},
		{"tick", id, "--start", "--step", "1"},
	} {
		if _, _, code := acta(t, dir, "", args...); code != 1 {
			t.Errorf("%v: exit %d, want 1", args, code)
		}
		same, _ := os.ReadFile(agentsFile(t, dir))
		if string(same) != string(record) {
			t.Errorf("%v wrote a record, want none", args)
		}
		refused, _ := os.ReadFile(plan)
		if string(refused) != string(after) {
			t.Errorf("%v changed the plan", args)
		}
	}
}
