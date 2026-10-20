---
id: DBT-0136
hash: g1vu153
parent: plans/2026-10-20-close-without-fast-suite
---
# Review NOTEs: Close builds without another fast suite

- [ ] (medium) The dispatch brief prints the plan whole **Tests:** line under GATES, fast command included (internal/cli/dispatch_brief.go), so a dispatched build can still run the broad suite before review. Decide whether the brief should name only each task focused tests.
