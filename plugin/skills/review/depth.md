# Review depth

This file holds the level rules. `SKILL.md` links here.

## Pick the level

Pick it in this order, and ask no question when one answers:
1. The argument of `/acta:review quick|standard|deep|auto`. It counts for this run only and is never saved.
2. `review_depth` from `acta config show`.
3. `auto`.

## The levels

- **quick:** a diff that is all light tier gets the inline light review (`## Review tiers`). Any other diff gets ONE reviewer subagent, from [code-reviewer.md](code-reviewer.md). It answers both the Spec axis and the Standards axis in one report.
- **standard:** the two reviewers in parallel, as `SKILL.md` says. A light-tier diff keeps the inline light review.
- **deep:** the two reviewers, then one read-only fact-check subagent over the BLOCKERs only. A light-tier diff keeps the inline light review. The orchestrator numbers the BLOCKERs before the fact-check. It uses [fact-check.md](fact-check.md). With no BLOCKER it does not run. A moved one becomes a NOTE, sorted into a bucket like any other.
- **auto:** the review picks between two levels, for this diff. A light-tier diff keeps the inline light review.
  - Two reviewers, as in `standard`: a risk signal holds, or you are in doubt whether one holds. Risk signals are never up for judgment: a security, auth, money, data migration or delete path; text an agent reads as instructions (skills, prompts, hook text); work the `dispatch` executor wrote; a fix round.
  - One reviewer, as in `quick`: no risk signal holds.

  Risk decides, never size: a small diff on a risk path still gets two reviewers. Outside the light tier, `auto` never reviews with no reviewer subagent.

## The floors

The setting can raise the depth, but not drop it below these. No level, from the argument or the setting, goes below them:
- A diff touching a security, auth, data migration, money or delete path runs at least `standard` with the deep lens.
- Fix rounds run at least `standard`.
- Work the `dispatch` executor wrote runs at least `standard` in the full tier.

## Rounds

Three rounds at most at every level. There is never a round 4. Rounds 2 and 3 hand the reviewers the earlier findings, so they look for others. When a reviewer cut BLOCKERs in the round before, the brief says so and tells the reviewers to look for them. After a BLOCKED round, write every finding that is not a fix-task line, each with its verdict, with `acta plan add <plan id> review-notes`, run in the worktree. Rounds 2 and 3 read those lines back from the plan's `## Review notes` and pass them as `[EARLIER_FINDINGS]`, with the lines of every `## Fix round` marked `fixed`. The plan keeps them across sessions and hand-offs, and the CLEAN round sorts its open NOTEs from them.

## Understanding

`review_understanding` is its own switch, false when missing. The review asks nothing about it. quick, standard, deep, the argument, a floor and the deep lens leave it as it is. When it is true, `acta:land` ends its report with a short explanation and asks no question. When it is false, there is none.

## Review notes

The orchestrator writes the level used, and any floor that raised it, into the summary line of the plan's `## Review notes`. Under `auto` the line also names the level picked and the reason. The line also gives `BLOCKER n, NOTE n, pre-existing n`: totals over rounds 1 to 3, after the merge. A light inline review counts as round 1; the polish review does not count. A BLOCKER the fact-check moved counts as a NOTE. `NOTE n` includes NOTEs cut by the cap; cut NOTEs are dropped on purpose and only their number stays.
