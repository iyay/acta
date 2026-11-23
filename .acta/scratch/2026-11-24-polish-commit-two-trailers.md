---
id: SCR-0100
hash: s93aziz
title: Polish commit carries the Polish task trailer too
status: raw
created: "2026-11-24 05:23:17"
schema: "1"
---
# Polish commit carries the Polish task trailer too

## Words

### 2026-11-24

Polish commits carry the trailer of the task they fix, but the land final plan check looks for the Polish task own number, so a plan with a Polish task fails the check (PLN-0251 land, 2026-11-24). User chose: a polish commit carries both trailers, Task: PLN-<hash>#<fixed task> and Task: PLN-<hash>#<polish task>. Fix the review skill text (After a CLEAN round) and the build Task loop line on trailers.

## Context

## Log

## Open questions
