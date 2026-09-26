package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillDispatch(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "dispatch",
		MaxLines: 620,
		Must: []string{
			"HERDR_ENV", "/goal", "REPLY-BACK", "references/house-rules.md", "pm:review", "pm:land", "pm:build",
			"PROPERTY", "ultrathink orchestrate", "Never wait", ".pm/plans/", "herdr-delivery.md", ".pm/bugs",
			"pm:bug", "already on the parent branch", "exactly one read", "checkpoint unconfirmed",
		},
		MustNot: []string{"superpowers:", "git-bug", "docs/superpowers", "/Users/", "herdr-pane-moves", "bugs.md", "Core Six",
			"Important/Minor", "per-task reviewer", "fix round R of 5", "WORKTREE LANDING", "--Users-",
			"until herdr agent read", "read it again"},
	})
}

// TestDispatchCheckpointOneRead reads SKILL.md and herdr-delivery.md one at
// a time. CheckSkill joins the folder, so reverting one file stayed green.
func TestDispatchCheckpointOneRead(t *testing.T) {
	for file, wants := range map[string][]string{
		"SKILL.md":          {"exactly one read", "checkpoint unconfirmed", "gets its own one read", "outside this rule"},
		"herdr-delivery.md": {"exactly one read", "checkpoint unconfirmed"},
	} {
		b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", file))
		if err != nil {
			t.Fatal(err)
		}
		for _, want := range wants {
			if !strings.Contains(string(b), want) {
				t.Errorf("dispatch/%s missing %q", file, want)
			}
		}
	}
	b, _ := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", "SKILL.md"))
	if n := strings.Count(string(b), "checkpoint unconfirmed"); n < 2 {
		t.Errorf("dispatch/SKILL.md says \"checkpoint unconfirmed\" %d times; the loop item and the section both need it", n)
	}
}
