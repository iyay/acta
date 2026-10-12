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
			// One test per plan case, at the layer that owns it.
			"test it once, at the layer that owns it",
			"wiring only: one success path and one error passed up from below",
			"only where the input crosses a trust boundary",
			"Every case the plan names has a test, and no other",
			// CheckSkill reads every markdown file in the folder, so this one is in writing-good-tests.md.
			"Missing validation the plan names",
		},
		MustNot: []string{
			"superpowers:", "- [ ] All tests pass", "- Configuration files", "## Final Rule", "Add tests for existing code",
			"Every new function/method has a test", "Edge cases and errors covered", "where the input can take them",
			"Missing validation for zero, empty, nil",
		},
	})
}
