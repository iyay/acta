---
parent: bugs/2026-10-19-debug-race-recipe-skipped
id: SPC-0160
created: "2026-10-19 11:54:58"
hash: fv7de21
---
# Debug recipes gate Phase 3

## Problem
The debug skill holds a Race recipe that asks for a forced interleaving probe
and a regression test that holds the workers the same way. In the debug-race
eval the agent skips it: it proves the cause with repeat runs and a lock
control, proposes a test that relies on thread volume, and reports. The case
scores 0.00 on main, and also on the skill before b986b0a.

## Approved design
Bounded, approved by the user in chat on 2026-10-19.

- Phase 1 names the matching recipe early: when the symptom fits one of the
  recipes (flaky, race, order-dependent, slow), pick it then.
- The Phase 3 exit adds one condition: when a recipe matches, Phase 3 does not
  exit until the evidence that recipe names was produced by a command that ran.
  For a race that is the forced-interleaving probe.
- Keep `plugin/skills/debug/SKILL.md` within its current byte cap by tightening
  other sentences; do not raise the cap.
- No Go change and no grader change.

## Acceptance
- Eval `debug-race` passes on 2 of 2 runs.
- Eval `debug-performance`, `debug-order-dependent` and `debug-exact-symptom`
  still pass.
- `scripts/test ./internal/plugincheck/` passes.
