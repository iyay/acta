package plugincheck

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSkillBuild(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "build",
		MaxLines: 1000,
		Must: []string{
			"## Executors", "`subagent`", "`dispatch`", "`inline`",
			`model: "sonnet"`, `agent="task"`, "Do not ask whether to create a worktree",
			"../<repo>-<slug>", "git rev-parse --show-toplevel", "git add -A",
			"acta:tdd", "acta:review", "acta:land", "no per-task reviewer",
			"acta tick plans/<stem>#task-N --step <n>", "acta tick [TASK_ID] --step <n>", "chore(plan): tick wave",
			"acta tick plans/<stem>#task-N --all",
			"acta show <plan id> --json", "progress.done",
			`git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"`, "`$PARENT` is the parent branch recorded above",
			"acta config show", "build_executor",
			"The spec and plan from acta:shape and acta:slice stay on main",
			"do not edit its plan on main",
			"The full suite waits for `acta:land`", "scripts/test",
			"through the executor picked by `/build <executor>`, else the one `acta config show` names, else asks which one",
			"On omp, `dispatch` runs as `subagent`", "Dispatch is only for harnesses other than omp.",
			"Copy in the house rules too",
			"`/build <executor>`", "It is for this run only; never save it.",
			"HERDR_ENV=1", "Build never refuses because herdr is missing.",
			"with `HERDR_ENV=1`, read [dispatch.md](dispatch.md) and follow it",
			"hand it to omp in another tab or pane",
			"[dispatch.md](dispatch.md)", "## Fix rounds", "## Close the tab",
			"A native worktree tool puts the worktree inside the repo",
			"or has `depth: minimal` in its frontmatter",
			"Refuse to start unless the spec is approved",
			// acta dispatch close reads the slug from the current branch, so it
			// only works in the worktree, before land deletes the branch.
			"With `dispatch`, run `acta dispatch close` from the worktree right before `acta:land`",
			"acta wiki check <parent>..HEAD", "[wiki.md](wiki.md)",
		},
		MustNot: []string{"superpowers:", "Would you like me to set up", "task-reviewer-prompt", "re-review-prompt", "## Final Review", "fix round R of 5",
			// The wiki is the one home for project knowledge. The old files stay
			// out of the skill text, and so does the name of the format its
			// fields came from.
			"CONTEXT.md", "docs/adr", "okf", "OKF",
			"default to `.worktrees/`", "Step 0 consent", "ls -d .worktrees", "acta:dispatch",
			"and no `herdr` on PATH",
			`"$LOCATION/$BRANCH_NAME"`, "acta tick <task-id>", "run it again with --all",
			"or `go test ./...`, whichever the project uses",
			"run the full test suite and the type checks, show the output, then use",
			"subagent (default", "`subagent` (default)",
			"EnterWorktree", "Native Worktree Tools", "native worktree tool available", "Git Worktree Fallback", "Step 1a",
			"Refuse to start without an approved spec and an approved plan. Say which one is missing.",
			"after `acta:land` close the tab",
		},
	})
}

// TestBuildTickRuleEverywhere checks implementer-prompt.md on its own, not
// folded into the whole build/ folder text. The implementer only ever reads
// this one file, so the --all rule has to live in it directly.
func TestBuildTickRuleEverywhere(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "acta tick [TASK_ID] --all") {
		t.Error("implementer-prompt.md missing \"acta tick [TASK_ID] --all\"")
	}
}

// TestBuildAgentFlag reads the implementer prompt on its own. The board shows
// who works on a task, and the omp harness sets no agent variable.
func TestBuildAgentFlag(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(b), "--agent omp") {
		t.Error("implementer-prompt.md missing \"--agent omp\"")
	}
}

// TestBuildSkillTickCommands reads SKILL.md on its own. CheckSkill looks at
// the whole build folder, so a revert of SKILL.md alone could stay green.
func TestBuildSkillTickCommands(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, want := range []string{"run `acta tick plans/<stem>#task-N --all`", "acta show <plan id> --json", "progress.done"} {
		if !strings.Contains(txt, want) {
			t.Errorf("build/SKILL.md missing %q", want)
		}
	}
	if strings.Contains(txt, "run it again with --all") {
		t.Error("build/SKILL.md still says \"run it again with --all\"")
	}
}

// TestBuildAgentFlagOnlyWhenTheHarnessSetsNothing guards the prompt text that
// the Claude Code subagent executor reads too. A bare --agent omp on the tick
// overrides AI_AGENT, so Claude's work would be recorded as omp.
func TestBuildAgentFlagOnlyWhenTheHarnessSetsNothing(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	for _, bad := range []string{"--step <n> --agent omp", "--all --agent omp"} {
		if strings.Contains(txt, bad) {
			t.Errorf("implementer-prompt.md passes the flag unconditionally: %q", bad)
		}
	}
	if !strings.Contains(txt, "AI_AGENT") {
		t.Error("implementer-prompt.md never says the flag is for a harness with no AI_AGENT")
	}
}

