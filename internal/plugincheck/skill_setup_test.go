package plugincheck

import "testing"

func TestSkillSetup(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "setup",
		MaxLines: 36,
		Must: []string{
			"AskUserQuestion", "## Every path", "4 questions", "4 options", "(default)", "Other",
			"acta setup", "! acta setup", "--plugin-dir",
			"acta config set", "acta doctor", "TTY",
			"mandatory", "no yes asked", "write to both", "Never create an AGENTS.md",
			"--unset", "overrides",
		},
		MustNot: []string{"superpowers:",
			"<!-- acta:begin -->", "<!-- acta:end -->",
			"HERDR_ENV", "--executor", "--subagent-models", "--plan-depth", "--questions",
			"--language", "--style", "--tone",
			"repo language", "subagent models",
		},
	})
}
