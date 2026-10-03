package plugincheck

import (
	"strings"
	"testing"
)

// leanMaxWords is the most words the lean skill may hold, frontmatter and all.
// Every session that loads the skill pays for each word.
const leanMaxWords = 400

// ponytailAllowed are the only plugin files that may name ponytail. NOTICE
// credits it, and the list of overlapping plugins has to name it to flag it.
var ponytailAllowed = map[string]bool{
	"NOTICE":                     true,
	"hooks/workflow-plugins.txt": true,
}

func TestSkillLean(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "lean",
		MaxLines: 60,
		Must: []string{
			// The stance.
			"never less care",
			// Read first.
			"every file", "real flow",
			// The ladder, one phrase for each rung.
			"first one that works wins",
			"needed at all", "repo already have it", "standard library",
			"native platform feature", "installed dependency", "Would a single line do", "minimum code",
			// The rules.
			"one user", "always the same", "may never come", "fewest files", "edge cases",
			// A bug fix goes to the root.
			"root cause", "every caller", "pass through",
			// What is never cut.
			"trust boundar", "data loss", "security", "accessibility", "asked for",
			// What was skipped is said out loud, not left as a marker.
			"in chat or in the plan", "marker",
		},
		// No intensity levels, and no word of the project the text was adapted from.
		MustNot: []string{"ponytail", "ultra", "intensity"},
	})
}

// TestSkillLeanWordCap counts the words of the whole file, so the skill cannot
// grow back toward the long text it was written from.
func TestSkillLeanWordCap(t *testing.T) {
	n := len(strings.Fields(readFile(t, "skills", "lean", "SKILL.md")))
	if n > leanMaxWords {
		t.Errorf("skills/lean/SKILL.md has %d words, cap is %d", n, leanMaxWords)
	}
}

// TestNoPonytail walks every file under plugin/ except node_modules and fails
// on the word in any case. The rules now live in acta:lean, so another copy
// anywhere else is the old text coming back.
func TestNoPonytail(t *testing.T) {
	scanned := 0
	walkPlugin(t, func(rel, text string) {
		scanned++
		if ponytailAllowed[rel] {
			return
		}
		for i, line := range strings.Split(text, "\n") {
			if strings.Contains(strings.ToLower(line), "ponytail") {
				t.Errorf("%s:%d names ponytail; only NOTICE and hooks/workflow-plugins.txt may", rel, i+1)
			}
		}
	})
	if scanned == 0 {
		t.Fatal("the walk found no file under plugin/")
	}
}
