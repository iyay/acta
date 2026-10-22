package plugincheck

import (
	"strings"
	"testing"
)

func TestSkillReview(t *testing.T) {
	CheckSkill(t, SkillRule{
		Name:     "review",
		MaxLines: 482,
		Must: []string{
			"Spec axis", "Standards axis", "BLOCKER", "NOTE", "three questions",
			"deep lens", "never a round 4", "code-reviewer.md", "CLEAN or BLOCKED",
			"## Receiving findings", "Fix round", "acta:land",
			"## Where findings go", "acta:bug", "already on the parent branch",
			"## Review tiers",
			"Light review: the orchestrator reads the full diff",
			"A fix round always takes the two reviewers.",
			"Reading the diff yourself never replaces the reviewers its depth names, however small the diff.",
			"review_depth", "depth.md", "fact-check.md", "/acta:review quick|standard|deep",
			"Fix rounds run at least `standard`",
			"In doubt: full.",
			"acta debt new",
			"drop the bucket tag",
			"the priority tag comes after the bucket tag",
			"A `[debt]` NOTE must start with `(medium) ` or `(high) `",
			"A NOTE that would be `(low)` or has no priority tag is a `[note]`.",
			"never the full suite",
			"[fix]", "[debt]", "[note]",
			"## Review notes",
			"moves those items to `[debt]`",
			"Land anyway: that round's `[fix]` NOTEs move to `[debt]`",
			"The `[fix]` NOTEs join that fix task.",
			// The whole CLEAN-round flow, in the one section that owns it.
			"\n## After a CLEAN round\n",
			"The polish commit counts as no round.",
			"Sort every NOTE into `[fix]`, `[debt]` or `[note]` with the bucket rules in `## Where findings go`.",
			"one polish commit holding all of them",
			"The orchestrator runs the narrow tests of the code the polish touched and the formatter check, with the output shown, never the full suite; the full suite waits for `acta:land`. Then the polish range gets the review its tier picks (`## Review tiers`).",
			"When the polish or the tests fail, revert that commit, and that moves those items to `[debt]`.",
			"The polish commit uses no round.",
			"A BLOCKER in the polish review reverts the polish commit and moves those items to `[debt]`. It starts no fix round.",
			"The polish-review NOTEs sort into `[debt]` or `[note]` only, so no second polish follows.",
			"acta debt new <plan id>",
			"The `[note]` NOTEs are not written one by one.",
			"one summary line under `## Review notes`",
			// Plan sections go in through acta, so no agent reads the whole
			// plan to add one.
			"`acta plan add <plan id> fix-round --title <title> --verify <line>`",
			"`acta plan add <plan id> polish`",
			"`acta plan add <plan id> review-notes`",
			"Then `acta:land`.",
			"A polish sent to another agent goes out with `acta dispatch send --round polish`, the `[fix]` NOTE list on `--note-file -`; when `/acta:review` arrives with round polish, run the narrow tests of the code the polish touched and the formatter check, with the output shown, then the review its tier picks (`## Review tiers`) over the polish range, then step 3 onward.",
			// Every other place names that section instead of restating it.
			"A CLEAN round runs the polish flow in `## After a CLEAN round`.",
			"run `## After a CLEAN round`, which ends in `acta:land`.",
			"CLEAN: run `## After a CLEAN round`. BLOCKER: one more fix task.",
			"CLEAN: run `## After a CLEAN round`. BLOCKER: stop and ask the user",
			"A round with no BLOCKER does not start; it runs `## After a CLEAN round`.",
			"No: no commit, one NOTE; the `[fix]` NOTEs are handled in `## After a CLEAN round`.",
			"and over a polish commit",
			"A CLEAN round: no new round, no new fix task. Run `## After a CLEAN round`.",
		},
		MustNot: []string{"Small means one file", "A NOTE may start with `(high) `, `(medium) ` or `(low) `", "one line each, and that section", "Every polish commit gets the two reviewers", "A polish commit is never a small change", "explicit instruction-file violation", "Acknowledge strengths", "Production readiness", "superpowers:", "Critical", "Important (Should Fix)", "Minor", "GitHub Thread Replies", "A change you judge small", "kept in memory", "one line in memory", "never a task",
			"land now with",
			`Before every commit after the deliverable is green, ask: "Without this, does a real user see a wrong result today?" No: no commit, one NOTE.`,
			"CLEAN: land.",
			"The agent reviews it itself",
			"no review round follows it",
			"start no reviewers",
			// The polish rule used to live as exceptions over many
			// sentences. It is one section now; these copies are gone.
			"except a `[fix]` NOTE, which rides the fix task or lands as the polish commit",
			"The `[fix]` NOTEs ride with it.",
			"(the polish commit when there are `[fix]` NOTEs)",
			"The polish commit is not a fix round",
			"on reply-back the two reviewers review the polish range; it is not a fix round",
			"Rounds count only fix rounds",
			"A polish sent to another agent is dispatched with --round polish.",
			"send one `/goal`",
			"the file is already in the diff, the fix adds no new logic, and the path is not security, auth, money, migration or delete; `[debt]` when leaving it costs something later that you can name; `[note]` otherwise.",
			"the `[fix]` NOTEs join that round's one fix task, so the next round reviews them",
			// The hand-written forms the commands replaced.
			"appended to the same plan as a `## Fix round <n>` section",
			"first append a `## Polish` task to the plan",
			"<one box per [fix] NOTE>",
			"note the move there",
			"that section is committed with the plan",
		},
	})

	// One section, not one sentence per pointer: every other place in the
	// plugin names this heading, so the heading itself must appear once.
	if n := strings.Count(readSkill(t, "skills/review/SKILL.md"), "\n## After a CLEAN round\n"); n != 1 {
		t.Errorf("review/SKILL.md has %d \"## After a CLEAN round\" headings; the CLEAN-round flow lives in one section", n)
	}

	// A polish reply-back runs the narrow checks and the tier's review before
	// step 3; "step 3 onward" on its own skips both.
	for _, line := range strings.Split(readSkill(t, "skills/review/SKILL.md"), "\n") {
		if strings.Contains(line, "step 3 onward") &&
			(!strings.Contains(line, "narrow tests") || !strings.Contains(line, "formatter") || !strings.Contains(line, "Review tiers")) {
			t.Errorf("review/SKILL.md sends a polish reply-back to step 3 onward without the narrow tests, the formatter and the tier review: %q", line)
		}
	}
	// The full suite runs only in acta:land. Land runs it right after the
	// polish, so a full run on the polish too would test the same code twice.
	for _, line := range strings.Split(readSkill(t, "skills/review/SKILL.md"), "\n") {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "polish") &&
			(strings.Contains(lower, "full test suite") || strings.Contains(lower, "run the full suite") || strings.Contains(lower, "runs the full suite")) {
			t.Errorf("review/SKILL.md asks for the full suite on a polish; land owns it: %q", line)
		}
	}
	// The fix-task rule is stated once, in budget step 2.
	if n := strings.Count(readSkill(t, "skills/review/SKILL.md"), "NOTEs join that fix task"); n != 1 {
		t.Errorf("review/SKILL.md says \"NOTEs join that fix task\" %d times; budget step 2 owns it", n)
	}
}

