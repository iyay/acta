---
parent: specs/2026-10-01-review-note-triage-design
depth: minimal
id: PLN-0070
created: "2026-10-01 15:10:51"
hash: wq5c6bl
started: "2026-10-01 15:14:18"
finished: "2026-10-01 15:16:31"
---
# Review NOTE Triage Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Review NOTEs are sorted into `[fix]`, `[debt]` and `[note]`, so only `[debt]` NOTEs reach `acta debt new`.

**Spec:** `.acta/specs/2026-10-01-review-note-triage-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Skill text is plain English with short words. Keep every phrase the plugincheck tests already require, unless this plan replaces it.
- The bucket tags are exactly `[fix]`, `[debt]` and `[note]`. The plan section is exactly `## Review notes`.
- The three-round budget and "a round with no BLOCKER does not start" stay as they are.

## Waves

- Wave 1: Task 1, Task 2 (no shared files).

### Task 1: Review skill sorts NOTEs into three buckets

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/review/code-reviewer.md`, `internal/plugincheck/skill_review_test.go`

**verify:** No text in the review skill or the reviewer template still says that every NOTE goes to debt, or that a NOTE is never a task without the `[fix]` exception. List each place checked. The bucket rules, the polish commit on a CLEAN round (self-review, full tests, revert and move to `[debt]` on failure), the `[fix]` NOTEs riding a BLOCKER fix round, and the `## Review notes` section all appear.

- [x] Failing test: `TestSkillReview` requires `[fix]`, `[debt]`, `[note]`, `## Review notes` and the polish commit phrase, and the template check requires the tag on NOTE lines. It fails because the skill text does not have them yet.
- [x] Change: rewrite "Finding bar" (the "never a task" line), "Budget" and "Where findings go" in SKILL.md with the spec's bucket rules and flow. Change the NOTEs output format in code-reviewer.md so each NOTE starts with its suggested bucket tag.
- [x] Commit: `Review sorts NOTEs into fix, debt and note buckets`

### Task 2: Hand-off text says NOTEs are sorted, not all filed

**Files:** `plugin/references/house-rules.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/plugin_test.go`, `internal/plugincheck/build_dispatch_test.go`

**verify:** No hand-off text (house rules, dispatch brief, dispatch memory sweep) still claims that the orchestrator files every NOTE with `acta debt new`. List each line checked. "never write NOTEs to memory" stays.

- [x] Failing test: the house-rules check and the dispatch check forbid "files them with acta debt new" and require "sorts them". They fail on the current text.
- [x] Change: house-rules.md line 8 and dispatch.md line 262 say the orchestrator sorts NOTEs through `acta:review` once the round is CLEAN. The memory sweep line in dispatch.md points to the same sort.
- [x] Commit: `Hand-off text says review NOTEs are sorted`

## Fix round 1

### Task 3: Close the three round-1 BLOCKERs and the [fix] NOTEs

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/review/code-reviewer.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/build_dispatch_test.go`

**verify:** An agent that follows the review skill word for word, on every path (round 1, 2 or 3 CLEAN with `[fix]` NOTEs, a BLOCKER round with `[fix]` NOTEs, a `[debt]` NOTE with and without a priority), makes the polish commit, keeps the priority, and stores no bucket tag in a debt line. List every sentence in plugin/ that tells an agent to land, to skip a commit, or to pipe NOTEs, and say how each now reads.

- [x] Failing test: `TestSkillReview` forbids "land now with" and the old line-47 wording with no exception, and requires that the skill says to drop the bucket tag before `acta debt new`, that the priority tag comes after the bucket tag, and that the `## Review notes` lines are committed with the plan; `TestCodeReviewerTemplateBucketTags` requires the tag order; `TestBuildDispatchDelivery` forbids "one NOTE (memory)". They fail on the current text.
- [x] Change: (1) Budget steps 1-4: "CLEAN" means run the CLEAN-round flow in "Where findings go" (polish commit when there are `[fix]` NOTEs), then `acta:land`. (2) The "Without this, does a real user see a wrong result today?" rule gets an exception for the `[fix]` NOTEs, in a fix round and in the polish commit. (3) `[debt]` lines go to `acta debt new` with the bucket tag removed; a priority tag, when present, comes right after the bucket tag in reviewer output, so after removal it starts the line. (4) Move the `(high)`/`(medium)`/`(low)` sentence to the `[debt]` bullet. (5) Fix the grammar of "revert that commit and it moves those items to `[debt]`" while keeping the pinned phrase. (6) Say the `## Review notes` section is committed with the plan before `acta:land`. (7) dispatch.md "Stopping rule": "logged as one NOTE (memory)" becomes a NOTE sorted through `acta:review`. (8) code-reviewer.md shows the order: bucket tag, then optional priority tag, then the note.
- [x] Commit: `Review NOTE flow: polish before land, debt lines lose the bucket tag`

## Polish

Round 2 came back CLEAN on both axes. These `[fix]` NOTEs land as one polish commit, checked by the orchestrator (full diff read, full suite run), with no new review round.

