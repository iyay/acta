package plugincheck

import "testing"

func TestSkillLand(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "land",
		MaxLines: 300,
		Must: []string{
			"Iron Law", "--no-ff", "Do not show a menu", "git worktree remove", "git branch -d",
			"Never `git push`", "120000", "fixed_in", "The Gate Function",
			"`acta id --fix-duplicates`", "renumber",
			"acta show <plan id>", "done` is less than `total`",
			"acta tick <DEBT-n.m> --all", "debt file",
			"git status --porcelain", "git rev-parse --abbrev-ref HEAD",
			"never check out, stash or reset",
			"in the plan's `closes:`",
			"scripts/eval", "plugin/skills/", "plugin/hooks/", "red eval",
			"acta: tick <plan>", "before the merge, so the ticks reach main",
		},
		MustNot: []string{"superpowers:", "Push and Create PR", "Keep As-Is", "Present Options", "discard the work",
			"each DEBT id the plan names as closed"},
	})
}
