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
			"each debt line starts with `(medium) ` or `(high) `",
			"Iron Law", "--no-ff", "Do not show a menu", "git worktree remove", "git branch -d",
			"Never `git push` unless the user asks", "120000", "fixed_in", "The Gate Function",
			"`acta id --fix-duplicates`", "renumber",
			"acta show <plan id>", "done` is less than `total`",
			"[baseline.md](baseline.md)", "`acta-baseline.md`", "absolute Git common directory",
			"acta tick <DBT-nnnn.nn> --all", "debt file",
			"git status --porcelain", "git rev-parse --abbrev-ref HEAD",
			"never check out, stash or reset",
			"in the plan's `closes:`",
			"`commit_history`", "`tidy`", "tidy.md",
			"Thinking \"just this once\"", "Tired and wanting work over",
			"| \"I'm tired\" | Tired is not an excuse |",
			"chore(plan): tick <plan>", `acta commit <plan path> -m "chore(plan): tick <plan>"`, "before the merge, so the ticks reach main",
			"git status --porcelain -- <plan path>",
			"acta run-one -- <full command>", "HEAD^{tree}", "tree same as branch, gates reused",
			// The full suite runs here and nowhere else in a plan, so a red
			// run must not turn into one full run per guess.
			"Run it once every change is in, polish included.",
			"Red: fix the cause through the narrow failing tests, then run it once more; never debug with the full suite.",
			"It reruns only then, or when step 6 or tidy needs a new gate.",
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
	// Step 1 already says a red gate case stops the land, so the stop has to sit in
	// the wiki sentence itself. The check reads the pages of the checkout it runs
	// in, so it runs in the worktree, where the pages of the branch are, and
	// before the merge, since it is a gate.
	const gate = "Wiki gate: from the worktree, before the merge, run `acta wiki check <base>..<head>` with the output shown; errors stop the land; stale warnings alone do not"
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

// Temp paths belong to a Git registry, not whichever repository the shell finds.
func TestLandGateWorktreeOwner(t *testing.T) {
	tidy := readFile(t, "skills/land/tidy.md")
	for _, want := range []string{
		`gate_owner="$(git rev-parse --show-toplevel)"`,
		`git -C "$gate_owner" worktree add --detach "$gate_path"`,
		`git -C "$gate_owner" worktree list --porcelain`,
		`git -C "$gate_owner" worktree remove "$gate_path"`,
		"distinct absolute path",
		"before moving the parent",
		"wrong repository is not a file-removal refusal",
	} {
		if !strings.Contains(tidy, want) {
			t.Errorf("temporary gate operations missing owner rule %q", want)
		}
	}
}

// TestLandBaselineEvidence reads baseline.md on its own. SKILL.md only points
// to it, so the shared record rules have to live there. The policy sits before
// the numbered steps, so an ordinary or tidy land cannot run the suite without
// it, and the record can never claim a run that did not pass.
func TestLandBaselineEvidence(t *testing.T) {
	skill := readSkill(t, "skills/land/SKILL.md")
	for _, want := range []string{
		"[baseline.md](baseline.md)",
		"`acta-baseline.md`",
		"absolute Git common directory",
		"shared by linked worktrees",
		"Only a complete passing run replaces that record",
		"A failed run, a reused gate or a write error keeps the last PASS",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("land/SKILL.md missing %q", want)
		}
	}
	policy := strings.Index(skill, "[baseline.md](baseline.md)")
	steps := strings.Index(skill, "1. Preconditions")
	mode := strings.Index(skill, "Mode: `commit_history`")
	if policy < 0 || steps < 0 || policy > steps || mode < 0 || policy > mode {
		t.Errorf("the baseline policy must sit before step 1 and the tidy mode switch (policy %d, step 1 %d, mode %d)", policy, steps, mode)
	}
	b, err := os.ReadFile(filepath.Join(pluginRoot(t), "skills", "land", "baseline.md"))
	if err != nil {
		t.Fatalf("land/baseline.md: %v", err)
	}
	base := string(b)
	for _, want := range []string{
		"git rev-parse --path-format=absolute --git-common-dir",
		"before every full suite run in a land: step 1, a step-6 rerun on the merge result, or a tidy-tip rerun",
		"in the checkout where the suite ran",
		"unique log",
		"acta-baseline-<utc-time>-<pid>.log",
		"Never pipe the suite into `tee`, `head` or a filter",
		"rc=$?",
		"already takes the run-one lock",
		"Keep the failed log",
		"the last PASS stays",
		"Never rerun the suite",
		"temp file",
		"rename it over `acta-baseline.md`",
		"commit:", "tree:", "command:", "when:", "runtime:", "exit:", "log:", "summary:",
		"writes no record",
	} {
		if !strings.Contains(base, want) {
			t.Errorf("land/baseline.md missing %q", want)
		}
	}
	for _, banned := range []string{"| tee", "| head"} {
		if strings.Contains(base, banned) {
			t.Errorf("land/baseline.md masks the exit status with %q", banned)
		}
	}
}

