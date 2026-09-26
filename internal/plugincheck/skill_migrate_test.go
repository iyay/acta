package plugincheck

import "testing"

func TestSkillMigrate(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "migrate",
		MaxLines: 120,
		Must: []string{
			"pmb migrate superpowers", "--apply", "Show the table", "Write nothing before that",
			"one commit", "Never push", "git log --diff-filter=A", "pmb list --all --json", "legacy",
		},
		MustNot: []string{"superpowers:"},
	})
}
