---
parent: specs/2026-10-05-plan-review-status
depth: minimal
id: PLN-0095
created: "2026-10-05 21:15:31"
hash: hhm9v3s
started: "2026-10-05 21:25:45"
finished: "2026-10-05 22:07:21"
---
# Plan review status Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A plan whose tasks are all ticked but whose branch is not merged shows as `review`, in the board, the CLI and the TUI.

**Spec:** `.acta/specs/2026-10-05-plan-review-status.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The status word is exactly `review`. It is not a closed status.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3

### Task 1: board status

**Files:** Modify `internal/board/board.go` and its tests; any CLI test that pins a plan status list.
**verify:** For every plan (all ticked in a worktree, all ticked in the main tree, some unticked in a worktree, no tasks) the derived status is exactly review, done, in-progress, and the status it has today; a spec or bug whose only plan is in review is never shown as finished; `review` is never treated as closed. List each case checked.
- [x] Failing test: a board with a plan whose tasks are all ticked, loaded from another worktree, must report `review`; it reports `done` today.
- [x] Code: where the plan status is derived, turn `done` into `review` when the item's Worktree is set; parent status counts review as not finished; `review` stays out of Closed.
- [x] Commit: `board: an unmerged plan with every task ticked is in review`.

### Task 2: TUI color and detail line

**Files:** Modify the TUI files that map a status to a color and draw the detail header (`internal/tui/styles.go`, `internal/tui/view.go` or `internal/tui/detail.go`, whichever holds them), `internal/theme` if status colors live there, and their tests.
**verify:** Every place the TUI colors a status gives `review` its own color, distinct from in-progress and done, in every theme; the detail of a plan in review shows its branch and last review round (or none); no other status changes color. List each place and theme checked.
- [x] Failing test: a rendered frame with a plan in review shows the review color on its row and a branch line in the detail; fails because review has no color today.
- [x] Code: add the review color to the status map and the detail line, reusing the round lookup from `internal/board` that `acta state` uses.
- [x] Commit: `tui: plans in review get their own color and a branch line`.

### Task 3: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.

## State

### Next

Task 3 starts from:
- review yellow (slot 3) on rows, workLine, dot ◐; detail WORKTREE + ROUND via board.PlanStates
- commit 016eb88 on review-status; tests internal/tui/review_test.go green
- no board logic touched; theme/* untouched (slot reuse, no new hex)
- next: bump patch in 3 plugin json files, run internal/plugincheck
  verify: version files agree, plugincheck green

## Polish

### Task 4: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] Put back the two comments the diff deleted with no task behind it: the comment above the dot constants in internal/tui/detail.go (near line 14) and "Each role always takes the same slot..." in internal/tui/styles.go (near line 31), as they read at 97634b1.
- [x] Commit: `polish: review notes for PLN-0095`

## Review notes

- slotReview = 3 sits out of number order with the other slots.
- The plan's own acta show inside the worktree reads done, since a worktree reads itself as the main tree; review shows from the main checkout.
