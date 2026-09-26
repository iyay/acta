---
id: BUG-1
hash: l015
---
# pmb id dies on a file with broken frontmatter

## Symptom
pmb id exits 3 with `yaml: line 1: did not find expected ...` and gives no IDs to any file.

## Root cause
internal/write/ids.go:47, SetField on specs/2026-09-17-broken.md (`ref: [unclosed`) returns the yaml error and AssignIDs fails the whole run.

## Repro
Copy the board basic fixture to a git repo, run `pmb id`.

## Found in
Worktree short-ids, while writing task-4 tests (pmb id on fixtureRepo).
