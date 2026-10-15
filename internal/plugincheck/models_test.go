package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modelsPara = "**Subagent models.** Follow the `Subagent models` line in the acta session note. With no such line, name no model and follow the user's own config."

func TestModelsParagraphInFiveSkills(t *testing.T) {
	for _, s := range []string{"slice", "shape", "debug", "build", "review"} {
		raw, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", s, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(string(raw), modelsPara); n != 1 {
			t.Errorf("skills/%s: models paragraph found %d times, want 1", s, n)
		}
	}
}
