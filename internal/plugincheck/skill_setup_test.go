package plugincheck

import "testing"

func TestSkillSetup(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "setup",
		MaxLines: 65,
		Must: []string{
			"acta voice set", "--language", "--style", "--tone", "--clear-tone", "--repo-language",
			"acta voice show", "adhd", "plain", "full English name",
			"acta doctor", "acta doctor --fix", "--executor", "HERDR_ENV=1", "herdr",
			"--subagent-models split", "Claude Code only",
			"<!-- acta:begin -->", "<!-- acta:end -->",
			"only after a yes", "only to files that already exist", "never edits settings",
			"which part to change",
		},
		MustNot: []string{"superpowers:", "It never edits CLAUDE.md"},
	})
}
