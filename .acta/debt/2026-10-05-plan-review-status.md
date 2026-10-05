---
id: DBT-0080
hash: w0rrj27
parent: plans/2026-10-05-plan-review-status
---
# Review NOTEs: Plan review status Implementation Plan

- [ ] (low) No test pins that a merged branch whose worktree is still there shows its plan as done; LoadTrees keeps the main copy on a tick tie, and a change there would quietly flip it to review.
- [ ] (low) No CLI test checks that acta show and acta list print `review`.
- [ ] (low) The review case in internal/board/board.go comes before the written frontmatter status, so a plan marked dropped in a worktree with every box ticked shows review; the spec does not say which wins.
- [ ] (low) The dot comment at internal/tui/detail.go:14-16 says finished work wears nothing, but done dots are green and the new review dot has its own color.
