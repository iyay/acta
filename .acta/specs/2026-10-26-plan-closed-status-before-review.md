---
parent: bugs/2026-10-13-derived-review-hides-done-plan
id: SPC-0180
created: "2026-10-26 11:36:53"
hash: trxp59z
---
# Keep a plan's written closed status before deriving review

## Scope

Bounded fix for BUG-0044, approved in WhatsApp on 2026-10-26.

A plan read from another branch can have every task done and an explicit
`status: done`, yet the board reports `review (derived)`. The same review
marker also makes the parent spec or bug count the completed tasks as work
still in progress.

## Design

In `internal/board/board.go`, omit plans with a written closed status from
the review map in `Board.derive`. Use the existing `Closed` helper. Valid
closed plan statuses are `done` and `dropped`. The normal status switch
then keeps that written status and reports `frontmatter` as its source.

Apply this before review affects either a plan or its parent's progress.
A plan's written closed status must survive whether it is loaded from the
current checkout, another worktree, or another branch. Plans with all tasks
done and no written closed status still enter review outside the current
checkout. Existing task and dropped-plan progress rules stay unchanged.

## Verification

Use `LoadTrees` in `internal/board/review_test.go` to cover written `done`
and `dropped` plans on another tree, including plan status, status source,
and parent progress. Keep existing tests for unclosed all-done plans in
review and completed plans in the current checkout.

The debug probe reproduced the bug with the current code:
`written=done actual=review source=derived branch=other-parent`.

Run the focused board tests, then `scripts/test` and `scripts/test --full`.
No new dependency, public API, file format, or status value is needed.
