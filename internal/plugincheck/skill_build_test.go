package plugincheck

import "testing"

func TestSkillBuild(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "build",
		MaxLines: 680,
		Must: []string{
			"## Executors", "`subagent`", "`dispatch`", "`inline`",
			`model: "sonnet"`, `agent="task"`, "Do not ask whether to create a worktree",
			"../<repo>-<slug>", "git rev-parse --show-toplevel", "git add -A",
			"pm:tdd", "pm:review", "pm:land", "pm:dispatch", "no per-task reviewer",
			"references/house-rules.md", "implementer-prompt.md", "## Waves",
			"pmb tick", "pm: tick wave",
		},
		MustNot: []string{"superpowers:", "Would you like me to set up", "task-reviewer-prompt", "re-review-prompt", "## Final Review", "fix round R of 5"},
	})
}
