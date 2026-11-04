---
parent: specs/2026-11-04-skill-token-diet-design
depth: minimal
id: PLN-0210
created: "2026-11-04 19:45:47"
hash: w4gjmeg
started: "2026-11-04 19:48:35"
finished: "2026-11-04 20:40:32"
---
# Diet: review skill implementation plan

**Goal:** Run the Standards-axis reviewer on a cheaper tier and load the receiving-findings text only when there is feedback to answer.
**Spec:** `.acta/specs/2026-11-04-skill-token-diet-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only `./internal/plugincheck/`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Skill frontmatter descriptions do not change.
Skill text names no repo path and no user name; it fits every harness.
`skills/review/SKILL.md` byte cap drops to 10000 or less.

## Waves

Wave 1: Task 1. Wave 2: Task 2. Both touch `plugin/skills/review/SKILL.md`.

### Task 1: Standards reviewer on a cheaper tier

**Files:** Modify `plugin/skills/review/SKILL.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the review skill text as `skill_review_test.go` reads it.
**verify:** Review never tells the agent to name no model for the Standards reviewer, and never pins a model. Cases: the text holds the spec's sentence (Spec axis on your own model; Standards axis on a cheaper, faster tier, sonnet in Claude Code as the example; harness default when no cheaper tier); the old "Name no model for any subagent" line is gone from review in any wording; slice keeps its line unchanged.
- [x] Red: a review test expects the cheaper-tier sentence and fails because review still says "Name no model for any subagent".
- [x] Change: replace the line in review with the spec's sentence, run `scripts/test ./internal/plugincheck/ -run 'Review|Budget'` to green.
- [x] Commit: `feat(review): Standards reviewer runs on a cheaper tier`.

### Task 2: receiving findings in its own file

**Files:** Create `plugin/skills/review/receiving.md`; modify `plugin/skills/review/SKILL.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the review skill folder as plugincheck reads it, plus `scripts/eval --case 'review-*' --case 'polish-*'`.
**verify:** No receiving rule is lost and none loads without feedback. Cases: every subsection of today's `## Receiving findings` is in `receiving.md` word for word; the body names all three triggers (any reviewer finding, review feedback from the user or an outside reviewer, a light inline review that finds something) in one pointer; the `SKILL.md` byte cap is 10000 or less and `receiving.md` has its own cap; the review and polish evals give the same verdicts on the branch as on main, and a CLEAN-run transcript has no Read of `receiving.md`.
- [x] Red: a plugincheck test expects `receiving.md` and the three triggers and fails because the file does not exist.
- [x] Change: move the section, add the pointer, set caps, run `scripts/test ./internal/plugincheck/` then the evals on branch and main.
- [x] Commit: `feat(review): load receiving findings only when there is feedback`.

## Fix round 1

### Task 3: Restore the dispatch full-tier rule

**verify:** No path lets dispatch work skip the full tier: Review tiers names it, a Must pins it, and the body stays at or under 10000 bytes

- [x] review/SKILL.md Review tiers lost the rule that work the dispatch executor wrote is always full tier; put it back in that section and pin it in skill_review_test.go
- [x] Commit with the trailer `Task: PLN-w4gjmeg#3`

## Polish

### Task 4: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] pin "It counts for this run only and is never saved." from review/depth.md in TestReviewDepthFiles
- [x] Commit: `polish: review notes for PLN-0210`

## Review notes

- Deep depth, full tier; Spec axis opus, Standards axis sonnet. Round 1 BLOCKED (dispatch full-tier rule lost in the trim, confirmed by fact-check), fixed in round 2, CLEAN. Polish pinned the run-only rule (reviewed CLEAN). One [debt]: Codex no-model text vs the cheaper tier. Other notes [note]: pointer placement, exact caps, duplicate wording between SKILL.md and depth.md.

## State

### Next

Paused 2026-11-04 by user. Review CLEAN, polish done, debt filed, full suite green at a3262d8 (4669 pass, baseline recorded).
Full eval set was started; log at the session scratchpad eval-review.log, may be cut short. Next: rerun scripts/eval (or only the red cases), then land per tidy.md.
