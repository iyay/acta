package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillBuild(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "build",
		MaxLines: 682,
		Must: []string{
			"## Executors", "`subagent`", "`dispatch`", "`inline`",
			`model: "sonnet"`, `agent="task"`, "Do not ask whether to create a worktree",
			"../<repo>-<slug>", "git rev-parse --show-toplevel", "git add -A",
			"acta:tdd", "acta:review", "acta:land", "acta:dispatch", "no per-task reviewer",
			"acta tick plans/<stem>#task-N --step <n>", "acta tick [TASK_ID] --step <n>", "acta: tick wave",
			"acta tick plans/<stem>#task-N --all",
			"acta show <plan id> --json", "progress.done",
			`git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"`, "`$PARENT` is the parent branch recorded above",
			"acta voice show", "build_executor",
			"The spec and plan from acta:brainstorm and acta:plan stay on main",
			"do not edit its plan on main",
			"The full suite waits for `acta:land`", "scripts/test",
			"the executor `acta voice show` names, else asks which one",
		},
		MustNot: []string{"superpowers:", "Would you like me to set up", "task-reviewer-prompt", "re-review-prompt", "## Final Review", "fix round R of 5",
			"default to `.worktrees/`", "Step 0 consent", "ls -d .worktrees",
			`"$LOCATION/$BRANCH_NAME"`, "acta tick <task-id>", "run it again with --all",
			"or `go test ./...`, whichever the project uses",
			"run the full test suite and the type checks, show the output, then use",
			"subagent (default", "`subagent` (default)",
		},
	})
}

// TestBuildTickRuleEverywhere checks implementer-prompt.md on its own, not
// folded into the whole build/ folder text. The implementer only ever reads
// this one file, so the --all rule has to live in it directly.
func TestBuildTickRuleEverywhere(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "acta tick [TASK_ID] --all") {
		t.Error("implementer-prompt.md missing \"acta tick [TASK_ID] --all\"")
	}
}

// TestBuildAgentFlag reads the implementer prompt on its own. The board shows
// who works on a task, and the omp harness sets no agent variable.
func TestBuildAgentFlag(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "--agent omp") {
		t.Error("implementer-prompt.md missing \"--agent omp\"")
	}
}

// TestBuildSkillTickCommands reads SKILL.md on its own. CheckSkill looks at
// the whole build folder, so a revert of SKILL.md alone could stay green.
func TestBuildSkillTickCommands(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{"run `acta tick plans/<stem>#task-N --all`", "acta show <plan id> --json", "progress.done"} {
		if !strings.Contains(txt, want) {
			t.Errorf("build/SKILL.md missing %q", want)
		}
	}
	if strings.Contains(txt, "run it again with --all") {
		t.Error("build/SKILL.md still says \"run it again with --all\"")
	}
}

// TestBuildAgentFlagOnlyWhenTheHarnessSetsNothing guards the prompt text that
// the Claude Code subagent executor reads too. A bare --agent omp on the tick
// overrides AI_AGENT, so Claude's work would be recorded as omp.
func TestBuildAgentFlagOnlyWhenTheHarnessSetsNothing(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, bad := range []string{"--step <n> --agent omp", "--all --agent omp"} {
		if strings.Contains(txt, bad) {
			t.Errorf("implementer-prompt.md passes the flag unconditionally: %q", bad)
		}
	}
	if !strings.Contains(txt, "AI_AGENT") {
		t.Error("implementer-prompt.md never says the flag is for a harness with no AI_AGENT")
	}
}

// TestBuildStartTickPrompt checks implementer-prompt.md on its own. The
// implementer only ever reads this one file, so the --start rule has to live
// in it directly, before the failing-test step.
func TestBuildStartTickPrompt(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	const want = "Very first action: run `acta tick [TASK_ID] --start`"
	if !strings.Contains(txt, want) {
		t.Errorf("implementer-prompt.md missing %q", want)
	}
	if strings.Index(txt, "--start") > strings.Index(txt, "failing test first") {
		t.Error("implementer-prompt.md names --start after the failing-test step; it must come first")
	}
}

// TestBuildStartTickSkill checks SKILL.md on its own. CheckSkill looks at the
// whole build folder, so a revert of SKILL.md alone could stay green.
func TestBuildStartTickSkill(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	const want = "very first action on every task: run `acta tick plans/<stem>#task-N --start`"
	if !strings.Contains(txt, want) {
		t.Errorf("build/SKILL.md missing %q", want)
	}
	if !strings.Contains(txt, "Claude: pass no flag; omp: add `--agent omp`") {
		t.Error("build/SKILL.md never states the --agent rule for --start")
	}
}

// TestImplementerPromptNoFullSuite reads implementer-prompt.md on its own. The
// implementer only reads this file, and the full suite belongs to acta:land.
func TestImplementerPromptNoFullSuite(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The prompt wraps lines, so join the words before matching a sentence.
	txt := strings.Join(strings.Fields(string(b)), " ")
	if strings.Contains(txt, "run the full suite once before committing") {
		t.Error("implementer-prompt.md still says \"run the full suite once before committing\"")
	}
	if !strings.Contains(txt, "never the full suite") {
		t.Error("implementer-prompt.md missing \"never the full suite\"")
	}
}
