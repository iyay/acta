package board

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"pm-board/internal/config"
)

func writeFile(t *testing.T, path, body string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

const planAgents = `# Plan A

**Spec:** .pm/specs/2026-09-20-a.md

### Task 1: One
- [x] a

### Task 2: Two
- [x] a
- [ ] b

### Task 3: Three
- [ ] a
`

const bugPlanAgents = `---
parent: bugs/2026-09-24-crash
---
# Fix the crash

### Task 1: Guard it
- [ ] a
`

// agentsOf is what every item of a board says about itself, so a test can name
// the ones it cares about and still see the rest.
func agentsOf(b *Board) string {
	var out []string
	for _, it := range b.Items {
		out = append(out, it.ID+"="+it.Agent)
	}
	return strings.Join(out, " ")
}

func TestAgentsOnOpenTasksAndTheirParents(t *testing.T) {
	main := tree(t, map[string]string{
		".pm/specs/2026-09-20-a.md":    specA,
		".pm/plans/2026-09-21-a.md":    planAgents,
		".pm/bugs/2026-09-24-crash.md": "# App crashes on start\n",
		".pm/plans/2026-09-25-fix.md":  bugPlanAgents,
		".pm/.agents.json": `{
  "plans/2026-09-21-a#task-1": {"agent": "bob", "at": "2026-09-26T09:00:00+07:00"},
  "plans/2026-09-21-a#task-2": {"agent": "claude", "at": "2026-09-26T10:00:00+07:00"},
  "plans/2026-09-21-a#task-3": {"agent": "omp", "at": "2026-09-26T11:00:00+07:00"},
  "plans/2026-09-25-fix#task-1": {"agent": "claude", "at": "2026-09-26T11:30:00+07:00"}
}`,
	})
	b, err := Load(main)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"plans/2026-09-21-a#task-1":   "", // done, so nobody works on it any more
		"plans/2026-09-21-a#task-2":   "claude",
		"plans/2026-09-21-a#task-3":   "omp",
		"plans/2026-09-21-a":          "claude,omp", // bob's task is done, so he is out
		"specs/2026-09-20-a":          "claude,omp",
		"plans/2026-09-25-fix#task-1": "claude",
		"plans/2026-09-25-fix":        "claude",
		"bugs/2026-09-24-crash":       "claude",
	}
	for id, agent := range want {
		it := b.Get(id)
		if it == nil {
			t.Fatalf("no item %s; board has %s", id, agentsOf(b))
		}
		if it.Agent != agent {
			t.Fatalf("%s agent = %q, want %q (all: %s)", id, it.Agent, agent, agentsOf(b))
		}
	}
}

func TestNoAgentsFileGivesNoAgentsAndNoError(t *testing.T) {
	main := tree(t, map[string]string{
		".pm/specs/2026-09-20-a.md": specA,
		".pm/plans/2026-09-21-a.md": planAhead,
	})
	b, err := Load(main)
	if err != nil {
		t.Fatal(err)
	}
	for _, it := range b.Items {
		if it.Agent != "" {
			t.Fatalf("%s agent = %q, want none without an agents file", it.ID, it.Agent)
		}
	}
}

func TestBrokenAgentsFileGivesNoAgentsAndNoError(t *testing.T) {
	main := tree(t, map[string]string{
		".pm/specs/2026-09-20-a.md": specA,
		".pm/plans/2026-09-21-a.md": planBehind,
		".pm/.agents.json":          "{not json at all",
	})
	b, err := Load(main)
	if err != nil {
		t.Fatalf("a broken agents file must not fail the load: %v", err)
	}
	for _, it := range b.Items {
		if it.Agent != "" {
			t.Fatalf("%s agent = %q, want none from a broken agents file", it.ID, it.Agent)
		}
	}
	// A broken file in one root must not blind the others.
	wt := tree(t, map[string]string{
		".pm/plans/2026-09-22-b.md": planBehind,
		".pm/.agents.json":          `{"plans/2026-09-22-b#task-1": {"agent": "omp", "at": "2026-09-26T12:00:00+07:00"}}`,
	})
	b, err = LoadTrees(main, []Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatalf("a broken agents file must not fail the load: %v", err)
	}
	if it := b.Get("plans/2026-09-22-b#task-1"); it == nil || it.Agent != "omp" {
		t.Fatalf("task of a healthy worktree = %+v, want it to show omp", it)
	}
	if it := b.Get("plans/2026-09-21-a#task-1"); it == nil || it.Agent != "" {
		t.Fatalf("task under the broken root = %+v, want it to show no agent", it)
	}
}

