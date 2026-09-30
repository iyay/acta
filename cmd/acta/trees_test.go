package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestListShowsWorktreeProgress(t *testing.T) {
	t.Parallel()

	repo := fixtureRepo(t)
	wt := repo + "-feat"
	if out, err := exec.Command("git", "-C", repo, "worktree", "add", "-q", wt, "-b", "feat").CombinedOutput(); err != nil {
		t.Fatalf("worktree add: %v %s", err, out)
	}
	t.Cleanup(func() { os.RemoveAll(wt) })
	plan := filepath.Join(wt, ".acta/plans/2026-09-21-alpha.md")
	b, _ := os.ReadFile(plan)
	ticked := []byte(string(b))
	for i := 0; i+5 <= len(ticked); i++ {
		if string(ticked[i:i+5]) == "- [ ]" {
			copy(ticked[i:i+5], "- [x]")
			break
		}
	}
	if err := os.WriteFile(plan, ticked, 0o644); err != nil {
		t.Fatal(err)
	}

	out, _, code := acta(t, repo, "", "show", "plans/2026-09-21-alpha#task-2", "--json")
	if code != 0 {
		t.Fatalf("exit %d", code)
	}
	var it struct {
		Status   string `json:"status"`
		Worktree string `json:"worktree"`
	}
	if err := json.Unmarshal([]byte(out), &it); err != nil {
		t.Fatal(err)
	}
	if it.Status != "done" || it.Worktree != "feat" {
		t.Fatalf("task from the worktree = %+v", it)
	}
	out, _, _ = acta(t, repo, "", "show", "specs/2026-09-22-beta", "--json")
	if err := json.Unmarshal([]byte(out), &it); err != nil || it.Worktree != "" {
		t.Fatalf("main item must say worktree \"\": %s", out)
	}
	if !strings.Contains(out, `"on_disk": true`) {
		t.Fatalf("main item must be on disk: %s", out)
	}
}

func TestListShowsBranchItems(t *testing.T) {
	t.Parallel()

	repo := fixtureRepo(t)
	gitDo := func(args ...string) {
		if out, err := exec.Command("git", append([]string{"-C", repo}, args...)...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v %s", args, err, out)
		}
	}
	gitDo("checkout", "-q", "-b", "feat-x")
	if err := os.WriteFile(filepath.Join(repo, ".acta/specs/2026-09-28-branch.md"), []byte("# Branch story\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitDo("add", ".")
	gitDo("commit", "-q", "-m", "spec on feat-x")
	gitDo("checkout", "-q", "main")

	out, _, code := acta(t, repo, "", "show", "specs/2026-09-28-branch", "--json")
	if code != 0 || !strings.Contains(out, `"worktree": "feat-x"`) || !strings.Contains(out, `"on_disk": false`) ||
		!strings.Contains(out, `"path": "feat-x:.acta/specs/2026-09-28-branch.md"`) {
		t.Fatalf("branch item: exit %d %s", code, out)
	}
}
