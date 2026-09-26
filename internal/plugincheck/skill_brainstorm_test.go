package plugincheck

import "testing"

func TestSkillBrainstorm(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "brainstorm",
		MaxLines: 360,
		Must: []string{
			"Spike", "Bounded", "Architectural", "HARD-GATE",
			".acta/specs/", "acta:plan", "CONTEXT.md", "docs/adr/",
			"trust boundary", "each gets its own yes", "acta list --json",
			"run " + "`acta id`" + ` right after`,
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "Visual Companion", "visual-companion", "writing-plans", "elements-of-style"},
	})
}
