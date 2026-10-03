package plugincheck

import "testing"

func TestSkillDebug(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "debug",
		MaxLines: 722,
		Must: []string{
			"Phase 1", "Read only until the hypothesis is proven", "acta:bug", "acta:shape",
			"regression test", "Three failed fixes", "root-cause-tracing.md",
		},
		MustNot: []string{"superpowers:", "CREATION-LOG", "test-pressure", "test-academic", "git-bug"},
	})
}
