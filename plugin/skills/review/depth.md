# Review depth

This file holds the level rules. `SKILL.md` links here.

## Pick the level

Pick it in this order, and ask no question when one answers:
1. The argument of `/acta:review quick|standard|deep`. It counts for this run only and is never saved.
2. `review_depth` from `acta config show`.
3. `standard`.

## The levels

- **quick:** a diff that is all light tier gets the inline light review (`## Review tiers`). Any other diff gets ONE reviewer subagent, from [code-reviewer.md](code-reviewer.md). It answers both the Spec axis and the Standards axis in one report.
- **standard:** the two reviewers in parallel, as `SKILL.md` says. A light-tier diff keeps the inline light review.
- **deep:** the two reviewers, then one read-only fact-check subagent over the BLOCKERs only. A light-tier diff keeps the inline light review. The orchestrator numbers the BLOCKERs before the fact-check. It uses [fact-check.md](fact-check.md). With no BLOCKER it does not run. A moved one becomes a NOTE, sorted into a bucket like any other.

## The floors

The setting can raise the depth, but not drop it below these. No level, from the argument or the setting, goes below them:
- A diff touching a security, auth, data migration, money or delete path runs at least `standard` with the deep lens.
- Fix rounds run at least `standard`.
- Work the `dispatch` executor wrote runs at least `standard` in the full tier.

## Rounds

Three rounds at most at every level. There is never a round 4. Rounds 2 and 3 hand the reviewers the earlier findings, so they look for others.

## Understanding

`review_understanding` is its own switch, false when missing. quick, standard, deep, the argument, a floor and the deep lens leave it as it is. When it is true, `acta:land` ends its report with a short explanation and asks no question. When it is false, there is none.

## Review notes

The orchestrator writes the level used, and any floor that raised it, into the one review line of the plan's `## Review notes`.
