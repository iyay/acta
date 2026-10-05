---
parent: bugs/2026-10-05-omp-new-session-gets-no-rules
depth: minimal
id: PLN-0086
created: "2026-10-05 13:01:46"
hash: kit90tv
started: "2026-10-05 13:06:16"
finished: "2026-10-05 13:16:30"
---
# omp session switch rules and doctor files check Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** After /new, resume or fork, omp gets the acta rules again, and `acta doctor` warns about spec and plan files acta never stamped.

**Spec:** `.acta/specs/2026-10-05-omp-session-switch-rules.md`

**Tests:** `scripts/test`, `scripts/test --full`; omp: `cd plugin/omp && bun test`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- doctor reads only the planning folders of the checkout it runs in, never other branches or worktrees.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: omp resets on session_switch

**Files:** Modify `plugin/omp/index.ts`, `plugin/omp/index.test.ts`.
**verify:** For every way omp starts a new session (startup, `session_switch` with reason new, resume and fork, compaction), the first agent turn after it carries the acta rules, and no later turn in the same session carries them again. List each event checked.
- [x] Failing test: with a fake `pi`, fire `session_start`, `before_agent_start`, then `session_switch` with each reason, then `before_agent_start`; the second context lacks the rules, so it fails.
- [x] Code: register `pi.on("session_switch", restart("startup"))` next to the existing two listeners.
- [x] Commit: `omp: send the acta rules again after /new, resume or fork`.

### Task 2: doctor files check

**Files:** Modify `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`.
**verify:** Every file in the checkout's `specs/` or `plans/` with no `id:` in its frontmatter, and every file in `plans/` with no `### Task` heading, is named in a `warn` with its fix; a repo where every file is stamped and every plan has tasks gives one `ok`; files in other branches or worktrees are never read. List each case checked.
- [x] Failing test: a temp planning root with a stamped plan with tasks, a plan with no id, a spec with no id and a plans file with no task heading; the `files` check does not exist, so it fails.
- [x] Code: add `checkFiles` to the list `Run` returns; warn per bad file with fix `acta id` (no id) or `move it out of plans/` (no task); `ok` when none.
- [x] Commit: `doctor: warn about spec and plan files acta never stamped`.

### Task 3: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.
