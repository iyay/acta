---
id: DBT-0148
hash: o531wz8
parent: plans/2026-10-31-flake-record
---
# Review NOTEs: Flake record implementation plan

- [ ] (medium) Bug-file writes (addFlake in internal/write/flake.go, and the older setBugPriority and markDate) take no lock and no tmp-then-rename, so acta flake record racing a TUI mark or priority change on the same bug can lose one write; a lock only in addFlake would not close it.
- [ ] (medium) An open flake bug means one with no fixed_in (spec Section 2), so a bug closed as dropped, or through a plan's closes: without fixed_in, still gets new ## Flakes lines instead of a new bug; changing it needs a spec change.
- [ ] (medium) The 60-character slug cap in internal/write/flake.go lets two long test names share one slug; the second first-record on the same day fails "already exists", and under the land rule that blocks the merge. A short hash of the full name or a counter on a clash fixes it.
- [ ] (medium) acta state set refuses to commit while the plan file holds uncommitted task ticks (exit 2, 'file had other uncommitted changes'), so in a subagent build every worker's state set fails until the orchestrator commits the plan; it should fold the ticks into the same planning commit.
