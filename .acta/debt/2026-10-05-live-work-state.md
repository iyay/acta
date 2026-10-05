---
id: DBT-0074
hash: m2pcf2z
parent: plans/2026-10-05-live-work-state
---
# Review NOTEs: Live work state Implementation Plan

- [ ] (medium) A plan counts as running only once its branch has committed the plan file, so a build after tick --start but before any commit is missing from acta state and the SessionStart summary (internal/board/state.go planStates).
- [ ] (medium) planStates runs a git log and board.Load for every linked worktree on every SessionStart; session start gets slower as worktrees grow, with no time cap.
- [ ] (low) TestSetStateFallsBackToTheIdWhenThePlanHasNoShortID passes the same string as it.ID, so it stays green when the fallback is reverted; pass a hash or another id form (internal/write/state_test.go:135).
