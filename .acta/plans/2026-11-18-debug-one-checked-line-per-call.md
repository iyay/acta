---
parent: specs/2026-11-18-fix-two-flaky-evals
depth: minimal
id: PLN-0247
created: "2026-11-18 08:01:07"
hash: ojbnh96
---
# Debug skill: one checked line per shell call

**Goal:** Agents on the debug skill put each probe write in its own checked shell call, so `debug-probe-isolation` passes at least 8 of 10 runs.

**Spec:** `.acta/specs/2026-11-18-fix-two-flaky-evals.md` (SPC-0215), Fix 2
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched and `scripts/eval --case <name>` for one case; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches `plugin/skills/debug/`, so the gate runs as `scripts/eval --tag=debug`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Never trim other words of the debug skill to fit its byte cap; raise the cap.
- Judge the wording on 10 runs per arm: new skill against old skill, each with an acta built from its own commit first on PATH in a temp dir. At most 3 wordings; after a third failure stop and report.

## Waves

- Wave 1: Task 1

### Task 1: One checked line per shell call

**Files:** `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`
**Seam:** `scripts/test ./internal/plugincheck -run 'TestDebug|TestBudget'`, and `scripts/eval --case debug-probe-isolation`
**verify:** No write in a debug probe can run outside its checked line: the skill text says each shell call holds one checked line and the next write goes in a new call with the check repeated, the test pins those words, and the case passes at least 8 of 10 runs and no fewer than the old skill. List both arms' scores and, for each failing run, the line in its trace that broke the rule.
- [ ] Red: the plugincheck test that pins the new sentences fails because the skill does not hold them yet.
- [ ] Change: append the spec's three sentences after the "Join the write" sentence, raise the debug skill byte cap by what they need, and run the 10 against 10 eval.
- [ ] Commit: fix(debug): one checked line per shell call
