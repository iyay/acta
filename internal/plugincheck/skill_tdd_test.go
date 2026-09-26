package plugincheck

import "testing"

func TestSkillTDD(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "tdd",
		MaxLines: 560,
		Must: []string{
			"The Iron Law", "## Cover the plan, not the universe", "## Quality bar",
			"reverted", "regression test", "flag-off", "Test tooling", "writing-good-tests.md",
		},
		MustNot: []string{"superpowers:"},
	})
}
