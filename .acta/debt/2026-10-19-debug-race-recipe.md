---
id: DBT-0131
hash: asyb8e4
parent: plans/2026-10-19-debug-race-recipe
---
# Review NOTEs: Debug recipes gate Phase 3 implementation plan

- [ ] (medium) debug-race still fails about 1 run in 3: the run never reads plugin/skills/debug/race.md and makes no barrier probe, so the land eval gate can go red on it.