// TestCodeReviewerTemplateBucketTags pins the reviewer template: every NOTE
// it prints carries a bucket tag, so the orchestrator can sort without asking.
func TestCodeReviewerTemplateBucketTags(t *testing.T) {
	txt := readSkill(t, "skills/review/code-reviewer.md")
	for _, want := range []string{"bucket tag", "[fix]", "[debt]", "[note]"} {
		if !strings.Contains(txt, want) {
			t.Errorf("code-reviewer.md missing %q", want)
		}
	}
	// The order is the contract: bucket tag, then the optional priority tag,
	// then the note. The orchestrator drops the bucket tag before
	// `acta debt new`, so a wrong order loses the priority too.
	bucket, priority, note := strings.Index(txt, "bucket tag"), strings.Index(txt, "priority tag"), strings.Index(txt, "then the note")
	if bucket < 0 || priority < 0 || note < 0 || !(bucket < priority && priority < note) {
		t.Errorf("code-reviewer.md must say bucket tag, then the priority tag, then the note; got indexes %d, %d, %d", bucket, priority, note)
	}
}

// TestReviewDepthFiles pins the three files that carry the review levels. The
// floors and the fact-check limits are the safety part, so each phrase that
// keeps a review from running too light or a BLOCKER from being dropped is
// checked here.
func TestReviewDepthFiles(t *testing.T) {
	files := map[string][]string{
		"skills/review/depth.md": {
			"quick", "standard", "deep",
			"Fix rounds run at least `standard`",
			"never a round 4",
			"security, auth, data migration, money or delete path runs at least `standard` with the deep lens",
			"`dispatch` executor wrote runs at least `standard` in the full tier",
			"but not drop it below these",
			"`## Review notes`",
			"With no BLOCKER it does not run",
			"Rounds 2 and 3 hand the reviewers the earlier findings",
			"No level, from the argument or the setting, goes below them",
			"A light-tier diff keeps the inline light review",
			"The orchestrator numbers the BLOCKERs before the fact-check",
			"becomes a NOTE, sorted into a bucket like any other",
		},
		"skills/review/fact-check.md": {
			"(A) the code it names is not in the diff under review",
			"(B) one line of the diff contradicts the claim",
			"In doubt, it stays a BLOCKER.",
			"security", "data loss", "concurrency", "behaviour change",
			"First line of your report: the model you ran on.",
		},
		"skills/review/code-reviewer.md": {
			"[EARLIER_FINDINGS]",
			"every changed file gets its own pass",
		},
	}
	for rel, wants := range files {
		txt := readSkill(t, rel)
		for _, want := range wants {
			if !strings.Contains(txt, want) {
				t.Errorf("%s missing %q", rel, want)
			}
		}
	}
}

// TestReviewUnderstanding keeps review_understanding apart from the review
// level. No level, argument, floor or the deep lens may turn it on, and the
// review hands the report part to acta:land instead of asking anything itself.
func TestReviewUnderstanding(t *testing.T) {
	files := map[string][]string{
		"skills/review/SKILL.md": {
			"`review_understanding` is a separate switch",
			"no level turns it on",
			"`acta:land` adds what it asks for",
		},
		"skills/review/depth.md": {
			"## Understanding",
			"`review_understanding` is its own switch, false when missing",
			"quick, standard, deep, the argument, a floor and the deep lens leave it as it is",
			"`acta:land` ends its report with a short explanation and asks no question",
			"When it is false, there is none",
		},
	}
	for rel, wants := range files {
		txt := readSkill(t, rel)
		for _, want := range wants {
			if !strings.Contains(txt, want) {
				t.Errorf("%s missing %q", rel, want)
			}
		}
	}
	// Nobody quizzes the user about the change, in review or after the land.
	for _, rel := range []string{"skills/review/SKILL.md", "skills/review/depth.md"} {
		txt := readSkill(t, rel)
		for _, banned := range []string{"predicts the outcome", "optional question"} {
			if strings.Contains(txt, banned) {
				t.Errorf("%s still quizzes the user: found %q", rel, banned)
			}
		}
	}
}
