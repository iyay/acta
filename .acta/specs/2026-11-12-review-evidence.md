---
id: SPC-0203
created: "2026-11-12 22:59:33"
hash: w48fwzl
started: "2026-11-13 00:04:27"
finished: "2026-11-13 00:27:14"
---
# Review verdicts carry their evidence

Bounded. Worked out with the user in the terminal on 2026-11-12. The user approved the design in chat.

## Problem

The review skill has strict rules for a BLOCKER and none for a CLEAN.

- A reviewer that reports CLEAN does not have to show which files it read or which commands it ran. The fact-check in the deep level can only move a BLOCKER down to a NOTE, so nothing checks a CLEAN that is wrong.
- The reviewer brief in `plugin/skills/review/code-reviewer.md` points at "the review skill" for the three questions, the finding bar and the bucket rules. The reviewer subagent has not read that skill.
- `plugin/skills/review/SKILL.md` takes the base as `git rev-parse <parent>`. The wiki page `review-base-merge-base.md` records that this shows the parent's newer commits as deletions when the parent moved.
- A reviewer lists five BLOCKERs at most and says nothing about the ones it cut. With three rounds at most, a cut BLOCKER can reach land.
- `plugin/skills/review/receiving.md` spends most of its lines on which words to avoid in a reply. It does not tell the orchestrator to run a BLOCKER's repro before the BLOCKER becomes a fix task. The third review question, whether the tests go red when the change is reverted, names nobody who runs that check.
- `plugin/evals/` has five review cases. They cover polish, the depth floor and the understanding report. None checks that a reviewer catches a real defect, that a nit stays a NOTE, or that a wrong BLOCKER is rejected.

## Changes

1. Eval cases first. Four new cases in `plugin/evals/`, written and run against today's skill text before any review text changes, so each has a baseline:
   - `review-planted-bug`: the diff has a seeded defect on a path the plan covers. The verdict must be BLOCKED and name the defect's `file:line`.
   - `review-nit-stays-note`: the diff has only naming and comment problems. The verdict must be CLEAN and no nit is a BLOCKER.
   - `review-false-blocker`: the diff is correct, and the scaffold hands the orchestrator a reviewer report whose BLOCKER is wrong. The orchestrator must run the repro, reject the BLOCKER with the output, and open no fix task.
   - `review-clean-has-evidence`: a CLEAN report must hold the `### Coverage` section of change 3.
2. A brief that stands alone. `code-reviewer.md` states the three questions, the finding bar and the bucket rules in its own text, with no pointer to the review skill. The opening persona line goes.
3. Evidence behind every verdict. The reviewer report gets a `### Coverage` section with three parts: each plan requirement with the `file:line` that meets it and the test that checks it; each changed file, marked as read; each command the reviewer ran with the output line that decides it. The orchestrator sends a report that lacks this section back to the same reviewer once. It does not dispatch a new reviewer for it. A second report without the section counts as BLOCKED for that axis, and the orchestrator tells the user.
4. The orchestrator proves each BLOCKER. `receiving.md` is rewritten around one rule: before a BLOCKER becomes a fix-task line, the orchestrator runs its repro in a temp clone and keeps the output. A BLOCKER whose repro does not show the wrong result gets the verdict `rejected: <reason>` with that output. The lines about gratitude and wording go. The lines about unclear feedback and about pushing back with technical reasons stay.
5. An owner for the revert check. The Standards-axis reviewer runs the third question: in a temp clone it reverts the change, runs the narrow tests, and shows that they go red. The brief says so, and the output goes under `### Coverage`.
6. Base from the merge base. `SKILL.md` takes `BASE_SHA=$(git merge-base <parent> HEAD)`.
7. No silent cut of BLOCKERs. The cap of five ranked BLOCKERs stays. A reviewer that cut some adds the line `N more BLOCKERs not listed`. The round is BLOCKED, and the next round's brief tells the reviewers to look for them.

## Rule for keeping a text change

Each change to text a reviewer or the orchestrator acts on is judged on the eval cases of change 1, 10 runs per arm, old text against new text. The cases are measured once for the whole plan first; a single change is measured on its own only when the whole scores below the old text on a case. A change that scores below the old text on any case is left out, and the plan's `## Review notes` say which. This is the rule the wiki page `skill-escape-lines-skip-proof.md` records.

## Files

- `plugin/skills/review/SKILL.md`, `code-reviewer.md`, `receiving.md`, `depth.md`
- Four new case folders in `plugin/evals/`
- `internal/plugincheck/skill_review_test.go` for the required strings, and `internal/plugincheck/budget_test.go` for the byte caps. `SKILL.md` is at 9.7K of a 9991 byte cap, so the cap goes up. No words an agent acts on are trimmed to fit.
- Wiki pages that the change makes wrong, found by `acta wiki check` at the close of the build.

## Testing

- The four eval cases, 10 runs per arm. That is 80 runs when no case gets worse, about 15 minutes at `scripts/eval -j 8`, and it costs quota.
- `internal/plugincheck` keeps the required strings and the size caps.

## Order

The plan `2026-11-12-review-clones-and-push-on-ask` (PLN-0231) is being built in another session and changes `review/SKILL.md` and `code-reviewer.md`. The build of this spec starts after that plan and PLN-0236 (SPC-0207, scoped land gates) land.

## Out of scope

A new reviewer seat, a fact-check that can raise findings, the number of rounds, the review tiers, and the debug skill. The debug skill gets its own spec.
