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

// TestDispatchNewThenGoal reads each dispatch file on its own. /new and
// /goal once landed as one message, and the goal never set.
func TestDispatchNewThenGoal(t *testing.T) {
	wants := []string{"HARD RULE", "New session started", "🎯 Goal", "never in one prompt", "fix round: `/goal` only"}
	for _, file := range []string{"SKILL.md", "herdr-delivery.md"} {
		b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", file))
		if err != nil {
			t.Fatal(err)
		}
		txt := string(b)
		for _, want := range wants {
			if !strings.Contains(txt, want) {
				t.Errorf("dispatch/%s missing %q", file, want)
			}
		}
		if strings.Contains(txt, "Back-to-back, no settle-wait") {
			t.Errorf("dispatch/%s still says \"Back-to-back, no settle-wait\"", file)
		}
	}
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", "herdr-delivery.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The fix-round command was once deleted by accident. It is the only one
	// with the fix range, so an agent on a fix round needs it.
	if !strings.Contains(string(b), "/pm:review <fixed-from>..<new-head>") {
		t.Error("dispatch/herdr-delivery.md lost the fix-round command with <fixed-from>..<new-head>")
	}
}

// TestDispatchAgentFlag reads herdr-delivery.md on its own. The recipient is
// an omp pane and the harness sets no agent variable, so the brief has to
// name it.
func TestDispatchAgentFlag(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "dispatch", "herdr-delivery.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "--agent omp") {
		t.Error("herdr-delivery.md missing \"--agent omp\"")
	}
}
