package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillScratch(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "scratch",
		MaxLines: 60,
		Must: []string{
			"acta scratch new", "acta scratch add", "verbatim", "images", "Filed SCR-0001",
			"note this", "idea for later", "any language", "File this in Scratchpad?", "never", "memory",
			"new session", "main branch", "status dropped",
			"--section context", "--section questions", "right after",
			"what work was going on", "what already exists", "file:line",
			"where the facts came from",
		},
		MustNot: []string{"superpowers:", "You may add your own lines", "below theirs", "catet", "nanti", "kepikiran"},
	})
}

// TestSkillScratchNoHandWrittenFences keeps the ban the only place in the
// skill that talks about `---` or "Agent notes". Pasted bodies once grew a
// pile of hand-made separators, so any other mention is that rule coming back.
func TestSkillScratchNoHandWrittenFences(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "scratch", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(string(b), "\n")
	// The frontmatter fences are structure, not skill text.
	start := 0
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			start = i + 1
			break
		}
	}
	ban := false
	for i := start; i < len(lines); i++ {
		line := lines[i]
		if !strings.Contains(line, "---") && !strings.Contains(line, "Agent notes") {
			continue
		}
		if strings.Contains(line, "Never write") {
			ban = true
			continue
		}
		t.Errorf("scratch/SKILL.md:%d mentions --- or Agent notes outside the ban: %q", i+1, line)
	}
	if !ban {
		t.Error("scratch/SKILL.md has no line banning hand-written --- or Agent notes")
	}
}
