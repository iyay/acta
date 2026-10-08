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
			"In doubt: full.",
			"acta debt new",
			"drop the bucket tag",
			"the priority tag comes after the bucket tag",
			"A NOTE may start with `(high) `, `(medium) ` or `(low) `",
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
			"The orchestrator runs the full test suite with the output shown, then the polish range gets the review its tier picks (`## Review tiers`).",
			"When the polish or the tests fail, revert that commit, and that moves those items to `[debt]`.",
			"The polish commit uses no round.",
			"A BLOCKER in the polish review reverts the polish commit and moves those items to `[debt]`. It starts no fix round.",
			"The polish-review NOTEs sort into `[debt]` or `[note]` only, so no second polish follows.",
			"acta debt new <plan id>",
			"a `## Review notes` section in the plan file",
			"that section is committed with the plan",
			"Then `acta:land`.",
			"A polish sent to another agent goes out with `acta dispatch send --round polish`, the `[fix]` NOTE list on `--note-file -`; when `/acta:review` arrives with round polish, run the full test suite with the output shown, the review its tier picks (`## Review tiers`) over the polish range, then step 3 onward.",
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
		MustNot: []string{"Small means one file", "Every polish commit gets the two reviewers", "A polish commit is never a small change", "explicit instruction-file violation", "Acknowledge strengths", "Production readiness", "superpowers:", "Critical", "Important (Should Fix)", "Minor", "GitHub Thread Replies", "A change you judge small", "kept in memory", "one line in memory", "never a task",
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
		},
	})

	// One section, not one sentence per pointer: every other place in the
	// plugin names this heading, so the heading itself must appear once.
	if n := strings.Count(readSkill(t, "skills/review/SKILL.md"), "\n## After a CLEAN round\n"); n != 1 {
		t.Errorf("review/SKILL.md has %d \"## After a CLEAN round\" headings; the CLEAN-round flow lives in one section", n)
	}

	// A polish reply-back runs the suite and the tier's review before step 3;
	// "step 3 onward" on its own skips both.
	for _, line := range strings.Split(readSkill(t, "skills/review/SKILL.md"), "\n") {
		if strings.Contains(line, "step 3 onward") &&
			(!strings.Contains(line, "full test suite") || !strings.Contains(line, "Review tiers")) {
			t.Errorf("review/SKILL.md sends a polish reply-back to step 3 onward without the full suite and the tier review: %q", line)
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
