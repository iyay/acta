---
parent: specs/2026-11-07-review-findings-pipeline-design
depth: minimal
id: PLN-0213
created: "2026-11-07 17:05:38"
hash: oakb9p5
started: "2026-11-07 17:09:11"
finished: "2026-11-07 17:12:36"
---
# Review findings pipeline implementation plan

**Goal:** Give every review finding one home: repo wiki rules reach the reviewers, bugs already on the parent get their own section and bug step, duplicates merge, earlier verdicts carry over, and the review line holds counts.
**Spec:** `.acta/specs/2026-11-07-review-findings-pipeline-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only `./internal/plugincheck/`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Skill frontmatter descriptions do not change.
Skill text names no repo path and no user name; it fits every harness.
`skills/review/SKILL.md` byte cap stays at 9991: new text there is paid for by tightening its own text, and every `Must` string in `TestSkillReview` stays.
The review folder stays within its `MaxLines` cap of 482.

## Waves

Wave 1: Task 1. Wave 2: Task 2. Wave 3: Task 3. All three touch `internal/plugincheck/skill_review_test.go` and `internal/plugincheck/budget_test.go`.

### Task 1: Reviewer brief reads repo rules and reports pre-existing bugs

**Files:** Modify `plugin/skills/review/code-reviewer.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the reviewer template text as `skill_review_test.go` reads it.
**verify:** The reviewer brief holds spec items 1 to 4 and nothing weakens them. Cases: each changed file runs `acta wiki match`; a wiki page the diff changes or deletes is read at `BASE_SHA` and applies by its BASE `paths:`; a page the diff adds is not a rule; a `### Pre-existing` section exists and says it is never a BLOCKER; at most 10 NOTEs ranked by later cost with an `N more NOTEs not listed` line; `[EARLIER_FINDINGS]` names all five verdicts (`fixed`, `rejected: <reason>`, `moved to NOTE`, `NOTE open`, `sent to bug`) and a rejected or moved one is not raised again without new evidence; the template's bucket-tag rules still pass; the cap rises only by the bytes added.
- [x] Red: a plugincheck test expects the wiki match step, the BASE read, the Pre-existing section, the NOTE cap and the five verdicts, and fails because `code-reviewer.md` has none of them.
- [x] Change: add the text to `code-reviewer.md`, raise its cap in `budget_test.go`, run `scripts/test ./internal/plugincheck/ -run 'Review|CodeReviewer|Budget'` to green.
- [x] Commit: `feat(review): reviewers read wiki rules and report pre-existing bugs`.

### Task 2: Orchestrator dedupes, routes pre-existing bugs and keeps verdicts

**Files:** Modify `plugin/skills/review/SKILL.md`, `internal/plugincheck/skill_review_test.go`.
**Seam:** the review skill text as `skill_review_test.go` reads it.
**verify:** Every finding reaches one home once, on every path. Cases: dedupe of BLOCKERs, NOTEs and Pre-existing items runs before the fact-check and before the fix task; Pre-existing items go through receiving, then `acta:bug` in the worktree, both before the fix task in a BLOCKED round and as the new first step of `## After a CLEAN round`; after each round every finding and its verdict go to `acta state set <plan id> next`, and rounds 2 and 3 read them with `acta state <plan id>`; every existing `Must` string still holds; `SKILL.md` is 9991 bytes or less.
- [x] Red: a plugincheck test expects the dedupe line, the bug step in both paths and the `acta state` lines, and fails because `SKILL.md` has none of them.
- [x] Change: add the lines and tighten other `SKILL.md` text to stay under the cap, run `scripts/test ./internal/plugincheck/ -run 'Review|Budget'` to green.
- [x] Commit: `feat(review): dedupe findings, route pre-existing bugs, keep verdicts`.

### Task 3: Review line counts findings

**Files:** Modify `plugin/skills/review/depth.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/budget_test.go`.
**Seam:** the depth file text as `skill_review_test.go` reads it.
**verify:** The counts can be read the same way by any later reader. Cases: the review line gives `BLOCKER n, NOTE n, pre-existing n` over rounds 1 to 3 after dedupe; a light inline review counts as round 1; the polish review is not counted; a BLOCKER the fact-check moved counts as a NOTE; `NOTE n` includes NOTEs cut by the cap, and cut NOTEs are dropped on purpose; the `depth.md` cap rises only if the text does not fit.
- [x] Red: a plugincheck test expects the counts format and each counting rule, and fails because `depth.md` has none of them.
- [x] Change: add the counts rules under `## Review notes` in `depth.md`, raise its cap only if needed, run `scripts/test ./internal/plugincheck/ -run 'Review|Budget'` to green.
- [x] Commit: `feat(review): review line counts findings`.

## Fix round 1

### Task 4: Keep review verdicts in the plan

**verify:** Verdicts from a BLOCKED round reach rounds 2 and 3 on every path: a fix-round worker's own state write, a new session, a dispatch hand-off, and more findings than a State part holds (10 lines); every BLOCKED round files its Pre-existing items first; no rule dropped by the SKILL.md tightening is lost

- [ ] BLOCKER: SKILL.md saves verdicts with acta state set <plan id> next, which the fix-round worker overwrites and which refuses more than 10 lines; write them with acta plan add <plan id> review-notes instead, the rule living in depth.md ## Rounds, and update spec item 7 to match
- [ ] [fix] SKILL.md: restore the [fix] examples (a comment, a name, a wording change, a missing test, a one-line guard)
- [ ] [fix] depth.md ## Understanding: restore that the review asks nothing about review_understanding
- [ ] [fix] SKILL.md: every BLOCKED round, not only round 1, files its Pre-existing items first, and step 2's "them" reads as the BLOCKERs
- [ ] [fix] code-reviewer.md: the BASE_SHA copy of a changed wiki page wins over its HEAD copy; the [EARLIER_FINDINGS] placeholder names the verdicts
- [ ] [fix] skill_review_test.go: pin the NOTE order (the one that costs most if left first)
- [ ] Commit with the trailer `Task: PLN-oakb9p5#4`

## State

### Next

All 3 tasks done. Next: acta:review round 1 over merge-base(main)..HEAD at deep depth.