// TestBuildStartTickPrompt checks implementer-prompt.md on its own. The
// implementer only ever reads this one file, so the --start rule has to live
// in it directly, before the failing-test step.
func TestBuildStartTickPrompt(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	const want = "Very first action: run `acta tick [TASK_ID] --start`"
	if !strings.Contains(txt, want) {
		t.Errorf("implementer-prompt.md missing %q", want)
	}
	if strings.Index(txt, "--start") > strings.Index(txt, "failing test first") {
		t.Error("implementer-prompt.md names --start after the failing-test step; it must come first")
	}
}

// TestBuildStartTickSkill checks SKILL.md on its own. CheckSkill looks at the
// whole build folder, so a revert of SKILL.md alone could stay green.
func TestBuildStartTickSkill(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	txt := string(b)
	const want = "very first action on every task: run `acta tick plans/<stem>#task-N --start`"
	if !strings.Contains(txt, want) {
		t.Errorf("build/SKILL.md missing %q", want)
	}
	if !strings.Contains(txt, "Claude: pass no flag; omp: add `--agent omp`") {
		t.Error("build/SKILL.md never states the --agent rule for --start")
	}
}

// TestImplementerPromptNoFullSuite reads implementer-prompt.md on its own. The
// implementer only reads this file, and the full suite belongs to acta:land.
func TestImplementerPromptNoFullSuite(t *testing.T) {
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "build", "implementer-prompt.md"))
	if err != nil {
		t.Fatal(err)
	}
	// The prompt wraps lines, so join the words before matching a sentence.
	txt := strings.Join(strings.Fields(string(b)), " ")
	if strings.Contains(txt, "run the full suite once before committing") {
		t.Error("implementer-prompt.md still says \"run the full suite once before committing\"")
	}
	if !strings.Contains(txt, "never the full suite") {
		t.Error("implementer-prompt.md missing \"never the full suite\"")
	}
}

// TestBuildExecutorOrder reads SKILL.md on its own. CheckSkill joins the
// folder, and dispatch.md also talks about executors.
func TestBuildExecutorOrder(t *testing.T) {
	txt := readBuildFile(t, "SKILL.md")
	arg := strings.Index(txt, "1. The argument of `/build <executor>`")
	cfg := strings.Index(txt, "2. `build_executor: <name>` from `acta config show`")
	ask := strings.Index(txt, "3. Neither: ask which executor to run")
	if arg < 0 || cfg < 0 || ask < 0 || !(arg < cfg && cfg < ask) {
		t.Errorf("build/SKILL.md must list the executor order argument, config, ask (got %d, %d, %d)", arg, cfg, ask)
	}
	const fallback = "`dispatch` without `HERDR_ENV=1` (this session is not inside a herdr pane) runs as `subagent`. Build never refuses because herdr is missing."
	if !strings.Contains(txt, fallback) {
		t.Errorf("build/SKILL.md missing the herdr fallback %q", fallback)
	}
	if strings.Contains(txt, "no `herdr` on PATH") {
		t.Error("build/SKILL.md still lets herdr on PATH alone pick dispatch; dispatch needs HERDR_ENV=1 and $HERDR_PANE_ID")
	}
}

// TestShowPathLineInSkills reads both skill files on their own. A skill that
// loses the line leaves its agents hunting for a file by hand, and the
// folder-wide Must lists would not notice, so each file is checked here.
func TestShowPathLineInSkills(t *testing.T) {
	for _, skill := range []string{"build", "shape"} {
		b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", skill, "SKILL.md"))
		if err != nil {
			t.Fatal(err)
		}
		if want := "acta show <id|hash> --path"; !strings.Contains(string(b), want) {
			t.Errorf("%s/SKILL.md missing %q", skill, want)
		}
	}
}

// TestBuildSkillNoBareGoTest reads SKILL.md on its own. A bare `go test`
// run uses every core on the machine, which is the bug this repo keeps
// hitting, so the skill has to name it and say the hook stops it.
func TestBuildSkillNoBareGoTest(t *testing.T) {
	txt := readBuildFile(t, "SKILL.md")
	if !strings.Contains(txt, "never run bare `go test`") {
		t.Error("build/SKILL.md missing \"never run bare `go test`\"")
	}
	if !strings.Contains(txt, "the pre-tool hook blocks it") {
		t.Error("build/SKILL.md does not say the pre-tool hook blocks it")
	}
}

