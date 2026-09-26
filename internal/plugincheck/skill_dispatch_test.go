package plugincheck

import "testing"

func TestSkillDispatch(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "dispatch",
		MaxLines: 620,
		Must: []string{
			"HERDR_ENV", "/goal", "REPLY-BACK", "references/house-rules.md", "pm:review", "pm:land", "pm:build",
			"PROPERTY", "ultrathink orchestrate", "Never wait", ".pm/plans/", "herdr-delivery.md", ".pm/bugs",
		},
		MustNot: []string{"superpowers:", "git-bug", "docs/superpowers", "/Users/", "herdr-pane-moves", "bugs.md", "Core Six"},
	})
}