// TestLandRecordsFlakes reads the land files on their own. A test that only
// passed on a rerun is green only once a bug records it, and that bug has to
// reach the parent. A tidy-tip rerun runs in a detached temp worktree, where
// a record cannot commit, and a step-6 rerun runs after the merge, so each
// run names its own checkout. The tested commit is read before a record
// commits, so the baseline names what was tested.
func TestLandRecordsFlakes(t *testing.T) {
	var step1, report string
	skill := readSkill(t, "skills/land/SKILL.md")
	for _, line := range strings.Split(skill, "\n") {
		switch {
		case strings.HasPrefix(line, "1. Preconditions"):
			step1 = line
		case strings.HasPrefix(line, "Report after landing:"):
			report = line
		}
	}
	for _, want := range []string{
		"Exit 0 with `FLAKY:` lines: before the merge, record each one as [baseline.md](baseline.md) step 4 says.",
		// A whole eval set costs quota, and a case red every time is a real red.
		"never rerun the whole set for one case",
		"Red on every rerun stops the land.",
		// An eval flake record has to commit where the bug can reach the parent.
		"in the checkout [baseline.md](baseline.md) step 4 names",
	} {
		if !strings.Contains(step1, want) {
			t.Errorf("land step 1 missing %q", want)
		}
	}
	if !strings.Contains(report, "each flake with its bug") {
		t.Error("the landing report does not name each flake with its bug")
	}
	base := readSkill(t, "skills/land/baseline.md")
	for _, want := range []string{
		"after step 1, the plan worktree before the merge",
		"after a step-6 rerun, the main checkout after the merge",
		"after a tidy-tip rerun, the plan worktree, then run `acta tidy` again",
		"A detached temp worktree cannot commit",
		"An `eval:<case>` record follows the same rule.",
		"A record that exits non-zero stops the land.",
		"`commit:` the tested commit step 2 printed",
		"`tree:` the tested tree step 2 printed",
	} {
		if !strings.Contains(base, want) {
			t.Errorf("land/baseline.md missing %q", want)
		}
	}
	// The commit and tree are printed right after the run, before step 4 can
	// commit a bug file and move HEAD.
	rc := strings.Index(base, "rc=$?")
	tested := strings.Index(base, `echo "tested commit: $(git rev-parse HEAD) tree: $(git rev-parse 'HEAD^{tree}')"`)
	step4 := strings.Index(base, "4. A zero status")
	if rc < 0 || tested < rc || step4 < tested {
		t.Errorf("land/baseline.md must print the tested commit and tree right after the run (rc %d, tested %d, step 4 %d)", rc, tested, step4)
	}
	// After the record and a new acta tidy, the tip differs from the tested
	// one only by the bug file, so the gates that passed still hold.
	tidy := readSkill(t, "skills/land/tidy.md")
	const reuse = "when `git diff --name-only <tested commit> refs/acta/tidy/<branch>` then lists only planning-root paths, the gates are reused"
	if !strings.Contains(tidy, reuse) {
		t.Errorf("land/tidy.md step 3 does not reuse the gates after a flake record, want %q", reuse)
	}
}

