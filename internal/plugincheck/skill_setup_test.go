package plugincheck

import "testing"

func TestSkillSetup(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "setup",
		MaxLines: 60,
		Must: []string{
			"acta voice set", "--language", "--style", "--tone", "--clear-tone", "--repo-language",
			"acta voice show", "adhd", "plain", "It never edits CLAUDE.md", "full English name",
		},
		MustNot: []string{"superpowers:"},
	})
}
