package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func readSkillFile(t *testing.T, skill, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", skill, name))
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

// TestPlanSplitText keeps the plan split rules where each step reads them:
// slice cuts and checks, build checks ready before it starts, land writes the
// spec dates before it moves last-land, and the setup fallback asks it.
func TestPlanSplitText(t *testing.T) {
	split := readSkillFile(t, "slice", "split.md")
	if first := strings.SplitN(split, "\n", 2)[0]; !strings.HasPrefix(first, "Read this") {
		t.Errorf("slice/split.md must start with one line saying when to read it, got %q", first)
	}
	checks := []struct{ file, text string }{
		{readSkillFile(t, "slice", "SKILL.md"), "split.md"},
		{readSkillFile(t, "slice", "SKILL.md"), "`/slice split`"},
		{readSkillFile(t, "slice", "SKILL.md"), "`plan_split`"},
		{split, "2 to 4 plans"},
		{split, "never touch the same file"},
		{split, "after: [PLN-"},
		{split, "acta ready"},
		{split, "fix the list with `acta set <plan id> after <ids>`"},
		{split, "do not start a build"},
		{readSkillFile(t, "build", "SKILL.md"), "acta ready <plan id> --parent <parent>"},
		{readSkillFile(t, "build", "SKILL.md"), "--head"},
		{readSkillFile(t, "build", "SKILL.md"), "Any non-zero exit stops the build"},
		{readSkillFile(t, "land", "tidy.md"), "acta spec-dates <plan hash>"},
		{readSkillFile(t, "land", "SKILL.md"), "acta spec-dates <plan hash>"},
		// A tick in a worktree never dates a spec now, so land dates every
		// spec, not only one with several plans.
		{readSkillFile(t, "land", "tidy.md"), "the dates of the plan's spec"},
		{readSkillFile(t, "land", "SKILL.md"), "the dates of the plan's spec"},
		{readSkillFile(t, "setup", "SKILL.md"), "`--plan-split` when it shows `plan_split: off (default)`"},
		{readSkillFile(t, "slice", "SKILL.md"), "A split (split.md) never starts a build"},
		{split, "Right after `acta id`"},
		{split, "## Build brief"},
		{readSkillFile(t, "build", "SKILL.md"), "## Build brief"},
	}
	for _, c := range checks {
		if !strings.Contains(c.file, c.text) {
			t.Errorf("missing %q", c.text)
		}
	}
	// spec-dates must run before last-land moves, or the next land meets a
	// spec that is not committed.
	tidy := readSkillFile(t, "land", "tidy.md")
	dates, last := strings.Index(tidy, "acta spec-dates"), strings.Index(tidy, "git update-ref refs/acta/last-land")
	if dates < 0 || last < 0 || dates > last {
		t.Errorf("land/tidy.md must run acta spec-dates before it moves refs/acta/last-land (got %d, %d)", dates, last)
	}
}
