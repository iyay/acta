---
id: DBT-0088
hash: dmczu3l
parent: plans/2026-10-06-dispatch-checkpoint-verdict
---
# dispatch checkpoint verdict review notes

- [ ] (low) internal/cli/dispatch_herdr.go checkpoint: card rows run to the end of the pane text, not the end of the card, so a chatter line below the card like 'Starting Task 4' still counts as a seen id. Finding the card's end needs new logic.
- [ ] (low) internal/cli/dispatch_herdr.go checkpoint reads 60 lines 20 s after the goal; if omp prints more than that after the goal, the goal mark scrolls out and the verdict is unconfirmed (safe, never a false drift).
- [ ] (low) internal/cli/dispatch_herdr_test.go: the 'Todo 1 task' case has no rows, so it passes with the old 'tasks' regex too; add a case with header 'Todo 1 task' and both rows for ids 1,2 (drift with the fix, ok without).
