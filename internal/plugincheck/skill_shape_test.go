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
		MaxLines: 177,
		Must: []string{
			"Spike", "Bounded", "Architectural", "HARD-GATE",
			".acta/specs/", "acta:slice",
			"glossary.md", "acta wiki ls --type Decision", "../build/wiki.md",
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
			"in the worktree", "create its worktree now", "as the first commit",
			// The wiki is the one home for project knowledge. The old files stay
			// out of the skill text, and so does the name of the format its
			// fields came from.
			"CONTEXT.md", "docs/adr", "okf", "OKF"},
	})
}

// TestNoSkillNamesTheOldKnowledgeFiles walks every skill file. Three skill
// tests refuse these names for their own skill. This one covers all of them,
// so no skill sends an agent to a place the wiki replaced.
func TestNoSkillNamesTheOldKnowledgeFiles(t *testing.T) {
	for _, p := range skillFiles(t, "skills") {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		rel, _ := filepath.Rel(pluginRoot(t), p)
		for _, bad := range []string{"CONTEXT.md", "docs/adr", "okf", "OKF"} {
			if strings.Contains(string(b), bad) {
				t.Errorf("%s names %q", rel, bad)
			}
		}
	}
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

// TestShapeBoundedReviewLoopsToShortSpec reads the Bounded checklist on its
// own. A Bounded "changes requested" once led into the Architectural design
// doc. The dot graph is gone, so the checklist is where the edge lives.
func TestShapeBoundedReviewLoopsToShortSpec(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "shape", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	if strings.Contains(txt, "digraph") {
		t.Error("shape text carries a dot graph; the checklists are the flow")
	}
	bounded := shapeChecklist(t, txt, "**Bounded:**", "**Architectural:**")
	if !strings.Contains(bounded, "changes requested") {
		t.Error("the Bounded checklist has no changes-requested step")
	}
	for _, line := range strings.Split(bounded, "\n") {
		if strings.Contains(line, "changes requested") && !strings.Contains(line, "short spec") {
			t.Errorf("Bounded changes requested leaves the short spec: %q", line)
		}
	}
	for _, old := range []string{"Write design doc", "spec review loop", "re-run the spec"} {
		if strings.Contains(bounded, old) {
			t.Errorf("Bounded review still points at the Architectural %q", old)
		}
	}
}

// shapeChecklist returns the slice of the shape skill between two markers. The
// checklist per path is the only flow text, so a rule has to live in the slice
// of its own path to count.
func shapeChecklist(t *testing.T, txt, from, to string) string {
	t.Helper()
	i := strings.Index(txt, from)
	if i < 0 {
		t.Fatalf("shape skill has no %s block", from)
	}
	rest := txt[i:]
	j := strings.Index(rest, to)
	if j < 0 {
		t.Fatalf("shape skill has no %s block after %s", to, from)
	}
	return rest[:j]
}

// TestShapeProbeModeNamesProbeFile guards the two hooks that open probe mode:
// the user says "probe" or "grill me", or the session note says
// `Questions: probe`. The file itself has to exist and carry the round shape.
func TestShapeProbeModeNamesProbeFile(t *testing.T) {
	skill, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "shape", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"probe.md", "grill me", "Questions: probe"} {
		if !strings.Contains(string(skill), want) {
			t.Errorf("shape/SKILL.md missing %q", want)
		}
	}
	probe, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "shape", "probe.md"))
	if err != nil {
		t.Fatalf("shape/probe.md: %v", err)
	}
	for _, want := range []string{"five", "Recommended:", "--section log", "**Q1."} {
		if !strings.Contains(string(probe), want) {
			t.Errorf("shape/probe.md missing %q", want)
		}
	}
}
