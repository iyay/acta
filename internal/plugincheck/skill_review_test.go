package plugincheck

import (
	"strings"
	"testing"
)

func TestSkillReview(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "review",
		MaxLines: 482,
		Must: []string{
			"Spec axis", "Standards axis", "BLOCKER", "NOTE", "three questions",
			"deep lens", "never a round 4", "code-reviewer.md", "CLEAN or BLOCKED",
			"## Receiving findings", "Fix round", "acta:land",
			"## Where findings go", "acta:bug", "already on the parent branch",
			"Small means one file, and only text or config with no code logic.",
			"acta debt new",
			"A NOTE may start with `(high) `, `(medium) ` or `(low) `",
			"never the full suite",
			"[fix]", "[debt]", "[note]",
			"## Review notes",
			"one polish commit",
			"moves those items to `[debt]`",
			"full test suite",
		},
		MustNot: []string{"superpowers:", "Critical", "Important (Should Fix)", "Minor", "GitHub Thread Replies", "A change you judge small", "kept in memory", "one line in memory", "never a task"},
	})
}

// TestCodeReviewerTemplateBucketTags pins the reviewer template: every NOTE
// it prints carries a bucket tag, so the orchestrator can sort without asking.
func TestCodeReviewerTemplateBucketTags(t *testing.T) {
	txt := readSkill(t, "skills/review/code-reviewer.md")
	for _, want := range []string{"bucket tag", "[fix]", "[debt]", "[note]"} {
		if !strings.Contains(txt, want) {
			t.Errorf("code-reviewer.md missing %q", want)
		}
	}
}
