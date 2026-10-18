---
parent: specs/2026-10-18-inline-default-executor-design
depth: minimal
id: PLN-0156
created: "2026-10-18 19:00:42"
hash: u9xqav4
started: "2026-10-18 19:02:47"
finished: "2026-10-18 19:08:30"
---
# Inline Default Executor Plan

**Goal:** An empty build executor means inline everywhere: setup, build, doctor, config show and the session note.

**Spec:** `.acta/specs/2026-10-18-inline-default-executor-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- A config that already names `subagent` or `dispatch` behaves exactly as before.

## Waves

- Wave 1: Task 1, Task 2, Task 3

### Task 1: Setup preselects inline

- [x] **Files:** `internal/setup/form.go`, `internal/setup/form_test.go`
- [x] **verify:** With no executor saved, every setup path (form and text mode) starts on inline, and inline is the first option listed; a saved executor is never replaced. Seam: `FormDefaults`, `buildFormState`.
- [x] Commit: `feat(setup): inline is the default build executor`

### Task 2: Build and doctor treat an empty executor as inline

- [x] **Files:** `plugin/skills/build/SKILL.md`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`
- [x] **verify:** No text in the build skill or doctor still says a build asks which executor when none is set; the build order ends in inline. Seam: `checkSetup`; `go test ./internal/plugincheck/` stays green.
- [x] Commit: `feat(build): run inline when no executor is set`

### Task 3: Config show and session note name the default

- [x] **Files:** `internal/cli/config_cmd.go`, `internal/hook/hook.go`, their tests in `internal/cli/` and `internal/hook/`
- [x] **verify:** With an empty executor, both `acta config show` (text and JSON) and the session note show inline marked as the default; a set executor prints as before with no default mark. Seam: `acta config show`, the session hook text.
- [x] Commit: `feat(config): show inline as the default executor`

## Fix round 1

### Task 4: Session note names the inline default

- [x] **Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`, `internal/plugincheck/budget_test.go`, `plugin/skills/slice/SKILL.md`
- [x] **verify:** With no executor set, the session note always carries `- Build executor: inline (default).` with the inline sentence, so the acta setting beats a user's CLAUDE.md; a set executor prints as before; a missing or broken config writes no settings heading. The session start cap rises only by the bytes this line adds. Slice's hand-off reads a `(default)` value as inline.
- [x] Commit: `fix(hook): session note names the inline default executor`

## Review notes

Round 1 BLOCKED (session note lacked the inline default), fixed in round 2, CLEAN. Notes only: config show JSON no longer tells unset from inline (same as plan_depth); the per-message prompt hook names no executor; setup no-TTY fallback no longer asks the executor.
