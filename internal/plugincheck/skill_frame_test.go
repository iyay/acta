package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillFrame(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "frame",
		MaxLines: 197,
		Must: []string{
			"what's your goal with this?", "startup.md", "premises", "assignment",
			"acta scratch add", "--section context", "--section log", "--section questions",
			"acta:shape",
			"one at a time", "Just do it", "optional", "never forced",
			"continue with shape SCR-xxxx now?",
		},
		MustNot: []string{"status brainstorming", "office hours", "user-invocable: false", "superpowers:"},
	})
}

// TestFrameDescriptionNamesTriggers guards the words a user types to reach
// this skill. The description is all the model sees before it opens the
// folder, so a trigger that goes missing means the skill never runs.
func TestFrameDescriptionNamesTriggers(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "frame", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	_, desc, _ := frontmatter(b)
	for _, want := range []string{
		"frame", "frame this idea", "is this worth building", "should I build this",
		"new product idea", "validate an idea", "who is this for",
	} {
		if !strings.Contains(desc, want) {
			t.Errorf("frame description missing %q: %q", want, desc)
		}
	}
}

// TestFrameStartupNamesForcingQuestions reads startup.md on its own. The six
// questions are the whole reason that file exists, so each one has to be named
// there rather than in the skill.
func TestFrameStartupNamesForcingQuestions(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "frame", "startup.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{
		"Demand reality", "Status quo", "Desperate specificity",
		"Narrowest wedge", "Observation", "Future-fit",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("frame/startup.md missing the question %q", want)
		}
	}
}

// TestFrameNeverSetsBrainstormingStatus keeps the scratch item out of the
// session's one brainstorm slot. The hook counts the `brainstorming` status as
// that slot, so frame setting it would block a later real brainstorm. The Must
// list only proves some words exist, so this reads the skill on its own.
func TestFrameNeverSetsBrainstormingStatus(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "frame", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	for i, line := range strings.Split(string(b), "\n") {
		if !strings.Contains(line, "brainstorming") {
			continue
		}
		low := strings.ToLower(line)
		if strings.Contains(low, "never") || strings.Contains(low, "no ") {
			continue
		}
		t.Errorf("frame/SKILL.md:%d sets the brainstorming status: %q", i+1, line)
	}
}
