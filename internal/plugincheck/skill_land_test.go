package plugincheck

import (
	"os"
	"path/filepath"
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
			"`commit_history`", "`tidy`", "tidy.md",
			"Thinking \"just this once\"", "Tired and wanting work over",
			"| \"I'm tired\" | Tired is not an excuse |",
			"scripts/eval", "plugin/skills/", "plugin/hooks/", "red eval",
			"chore(plan): tick <plan>", `acta commit <plan path> -m "chore(plan): tick <plan>"`, "before the merge, so the ticks reach main",
			"git status --porcelain -- <plan path>",
			"acta run-one -- <full command>", "HEAD^{tree}", "tree same as branch, gates reused",
			"acta wiki check <base>..<head>", "the wiki pages the branch added or changed",
			"git diff --name-only <merge-sha>^1 <merge-sha> -- .acta/wiki/",
		},
		MustNot: []string{"superpowers:", "Push and Create PR", "Keep As-Is", "Present Options", "discard the work",
			"Build never commits the plan file", "First commit the plan file",
			"each DEBT id the plan names as closed",
			// Once the branch is merged, main already holds all of it, so this
			// range is empty and the report would list no page.
			"<base>..<head> -- .acta/wiki/",
			// The wiki is the one home for project knowledge. The old files stay
			// out of the skill text, and so does the name of the format its
			// fields came from.
			"CONTEXT.md", "docs/adr", "okf", "OKF",
			// The branch tree no longer stands for the tidy tip once the parent moves.
			"the trees match"},
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
	// the wiki sentence itself. The check reads the pages of the checkout it runs
	// in, so it runs in the worktree, where the pages of the branch are, and
	// before the merge, since it is a gate.
	const gate = "Wiki gate: from the worktree, before the merge, run `acta wiki check <base>..<head>` with the output shown; a problem stops the land the same as a red test"
	if !strings.Contains(step1, gate) {
		t.Errorf("land step 1 does not stop the merge on acta wiki check from the worktree, want %q in %q", gate, step1)
	}
	// The branch is gone and main may have moved by the time the report is
	// written, so the pages come from the merge commit itself: its first parent
	// is main as it was, and the merge is main with the branch in it.
	const pages = "the wiki pages the branch added or changed (`git diff --name-only <merge-sha>^1 <merge-sha> -- .acta/wiki/`, or none)"
	if !strings.Contains(report, pages) {
		t.Errorf("the landing report does not list the wiki pages from the merge commit, want %q in %q", pages, report)
	}
}

// TestLandTidyPath reads tidy.md on its own. SKILL.md only points to it, so
// the tidy steps and their rules have to live there, and SKILL.md has to send
// the agent to the file when commit_history is tidy.
func TestLandTidyPath(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "land", "tidy.md"))
	if err != nil {
		t.Fatalf("land/tidy.md: %v", err)
	}
	tidy := string(b)
	first := strings.SplitN(tidy, "\n", 2)[0]
	if !strings.HasPrefix(first, "Read this") {
		t.Errorf("land/tidy.md must start with one line saying when to read it, got %q", first)
	}
	for _, want := range []string{
		"acta tidy", "refs/acta/last-land", "--ff-only", "reset --keep",
		"never fall back to a plain merge", "git branch -D",
		"Pick the fold point (step 2) first", "folded <n>", "must equal the `parent` sha",
		"`folded` is 0", "`folded` is above 0",
		"git diff --name-only <branch> refs/acta/tidy/<branch>", "detached temp worktree",
		"tidy proved the tip equals the merge of parent and branch",
		"only planning files differ, gates reused", "fixed_in", "git update-ref -d refs/acta/tidy/<branch>",
		"the new parent tip, and the `tidy:` line",
	} {
		if !strings.Contains(tidy, want) {
			t.Errorf("land/tidy.md missing %q", want)
		}
	}
	skill := readSkill(t, "skills/land/SKILL.md")
	const pointer = "`tidy` (default): follow [tidy.md](tidy.md)"
	if !strings.Contains(skill, pointer) {
		t.Errorf("land SKILL.md does not send tidy mode to tidy.md, want %q", pointer)
	}
	for _, gone := range []string{"acta tidy", "--ff-only", "reset --keep", "git branch -D"} {
		if strings.Contains(skill, gone) {
			t.Errorf("land SKILL.md still holds tidy text %q; it belongs in tidy.md", gone)
		}
	}
}