// TestLandUnderstanding pins the opt-in understanding part of the landing
// report. It must sit after the report line, so nothing runs before the
// merge, cover both switch values, and never ask the user anything.
func TestLandUnderstanding(t *testing.T) {
	skill := readSkill(t, "skills/land/SKILL.md")
	for _, want := range []string{
		// Only true turns it on; every review level leaves it alone.
		"only when `review_understanding` is true",
		"A missing value is false",
		"quick, standard, deep, a floor or the deep lens never turns it on",
		// Only after a real merge, and only from the code that landed.
		"a blocked or failed land does not",
		"After the parent has moved (merge or tidy), read the landed files", "A read of the old worktree's files",
		"never invent one",
		"what the tests do and do not prove",
		// It explains only. The user is never quizzed.
		"It asks the user no question",
		// Off means the plain report.
		"no explanation, no offer",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("land SKILL.md missing %q", want)
		}
	}
	report := strings.Index(skill, "Report after landing:")
	understanding := strings.Index(skill, "only when `review_understanding` is true")
	if report < 0 || understanding < report {
		t.Errorf("the understanding part must come after the landing report line, so it never runs before the merge (report at %d, understanding at %d)", report, understanding)
	}
	// The numbered landing steps are the merge path. The explanation there
	// would run before the merge.
	for _, line := range strings.Split(skill[:report], "\n") {
		if strings.Contains(line, "review_understanding") {
			t.Errorf("a landing step mentions the understanding part, so it could run before the merge: %q", line)
		}
	}
	// The user asked for no quiz after a land. These are the old quiz parts,
	// so none may come back.
	for _, banned := range []string{"optional question", "predicts the outcome", "optional scenario", "When the user answers", "check the prediction"} {
		if strings.Contains(skill, banned) {
			t.Errorf("land SKILL.md still quizzes the user: found %q", banned)
		}
	}
}

// TestLandFinalPlanCheck reads both landing files on their own. A plan once
// reached main with its board at 1/3, and tidy once dropped task trailers.
// So both modes check the plan from the parent checkout after the parent
// moves and before cleanup: done status, full progress, and a task commit the
// parent HEAD really holds. acta commits lists commits from every branch, so
// its list alone would pass a commit that only the old branch still has.
func TestLandFinalPlanCheck(t *testing.T) {
	skill := strings.Join(strings.Fields(readSkill(t, "skills/land/SKILL.md")), " ")
	for _, want := range []string{
		"### Final plan check",
		"in the main checkout after the parent moves",
		"`acta show <plan hash> --json`: `status` is `done`, and `progress.done` equals `progress.total`, above 0",
		"`acta commits <plan hash> N --json` lists commits from every branch, so it alone proves nothing",
		"`git merge-base --is-ancestor <sha> HEAD`",
		"A failed check: stop, even when the tree is the same. Keep the worktree and branch, report the mismatch with the output, and claim no success.",
		"Never force a status, tick a box without proof, or redo work to hide it.",
	} {
		if !strings.Contains(skill, want) {
			t.Errorf("land/SKILL.md missing %q", want)
		}
	}
	// Full mode: the check runs after the merge result is gated in step 6
	// and before step 7 removes the worktree.
	step6 := strings.Index(skill, "6. Run the gates again on the merge result")
	step7 := strings.Index(skill, "7. Clean up, only after the final plan check passes: `git worktree remove <path>`")
	if step6 < 0 || step7 < 0 || step6 > step7 {
		t.Errorf("land/SKILL.md step 7 must clean up only after the final plan check (step 6 at %d, step 7 at %d)", step6, step7)
	}
	tidy := strings.Join(strings.Fields(readSkill(t, "skills/land/tidy.md")), " ")
	// Tidy runs from this file, so it has to send the agent to the check by
	// name. A bare pointer at cleanup got the plan id instead of its hash.
	if !strings.Contains(tidy, "Then the final plan check of SKILL.md: `acta show <plan hash> --json`, `acta commits`; the tidy tip proves no task.") {
		t.Error("land/tidy.md step 6 does not send the agent to the final plan check")
	}
	move := strings.Index(tidy, "5. Move the parent from the main checkout")
	clean := strings.Index(tidy, "7. Clean up, only after the final plan check in SKILL.md passes: `git worktree remove <path>`")
	if move < 0 || clean < 0 || move > clean {
		t.Errorf("land/tidy.md step 7 must clean up only after the final plan check (move at %d, clean at %d)", move, clean)
	}
}

