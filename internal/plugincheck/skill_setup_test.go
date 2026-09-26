package plugincheck

import "testing"

func TestSkillSetup(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "setup",
		MaxLines: 60,
		Must: []string{
			"pmb voice set", "--language", "--style", "--tone", "--clear-tone", "--repo-language",
			"pmb voice show", "adhd", "plain", "It never edits CLAUDE.md", "full English name",
		},
		MustNot: []string{"superpowers:"},
	})
}
