---
parent: specs/2026-10-20-close-without-fast-suite-design
depth: minimal
id: PLN-0180
created: "2026-10-20 07:15:42"
hash: i9x3d61
---
# Close builds without another fast suite

**Goal:** Send completed builds to review without an extra repository-wide fast suite.
**Spec:** `.acta/specs/2026-10-20-close-without-fast-suite-design.md`
**Tests:** `scripts/test` (manual fast command, not a mandatory close step), `scripts/test --full` (full, land only); task regression command: `scripts/test ./internal/plugincheck/ -run 'Test(BuildClose|BuildDispatch|Eval)'`; behavior gate: `scripts/eval` at landing.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Remove only the mandatory repository-wide fast suite at close and dispatch reply verification; preserve task TDD, relevant tests after new changes, formatter checks, outstanding type checks, wiki checks, progress checks, review, and full landing gates.
- Reuse task evidence when it still covers the current changes; independent dispatch checks stay narrow, and decisive mutation checks on sensitive changes remain required.
- Keep startup without baseline tests, saved landing baselines, and existing failure/integration rerun rules; leave `scripts/test` unchanged and available for manual use.
- No new setting, dependency, CLI command, or unrelated fixes; use the inline executor in a worktree, write repo content in English, and never push.

## Waves

- Wave 1: Task 1. This is one coupled workflow change, so no plan split is needed.

### Task 1: Remove the extra close suite across build executors

**Files:** Modify `plugin/skills/build/SKILL.md`, `plugin/skills/build/dispatch.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/build_dispatch_test.go`, `internal/plugincheck/evals_test.go`, and `.acta/wiki/skill-text-fits-everyone.md`; create `plugin/evals/build-close-no-suite/case.yaml`, `prompt.md`, `scaffold.sh`, and deterministic graders under `graders/`.
**Seam:** The build skill's shared close transition and dispatch reply verification, `CheckSkill` and focused build policy checks, and the sandboxed build-close eval's commands and artifacts.
**verify:** Inline, subagent, and dispatch close paths never require a broad fast suite before review; wiki updates still precede review, task evidence is retained, outstanding type checks run, missing evidence requires only relevant checks, and dispatch still independently verifies sensitive changes; enumerate the checked paths and preserve all existing landing gates. In the behavior eval, a completed plan reaches the review hand-off, a known required type check runs, and the broad suite's persistent invocation marker remains absent; the prompt must not tell the agent to skip tests, and reading or quoting a command must not count as executing it.
- [ ] Red: Update the close-order regression and add shared-close/dispatch checks plus the registered behavior case with a counted broad-suite trap and type-check artifact; observe `scripts/test ./internal/plugincheck/ -run 'Test(BuildClose|BuildDispatch|Eval)'` fail because the current skills still require the extra suite.
- [ ] Change: Remove the broad close run and dispatch fast-command rerun, preserve outstanding type checks and verification evidence, finish the eval using existing scaffold/grader patterns and no shell-command regex as proof of execution, update the wiki, and pass the focused checks and new behavior case.
- [ ] Commit: `perf(build): skip the extra fast suite before review`.

## Close

Review the complete branch once at the configured deep level, then run the normal full suite, behavior eval, and wiki gates before landing. Keep the existing no-baseline startup rule. The changed close policy removes the broad fast run; focused regression checks above still prove this change.