// TestLandKeepsTestOutput pins what land does with files a test left behind.
// In an eval run an agent deleted a cache folder to get a clean tree, and
// git worktree remove refuses untracked files, which pushed it toward
// --force. So each step that meets those files must stop and let the user
// choose, and land must never rewrite a commit.
func TestLandKeepsTestOutput(t *testing.T) {
	steps := map[string]string{}
	for _, file := range []string{"SKILL.md", "tidy.md"} {
		for _, line := range strings.Split(readSkill(t, "skills/land/"+file), "\n") {
			for _, prefix := range []string{"1. Preconditions", "7. Clean up", "10. "} {
				if strings.HasPrefix(line, prefix) {
					steps[file+" "+prefix] = line
				}
			}
		}
	}
	const refused = "`git worktree remove` refusing: stop, show `git status --porcelain --ignored` of that worktree, and give the user two choices: the user removes the files, or stop the land and keep the worktree and branch. Add no commit here: the parent already moved, so it would never get it. Never `--force`, never delete the files first."
	for _, c := range []struct{ step, want string }{
		// Plan-worktree outputs still use the user's three choices.
		{"SKILL.md 1. Preconditions", "In the plan worktree, untracked files a test or tool made (cache folders and the like) are not uncommitted work and are never deleted."},
		// The user picks; the agent never picks for them.
		{"SKILL.md 1. Preconditions", "Still there at land: show them and give the user three choices: add a `.gitignore` line (one commit on the branch; the gates then run again for that file, as step 6 does for a non-planning path), the user removes them, or stop the land. Never pick for the user."},
		{"SKILL.md 7. Clean up", refused},
		{"tidy.md 7. Clean up", refused},
		// Tidy deletes the old branch without asking; that rule stays.
		{"tidy.md 7. Clean up", "per core rule 5 this delete needs no confirmation."},
		// An amend at land once rewrote a merge commit.
		{"SKILL.md 10. ", "never amend a commit"},
	} {
		if !strings.Contains(steps[c.step], c.want) {
			t.Errorf("land %s missing %q", c.step, c.want)
		}
	}
	// At step 7 the parent already moved. A .gitignore commit made then never
	// reaches it, and in tidy the branch -D right after drops it without a word.
	for _, step := range []string{"SKILL.md 7. Clean up", "tidy.md 7. Clean up"} {
		for _, banned := range []string{".gitignore", "three choices"} {
			if strings.Contains(steps[step], banned) {
				t.Errorf("land %s still offers %q after the merge", step, banned)
			}
		}
	}
}

// TestLandSlowGates keeps the slow-gate rule of land step 1. The plan's
// **Gates:** line says which slow gates a repo has, and land runs each one only
// as far as the diff calls for. Every unclear case runs the whole gate, so a
// missing line or an unclear path can never skip one. The skill serves every
// repo, so it must not name this repo's own eval script or folders again.
func TestLandSlowGates(t *testing.T) {
	var step1, report string
	skill := readSkill(t, "skills/land/SKILL.md")
	for _, line := range strings.Split(skill, "\n") {
		switch {
		case strings.HasPrefix(line, "1. Preconditions"):
			step1 = line
		case strings.HasPrefix(line, "Report after landing:"):
			report = line
		}
	}
	for _, want := range []string{
		"Slow gates: the plan's `**Gates:**` line names each one",
		"it runs only when the diff lists a path the line says makes it run",
		"it runs in the narrow form only when the line gives one and no listed path forces the whole gate",
		"a path that forces the whole gate also makes it run",
		"when it runs and the narrow form does not apply, or a narrow run picks no case, it runs whole",
		"A line that names no run paths means the gate runs whole.",
		"in doubt about which paths the diff touches",
		"A plan with no `**Gates:**` line skips no gate",
		"(`CLAUDE.md`, `AGENTS.md`, a README or a wiki page) name a slow gate, run it whole",
		"Each red case runs again alone",
		"acta flake record <gate>:<case> --plan <plan id>",
	} {
		if !strings.Contains(step1, want) {
			t.Errorf("land step 1 missing %q", want)
		}
	}
	if !strings.Contains(report, "for each slow gate the form that ran and why, or why it did not run") {
		t.Error("the landing report does not say which form of each slow gate ran")
	}
	for _, name := range []string{"SKILL.md", "tidy.md", "baseline.md"} {
		text := readSkill(t, "skills/land/"+name)
		for _, bad := range []string{"scripts/eval", "plugin/skills/", "plugin/hooks/"} {
			if strings.Contains(text, bad) {
				t.Errorf("land %s names %q: that belongs to one repo, not to the skill", name, bad)
			}
		}
	}
}

