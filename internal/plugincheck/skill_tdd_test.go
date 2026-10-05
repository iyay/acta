package plugincheck

import "testing"

func TestSkillTDD(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "tdd",
		MaxLines: 560,
		Must: []string{
			"The Iron Law", "## Cover the plan, not the universe", "## Quality bar",
			"reverted", "regression test", "flag-off", "Test tooling", "writing-good-tests.md",
			"the full suite runs in `acta:land`",
		},
		MustNot: []string{"superpowers:", "- [ ] All tests pass", "- Configuration files", "## Final Rule", "Add tests for existing code"},
	})
}
