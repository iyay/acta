---
parent: specs/2026-10-22-plan-completion-skill-gates-design
depth: minimal
id: PLN-0192
created: "2026-10-22 23:50:15"
hash: vahdco8
---
# Verify plan completion through build and land
**Goal:** Prevent build and land from claiming completion while plan progress or implementation commit links remain incomplete.
**Spec:** .acta/specs/2026-10-22-plan-completion-skill-gates-design.md
**Tests:** Fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`, `go vet ./...`, and `scripts/eval` at land with the branch binary on PATH.
## Global Constraints
- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Change skills, focused contract tests, and behavior evals only; keep CLI status derivation and tidy implementation unchanged. Do not repair or close existing plans.
- Preserve all existing review, test, wiki, clean-tree, merge, and approval gates. Keep skill text portable and within plugincheck size caps.
- Eval fixtures use disposable repositories and homes. Grade repository state and executed checks, not response keywords alone.
- One plan covers the completion invariant shared by build and both landing modes; no independent subsystem requires a separate plan.
## Waves
- Wave 1: Task 1 and Task 2, executed inline.
### Task 1: Stop build when task progress cannot be recorded
**Files:** Modify `plugin/skills/build/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, `internal/plugincheck/skill_build_test.go`; create `plugin/evals/build-tick-failure/`, `plugin/evals/build-missing-checkboxes/`, and `plugin/evals/build-open-progress/`, each with `case.yaml`, `prompt.md`, `scaffold.sh`, and `graders/progress-gate.md`.
**Seam:** Build skill instructions used by the executor and the implementer prompt; behavior evals running the real acta CLI in disposable repositories.
**verify:** Every progress-write error blocks subsequent task work, review and land until resolved; a completed task is read back after commit and tick; review requires a nonempty plan with every task complete. Cover a failed tick, missing checkboxes, and open task progress. Preserve evidence and never force completion or repeat implementation to hide a tracking error. Verify with `scripts/test ./internal/plugincheck -run 'Test.*Build'` and the new behavior evals through the existing runner.
- [ ] Red: Add focused contract checks and behavior evals that expose missing stop conditions and task readback in the current build instructions; run the focused checks and observe the intended failures before editing skills.
- [ ] Change: Strengthen the existing tick and pre-review gates in the build skill and implementer prompt, then run the focused checks and the new evals to verify failure handling and successful progress recording.
- [ ] Commit: `fix(build): stop on incomplete task progress` with this task's stable Task trailer.
### Task 2: Verify completion in the parent before cleanup and success
**Files:** Modify `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, `internal/plugincheck/skill_land_test.go`; create `plugin/evals/land-full-progress/`, `plugin/evals/land-tidy-progress/`, and `plugin/evals/land-parent-progress-mismatch/`, each with `case.yaml`, `prompt.md`, `scaffold.sh`, and `graders/progress-gate.md`.
**Seam:** Full-history and tidy landing skill flows, using existing plan/task show and commit-list commands from the parent checkout.
**verify:** Both landing modes verify the stable plan hash from the parent after moving it and before cleanup or completion claims; require done status, nonempty complete task progress, and linked implementation commits for every task. Cover full-history success, tidy success with preserved task links, and a final-parent mismatch that retains the worktree and branch and blocks success. Verify with `scripts/test ./internal/plugincheck -run 'Test.*Land'` and the new behavior evals through the existing runner. Build-side failures belong to Task 1 only.
- [ ] Red: Add focused checks for ordering in both landing modes and behavior evals for final parent state; run the focused checks and observe the missing final verification before editing skills.
- [ ] Change: Add a shared final completion gate and make both landing modes run it before cleanup, then run the focused checks and behavior evals; preserve failed-state evidence without forcing plan status.
- [ ] Commit: `fix(land): verify parent plan before cleanup` with this task's stable Task trailer.