func TestNewestAgentWinsAcrossRoots(t *testing.T) {
	plan := `# Plan A

### Task 1: One
- [ ] a

### Task 2: Two
- [ ] a

### Task 3: Three
- [ ] a
`
	main := tree(t, map[string]string{
		".pm/plans/2026-09-21-a.md": plan,
		".pm/.agents.json": `{
  "plans/2026-09-21-a#task-1": {"agent": "omp", "at": "2026-09-26T12:00:00+07:00"},
  "plans/2026-09-21-a#task-2": {"agent": "claude", "at": "2026-09-26T12:00:00+07:00"}
}`,
	})
	wt := tree(t, map[string]string{
		".pm/plans/2026-09-21-a.md": plan,
		".pm/plans/2026-09-22-b.md": plan,
		".pm/.agents.json": `{
  "plans/2026-09-21-a#task-1": {"agent": "claude", "at": "2026-09-26T09:00:00+07:00"},
  "plans/2026-09-21-a#task-2": {"agent": "omp", "at": "2026-09-26T13:00:00+07:00"},
  "plans/2026-09-22-b#task-1": {"agent": "claude", "at": "2026-09-26T13:00:00+07:00"}
}`,
	})
	b, err := LoadTrees(main, []Tree{{Cfg: wt, Branch: "feat"}})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"plans/2026-09-21-a#task-1": "omp",    // main ticked it last
		"plans/2026-09-21-a#task-2": "omp",    // the worktree ticked it last
		"plans/2026-09-22-b#task-1": "claude", // only the worktree knows it
		"plans/2026-09-21-a":        "omp",
		"plans/2026-09-22-b":        "claude",
	}
	for id, agent := range want {
		it := b.Get(id)
		if it == nil {
			t.Fatalf("no item %s; board has %s", id, agentsOf(b))
		}
		if it.Agent != agent {
			t.Fatalf("%s agent = %q, want %q (all: %s)", id, it.Agent, agent, agentsOf(b))
		}
	}
}

func TestBranchFromGitHasNoAgents(t *testing.T) {
	plan := `# Plan A

### Task 1: One
- [ ] a
`
	main := tree(t, map[string]string{".pm/plans/2026-09-21-a.md": plan})
	// The branch folder holds a record for the branch's own plan, which must
	// stay unread: a branch read from git is not on disk.
	gone := filepath.Join(t.TempDir(), "gone")
	writeFile(t, filepath.Join(gone, ".pm", ".agents.json"), `{
  "plans/2026-09-27-x#task-1": {"agent": "claude", "at": "2026-09-26T12:00:00+07:00"}
}`)
	b, err := LoadTrees(main, []Tree{{Cfg: config.Default(gone), Branch: "feat-x", Files: map[string][]byte{
		".pm/plans/2026-09-27-x.md": []byte(plan),
	}}})
	if err != nil {
		t.Fatal(err)
	}
	it := b.Get("plans/2026-09-27-x#task-1")
	if it == nil {
		t.Fatalf("no item from the branch; board has %s", agentsOf(b))
	}
	if it.Agent != "" {
		t.Fatalf("a task read from git has agent %q, want none", it.Agent)
	}
	if p := b.Get("plans/2026-09-27-x"); p.Agent != "" {
		t.Fatalf("a plan read from git has agent %q, want none", p.Agent)
	}
}
