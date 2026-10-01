---
parent: bugs/2026-10-01-fixed-in-does-not-close-bug
id: SPC-0059
created: "2026-10-01 14:21:04"
hash: ohv7knd
started: "2026-10-01 14:27:21"
finished: "2026-10-01 14:33:03"
---
Status: approved by the user on 2026-10-01 (Bounded). Fixes bug `.acta/bugs/2026-10-01-fixed-in-does-not-close-bug.md`.

# A bug with fixed_in is fixed

## Problem
A bug's status comes only from plans whose parent is that bug. A bug fixed by a plan whose parent is something else gets only `fixed_in` written on it, and `fixed_in` does not change its status. On main, BUG-0015, BUG-0020 and BUG-0002 show `open` even though their fixes have landed.

## Rule
A bug with no written `status` and a non-empty `fixed_in` has status `fixed`, source `derived`.

- `fixed_in` wins over child plans. It is written only after a merge (`acta:land` step 8), so it is the stronger proof. A child plan that is not fully ticked does not pull the bug back to `fixing` or `open`.
- A written status (for example `wontfix`) still wins over `fixed_in`.
- The sha is not checked against git. The board does not call git to find a status, and this rule keeps it that way.

## Change
`internal/board/board.go`, the status switch in the derive step: one new case before the `it.fmStatus == ""` case, for `it.Kind == KindBug && it.fmStatus == "" && it.FixedIn != ""`, which sets `fixed` and `derived`. No new helper.

Specs that a bug closes follow for free: `closedByStatus` reads the bug's status.

## Tests
New tests in `internal/board/`:
1. A bug with `fixed_in` and no plan is `fixed`, `derived`.
2. A bug with `fixed_in` and a child plan with an open task is `fixed`.
3. A bug with `fixed_in` and `status: wontfix` is `wontfix`, `frontmatter`.

## Out of scope
- Checking that the `fixed_in` sha exists.
- Skill text changes. `acta:bug` and `acta:land` already say to write `fixed_in`.
