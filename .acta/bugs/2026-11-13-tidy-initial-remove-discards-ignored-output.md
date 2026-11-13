---
id: BUG-0092
hash: rz3b2lg
---
# Initial temporary-worktree removal discards ignored outputs

## Symptom

A temporary gate worktree containing only ignored output files can be removed without approval, silently discarding those files.

## Root cause

plugin/skills/land/tidy.md:9 attempts plain git worktree remove before the cleanup procedure's ignored-file guard. Git accepts an otherwise clean worktree with ignored-only files. The same unconditional first removal is present at parent b3cdaa408a302685e0ce3f8900edc56e35c24f48; the reviewed change guards the retry, not this inherited first removal.

## Repro

In an isolated clone with fresh HOME and TMPDIR, add report.cache to .git/info/exclude, create a detached worktree at b3cdaa408a302685e0ce3f8900edc56e35c24f48, and write a synthetic report.cache. git status --short --ignored prints !! report.cache. Plain git worktree remove exits 0 and the file no longer exists. No force was used. The synthetic worktree was removed; the real checkouts were untouched.

## Found in

PLN-0242 Standards review round 1, confirmed by the orchestrator in an isolated clone. This is a pre-existing defect, outside the reviewed policy-change fix. Repro retained under the session temp directory acta-ignored-remove-proof-20261113/repo.
