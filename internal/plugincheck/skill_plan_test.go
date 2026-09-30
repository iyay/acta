package plugincheck

import "testing"

func TestSkillPlan(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "plan",
		MaxLines: 262,
		Must: []string{
			".acta/plans/", "acta:build", "verify:", "## Waves", "ponytail-lazy",
			"never as the one case", "wait for a yes", "`subagent`", "`dispatch`", "`inline`",
			"No Placeholders", "Global Constraints", "**Spec:** none (Bounded, approved in chat on <date>)",
			"run `acta id` right after", "parent: debt/",
			"never holds a step that can only happen after landing",
			"closes:", "parent: debt/<stem>",
			"Commit the plan on main", "only after the plan is approved",
			"A fix round, or any change to a plan whose build is running, goes in that build's worktree",
			"## Plan depth", "`/plan minimal`", "depth: minimal", "`plan_depth`",
			"apply to `full` plans only",
			"invoke `acta:build` in the same turn",
			"`/plan full`", "do not ask when one answers", "then `full`.",
			"no Architecture, Tech Stack, File map or Interfaces", "No code blocks.",
			"run `acta config show`", "When it prints `build_executor: <name>`, that executor is chosen",
			"do not ask",
			"## Test commands", "Never `./...` in a task", "**Tests:**",
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "executing-plans", "Two execution options", "plan-document-reviewer",
			"names the DEBT ids it closes",
			"If working in an isolated worktree",
			"Ask which executor only if the user has not said",
		},
	})
}
