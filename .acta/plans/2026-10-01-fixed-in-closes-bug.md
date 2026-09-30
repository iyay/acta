---
parent: specs/2026-10-01-fixed-in-closes-bug-design
depth: minimal
id: PLN-0068
created: "2026-10-01 14:24:09"
hash: hbk550d
started: "2026-10-01 14:27:21"
finished: "2026-10-01 14:33:03"
---
# Fixed In Closes Bug Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A bug with no written status and a non-empty `fixed_in` shows as `fixed`, whatever its child plans say.

**Spec:** `.acta/specs/2026-10-01-fixed-in-closes-bug-design.md`

**Tests:** fast `scripts/test ./internal/board`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- A written status always wins over `fixed_in`. The sha is never checked against git.
- Comments are plain English a 10-year-old can read. They say why, not what.

## Waves

- Wave 1: Task 1.

### Task 1: fixed_in makes a bug fixed

**Files:**
- Modify: `internal/board/board.go`
- Test: `internal/board/fixed_in_test.go` (new)

**verify:** A bug is `fixed` exactly when it has no written status and a non-empty `fixed_in`, on every path. List every case checked (fixed_in with no plan, fixed_in with a child plan that has an open task, fixed_in with a done child plan, fixed_in with `status: wontfix`, no fixed_in with no plan, no fixed_in with a done child plan) and the status and source each gives.

- [x] Failing test: a new table test loads a board for each case above and checks status and source; the fixed_in cases fail, because the status switch derives a bug's status only from its child plans and ignores `FixedIn`.
- [x] Code: in the status switch of the derive step, add one case before `it.fmStatus == ""` for `it.Kind == KindBug && it.fmStatus == "" && it.FixedIn != ""` that sets `fixed` and `derived`, with a short comment that says fixed_in is written only after a merge.
- [x] Commit: `bug with fixed_in shows as fixed (BUG fixed-in-does-not-close-bug)`
