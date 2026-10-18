---
parent: specs/2026-10-18-inline-default-executor-design
depth: minimal
id: PLN-0156
created: "2026-10-18 19:00:42"
hash: u9xqav4
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

- [ ] **Files:** `internal/setup/form.go`, `internal/setup/form_test.go`
- [ ] **verify:** With no executor saved, every setup path (form and text mode) starts on inline, and inline is the first option listed; a saved executor is never replaced. Seam: `FormDefaults`, `buildFormState`.
- [ ] Commit: `feat(setup): inline is the default build executor`

### Task 2: Build and doctor treat an empty executor as inline

- [ ] **Files:** `plugin/skills/build/SKILL.md`, `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`
- [ ] **verify:** No text in the build skill or doctor still says a build asks which executor when none is set; the build order ends in inline. Seam: `checkSetup`; `go test ./internal/plugincheck/` stays green.
- [ ] Commit: `feat(build): run inline when no executor is set`

### Task 3: Config show and session note name the default

- [ ] **Files:** `internal/cli/config_cmd.go`, `internal/hook/hook.go`, their tests in `internal/cli/` and `internal/hook/`
- [ ] **verify:** With an empty executor, both `acta config show` (text and JSON) and the session note show inline marked as the default; a set executor prints as before with no default mark. Seam: `acta config show`, the session hook text.
- [ ] Commit: `feat(config): show inline as the default executor`
