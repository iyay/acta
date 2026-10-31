---
id: DBT-0147
hash: vuhula3
parent: plans/2026-10-31-cleanup-skill
---
# Review NOTEs: Cleanup skill implementation plan

- [ ] (medium) No eval covers cleanup skill step 1: the cleanup-proposes-first scaffold has no plan or spec, so acta cleanup lists no --fix change, and a skill edit that runs --fix without showing it and waiting for a yes would pass every test. Add a case whose scaffold holds a fixable plan.
- [ ] (medium) TUI tests flake under the full suite: TestWheelNotchesBelongToTheScreenTheyWereGatheredOn (internal/tui/scroll_test.go:918) failed once at the PLN-0202 land and passed 5 of 5 alone, like the known click test. Each flake costs a full-suite rerun at land.
