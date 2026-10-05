---
parent: scratch/2026-10-05-tui-review-indicator
id: SPC-0085
created: "2026-10-05 21:15:31"
hash: zpbhwwe
---
# A plan with every task ticked but not merged shows as review

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: while a plan's review, fix rounds or polish run, every task can be ticked and the board shows the plan as `done`, though nothing has landed. The user cannot tell a finished plan from one still under review.

Design:
- `internal/board`: a plan whose derived status is `done` and whose file comes from another worktree (the item's `Worktree` branch is set, so the branch is not merged yet) gets the derived status `review` instead. Once the branch lands and the worktree is gone, the plan comes from the main tree and shows `done` as today.
- A fix round or polish task that is not ticked turns the plan back to `in-progress` on its own, as today.
- `review` is not closed: the plan stays on the active list. A spec or bug whose plan is in review counts it as not finished, so the parent shows in progress.
- `internal/tui`: `review` gets its own status color in the list and detail, and the detail shows the branch and the last review round (the round `acta state` already reads).
- `acta show` and `acta list` print `review` as the status.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: a live "reviewers running now" signal.

Tests: a plan with all tasks ticked from a worktree is `review`; the same plan from the main tree is `done`; an unticked fix task gives `in-progress`; the parent spec of a plan in review is in progress; the TUI draws the review color and the branch line.
