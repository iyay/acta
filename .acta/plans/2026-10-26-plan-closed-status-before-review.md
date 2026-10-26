---
parent: bugs/2026-10-13-derived-review-hides-done-plan
depth: minimal
id: PLN-0197
created: "2026-10-26 11:38:08"
hash: v7pe7gq
started: "2026-10-26 11:39:08"
finished: "2026-10-26 11:40:18"
---
# Preserve closed plan status before review

**Goal:** Keep explicit closed plan statuses and parent progress correct when a plan is read from another tree.
**Spec:** `.acta/specs/2026-10-26-plan-closed-status-before-review.md`
**Tests:** fast `scripts/test`; full `scripts/test --full`.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Use the existing `Closed` helper; add no dependency, API, file format, or status value.
Keep task progress and dropped-plan progress rules unchanged.

## Waves

Wave 1: Task 1.

### Task 1: Keep written closed statuses out of derived review

**Files:** Modify `internal/board/board.go`; test `internal/board/review_test.go`.
**Seam:** `board.LoadTrees`, the board loader shared by CLI and TUI.
**verify:** Explicit `done` and `dropped` plans retain their written status and frontmatter source across the current checkout and another tree; their parent spec or bug receives normal task progress without the review marker; unclosed all-done plans on another tree remain review and merged all-done plans remain done, using the existing controls.
- [x] Red: Add table-driven closed-plan regressions through `LoadTrees` and run `scripts/test ./internal/board -run 'Test.*(ClosedPlan|AllTickedPlan|BugWithOnlyPlanInReview)' -v`; another-tree cases fail because the review map overrides written closed statuses and parent progress.
- [x] Change: Exclude plans whose frontmatter status is `Closed` when building the review map in `Board.derive`, then rerun the focused tests and `scripts/test ./internal/board`.
- [x] Commit: `fix(board): preserve closed plan status before review`.

## Review notes

- Deep review: Spec CLEAN, Standards CLEAN; no findings. Both reviewers independently proved the regression fails with the fix reverted. Focused board tests and go vet ./... pass.

## State

### Next

Task 1 is committed; run the closing review, then landing gates.
