---
id: BUG-2
hash: jrky
---
# pmb id flip-flops the number of a type-bug file in specs/

## Symptom
First `pmb id` writes `id: BUG-3` to specs/2026-09-15-really-bug.md (`type: bug` in the specs folder); the second run rewrites it to `id: SPEC-6`, prints a change line and commits again. `pmb id` is never idempotent on this file.

## Root cause
internal/board/board.go fileItem validates the `id` field against Prefix(dir kind) = SPEC, so run 2 reads BUG-3 as bad id and leaves ShortID empty while Hash stays SPEC-.... internal/write/ids.go prefixOf then derives the prefix from the hash (SPEC) instead of the item kind (bug), and assigns a new SPEC number. The two disagree forever.

## Repro
Copy the board basic fixture to a git repo, drop the broken spec, run `pmb id` twice; the second run prints `SPEC-6 (specs/2026-09-15-really-bug.md)`.

## Found in
Worktree short-ids, while writing task-4 tests.