### Task 4: Polish from the round 1 and round 2 `[fix]` NOTEs

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/skill_review_test.go`

**verify:** No sentence in the review skill sends a `[fix]` NOTE anywhere but the fix task or the polish commit, and none leaves one with no home. This includes round 3 BLOCKED where the user says land anyway. Reverting any one of these sentences turns a plugincheck test red. List every sentence checked.

- [x] Failing test: `TestSkillReview` requires "except a `[fix]` NOTE, which rides the fix task or lands as the polish commit" and the land-anyway sentence, and forbids "CLEAN: land." and "The agent reviews it itself". It fails on the current text.
- [x] Change: (1) SKILL.md, the line-47 rule ends "except a `[fix]` NOTE, which rides the fix task or lands as the polish commit". (2) Round 3 BLOCKED with land anyway: that round's `[fix]` NOTEs move to `[debt]`. (3) "The agent reviews it itself" becomes "The orchestrator reviews it itself". (4) dispatch.md "Stopping rule": "Everything else is one NOTE, which the orchestrator sorts through `acta:review` into `[fix]`, `[debt]` or `[note]`, or one follow-up ticket; the branch lands."
- [x] Commit: `Review polish: fix NOTE homes on every path`

## Fix round 2

The polish carried the user's own rulings (polish is never a fix round; polish still gets the two reviewers), so its review BLOCKER is fixed here as fix round 2 instead of reverting the polish. User approved on 2026-10-01.

### Task 5: One place for the CLEAN-round flow

User ruling 2026-10-01: the polish rule was spread as exceptions over many sentences, which confuses agents. Put it in one section; every other place points there.

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/build/dispatch.md`, `plugin/references/house-rules.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/build_dispatch_test.go`, `internal/plugincheck/plugin_test.go`

**verify:** The polish rule lives in exactly one section, `## After a CLEAN round`, in the review skill. No other sentence in plugin/ restates or makes an exception for the polish; each one that touches the CLEAN-round path points to that section by name. On every path (budget steps 1-4, dispatch loop steps 9-11, small changes, the opening line, a polish reply-back), an agent following the text gives every commit the two reviewers, never counts the polish as a round, and on a polish BLOCKER reverts the polish and moves those items to `[debt]`. List every sentence checked. Reverting the section or a pointer turns a plugincheck test red.

- [x] Failing test: pins require the `## After a CLEAN round` section with its steps and checklist, the round rule sentence, and the pointers in dispatch.md steps 10-11 and house-rules.md; MustNot forbids the scattered polish exceptions. They fail on the current text.
- [x] Change: (1) Add `## After a CLEAN round` to the review skill as numbered steps: sort each NOTE with the checklist (file already in the diff, no new logic, not a security, auth, money, migration or delete path: `[fix]`; a real later cost you can name: `[debt]`; else `[note]`); `[fix]` NOTEs exist: one polish commit, full test suite with output shown, then the two reviewers over the polish range; polish review BLOCKER: revert the polish, move those items to `[debt]`; polish-review NOTEs sort into `[debt]` or `[note]` only; `[debt]` lines lose the bucket tag and keep the priority, then `acta debt new`; `[note]` lines go to `## Review notes` in the plan, committed; then `acta:land`. A polish sent to another agent uses `acta dispatch init --round polish`; when `/acta:review` arrives with round polish, run step 3 onward. (2) One round rule, stated once: "Rounds count only fix rounds. Every commit gets the two reviewers." (3) Budget steps 1-4, the "real user" rule, "a round with no BLOCKER does not start", the small-change exception and the opening line drop their polish exceptions and point to the section. (4) dispatch.md steps 10-11, the Fix rounds polish line, the Stopping rule, and house-rules.md NOTES point to the section instead of restating it. (5) Keep the bucket rules and BLOCKER fix-round flow; move, do not duplicate.
- [x] Commit: `Review: one After a CLEAN round section, everything else points there`

## Polish 2

Round 3 came back CLEAN on both axes. These `[fix]` NOTEs land as one polish commit. The polish counts as no round. The orchestrator runs the full suite, then the two reviewers review the polish range.

### Task 6: Polish from the round 3 `[fix]` NOTEs

**Files:** `plugin/skills/review/SKILL.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/build_dispatch_test.go`

**verify:** Every rule in `## After a CLEAN round` and its pointers is stated once, and a polish reply-back runs the full suite and the two reviewers before step 3. List every sentence checked. Reverting any change turns a plugincheck test red.

- [x] Failing test: pins for each change below; MustNot for "run step 3 onward" without the reviewers and for "Rounds count only fix rounds".
- [x] Change: (1) SKILL.md polish reply-back line: one sentence only, "A polish sent to another agent goes out with `acta dispatch init --round polish`; when `/acta:review` arrives with round polish, run the full test suite with the output shown, the two reviewers over the polish range, then step 3 onward." (2) Round rule: "The polish commit counts as no round. Every commit gets the two reviewers." (3) Step 1 of the section points to the bucket rules in "Where findings go" instead of restating them. (4) "`[fix]` NOTEs join the fix task" stays in budget step 2 only; drop the copy in "The flow". (5) dispatch.md steps 10-11: one land only; step 11 says the section's `acta:land` is build's `## Close`, step 10 drops its land tail. (6) Finish the cut-off comment in `build_dispatch_test.go` near line 65.
- [x] Commit: `Review polish 2: one statement per rule, polish reply-back runs the reviewers`

## Review notes

- `SplitPriority` matches only the lowercase tag with one space after it; `(High) x` keeps the tag as text. The skill and template show the exact form.
- The order check in `skill_review_test.go` uses the first `strings.Index` of each phrase; a later repeat of "bucket tag" could make it pass or fail for the wrong reason.
- `plugin/skills/review/SKILL.md` says the polish counts as no round twice in `## After a CLEAN round`; both lines are pinned.
- `plugin/skills/build/dispatch.md` step 11, "its `acta:land` is build's `## Close`", reads oddly: build's Close leads to land through review.
- "The flow." in `## Where findings go` now holds one bullet; it could be one sentence.
- `plugin/skills/build/dispatch.md` memory sweep says "once the round is CLEAN", which is already past when it runs. This was there before this plan.
- "When the polish or the tests fail" does not say what a failed polish is.
- The dispatch.md Stopping-rule rewording is not pinned; its meaning did not change.
- Task 4 text still says the orchestrator reviews the polish itself; the user overruled that (the polish gets the two reviewers), and the spec records the change.
