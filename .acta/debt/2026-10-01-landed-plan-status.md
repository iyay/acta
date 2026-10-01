---
id: DBT-0063
hash: ve83rxr
parent: plans/2026-10-01-landed-plan-status
---
# Review NOTEs: A plan's written approved or in-progress follows its task boxes Implementation Plan

- [ ] (low) Setting a plan to in-progress from the TUI status popup or acta set, with no task started, still shows approved and nothing on screen says why (internal/tui/model.go status popup).
- [ ] (low) A plan with zero tasks written in-progress now shows approved; before it showed in-progress.
- [ ] TestPlanWrittenStatusFollowsBoxes has no case for a written done plan or for a plan with zero tasks.
