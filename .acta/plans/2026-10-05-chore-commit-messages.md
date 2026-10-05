---
parent: specs/2026-10-05-chore-commit-messages
depth: minimal
id: PLN-0090
created: "2026-10-05 15:18:54"
hash: ub9flsj
started: "2026-10-05 15:24:25"
---
# chore commit messages Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Planning commits read `chore(<kind>): <what>` or `chore: <what>`, never `acta: `.

**Spec:** `.acta/specs/2026-10-05-chore-commit-messages.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Scopes, exactly: `scratch`, `spec`, `plan`, `bug`, `debt`, `wiki`; more than one kind, doctor fix and migrate use no scope.
- Write commands auto-commit: tests run them in temp clones with temp HOME and TMPDIR.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: auto-commit subjects and review round

**Files:** Modify the commit subjects in `internal/write/*.go`, `internal/cli/doctor.go`, `internal/cli/migrate.go`, and any other non-test Go file that writes `acta: ` into a commit subject; `internal/board/state.go`; their tests.
**verify:** No commit acta makes has a subject starting with `acta: `; each single-kind commit is scoped by that kind and each multi-kind, doctor or migrate commit has no scope; the review round is found from both `chore(plan): tick fix round N` and the old `acta: tick fix round N`. List every write command and commit path checked.
- [x] Failing test: a write command in a temp clone (for example `acta scratch new`) must commit `chore(scratch): ...`; fails because it commits `acta: ...`.
- [x] Code: one small helper that builds the subject from the kinds touched, used by every commit path; state.go matches both round subjects.
- [x] Commit: `write: planning commits use chore(<kind>) subjects`.

### Task 2: skill text

**Files:** Modify `plugin/skills/land/SKILL.md`, `plugin/skills/migrate/SKILL.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/review/SKILL.md`, `plugin/skills/build/dispatch.md`, `plugin/references/house-rules.md` where they name a commit subject.
**verify:** No skill or reference line tells an agent to commit with an `acta: ` subject, and each names the new form; `internal/plugincheck` passes. List each file and line checked.
- [x] Failing test: grep the skill and reference files for commit subjects starting with `acta: ` and see hits; `scripts/test ./internal/plugincheck` green before.
- [x] Code: change each named subject to the chore form, for example `chore(plan): tick wave <n>`, `chore(plan): tick fix round <n>`, `chore(plan): review notes for <plan id>`.
- [x] Commit: `skills: commit subjects use chore(<kind>)`.

### Task 3: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.
