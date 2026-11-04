package plugincheck

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modelsPara = "**Subagent models.** Name no model for any subagent; the harness default picks it."

func TestModelsParagraphInFourSkills(t *testing.T) {
	for _, s := range []string{"slice", "shape", "debug", "build"} {
		// Build keeps the paragraph in subagent.md, the only executor that
		// starts subagents.
		file := "SKILL.md"
		if s == "build" {
			file = "subagent.md"
		}
		raw, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", s, file))
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(string(raw), modelsPara); n != 1 {
			t.Errorf("skills/%s: models paragraph found %d times, want 1", s, n)
		}
	}
}

// The old setting picked models per role. None of its wording may come back,
// or an agent would pick a model again.
func TestNoSkillPicksAModel(t *testing.T) {
	banned := []string{"`Subagent models` line", "subagent_models", "own model", "never lower", "reviewer model", "sonnet"}
	root := pluginRoot(t)
	files := []string{filepath.Join(root, "README.md")}
	err := filepath.WalkDir(filepath.Join(root, "skills"), func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".md") {
			files = append(files, p)
		}
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range files {
		raw, err := os.ReadFile(f)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range banned {
			// Review names a cheaper tier for its Standards reviewer, so it
			// may say "own model" and give "sonnet" as the example.
			if filepath.Base(filepath.Dir(f)) == "review" && filepath.Base(f) == "SKILL.md" && (b == "own model" || b == "sonnet") {
				continue
			}
			if strings.Contains(string(raw), b) {
				t.Errorf("%s: still says %q", f, b)
			}
		}
	}
}
