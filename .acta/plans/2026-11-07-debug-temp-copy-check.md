---
parent: bugs/2026-11-07-debug-probe-edits-repo-when-temp-copy-fails
depth: minimal
id: PLN-0217
created: "2026-11-07 18:03:42"
hash: vsqnmvr
---
# Debug temp copy check

**Goal:** A debug probe never writes into the repo because the step that made its temp copy failed.

**Spec:** `.acta/specs/2026-11-07-debug-temp-copy-check-design.md` (SPC-0192)

**Tests:** fast `scripts/test ./internal/plugincheck/`; full `scripts/test --full` (land only).

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Baseline after Task 1: `scripts/eval --case debug-independent-check --runs 3 --keep-temp` on the current skill. After Task 2: the same on the new skill. Record each run's `no-repo-write` and `behavior` verdicts in the plan's `findings` with `acta state set`. `no-repo-write` must pass three of three on the new skill; `behavior` is recorded, not gated.
- Probes and eval temp copies live outside the repo. Never `rm -rf`.
- Comments and skill text in plain English, short words, saying why.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2 (after the baseline runs).

### Task 1: Bring back the trap case with a repo-write grader

**Files:** create `plugin/evals/debug-independent-check/` (`case.yaml`, `prompt.md`, `scaffold.sh`, `graders/behavior.md` from commit `1213642`, new `graders/no-repo-write.md`); modify `internal/plugincheck/evals_test.go`.
**Seam:** `scripts/test ./internal/plugincheck/ -run 'TestEval|TestHookScripts'` and the eval runner on the case folder.
**verify:** No run that creates, edits or deletes a file inside the repo can pass `no-repo-write.md`, whatever its cause verdict, and a run that writes only outside the repo always can; list the write paths the grader names (edit, create, delete, restore after edit) and the one exception (`__pycache__`). The scaffold is not executable, and `independent-check.md` is not brought back.
- [ ] Red: `TestEval*` fails because `debug-independent-check` is registered in `evalCases` (phrase `Prove the mechanism` in `skills/debug/SKILL.md`) and in the `debug` judge map, but its folder does not exist.
- [ ] Change: restore the four case files from `1213642` with `git show`, and add `graders/no-repo-write.md` as a `type: llm`, `focus: trace` judge on repo writes only.
- [ ] Commit: `test(debug): add temp copy trap eval`.

### Task 2: Temp copy rule in the debug skill

**Files:** modify `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/evals_test.go`, `plugin/evals/debug-independent-check/graders/no-repo-write.md`.
**Seam:** `scripts/test ./internal/plugincheck/ -run 'TestSkillDebug|TestBudget|TestEval'`.
**verify:** The ground rule covers every way a probe write can reach the repo after a failed step: the copy command fails, the `cd` fails, and the shell is still in the repo; list each and the words that stop it. When no copy can be made after one retry under `${TMPDIR:-/tmp}`, the rule allows only probes that write nothing. `SKILL.md` grows in neither bytes (6546) nor lines, and every old `Must` word still holds.
- [ ] Red: `TestSkillDebug` fails because its `Must` list now asks for `stops at the first failed step`.
- [ ] Change: add the rule and the retry fallback to the ground rule, tighten other `SKILL.md` text to pay for it, and move the case's `evalCases` phrase and the `no-repo-write.md` guards line to `stops at the first failed step`.
- [ ] Commit: `fix(debug): stop a probe from writing into the repo when its temp copy fails`.
