---
parent: scratch/2026-10-15-tdd-test-count-cap
id: SPC-0134
created: "2026-10-15 09:56:04"
hash: d9cgg48
---
# Lean tests: seams, behavior cases, table tests

Status: Bounded, approved by the user in chat on 2026-10-15.

## Problem

`acta:tdd` already limits tests to the plan's cases. Two gaps stay open: nothing stops a plan from naming many cases that run the same code path, and nothing folds such cases into one table test. The user also asked to merge ideas from an outside tdd skill: test at seams (public interfaces), not internals; one test then one implementation, never all tests first; test names in the project's domain words. Expected values from an independent source and no tautological tests are already in `writing-good-tests.md`.

## Design

- `plugin/skills/slice/SKILL.md`: every task names its seam, the public interface its tests go through. The user's yes to the plan is the yes to the seams; tdd asks no second time. A verify line names a case only when it needs different behavior from the code. Inputs that run the same path are one case.
- `plugin/skills/tdd/SKILL.md`:
  - Tests go through the seam the plan names, never through private functions or internal state.
  - Cases that share one code path and differ only by input go in one table-driven test, one row each.
  - One test, then the code for it, then the next test. Never write all tests first.
- `plugin/skills/tdd/writing-good-tests.md`: test names use the words in `.acta/wiki/glossary.md` when it exists.
- Not taken from the outside skill: a separate seam gate inside tdd, and moving refactor out of the loop. Refactor stays in red-green-refactor.
- Wording stays general, fit for every user of the plugin.

## Testing

Add the new rule phrases to `Must` in `internal/plugincheck/skill_tdd_test.go` and to the slice skill check, red first. This is the repo's existing pattern for required skill rules. The tdd folder stays under its 560-line cap.
