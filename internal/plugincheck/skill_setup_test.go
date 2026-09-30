package plugincheck

import "testing"

func TestSkillSetup(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "setup",
		MaxLines: 76,
		Must: []string{
			"acta config set", "--language", "--style", "--tone", "--clear-tone", "--repo-language",
			"acta config show", "adhd", "plain", "full English name",
			"acta doctor", "acta doctor --fix", "--executor", "HERDR_ENV=1", "herdr",
			"Offer `dispatch` only when `HERDR_ENV=1` is in the environment",
			"--subagent-models split", "--subagent-models default", "Claude Code only",
			"<!-- acta:begin -->", "<!-- acta:end -->",
			"only after a yes", "never edits settings",
			"which part to change",
			"`~/.acta/config.yaml`",
			"When only AGENTS.md exists, write the block there",
			"run `/init` first",
			"create a CLAUDE.md that holds only the block",
			"--plan-depth minimal", "--plan-depth full", "acta config set --repo",
			"every repo or this repo only",
			"acta config set --repo --executor inline --plan-depth minimal",
		},
		MustNot: []string{"superpowers:", "It never edits CLAUDE.md",
			"or `herdr` on PATH",
			"~/.acta/voice.yaml", "only to files that already exist", "never create a CLAUDE.md",
		},
	})
}
