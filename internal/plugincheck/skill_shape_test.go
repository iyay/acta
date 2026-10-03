package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillShape(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "shape",
		MaxLines: 395,
		Must: []string{
			"Spike", "Bounded", "Architectural", "HARD-GATE",
			".acta/specs/", "acta:slice", "CONTEXT.md", "docs/adr/",
			"trust boundary", "each gets its own yes", "acta list --json",
			"run " + "`acta id`" + ` right after`, "parent: debt/",
			"acta scratch new", "acta scratch add", "status brainstorming",
			"parent: scratch/", "One Architectural brainstorm per session",
			"claude --bg 'brainstorm SCR-0001'",
			"only when the session text has the " + "`herdr:`" + " line",
			"pbcopy", "wl-copy", "xclip", "OSC 52", "press ←", "Spike and Bounded",
			"closes:",
			"commit it on the branch that is checked out (usually main)",
			"`acta:build` makes the worktree",
			"acta scratch add SCRATCH-n --section log",
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "Visual Companion", "visual-companion", "writing-plans", "elements-of-style",
			"HERDR_ENV",
			"names the DEBT ids it closes",
			"in the worktree", "create its worktree now", "as the first commit"},
	})
}

// TestShapeDescriptionKeepsBrainstorm guards the word users still type.
// "brainstorm SCR-0001" reaches this skill through its description, so a
// rewrite that drops the word leaves the skill without its trigger.
func TestShapeDescriptionKeepsBrainstorm(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "shape", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, desc, _ := frontmatter(b)
	if !strings.Contains(desc, "brainstorm") {
		t.Errorf("shape description has no word brainstorm: %q", desc)
	}
}

// TestShapeEveryScratchAddLogs catches one line that falls back to a
// plain append. The Must list only proves the string exists somewhere, so a
// single answer or section could quietly lose the section flag.
func TestShapeEveryScratchAddLogs(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "shape", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(b), "\n") {
		if !strings.Contains(line, "acta scratch add") {
			continue
		}
		if !strings.Contains(line, "--section log") {
			t.Errorf("shape/SKILL.md:%d appends without --section log: %q", i+1, line)
		}
	}
}

// TestShapeBoundedWritesSpec reads the shape skill on its own.
// A Bounded design once lived only in chat, so the user never saw a spec.
func TestShapeBoundedWritesSpec(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "shape", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{"Write short spec", "short spec (about half a page)"} {
		if !strings.Contains(txt, want) {
			t.Errorf("shape/SKILL.md missing %q", want)
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
			t.Errorf("shape/SKILL.md still says %q", old)
		}
	}
}

// TestShapeBoundedReviewLoopsToShortSpec reads the graph on its own.
// A Bounded "changes requested" once led into the Architectural design doc.
func TestShapeBoundedReviewLoopsToShortSpec(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "shape", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{
		`"Write short spec" -> "User reviews short spec?"`,
		`"User reviews short spec?" -> "Write short spec" [label="changes requested"]`,
		`"User reviews short spec?" -> "Invoke acta:slice skill" [label="approved"]`,
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("shape graph missing %s", want)
		}
	}
	if strings.Contains(txt, `"Write short spec" -> "User reviews spec?"`) {
		t.Error("Bounded still joins the Architectural review node")
	}
}
