package plugincheck

import "testing"

func TestSkillMigrate(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "migrate",
		MaxLines: 120,
		Must: []string{
			"git mv", "Show the table", "Write nothing before that",
			"one commit", "Never push", "git log --diff-filter=A", "acta list --all --json", "legacy",
		},
		MustNot: []string{"superpowers:", "acta migrate superpowers"},
	})
}
