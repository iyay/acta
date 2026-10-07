---
id: DBT-0046
hash: gtcs5lu
parent: plans/2026-09-30-tui-tick-items
---
# Review NOTEs: Mark Tasks and Debt Lines Done from the TUI Implementation Plan

- [ ] (low) internal/write/mark.go: + on a task already done (or - on one already open) changes nothing, but CommitPaths still runs and the TUI status says "commit failed" although nothing was written.
- [ ] (low) internal/tui/view.go: the help does not name ctrl+c or the down/up arrow aliases that the key switch handles.
- [x] Commit 4dba25c adds a 30-row help test that the plan did not ask for; it is test-only and changes no behaviour. (stale)
- [ ] (medium) internal/write/mark.go: no committed test covers the dirty-spec guard (a dirty spec file on + of a task); removing the spec append keeps the suite green.
- [ ] (low) internal/write/mark.go: when TaskDates fails after Tick, the plan stays ticked and uncommitted and the TUI shows only the error.
- [ ] (medium) internal/gitc: CommitPaths leaves files staged if a pre-commit hook rejects the commit (same old gap as gitc.Commit).
- [ ] (low) internal/write/mark.go: the date write in TaskDates and markDate runs outside the tick lock and writes in place, not through a temp file (same gap as the old cli.markDate).
- [x] internal/write/mark.go: TaskDates now returns early when the plan is missing; the old markTaskDates would have crashed there. (stale)
- [ ] (low) internal/tui/model.go: markRow runs the write and the git commit inside Update, so the UI waits on git (same as the status popup).
- [ ] (low) internal/cli/tick.go: --undo has no worktree refusal, the same as the other CLI tick actions.
- [x] internal/write/mark.go: the comment on unique says what, not why.
