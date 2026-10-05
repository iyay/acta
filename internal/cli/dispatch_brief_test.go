package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const briefPlanHead = `---
depth: minimal
---
# Demo Plan

**Spec:** ` + "`.acta/specs/demo-design.md`" + `

**Tests:** fast ` + "`scripts/test ./internal/cli`" + `, full ` + "`scripts/test --full`" + `

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: First thing

**verify:** property one holds on every path

- [ ] Failing test

### Task 2: Second thing

**verify:** property two holds on every path

### Task 3: Third thing

**verify:** property three holds on every path
`

const briefFixOne = `
## Fix round 1

### Task 4: Fix the first finding

**verify:** fix one holds
`

const briefFixTwo = `
## Fix round 2

### Task 5: Fix the second finding

**verify:** fix two holds

### Task 6: Fix the third finding

**verify:** fix three holds
`

// briefEnv makes a temp home with both MEMORY.md files and a rules file.
func briefEnv(t *testing.T) briefInput {
	t.Helper()
	dir := t.TempDir()
	home := filepath.Join(dir, "home")
	main := "/Users/me/proj.x"
	for _, p := range []string{
		filepath.Join(home, ".claude", "memory", "MEMORY.md"),
		filepath.Join(home, ".claude", "projects", "-Users-me-proj-x", "memory", "MEMORY.md"),
	} {
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rules := filepath.Join(dir, "house-rules.md")
	if err := os.WriteFile(rules, []byte("rules"), 0o644); err != nil {
		t.Fatal(err)
	}
	return briefInput{
		Rules: rules, Worktree: "/Users/me/proj-wt", Branch: "demo-branch",
		Parent: "main", Base: strings.Repeat("a", 40), Home: home, MainCheckout: main,
	}
}

// briefHasState checks the line every round kind carries: the recipient reads
// the plan's State before anything else. It fails when the line is missing,
// doubled, or placed after the tickets it must come before.
func briefHasState(t *testing.T, got string) {
	t.Helper()
	const line = "STATE: read `acta state <plan>` FIRST, before the tickets"
	if n := strings.Count(got, line); n != 1 {
		t.Errorf("want one %q line, got %d\n%s", line, n, got)
		return
	}
	if j := strings.Index(got, "TICKETS"); j >= 0 && strings.Index(got, line) > j {
		t.Errorf("the state line sits after the tickets\n%s", got)
	}
}

func TestBuildBriefFirstDispatch(t *testing.T) {
	in := briefEnv(t)
	in.Note = "mind the cache"
	src := []byte(briefPlanHead + briefFixOne)
	got, err := buildBrief(".acta/plans/demo.md", src, in)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"PLAN: .acta/plans/demo.md",
		".acta/specs/demo-design.md",
		"plans/demo#task-1", "property one holds on every path",
		"plans/demo#task-2", "property two holds on every path",
		"plans/demo#task-3", "property three holds on every path",
		"- Wave 1: Task 1, Task 2\n- Wave 2: Task 3",
		"SKILL: load build",
		"WORKTREE: /Users/me/proj-wt",
		"branch demo-branch", "parent main", "base " + in.Base,
		in.Rules,
		filepath.Join(in.Home, ".claude", "memory", "MEMORY.md"),
		filepath.Join(in.Home, ".claude", "projects", "-Users-me-proj-x", "memory", "MEMORY.md"),
		"scripts/test ./internal/cli",
		"never run bare `go test`", "pre-tool hook",
		"NOTE: mind the cache",
		"REPLY-BACK: after the last commit the build skill runs `acta reply-back`.",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("brief misses %q\n%s", want, got)
		}
	}
	// A fix round task must not leak into the first dispatch.
	if strings.Contains(got, "task-4") || strings.Contains(got, "fix one holds") {
		t.Errorf("first dispatch took a fix round task\n%s", got)
	}
	if n := strings.Count(got, "REPLY-BACK:"); n != 1 {
		t.Errorf("want one REPLY-BACK line, got %d", n)
	}
	briefHasState(t, got)
}

func TestBuildBriefFixRounds(t *testing.T) {
	in := briefEnv(t)
	in.Round = "r2"
	src := []byte(briefPlanHead + briefFixOne + briefFixTwo)
	got, err := buildBrief(".acta/plans/demo.md", src, in)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"task-5", "fix two holds", "task-6", "fix three holds", "- Wave 1: Task 1, Task 2"} {
		if !strings.Contains(got, want) {
			t.Errorf("brief misses %q\n%s", want, got)
		}
	}
	for _, bad := range []string{"task-1", "task-2", "task-3", "task-4", "property one", "fix one holds"} {
		if strings.Contains(got, bad) {
			t.Errorf("last fix round took %q\n%s", bad, got)
		}
	}
	// One fix round: that one is the last.
	got, err = buildBrief(".acta/plans/demo.md", []byte(briefPlanHead+briefFixOne), in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "task-4") || strings.Contains(got, "task-1") {
		t.Errorf("single fix round wrong\n%s", got)
	}
	briefHasState(t, got)
}

