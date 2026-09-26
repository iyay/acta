package plugincheck

import "testing"

func TestSkillPlan(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "plan",
		MaxLines: 260,
		Must: []string{
			".acta/plans/", "acta:build", "verify:", "## Waves", "ponytail-lazy",
			"never as the one case", "wait for a yes", "`subagent`", "`dispatch`", "`inline`",
			"No Placeholders", "Global Constraints", "**Spec:** none (Bounded, approved in chat on <date>)",
			"run `acta id` right after",
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "executing-plans", "Two execution options", "plan-document-reviewer"},
	})
}
