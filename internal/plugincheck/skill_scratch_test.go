package plugincheck

import "testing"

func TestSkillScratch(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "scratch",
		MaxLines: 60,
		Must: []string{
			"acta scratch new", "acta scratch add", "verbatim", "images", "Filed SCRATCH-",
			"catet", "nanti", "kepikiran", "File this in Scratchpad?", "never", "memory",
			"new session", "main branch", "status dropped",
		},
		MustNot: []string{"superpowers:"},
	})
}
