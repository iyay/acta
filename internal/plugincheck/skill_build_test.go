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
		MaxLines: 680,
		Must: []string{
			"## Executors", "`subagent`", "`dispatch`", "`inline`",
			`model: "sonnet"`, `agent="task"`, "Do not ask whether to create a worktree",
			"../<repo>-<slug>", "git rev-parse --show-toplevel", "git add -A",
			"pm:tdd", "pm:review", "pm:land", "pm:dispatch", "no per-task reviewer",
			"pmb tick plans/<stem>#task-N --step <n>", "pmb tick [TASK_ID] --step <n>", "pm: tick wave",
			"pmb tick plans/<stem>#task-N --all",
			"pmb show <plan id> --json", "progress.done",
			`git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"`, "`$PARENT` is the parent branch recorded above",
		},
		MustNot: []string{"superpowers:", "Would you like me to set up", "task-reviewer-prompt", "re-review-prompt", "## Final Review", "fix round R of 5",
			"default to `.worktrees/`", "Step 0 consent", "ls -d .worktrees",
			`"$LOCATION/$BRANCH_NAME"`, "pmb tick <task-id>", "run it again with --all"},
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
	if !strings.Contains(string(b), "pmb tick [TASK_ID] --all") {
		t.Error("implementer-prompt.md missing \"pmb tick [TASK_ID] --all\"")
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
	for _, want := range []string{"run `pmb tick plans/<stem>#task-N --all`", "pmb show <plan id> --json", "progress.done"} {
		if !strings.Contains(txt, want) {
			t.Errorf("build/SKILL.md missing %q", want)
		}
	}
	if strings.Contains(txt, "run it again with --all") {
		t.Error("build/SKILL.md still says \"run it again with --all\"")
	}
}
