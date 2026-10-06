---
id: DBT-0061
hash: ru37fuj
parent: plans/2026-10-01-fixed-in-closes-bug
---
# Review NOTEs: Fixed In Closes Bug Implementation Plan

- [ ] (low) No test covers a spec whose closes list names a bug with fixed_in turning done; the closedByStatus path is covered only indirectly.
- [ ] (low) The board comment says fixed_in is written only after a merge, but internal/write/ops.go only checks the value looks like a sha, so an early fixed_in shows the bug fixed too soon.
- [ ] (low) Two of the six table cases (finished child plan, with and without fixed_in) pass before the change too; they guard against regression, not prove the new case.
