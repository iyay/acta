---
id: DBT-0059
hash: gqnbffb
parent: plans/2026-10-01-tui-priority-key
---
# Review NOTEs: TUI Key p Sets Priority Implementation Plan

- [x] The ? close this help line was dropped from helpLines, which the plan did not ask for, most likely to stay under the 24-line help cap; ? still closes help and the esc line still says so. (stale)
- [ ] (low) Priority: p sits before Edit, Copy id, New bug and Sort, so on a narrow screen those hints drop out sooner on bug and debt rows.
- [ ] (low) The new p help line is now the longest help line, so it sets the help popup width.
- [x] The test file adds TestPriorityKeyRefusesTheRowsWithNoItemOfTheirOwn, which the plan did not list; it covers group, worktree and legacy rows. (stale)
- [ ] (medium) acta tick left the started and finished stamps on the spec uncommitted, so the orchestrator had to commit them before land.
