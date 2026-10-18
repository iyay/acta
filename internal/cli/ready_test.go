package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readyPlan is a one-task plan file. box is " " for an open task and "x"
// for a done one; after is the plan's after list, or "" for none.
func readyPlan(id, after, box string) string {
	front := "id: " + id + "\n"
	if after != "" {
		front += "after: [" + after + "]\n"
	}
	return "---\n" + front + "---\n# Plan " + id + "\n\n### Task 1: Do it\n\n- [" + box + "] Red: it fails\n"
}

// readyRepo makes a repo on main holding the given planning files in one
// commit, and returns its folder.
func readyRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	dir := tidyRepo(t)
	// tidyRepo leaves its work branch checked out; the plans belong on main.
	tidyGit(t, dir, "checkout", "-q", "main")
	for name, body := range files {
		writeReady(t, dir, name, body)
	}
	tidyGit(t, dir, "add", "-A")
	tidyGit(t, dir, "commit", "-q", "-m", "plans")
	return dir
}

func writeReady(t *testing.T, dir, name, body string) {
	t.Helper()
	path := filepath.Join(dir, ".acta", name)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// runReady runs acta ready in dir and gives back the exit code and both
// outputs joined, so a test can look for the cause wherever it is printed.
func runReady(t *testing.T, dir string, args ...string) (int, string) {
	t.Helper()
	var stdout, stderr strings.Builder
	var code int
	inDir(t, dir, func() {
		code = Run(append([]string{"ready"}, args...), strings.NewReader(""), false, &stdout, &stderr)
	})
	return code, stdout.String() + stderr.String()
}

// A plan is ready only when every plan in its after list is in the parent's
// committed tree with every box ticked. No after list is ready at once.
func TestReadyOnTheParent(t *testing.T) {
	dir := readyRepo(t, map[string]string{
		"plans/2026-10-01-a.md": readyPlan("PLN-0001", "", "x"),
		"plans/2026-10-01-c.md": readyPlan("PLN-0003", "", " "),
		"plans/2026-10-01-b.md": readyPlan("PLN-0002", "PLN-0001", " "),
		"plans/2026-10-01-d.md": readyPlan("PLN-0004", "PLN-0001, PLN-0003", " "),
		"plans/2026-10-01-e.md": readyPlan("PLN-0005", "", " "),
	})
	if code, out := runReady(t, dir, "PLN-0002", "--parent", "main"); code != 0 {
		t.Errorf("after a landed plan: exit %d, want 0\n%s", code, out)
	}
	if code, out := runReady(t, dir, "PLN-0005", "--parent", "main"); code != 0 {
		t.Errorf("no after list: exit %d, want 0\n%s", code, out)
	}
	code, out := runReady(t, dir, "PLN-0004", "--parent", "main")
	if code != 1 || !strings.Contains(out, "PLN-0003") || strings.Contains(out, "PLN-0001 ") {
		t.Errorf("after one open plan: exit %d, want 1 naming only PLN-0003\n%s", code, out)
	}
}

// Done in a worktree or on another branch is not landed: only the committed
// tree of the named parent counts.
func TestReadyIgnoresWorkNotOnTheParent(t *testing.T) {
	dir := readyRepo(t, map[string]string{
		"plans/2026-10-01-a.md": readyPlan("PLN-0001", "", " "),
		"plans/2026-10-01-b.md": readyPlan("PLN-0002", "PLN-0001", " "),
	})
	// Ticked on disk but not committed: the parent tree still has it open.
	writeReady(t, dir, "plans/2026-10-01-a.md", readyPlan("PLN-0001", "", "x"))
	if code, out := runReady(t, dir, "PLN-0002", "--parent", "main"); code != 1 {
		t.Errorf("ticked only on disk: exit %d, want 1\n%s", code, out)
	}
	// Landed on another branch, never on main.
	tidyGit(t, dir, "checkout", "-q", "-b", "other")
	tidyGit(t, dir, "commit", "-q", "-am", "land on other")
	tidyGit(t, dir, "checkout", "-q", "main")
	if code, out := runReady(t, dir, "PLN-0002", "--parent", "main"); code != 1 {
		t.Errorf("landed only on other: exit %d, want 1\n%s", code, out)
	}
	if code, out := runReady(t, dir, "PLN-0002", "--parent", "other"); code != 0 {
		t.Errorf("parent other: exit %d, want 0\n%s", code, out)
	}
}

// A bad after list is refused with exit 2 and the cause named: an unknown
// id, an id that is not a plan, the plan itself, and a cycle with its chain.
func TestReadyRefusesBadAfter(t *testing.T) {
	dir := readyRepo(t, map[string]string{
		"specs/2026-10-01-s.md":  "---\nid: SPC-0009\n---\n# S\n",
		"plans/2026-10-01-u.md":  readyPlan("PLN-0011", "PLN-0999", " "),
		"plans/2026-10-01-n.md":  readyPlan("PLN-0012", "SPC-0009", " "),
		"plans/2026-10-01-s.md":  readyPlan("PLN-0013", "PLN-0013", " "),
		"plans/2026-10-01-c1.md": readyPlan("PLN-0021", "PLN-0022", " "),
		"plans/2026-10-01-c2.md": readyPlan("PLN-0022", "PLN-0021", " "),
		"plans/2026-10-01-t1.md": readyPlan("PLN-0031", "PLN-0032", " "),
		"plans/2026-10-01-t2.md": readyPlan("PLN-0032", "PLN-0033", " "),
		"plans/2026-10-01-t3.md": readyPlan("PLN-0033", "PLN-0031", " "),
	})
	for _, c := range []struct{ plan, want string }{
		{"PLN-0011", "PLN-0999"},
		{"PLN-0012", "SPC-0009 is not a plan"},
		{"PLN-0013", "PLN-0013 waits on itself"},
		{"PLN-0021", "PLN-0021 -> PLN-0022 -> PLN-0021"},
		{"PLN-0031", "PLN-0031 -> PLN-0032 -> PLN-0033 -> PLN-0031"},
	} {
		code, out := runReady(t, dir, c.plan, "--parent", "main")
		if code != 2 || !strings.Contains(out, c.want) {
			t.Errorf("%s: exit %d, want 2 with %q\n%s", c.plan, code, c.want, out)
		}
	}
}

// A worktree made before its after plan landed is not ready with --head,
// even though the parent now holds the plan; once the worktree takes in the
// parent it is.
func TestReadyHeadMustHoldTheLand(t *testing.T) {
	dir := readyRepo(t, map[string]string{
		"plans/2026-10-01-a.md": readyPlan("PLN-0001", "", " "),
		"plans/2026-10-01-b.md": readyPlan("PLN-0002", "PLN-0001", " "),
	})
	tidyGit(t, dir, "branch", "b-work")
	writeReady(t, dir, "plans/2026-10-01-a.md", readyPlan("PLN-0001", "", "x"))
	tidyGit(t, dir, "commit", "-q", "-am", "land a")
	code, out := runReady(t, dir, "PLN-0002", "--parent", "main", "--head", "b-work")
	if code != 1 || !strings.Contains(out, "PLN-0001") {
		t.Errorf("stale head: exit %d, want 1 naming PLN-0001\n%s", code, out)
	}
	tidyGit(t, dir, "checkout", "-q", "b-work")
	tidyGit(t, dir, "merge", "-q", "main")
	tidyGit(t, dir, "checkout", "-q", "main")
	if code, out := runReady(t, dir, "PLN-0002", "--parent", "main", "--head", "b-work"); code != 0 {
		t.Errorf("head took in the parent: exit %d, want 0\n%s", code, out)
	}
}

// Boxes outside task sections, like an example in a code fence, do not count:
// ready reads the plan the way the board does.
func TestReadyIgnoresBoxesOutsideTasks(t *testing.T) {
	fenced := "---\nid: PLN-0001\n---\n# Plan\n\n### Task 1: Do it\n\n- [x] Red: it fails\n\n```md\n- [ ] Red: example\n```\n\n## Notes\n\n- [ ] a note, not a task box\n"
	dir := readyRepo(t, map[string]string{
		"plans/2026-10-01-a.md": fenced,
		"plans/2026-10-01-b.md": readyPlan("PLN-0002", "PLN-0001", " "),
	})
	if code, out := runReady(t, dir, "PLN-0002", "--parent", "main"); code != 0 {
		t.Errorf("fenced example box: exit %d, want 0\n%s", code, out)
	}
}

// Input ready cannot judge is never "waiting": a parent ref that does not
// exist and a dropped after plan exit 2 with the cause named.
func TestReadyRefusesWhatCanNeverLand(t *testing.T) {
	dir := readyRepo(t, map[string]string{
		"plans/2026-10-01-a.md": "---\nid: PLN-0001\nstatus: dropped\n---\n# Plan\n\n### Task 1: Do it\n\n- [ ] Red: it fails\n",
		"plans/2026-10-01-b.md": readyPlan("PLN-0002", "PLN-0001", " "),
		"plans/2026-10-01-c.md": readyPlan("PLN-0003", "", " "),
	})
	for _, c := range []struct {
		args []string
		want string
	}{
		{[]string{"PLN-0003", "--parent", "mian"}, "mian"},
		{[]string{"PLN-0002", "--parent", "main"}, "PLN-0001 was dropped"},
	} {
		code, out := runReady(t, dir, c.args...)
		if code != 2 || !strings.Contains(out, c.want) {
			t.Errorf("%v: exit %d, want 2 with %q\n%s", c.args, code, c.want, out)
		}
	}
}

// A boxless task counts as done only when finished holds a real date, the
// way the board reads it. An empty or null finished does not land the plan.
func TestReadyBoxlessNeedsARealFinishedDate(t *testing.T) {
	boxless := func(finished string) string {
		return "---\nid: PLN-0001\n" + finished + "---\n# Plan\n\n### Task 1: Do it\n\n**verify:** it works\n"
	}
	for _, c := range []struct {
		finished string
		want     int
	}{
		{"finished: \"2026-10-01 11:00:00\"\n", 0},
		{"finished: \"\"\n", 1},
		{"finished: null\n", 1},
		{"finished: soon\n", 1},
	} {
		dir := readyRepo(t, map[string]string{
			"plans/2026-10-01-a.md": boxless(c.finished),
			"plans/2026-10-01-b.md": readyPlan("PLN-0002", "PLN-0001", " "),
		})
		if code, out := runReady(t, dir, "PLN-0002", "--parent", "main"); code != c.want {
			t.Errorf("%q: exit %d, want %d\n%s", c.finished, code, c.want, out)
		}
	}
}
