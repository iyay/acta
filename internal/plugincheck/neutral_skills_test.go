package plugincheck

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Skills run in Claude Code, omp and Codex, so their text must not name a
// tool or a model only one harness has. The session note says which subagent
// tool and model to use in the harness at hand.
var claudeOnly = regexp.MustCompile(`(?i)\bagent tool|subagent_type|model:\s*\\?"(sonnet|opus|haiku)`)

func TestSkillsNameNoClaudeOnlyTool(t *testing.T) {
	root := filepath.Join(pluginRoot(t), "skills")
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(p, ".md") {
			return err
		}
		raw, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if m := claudeOnly.FindString(string(raw)); m != "" {
			rel, _ := filepath.Rel(root, p)
			t.Errorf("skills/%s names %q, which only Claude Code has", rel, m)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
