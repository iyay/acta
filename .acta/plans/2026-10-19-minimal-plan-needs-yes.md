---
parent: specs/2026-10-19-minimal-plan-needs-yes-design
depth: minimal
id: PLN-0166
created: "2026-10-19 07:14:32"
hash: et54bvp
started: "2026-10-19 07:16:13"
finished: "2026-10-19 07:17:16"
---
# Minimal plan needs yes plan

**Goal:** A minimal plan waits for the user's yes before build, the same as a full plan.
**Spec:** `.acta/specs/2026-10-19-minimal-plan-needs-yes-design.md`
**Tests:** `scripts/test` (fast), `scripts/test --full` (full, land only); task loop: `scripts/test ./internal/plugincheck/ ./internal/hook/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Depth changes only what a plan holds, never the gate. A direct ask to build a plan still counts as its yes. No board or Go product change beyond the hook text.
- Inline in a worktree, English in the repo, never push.

## Waves

- Wave 1: Task 1, Task 2.

### Task 1: Slice and build skills keep the plan gate for minimal plans

**Files:** Modify `plugin/skills/slice/SKILL.md`, `plugin/skills/build/SKILL.md`, `internal/plugincheck/skill_slice_test.go`, `internal/plugincheck/skill_build_test.go`, and `internal/plugincheck/budget_test.go` only if a cap changes.
**Seam:** `plugincheck.CheckSkill` with the slice and build `SkillRule`s.
**verify:** No skill text lets any plan depth skip the yes or start build in the same turn as slice; list every place in slice and build that names depth or the gate, and what each says now. The old phrases cannot come back reworded next to the new ones.
- [x] Red: Require the gate wording and ban "needs no yes", "skips the wait", "in the same turn" and "or has `depth: minimal`"; observe `scripts/test ./internal/plugincheck/` fail.
- [x] Change: Rewrite the slice Plan depth and Hand-off text and the build refuse line; pass the same command.
- [x] Commit: `fix(slice): a minimal plan waits for the user's yes`.

### Task 2: Core rule 2 drops the minimal exemption

**Files:** Modify `internal/hook/hook.go`, `plugin/hooks/default-rules.md`, `internal/hook/hook_test.go`.
**Seam:** the session start text from `internal/hook` and the fallback `default-rules.md`.
**verify:** Both copies of core rule 2 (the hook text and the fallback file) say no code before an approved design and an approved plan, with no depth exemption, and they still match each other. List both and any test that pins them.
- [x] Red: Change the expected rule in `hook_test.go`; observe `scripts/test ./internal/hook/` fail.
- [x] Change: Drop the exemption sentence from `hook.go` and `default-rules.md`; pass `scripts/test ./internal/hook/ ./internal/plugincheck/`.
- [x] Commit: `fix(hook): core rule 2 has no minimal plan exemption`.

## Fix round 1

### Task 3: Fix round 1

**verify:** No surface the user sees says a minimal plan builds at once: list every one checked (setup wizard, slice, split, build, core rule 2).
- [x] BLOCKER: setup wizard Plan detail said "minimal: short steps, build starts at once"; now "short steps, no code blocks", and full drops "waits for your yes" (internal/setup/form.go:158-159).
- [x] [fix] build test requires "a direct ask to build that plan counts as its yes".
- [x] [fix] hook_test: the new check moved above the index comment it had split.
- [x] [fix] split.md drops "even for a minimal plan".
- [x] Commit: `fix(setup): plan detail no longer promises an instant build`

## Review notes

- Round 1: deep, full tier, Spec and Standards reviewers over eb65b4a..ed61f10: BLOCKED (setup wizard said minimal builds at once); fix round 1 96a83d6. Round 2 over ed61f10..96a83d6: CLEAN. Two (medium) NOTEs filed as debt; notes left: old spec 2026-10-07 shows the old wizard labels, phrase bans miss new paraphrases.