func TestBuildBriefPolish(t *testing.T) {
	in := briefEnv(t)
	in.Round = "polish"
	src := []byte(briefPlanHead + briefFixOne)

	in.Note = "  \n"
	if _, err := buildBrief("p.md", src, in); err == nil {
		t.Error("polish with a blank note must fail")
	}
	in.Note = "[fix] rename the helper"
	got, err := buildBrief(".acta/plans/demo.md", src, in)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "NOTE: [fix] rename the helper") {
		t.Errorf("polish brief misses the note\n%s", got)
	}
	if strings.Contains(got, "TICKETS") || strings.Contains(got, "task-") {
		t.Errorf("polish must take no tasks\n%s", got)
	}
	briefHasState(t, got)
}

func TestBuildBriefRefusals(t *testing.T) {
	good := briefPlanHead
	cases := []struct {
		name   string
		src    string
		mut    func(*briefInput)
		wantIn string
	}{
		{"fix round missing", good, func(in *briefInput) { in.Round = "r1" }, "Fix round"},
		{"bad round slug", good, func(in *briefInput) { in.Round = "Bad Round" }, "round"},
		{"task without verify", strings.Replace(good, "**verify:** property two holds on every path", "no claim here", 1), nil, "task 2"},
		{"fix task without verify", good + "\n## Fix round 1\n\n### Task 4: Fix\n\nnothing\n", func(in *briefInput) { in.Round = "r1" }, "task 4"},
		{"no tasks", "# Empty\n\n**Tests:** `x`\n\n## Waves\n\nnone\n", nil, "no tasks"},
		{"no waves", strings.Replace(good, "## Waves", "## Order", 1), nil, "Waves"},
		{"no tests line", strings.Replace(good, "**Tests:**", "Tests:", 1), nil, "Tests"},
		{"rules file missing", good, func(in *briefInput) { in.Rules = filepath.Join(in.Home, "nope.md") }, "house rules"},
		{"rules path relative", good, func(in *briefInput) { in.Rules = "references/house-rules.md" }, "absolute"},
		{"worktree relative", good, func(in *briefInput) { in.Worktree = "wt" }, "worktree"},
		{"branch empty", good, func(in *briefInput) { in.Branch = "" }, "branch"},
		{"parent empty", good, func(in *briefInput) { in.Parent = "" }, "parent"},
		{"base not a sha", good, func(in *briefInput) { in.Base = "abc" }, "base"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			in := briefEnv(t)
			if c.mut != nil {
				c.mut(&in)
			}
			got, err := buildBrief(".acta/plans/demo.md", []byte(c.src), in)
			if err == nil {
				t.Fatalf("want an error, got brief:\n%s", got)
			}
			if got != "" {
				t.Errorf("an error must come with no text, got %q", got)
			}
			if !strings.Contains(err.Error(), c.wantIn) {
				t.Errorf("error %q misses %q", err, c.wantIn)
			}
		})
	}
}

func TestBuildBriefMemoryOnlyWhenPresent(t *testing.T) {
	in := briefEnv(t)
	if err := os.Remove(filepath.Join(in.Home, ".claude", "memory", "MEMORY.md")); err != nil {
		t.Fatal(err)
	}
	got, err := buildBrief("p.md", []byte(briefPlanHead), in)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got, filepath.Join(in.Home, ".claude", "memory", "MEMORY.md")) {
		t.Errorf("brief names a MEMORY.md that does not exist\n%s", got)
	}
	if !strings.Contains(got, "-Users-me-proj-x") {
		t.Errorf("brief misses the project memory\n%s", got)
	}
}

func TestBriefTasksIgnoresFencedHeadings(t *testing.T) {
	// A fix round heading inside a code block is an example, not a section.
	src := briefPlanHead + "\n```\n## Fix round 9\n```\n"
	if _, err := briefTasks([]byte(src), "r1"); err == nil {
		t.Error("a fenced heading must not count as a fix round")
	}
	tasks, err := briefTasks([]byte(briefPlanHead+briefFixOne+briefFixTwo), "")
	if err != nil || len(tasks) != 3 {
		t.Errorf("first dispatch: got %d tasks, err %v", len(tasks), err)
	}
}
