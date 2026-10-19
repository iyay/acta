---
parent: specs/2026-10-19-debug-race-recipe-design
depth: minimal
id: PLN-0172
created: "2026-10-19 12:01:11"
hash: j0shu31
started: "2026-10-19 12:02:38"
---
# Debug recipes gate Phase 3 implementation plan

**Goal:** A debug run on a race produces the forced-interleaving probe and test the Race recipe asks for, so eval debug-race passes.
**Spec:** `.acta/specs/2026-10-19-debug-race-recipe-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full` (land only); task checks: `scripts/test ./internal/plugincheck/` and `scripts/eval --case <name>`.

## Global Constraints

- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- `plugin/skills/debug/SKILL.md` stays within its current byte cap in `internal/plugincheck/budget_test.go`; do not raise it. No Go change, no grader change.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2, after Task 1; both edit the debug skill.

### Task 1: Make the matching recipe a Phase 3 exit condition

**Files:** Modify `plugin/skills/debug/SKILL.md`.
**Seam:** The debug skill as run by the eval harness on cases debug-race, debug-performance, debug-order-dependent and debug-exact-symptom.
**verify:** Whenever a symptom fits a recipe, the run picks it in Phase 1 and does not report a cause until the evidence that recipe names came from a command it ran; for a race that is a forced-interleaving probe, and the proposed test holds the workers the same way. List the cases run and their scores; debug-race passes 2 of 2 and the other three still pass.
- [x] Red: `scripts/eval --case debug-race` on the unchanged skill fails (already seen 0.00 on main, 3 of 3).
- [ ] Change: add the Phase 1 recipe pick and the Phase 3 exit condition, tightening other sentences to stay in the byte cap; pass `scripts/test ./internal/plugincheck/`, then run debug-race twice and the other three cases once.
- [ ] Commit: `fix(debug): matching recipe gates the Phase 3 exit`

### Task 2: Copyable race probe and artifact grader

**Files:** Create `plugin/skills/debug/race.md`; modify `plugin/skills/debug/SKILL.md`, `internal/plugincheck/budget_test.go`, `internal/plugincheck/evals_test.go`; create `plugin/evals/debug-race/graders/barrier-probe.md`.
**Seam:** The debug skill run by the eval harness on debug-race; plugincheck budget and eval-case rules.
**verify:** A run on a race executes a probe that holds both workers at a barrier between read and write, and a run that skips it fails a deterministic grader, not only the judge; every debug case keeps its judge, and only the debug-race case gains a non-judge grader. List the cases run; debug-race passes 2 of 2, the other debug cases still pass.
- [ ] Red: add the Barrier grader and the plugincheck change; run debug-race on the Task 1 skill and see the new grader fail.
- [ ] Change: write race.md, point the Race recipe at it, add its byte cap; pass `scripts/test ./internal/plugincheck/`, then debug-race twice and the other debug cases once.
- [ ] Commit: `fix(debug): copyable race probe and barrier grader`
