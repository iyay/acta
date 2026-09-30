package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const modelsPara = "**Subagent models.** Only when `acta config show` lists `subagent_models: split` and you run in Claude Code: subagents that write code use `model: \"sonnet\"`; all other subagents (mapping, explore, planning help, debug investigation, spikes) use `model: \"opus\"`; reviewers use your own model alias. Otherwise name no model and follow the user's own config."

func TestModelsParagraphInFiveSkills(t *testing.T) {
	for _, s := range []string{"plan", "brainstorm", "debug", "build", "review"} {
		raw, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", s, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if n := strings.Count(string(raw), modelsPara); n != 1 {
			t.Errorf("skills/%s: models paragraph found %d times, want 1", s, n)
		}
	}
}
