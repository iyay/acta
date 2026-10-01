---
id: BUG-0001
hash: hqd7ylj
fixed_in: 080bc63
finished: "2026-10-01 16:14:07"
---
# TestFixDuplicatesKeepsTheFileThatReachedTheBranchFirst fails now and then

## Symptom
`go test ./...` sometimes fails in `internal/write` with:

    --- FAIL: TestFixDuplicatesKeepsTheFileThatReachedTheBranchFirst (0.40s)
    testing.go:1617: TempDir RemoveAll cleanup: unlinkat .../001/.git/objects: ...

The test logic passes; only the temp folder cleanup fails.

## Root cause
Not proven yet. Likely a git process the test starts (or a background git gc / maintenance it triggers) still writes to `.git/objects` when `t.TempDir` removes the folder.

## Repro
Seen once in a full `go test ./...` on main right after merge 892edbd (2026-09-27). The same test then passed 6 times in a row with `go test -count=1 -run TestFixDuplicatesKeepsTheFileThatReachedTheBranchFirst ./internal/write/`. Full suite under load is the likely trigger.

## Found in
main, post-merge gate for PLAN-16 (892edbd). The test and `internal/write/ids.go` were not touched by that branch.
