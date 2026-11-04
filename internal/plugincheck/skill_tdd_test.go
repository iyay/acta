package plugincheck

import "testing"

func TestSkillTDD(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "tdd",
		MaxLines: 560,
		Must: []string{
			"The Iron Law", "## Cover the plan, not the universe", "## Quality bar",
			"reverted", "regression test", "flag-off", "Test tooling", "writing-good-tests.md",
			// One test per plan case, at the layer that owns it.
			"test it once, at the layer that owns it",
			"wiring only: one success path and one error passed up from below",
			"only where the input crosses a trust boundary",
			"Write the plan's cases and no others",
			"the plan names which adversarial inputs to try, and you write the ones it names",
			// One block replaces the rationalizations table and the red flags list.
			"## Rationalizations - STOP and Start Over", "Skipping the red run", "Code before the test",
			"delete the code, start over",
			// CheckSkill reads every markdown file in the folder, so this one is in writing-good-tests.md.
			"Missing validation the plan names",
			// Tests go through the seam the plan names, and same-path inputs share one table test.
			"through the seam the plan names", "one table-driven test, one row each",
			"Never write all the tests before any code", "with none named the public interface callers use",
			// writing-good-tests.md: names follow the project's own words.
			"`.acta/wiki/glossary.md`",
		},
		MustNot: []string{
			"superpowers:", "- [ ] All tests pass", "- Configuration files", "## Final Rule", "Add tests for existing code",
			"Every new function/method has a test", "Edge cases and errors covered", "where the input can take them",
			"Missing validation for zero, empty, nil",
			"## Common Rationalizations", "## Red Flags", "## Example: Bug Fix", "## Verification Checklist",
		},
	})
}
