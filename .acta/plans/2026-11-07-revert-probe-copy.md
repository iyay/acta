---
parent: specs/2026-11-07-revert-probe-copy-design
depth: minimal
id: PLN-0221
created: "2026-11-07 21:06:05"
hash: k48qxgd
started: "2026-11-07 21:07:29"
finished: "2026-11-07 21:07:49"
---
# Revert the probe copy rule

**Goal:** The debug skill reads as before PLN-0217, with no unused `probe-copy.sh` left behind.

**Spec:** `.acta/specs/2026-11-07-revert-probe-copy-design.md`

**Tests:** fast `scripts/test ./internal/plugincheck/`; full `scripts/test --full` (land only).

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Delete files with `git rm`, never `rm -rf`.

## Waves

- Wave 1: Task 1.

### Task 1: Revert the skill line and remove the helper

**Files:** modify `plugin/skills/debug/SKILL.md`, `internal/plugincheck/skill_debug_test.go`, `internal/plugincheck/budget_test.go`, `internal/plugincheck/final_test.go`, `internal/plugincheck/evals_test.go`, `plugin/evals/debug-independent-check/graders/no-repo-write.md`, `.acta/wiki/skill-escape-lines-skip-proof.md`; delete `plugin/skills/debug/probe-copy.sh`, `internal/plugincheck/probe_copy_test.go`.
**Seam:** `scripts/test ./internal/plugincheck/ ./internal/setup/ ./plugin/`.
**verify:** No trace of the helper or its rule is left in the plugin or its checks: `git diff 20a15a7 -- plugin/skills/debug/` is empty, and no file under `plugin/` or `internal/` names `probe-copy` or `stops at the first failed step`; list every file that named them and what each says now.
- [x] Red: `TestSkillDebug` fails because its `Must` list no longer holds the two helper phrases while `SKILL.md` still does, and the `SKILL.md` cap is back at 6546.
- [x] Change: restore `SKILL.md` from `20a15a7`, `git rm` the helper and its test, move the case phrase and guards line back to `Prove the mechanism`, drop the executable allowance, and add the result line to the wiki page.
- [x] Commit: `revert(debug): drop the probe-copy rule and helper`.

## Review notes

- Review: deep, 1 round, Spec and Standards CLEAN. BLOCKER 0, NOTE 5, pre-existing 1 (bug debug-bisect-breaks-read-only-rule still names the removed probe-copy.sh in its repro). Bug debug-probe-edits-repo-when-temp-copy-fails is reopened at land. Open notes: no test checks the no-repo-write guards line; the wiki line adds one claim the spec does not hold.
