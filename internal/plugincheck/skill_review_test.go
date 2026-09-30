package plugincheck

import "testing"

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
		},
		MustNot: []string{"superpowers:", "Critical", "Important (Should Fix)", "Minor", "GitHub Thread Replies", "A change you judge small", "kept in memory", "one line in memory"},
	})
}
