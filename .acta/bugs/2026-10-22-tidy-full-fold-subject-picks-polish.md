---
id: BUG-0055
hash: yaaxxne
started: "2026-10-22 23:55:45"
fixed_in: 86e991d67b18e2c486c917aeefa9af6e34ecc545
finished: "2026-10-23 00:07:26"
---
# Tidy names a planning-only branch after its polish commit

## Symptom
When every file a branch changes sits under the planning root and the branch has a polish commit, `acta tidy` folds the whole branch into one commit and gives it the polish subject. The real change subject is lost. PLN-0189 (a wiki-only branch) came out as `d0e4dec polish: review notes PLN-0189` and not as `docs(wiki): narrow plugin page paths so releases do not stale them`.

## Root cause
`internal/tidy/tidy.go:224-240`, `fullFoldSource`. The first loop wants the newest commit that touches files outside the planning root; a planning-only branch has none. The second loop (`tidy.go:233-237`) then takes the newest commit whose type is not `chore`. `polish` and `wiki` are fold types (`tidy.go:41`) too, but this loop does not skip them, so a late polish commit beats the earlier `docs(wiki)` commit that holds the real change.

## Repro
Branch from main with these commits, oldest first: `docs(wiki): change a page` (touches only `.acta/wiki/x.md`), `chore(plan): record evidence`, `polish: review notes PLN-x`, `chore(plan): tick polish`. Run `acta tidy main <branch>`. The single new commit is titled `polish: review notes PLN-x`; expected `docs(wiki): change a page`.

## Found in
Land of PLN-0189, worktree `acta-release-wiki-warning-noise`, while reading the tidy output before moving main.
