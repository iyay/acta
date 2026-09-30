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
			"Nothing is set (no config file, or every value marked `(default)`)",
			"show the current setting first",
			"When the user already said which part to change, change only that part and stop there",
			"ask only the parts not set yet",
			"Change anything already set? (voice, executor, plan depth, subagent models, acta block)",
			"`~/.acta/config.yaml`",
			"When only AGENTS.md exists, write the block there",
			"run `/init` first",
			"create a CLAUDE.md that holds only the block",
			"--plan-depth minimal", "--plan-depth full", "acta config set --repo",
			"every repo or this repo only",
			"acta config set --repo --executor inline --plan-depth minimal",
		},
		MustNot: []string{"superpowers:", "It never edits CLAUDE.md",
			"only for what `acta config show` says is not set yet",
			"or `herdr` on PATH",
			"~/.acta/voice.yaml", "only to files that already exist", "never create a CLAUDE.md",
		},
	})
}
