package plugincheck

import (
	"os"
	"regexp"
	"strings"
	"testing"
)

func TestSkillSlice(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "slice",
		MaxLines: 262,
		Must: []string{
			".acta/plans/", "acta:build", "verify:", "## Waves", "Follow acta:lean",
			"the lean line, unless coding_guide is off",
			"Unless `acta config show` says `coding_guide: off`, every plan's Global Constraints carry this line",
			"never as the one case", "wait for a yes", "`subagent`", "`dispatch`", "`inline`",
			"No Placeholders", "Global Constraints", "**Spec:** none (Bounded, approved in chat on <date>)",
			"run `acta id` right after", "parent: debt/",
			"never holds a step that can only happen after landing",
			"closes:", "parent: debt/<stem>",
			"Commit the plan on main", `acta commit <path> -m "<message>"`,
			"only after the plan is approved",
			"A fix round, or any change to a plan whose build is running, goes in that build's worktree",
			"## Plan depth", "`/slice minimal`", "depth: minimal", "`plan_depth`",
			"apply to `full` plans only",
			"invoke `acta:build` in the same turn",
			"`/slice full`", "do not ask when one answers", "then `full`.",
			"no Architecture, Tech Stack, File map or Interfaces", "No code blocks.",
			"run `acta config show`", "When it prints `build_executor: <name>`, that executor is chosen",
			"do not ask",
			"## Test commands", "Never `./...` in a task", "**Tests:**",
		},
		MustNot: []string{"superpowers:", "docs/superpowers", "executing-plans", "Two execution options", "plan-document-reviewer",
			"names the DEBT ids it closes",
			"If working in an isolated worktree",
			"Ask which executor only if the user has not said",
			"Every plan MUST start", "Announce at start", "questionable taste", "Frequent commits",
		},
	})
}

// gitCommitSpan matches a `git commit ...` command written in backticks, so a
// sentence that says "never `git commit`" is not a command.
var gitCommitSpan = regexp.MustCompile("`git commit [^`]*`")

// planningWord matches a planning file or a planning commit subject inside it.
var planningWord = regexp.MustCompile(`(?i)\b(plan|spec)\b|chore\(`)

// TestNoSkillCommitsPlanningFilesWithGit walks every skill file. A spec, a
// plan or a tick commit goes through acta commit, so acta can fold it into the
// planning commit before it. Code commits keep git commit.
func TestNoSkillCommitsPlanningFilesWithGit(t *testing.T) {
	for _, p := range skillFiles(t, "skills") {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		for i, line := range strings.Split(string(b), "\n") {
			for _, span := range gitCommitSpan.FindAllString(line, -1) {
				if planningWord.MatchString(span) {
					t.Errorf("%s:%d commits a planning file with git commit: %s", p, i+1, span)
				}
			}
		}
	}
}
