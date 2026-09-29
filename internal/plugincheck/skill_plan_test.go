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
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "executing-plans", "Two execution options", "plan-document-reviewer",
			"names the DEBT ids it closes",
			"If working in an isolated worktree",
		},
	})
}
