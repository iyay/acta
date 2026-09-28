package plugincheck

import "testing"

func TestSkillBrainstorm(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "brainstorm",
		MaxLines: 395,
		Must: []string{
			"Spike", "Bounded", "Architectural", "HARD-GATE",
			".acta/specs/", "acta:plan", "CONTEXT.md", "docs/adr/",
			"trust boundary", "each gets its own yes", "acta list --json",
			"run " + "`acta id`" + ` right after`, "parent: debt/",
			"acta scratch new", "acta scratch add", "status brainstorming",
			"parent: scratch/", "One Architectural brainstorm per session",
			"claude --bg 'brainstorm SCRATCH-", "HERDR_ENV=1",
			"pbcopy", "wl-copy", "xclip", "OSC 52", "press ←", "Spike and Bounded",
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "Visual Companion", "visual-companion", "writing-plans", "elements-of-style"},
	})
}
