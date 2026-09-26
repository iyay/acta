package plugincheck

import "testing"

func TestSkillReview(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "review",
		MaxLines: 480,
		Must: []string{
			"Spec axis", "Standards axis", "BLOCKER", "NOTE", "three questions",
			"deep lens", "never a round 4", "code-reviewer.md", "CLEAN or BLOCKED",
			"## Receiving findings", "Fix round", "pm:land",
			"## Where findings go", "pm:bug", "already on the parent branch",
		},
		MustNot: []string{"superpowers:", "Critical", "Important (Should Fix)", "Minor", "GitHub Thread Replies"},
	})
}
