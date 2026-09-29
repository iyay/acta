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
		},
		MustNot: []string{"superpowers:", "Push and Create PR", "Keep As-Is", "Present Options", "discard the work"},
	})
}
