---
parent: bugs/2026-10-19-debug-race-recipe-skipped
id: SPC-0160
created: "2026-10-19 11:54:58"
hash: fv7de21
started: "2026-10-19 12:02:38"
finished: "2026-10-19 12:51:02"
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
- Phase 1 builds its loop in a temp copy, so a user command that writes files
  never runs in the repo.

## Revision 2026-10-19: copyable probe and artifact grader
Three text-only fixes left debug-race at 0/2, 1/2, 0/2: the run made no
barrier probe. Approved by the user in chat (C2 + C1):

- Add `plugin/skills/debug/race.md`: a short, copyable probe that patches the
  read to wait on a two-party barrier, so both workers read before either
  writes, runs it, and prints the lost count; plus the matching regression
  test shape. The Race recipe in SKILL.md points to it. It gets its own byte
  cap; the SKILL.md cap stays.
- Add a deterministic grader to debug-race: a Bash call whose input contains
  `Barrier` must exist. The LLM judge stays. Debug cases may then mix one
  deterministic grader with their judge; the plugincheck rule changes only
  that far.

## Acceptance
- Eval `debug-race` passes on at least 2 of 3 runs, including the new Barrier
  grader (lowered from 2 of 2 by the user on 2026-10-19; a failing run is one
  that never reads `race.md`, and that gap is filed as debt).
- Eval `debug-performance`, `debug-order-dependent` and `debug-exact-symptom`
  still pass.
- `scripts/test ./internal/plugincheck/` passes.
