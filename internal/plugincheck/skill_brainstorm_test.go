package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillBrainstorm(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "brainstorm",
		MaxLines: 395,
		Must: []string{
			"Spike", "Bounded", "Architectural", "HARD-GATE",
			".acta/specs/", "acta:plan", "CONTEXT.md", "docs/adr/",
			"trust boundary", "each gets its own yes", "acta list --json",
			"run " + "`acta id`" + ` right after`, "parent: debt/",
			"acta scratch new", "acta scratch add", "status brainstorming",
			"parent: scratch/", "One Architectural brainstorm per session",
			"claude --bg 'brainstorm SCRATCH-", "HERDR_ENV=1",
			"pbcopy", "wl-copy", "xclip", "OSC 52", "press ←", "Spike and Bounded",
			"closes:",
			"commit it on the branch that is checked out (usually main)",
			"`acta:build` makes the worktree",
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "Visual Companion", "visual-companion", "writing-plans", "elements-of-style",
			"names the DEBT ids it closes",
			"in the worktree", "create its worktree now", "as the first commit"},
	})
}

// TestBrainstormBoundedWritesSpec reads the brainstorm skill on its own.
// A Bounded design once lived only in chat, so the user never saw a spec.
func TestBrainstormBoundedWritesSpec(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "brainstorm", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{"Write short spec", "short spec (about half a page)"} {
		if !strings.Contains(txt, want) {
			t.Errorf("brainstorm/SKILL.md missing %q", want)
		}
	}
	for _, old := range []string{
		"No spec file, no implementation plan document",
		"no plan document",
		"no plan doc",
		"Implement via normal workflow",
		"implementation proceeds directly",
		"short in-chat design",
	} {
		if strings.Contains(txt, old) {
			t.Errorf("brainstorm/SKILL.md still says %q", old)
		}
	}
}

// TestBrainstormBoundedReviewLoopsToShortSpec reads the graph on its own.
// A Bounded "changes requested" once led into the Architectural design doc.
func TestBrainstormBoundedReviewLoopsToShortSpec(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "brainstorm", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{
		`"Write short spec" -> "User reviews short spec?"`,
		`"User reviews short spec?" -> "Write short spec" [label="changes requested"]`,
		`"User reviews short spec?" -> "Invoke acta:plan skill" [label="approved"]`,
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("brainstorm graph missing %s", want)
		}
	}
	if strings.Contains(txt, `"Write short spec" -> "User reviews spec?"`) {
		t.Error("Bounded still joins the Architectural review node")
	}
}
