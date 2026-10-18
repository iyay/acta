---
id: SPC-0145
created: "2026-10-18 19:07:05"
hash: osse8at
---
# Run the full suite once, at land

## Why

A plan with review polish runs the full suite twice in a row: once after
the polish commit (`acta:review`, After a CLEAN round) and again at
`acta:land` step 1. Between the two, only the plan file or review notes
change, so the second run tests the same code. Each full run holds the
machine-wide `acta run-one` lock and slows every other worktree.

Reusing the polish run was considered and dropped: it needs a clean
worktree before the run, a commit noted before and checked after, and
proof for each gate (type checks are a separate command in some repos).
Running the suite once, at the end, needs none of that.

## Design

One rule: the full suite runs only in `acta:land`, after every change is
in (tasks, fix rounds, polish, debt and review notes). It runs again only
after a fix for a red full run, or when the merged or tidy result needs
a new gate under the step 6 and tidy rules.

1. Polish runs narrow checks; land owns the full suite.
   `plugin/skills/review/SKILL.md`, polish steps (the CLEAN-round step 2
   and the `--round polish` paragraph): after the polish commit the
   orchestrator runs the tests of the code the polish touched and the
   formatter, with the output shown, never the full suite. "When the
   polish or the tests fail, revert that commit" now means those narrow
   tests.
2. `plugin/skills/land/SKILL.md` step 1 adds: when the full suite fails,
   find and fix the cause with the narrow failing tests, then run the
   full suite once more. The full suite is never a debugging loop. The
   rest of step 1 (eval gate, wiki gate, report and do not merge when a
   gate is red) stays as it is.
3. `CLAUDE.md` Commands section: "Agents run `scripts/test --full` only
   as the land gate (`acta:land`), never per task or per review round.
   It runs again after a fix for a red run, or when the merge result
   needs a new gate."
4. `internal/plugincheck/skill_review_test.go`, three checks:
   - the required polish sentence (CLEAN-round step 2) and the required
     `--round polish` reply-back sentence are replaced with the new ones;
   - the "step 3 onward" guard (each line that sends a polish reply-back
     to step 3 onward) now requires the narrow tests and `Review tiers`
     on that line, and fails when the line says "full test suite";
   - the review skill fails the test when any polish line asks for the
     full suite.
   `skill_land_test.go` requires the no-debugging-loop sentence and the
   run-again rule.

Unchanged: land step 6 and the tidy gates (reuse when only planning-root
files differ), the eval gate, the wiki gate, task and fix-round test
scope (already narrow).

Out of scope: `scripts/test --full` running both the whole repo and the
`notui` build of `internal/cli`.

## Testing

plugincheck string tests go red first, then the skill text turns them
green. Land runs `scripts/eval`, since skill files change.
