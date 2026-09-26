package plugincheck

import "testing"

func TestSkillBug(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "bug",
		MaxLines: 80,
		Must:     []string{"pmb bug new", "## Symptom", "fixed_in", "parent: bugs/", "Exit codes", "Never push"},
		MustNot:  []string{"git-bug", "superpowers:"},
	})
}