// TestBuildCloseWritesWikiFirst reads SKILL.md on its own. CheckSkill joins
// the whole folder, and wiki.md says "acta wiki check" too, so a revert of
// SKILL.md alone could stay green. The wiki step has to come before the fast
// tests and the review, or the review cannot see the wiki diff. The check runs
// first, in the worktree, and wiki.md is read only after it, when a page has
// to be fixed or added.
func TestBuildCloseWritesWikiFirst(t *testing.T) {
	_, closing, ok := strings.Cut(readBuildFile(t, "SKILL.md"), "\n## Close\n")
	if !ok {
		t.Fatal("build/SKILL.md has no ## Close section")
	}
	last := -1
	for _, step := range []string{
		"In the worktree, run `acta wiki check <parent>..HEAD`",
		"Read [wiki.md](wiki.md) only when you will fix or add a page",
		"Fix each touched page and bump its `timestamp`",
		"Write each page the spec approved (shape notes each yes in the spec)",
		"Add any other page only for a lesson a fresh agent would lose time without",
		"Make one commit",
		"run the fast tests and the type checks",
		"then use `acta:review` over `<parent>..HEAD`",
	} {
		i := strings.Index(closing, step)
		switch {
		case i < 0:
			t.Errorf("build/SKILL.md ## Close is missing %q", step)
		case i < last:
			t.Errorf("build/SKILL.md ## Close names %q before an earlier step", step)
		default:
			last = i
		}
	}
	// A dispatch recipient replies back, and the review reads what it sent.
	// So its wiki commit has to come before the reply.
	if !strings.Contains(closing, "`acta reply-back` after the last task commit and the wiki step") {
		t.Error("build/SKILL.md ## Close lets a dispatch recipient reply back before the wiki step")
	}
}

// TestBuildWikiFileHoldsTheRules reads wiki.md on its own. It is the one place
// that says how a page looks and when one is written, so each rule of the
// design has to be in it. The words are joined first, since a line may wrap.
func TestBuildWikiFileHoldsTheRules(t *testing.T) {
	txt := strings.Join(strings.Fields(readBuildFile(t, "wiki.md")), " ")
	for _, r := range []struct{ rule, want string }{
		{"where pages live", "`.acta/wiki/` under the acta root"},
		{"one concept per file", "One concept per file"},
		{"subfolders", "Subfolders are fine"},
		{"page id", "its path minus `.md`"},
		{"no acta id, no status", "no acta id and no status"},
		{"the board skips it", "the board skips the wiki"},
		{"frontmatter type", "type: Gotcha"},
		{"frontmatter title", "title: "},
		{"frontmatter description", "description: "},
		{"frontmatter paths", "paths: ["},
		{"frontmatter timestamp", "timestamp: "},
		{"the five types", "`Decision`"},
		{"the five types", "`Gotcha`"},
		{"the five types", "`Runbook`"},
		{"the five types", "`Reference`"},
		{"the five types", "`Glossary`"},
		{"one glossary page", "`glossary.md`"},
		{"description size", "one line, 120 characters at most"},
		{"description is the hint", "the hint text"},
		{"folder paths", "A folder ends with `/`"},
		{"prefix, no globs", "by prefix, with no globs"},
		{"timestamp format", "ISO 8601"},
		{"timestamp bumps", "on every change and on every re-check"},
		{"body size", "250 words at most"},
		{"split long pages", "Split a longer page"},
		{"no index or log", "no `index.md` and no `log.md`"},
		{"feature off with no wiki", "pays 0 tokens"},
		{"when pages are written", "last step of `acta:build`, before review"},
		{"who writes them", "The agent that ran the plan"},
		{"the check", "`acta wiki check <parent>..HEAD`"},
		{"fix lines that are now wrong", "fix every line that is now wrong"},
		{"bump the timestamp", "Bump `timestamp`"},
		{"the pages the spec approved", "Write each page the spec approved"},
		{"shape notes each yes", "Shape notes each yes in the spec"},
		{"when to add a page", "a lesson a fresh agent would lose time without"},
		{"one commit", "one commit in the worktree"},
		{"no wiki task in a plan", "A plan has no wiki task"},
		{"once per plan", "at most once per plan"},
		{"update changes lines", "An update changes lines. It does not rewrite the page."},
		{"nothing git already holds", "no LANDED status and no review NOTEs"},
	} {
		if !strings.Contains(txt, r.want) {
			t.Errorf("build/wiki.md is missing the rule %q: %q", r.rule, r.want)
		}
	}
}

// TestImplementerPromptNoReportFile reads implementer-prompt.md on its own. The
// implementer only reads this file, so no report-file step may live in it. A
// subagent write to a report file gets refused, and the whole report then
// floods the orchestrator reply instead.
func TestImplementerPromptNoReportFile(t *testing.T) {
	txt := strings.Join(strings.Fields(readBuildFile(t, "implementer-prompt.md")), " ")
	for _, bad := range []string{
		"[REPORT_FILE]",
		"Write your full report to",
		"report file path",
		"detail lives in the report file",
	} {
		if strings.Contains(txt, bad) {
			t.Errorf("implementer-prompt.md still names a report file: %q", bad)
		}
	}
	for _, want := range []string{
		"at most 25 lines",
		"RED",
		"GREEN",
		"Files changed",
		"Self-review",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("implementer-prompt.md is missing the capped-reply ask %q", want)
		}
	}
}
