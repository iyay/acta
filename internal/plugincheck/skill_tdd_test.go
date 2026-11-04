package plugincheck

import (
	"strings"
	"testing"
)

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
			// Only the items that mean delete say delete; explore, hard to test and no tests do not.
			"Each of these means: delete the code, start over with a test first.",
			"Other answers:",
			"Throw the exploration away, then write the test.",
			// The suite is one run in land, never per task.
			"The full suite runs only in `acta:land`",
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

// TestSkillTDDItemOrder keeps the delete header first and the three non-delete
// items under Other answers, so none of them reads as a delete item.
func TestSkillTDDItemOrder(t *testing.T) {
	body := readFile(t, "skills", "tdd", "SKILL.md")
	at := func(s string) int {
		i := strings.Index(body, s)
		if i < 0 {
			t.Fatalf("SKILL.md is missing %q", s)
		}
		return i
	}
	deleteHeader := at("Each of these means: delete the code")
	lastDelete := at("Deleting X hours")
	other := at("Other answers:")
	if !(deleteHeader < lastDelete && lastDelete < other) {
		t.Errorf("Other answers: must come after the delete header and the Deleting X hours item")
	}
	for _, item := range []string{"Need to explore first", "Hard to test", "Existing code has no tests"} {
		if at(item) < other {
			t.Errorf("%q must sit under Other answers:", item)
		}
	}
}
