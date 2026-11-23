---
id: BUG-0103
hash: jybdqm8
---
# Tick lock files pile up in the user cache folder

## Symptom
Every `acta tick` on a plan path never seen before leaves one empty lock file under `~/Library/Caches/acta/locks` (or the platform cache folder). Nothing removes them. `acta tidy --plan` ticks debt in a new temp folder each run, so it adds one file per land.

## Root cause
`internal/write/tick.go:210-240` (`lockDir`, `lockPath`) names the lock from the plan's real path and never deletes it after the tick.

## Repro
Run `acta tidy --plan <hash>` for a plan with a debt item in `closes:` a few times, then list the cache lock folder: one new file per run.

## Found in
Branch land-bookkeeping-in-tidy, PLN-0250 review round 1 (Standards axis, Pre-existing).
