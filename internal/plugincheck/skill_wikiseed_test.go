package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestSkillWikiSeed pins the first-time wiki seed flow: migrate holds a
// "Seed the wiki" section with the five spec steps, and setup offers it in
// one line only when the repo has code and no wiki page yet. Losing the
// section, the table or the yes must fail this test.
func TestSkillWikiSeed(t *testing.T) {
	migrate, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "migrate", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	setup, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "setup", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"## Seed the wiki",
		"how many sources will be read",
		"by convention, read-only",
		"Keep only facts the code cannot tell",
		"Show one table",
		"drop with the reason",
		"Wait for a yes",
		"in a worktree",
		"acta wiki check",
	} {
		if !strings.Contains(string(migrate), want) {
			t.Errorf("migrate SKILL.md missing %q", want)
		}
	}
	for _, want := range []string{
		"has code",
		"no page",
		"seed the wiki",
		"yes",
		"migrate",
	} {
		if !strings.Contains(string(setup), want) {
			t.Errorf("setup SKILL.md missing %q", want)
		}
	}
}
