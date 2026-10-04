package plugincheck

import (
	"strings"
	"testing"
)

func TestSkillLand(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "land",
		MaxLines: 300,
		Must: []string{
			"Iron Law", "--no-ff", "Do not show a menu", "git worktree remove", "git branch -d",
			"Never `git push`", "120000", "fixed_in", "The Gate Function",
			"`acta id --fix-duplicates`", "renumber",
			"acta show <plan id>", "done` is less than `total`",
			"acta tick <DBT-nnnn.nn> --all", "debt file",
			"git status --porcelain", "git rev-parse --abbrev-ref HEAD",
			"never check out, stash or reset",
			"in the plan's `closes:`",
			"scripts/eval", "plugin/skills/", "plugin/hooks/", "red eval",
			"acta: tick <plan>", "before the merge, so the ticks reach main",
			"git status --porcelain -- <plan path>",
			"acta run-one -- <full command>", "HEAD^{tree}", "tree same as branch, gates reused",
			"acta wiki check <base>..<head>", "the wiki pages the branch added or changed",
		},
		MustNot: []string{"superpowers:", "Push and Create PR", "Keep As-Is", "Present Options", "discard the work",
			"Build never commits the plan file", "First commit the plan file",
			"each DEBT id the plan names as closed",
			// The wiki is the one home for project knowledge. The old files stay
			// out of the skill text, and so does the name of the format its
			// fields came from.
			"CONTEXT.md", "docs/adr", "okf", "OKF"},
	})
}

// TestLandWikiGate checks where the two wiki lines sit. The Must list only
// proves the words exist somewhere. The check has to be in step 1, where a
// problem stops the merge, and the pages have to be in the landing report.
func TestLandWikiGate(t *testing.T) {
	var step1, report string
	for _, line := range strings.Split(readSkill(t, "skills/land/SKILL.md"), "\n") {
		switch {
		case strings.HasPrefix(line, "1. Preconditions"):
			step1 = line
		case strings.HasPrefix(line, "Report after landing:"):
			report = line
		}
	}
	// Step 1 already says a red eval stops the land, so the stop has to sit in
	// the wiki sentence itself.
	const gate = "run `acta wiki check <base>..<head>` with the output shown; a problem stops the land the same as a red test"
	if !strings.Contains(step1, gate) {
		t.Errorf("land step 1 does not stop the merge on acta wiki check, want %q in %q", gate, step1)
	}
	if !strings.Contains(report, "the wiki pages the branch added or changed") {
		t.Errorf("the landing report does not list the wiki pages: %q", report)
	}
}
