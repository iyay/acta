---
id: SPC-0190
created: "2026-11-07 16:57:24"
hash: bzyfazw
started: "2026-11-07 17:09:11"
finished: "2026-11-07 18:07:25"
---
# Review findings pipeline

## Problem

The Claude Code Review docs (code.claude.com/docs/en/code-review and its sub-pages) run a review in four steps: find, verify, rank with dedupe, post. They mark bugs the change did not bring in as their own severity, read repo review rules per changed path, and let a later review see the replies to earlier findings. Set against `acta:review`, five gaps show:

- A reviewer that finds a bug already on the parent has no place to put it in its report, so it can land as a BLOCKER and become a fix task.
- Repo gotchas in `.acta/wiki/` (like `review-base-merge-base` or `reviewer-temp-clone`) never reach the reviewers.
- The two reviewers often report the same BLOCKER, and nothing merges them before the fact-check and the fix task.
- Rounds 2 and 3 get the earlier findings without what happened to them, so a finding the orchestrator pushed back on can come back.
- NOTE lists have no cap, and the review-notes line holds no counts that later stats (SCR-0085) could read.

## Goal

Each finding reaches the right home once, with the repo's own rules in front of the reviewer.

## Out of scope

- No new review rules file like `REVIEW.md`. The wiki already maps pages to paths.
- No fact-check at `standard`; it stays at `deep`.
- No PR comments, no merge gate change, no new CLI command.
- No new plugin eval. The user can add one later.

## Design

In `plugin/skills/review/code-reviewer.md`:
1. **Repo rules.** For each changed file, the reviewer runs `acta wiki match <file>` and reads each page it names as review rules for that file. A change must not weaken its own review, so every wiki page the diff changes or deletes is read at `BASE_SHA` with `git show`. That BASE copy applies when its own `paths:` cover a changed file, even when the HEAD copy no longer matches or is gone. A page the diff adds is not a rule yet.
2. **Pre-existing section.** A new `### Pre-existing` output section: a defect in code the diff did not bring in, with `file:line` and how to see it. It is never a BLOCKER.
3. **NOTE cap.** At most 10 NOTEs, ranked by what leaving each one costs later, highest first, then one line `N more NOTEs not listed` when some were cut.
4. **Earlier findings with verdicts.** `[EARLIER_FINDINGS]` lists each earlier finding with one verdict: `fixed`, `rejected: <reason>`, `moved to NOTE` (by the fact-check), `NOTE open` (not sorted yet), or `sent to bug`. The reviewer does not raise a rejected or moved one again unless the diff gives new evidence.

In `plugin/skills/review/SKILL.md`:
5. **Dedupe.** Before the fact-check and before the fix task, the orchestrator merges BLOCKERs, NOTEs and Pre-existing items from the two reviewers that name the same defect, and keeps the one with the stronger proof.
6. **Pre-existing routing.** Items in `### Pre-existing` go through `receiving.md`, then to `acta:bug`, run in the worktree, so the bug file commits on the branch and lands with it. This runs in both paths: before the fix task in a BLOCKED round, and as a new first step of `## After a CLEAN round`. They never widen the plan.
7. **Verdicts kept.** After a BLOCKED round the orchestrator writes every finding that is not a fix-task line, each with its verdict, with `acta plan add <plan id> review-notes`. Rounds 2 and 3 pass those lines, plus the fix-task lines as `fixed`, as `[EARLIER_FINDINGS]`. The rule lives in `depth.md` `## Rounds`, and `SKILL.md` points to it. The plan keeps the lines, so a round in a later session or through `acta dispatch` still has them. The State is not used: a fix-round worker rewrites its own `next` part, and a State part holds only 10 lines. (Changed after review round 1 found both problems.)

The added text in `SKILL.md` is paid for by tightening its existing text, so the file stays at or under its 9991-byte cap.

In `plugin/skills/review/depth.md`:
8. **Counts.** The review-notes line also gives totals over rounds 1 to 3, after dedupe: `BLOCKER n, NOTE n, pre-existing n`. A light inline review counts as round 1. The polish review is not counted. A BLOCKER the fact-check moved counts as a NOTE. `NOTE n` includes the NOTEs cut by the cap. Cut NOTEs are dropped on purpose: they rank lowest, and only their number is kept.

### Budget

`code-reviewer.md` cap in `internal/plugincheck/budget_test.go` rises from 4042 to fit the new text, about 600 bytes. `depth.md` cap rises only if the counts line does not fit in its 85 bytes of slack. `SKILL.md` cap stays at 9991.

### Testing

- Red first in `internal/plugincheck/skill_review_test.go`: required phrases for each design item (the `### Pre-existing` section, `acta wiki match`, the `BASE_SHA` read of a changed or deleted page, the NOTE cap, the five verdicts, `acta state set` for verdicts, dedupe before the fact-check, the bug step in `## After a CLEAN round`, the counts line).
- The budget test passes with the new caps.
- `scripts/test` green; existing review evals unchanged.