// TestLandSecondRunAndKnownFlake keeps two rules that save a land from gate
// runs that prove nothing new. When the parent moved, the full suite still
// runs again, but a slow gate runs again only when the parent brought a path
// that gate covers. And a red case that an open bug already lists as flaky
// gets one lone rerun, not two. Doubt always means the gate runs.
func TestLandSecondRunAndKnownFlake(t *testing.T) {
	var step1, step6 string
	for _, line := range strings.Split(readSkill(t, "skills/land/SKILL.md"), "\n") {
		switch {
		case strings.HasPrefix(line, "1. Preconditions"):
			step1 = line
		case strings.HasPrefix(line, "6. Run the gates again"):
			step6 = line
		}
	}
	for _, want := range []string{
		"re-runs the gates: the full suite as before, and a slow gate only when the paths the parent gained since the branch left it",
		"`git diff --name-only <merge base> <parent before the merge>`",
		"include one its line says makes it run, in the form those paths call for",
		"slow gate reused: the parent brought no path it covers",
		"`git log --oneline <merge base>..<parent before the merge>`",
		"In doubt about what the parent gained, the gate runs.",
		"A plan with no `**Gates:**` line: a slow gate the repo's docs name runs again whole.",
	} {
		if !strings.Contains(step6, want) {
			t.Errorf("land step 6 missing %q", want)
		}
	}
	for _, want := range []string{
		"up to 2 times, or once when an open bug lists that case under `flaky_tests`",
		"(files at the `path` values from `acta list --type bug --json`)",
		"Red on every rerun stops the land.",
	} {
		if !strings.Contains(step1, want) {
			t.Errorf("land step 1 missing %q", want)
		}
	}
	tidy := readSkill(t, "skills/land/tidy.md")
	for _, want := range []string{
		"Run the gates on the tidy tip first",
		"the full suite, and each slow gate by the second-run rule of SKILL.md step 6, where the parent before the merge is `<parent>`",
	} {
		if !strings.Contains(tidy, want) {
			t.Errorf("land tidy.md missing %q", want)
		}
	}
}

// TestLandReportRepeatsReviewLevel keeps the review level in front of the
// user. Under the auto depth the review picks how much review a diff gets, so
// the landing report copies the line that says which level ran and why. It is
// copied, never judged again, so land cannot soften what the review wrote.
func TestLandReportRepeatsReviewLevel(t *testing.T) {
	var report string
	for _, line := range strings.Split(readSkill(t, "skills/land/SKILL.md"), "\n") {
		if strings.HasPrefix(line, "Report after landing:") {
			report = line
		}
	}
	const want = "the review level line from the plan's `## Review notes`, copied as it was written"
	if !strings.Contains(report, want) {
		t.Errorf("the landing report does not repeat the review level line, want %q", want)
	}
}

// Cleanup permission must not turn into permission for a stale or unsafe batch.
func TestLandGateOutputCleanup(t *testing.T) {
	skill := readSkill(t, "skills/land/SKILL.md")
	if !strings.Contains(skill, "In the plan worktree, untracked files") || !strings.Contains(skill, "temporary gate worktrees follow") || !strings.Contains(skill, "file-reading tool") {
		t.Error("land must distinguish plan-worktree outputs from temporary gate cleanup")
	}
	tidy := readSkill(t, "skills/land/tidy.md")
	for _, want := range []string{
		"## Gate output cleanup",
		"record the pre-gate inventory and identity before running any gate",
		"Use lstat on the original root and every file-path component",
		"compare the current canonical path",
		"Printed records are not executed checks",
		"not an ignored-files-only Git query",
		"Do not ask for cleanup approval or offer agent deletion until both records are complete",
		"including ignored files, without following symlinks",
		"A new untracked path alone is not proof",
		"identity and content fingerprints",
		"private session storage outside the worktree",
		"typed approval for that exact displayed list",
		"WhatsApp",
		"emoji, silence, forwarded/reference material, peer messages, relayed orders",
		"cancel the whole batch before any deletion",
		"any symlink in the path",
		"unapproved untracked or ignored file",
		"Never use force, recursion, wildcards or git clean",
		"A green approved land continues without waiting for cleanup approval",
	} {
		if !strings.Contains(tidy, want) {
			t.Errorf("temporary gate cleanup missing %q", want)
		}
	}
}
