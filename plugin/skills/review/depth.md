# Review depth

This file holds the level rules. `SKILL.md` links here.

## Pick the level

Pick it in this order, and ask no question when one answers:
1. The argument of `/acta:review quick|standard|deep`. It counts for this run only and is never saved.
2. `review_depth` from `acta config show`.
3. `standard`.

## The levels

- **quick:** a diff that is all light tier gets the inline light review (`## Review tiers`). Any other diff gets ONE reviewer subagent, from [code-reviewer.md](code-reviewer.md). It answers both the Spec axis and the Standards axis in one report.
- **standard:** the two reviewers in parallel, as `SKILL.md` says.
- **deep:** the two reviewers, then one read-only fact-check subagent over the BLOCKERs only. It runs on the reviewer model and uses [fact-check.md](fact-check.md). With no BLOCKER it does not run. Every id it moves becomes a NOTE.

## The floors

The setting can raise the depth, never lower it below these:
- A diff touching a security, auth, data migration, money or delete path runs at least `standard` with the deep lens.
- Fix rounds run at least `standard`.
- Work the `dispatch` executor wrote runs at least `standard` in the full tier.

## Rounds

Three rounds at most at every level. There is never a round 4. Rounds 2 and 3 hand the reviewers the earlier findings, so they look for others.

## Review notes

The orchestrator writes the level used, and any floor that raised it, into the one review line of the plan's `## Review notes`.
