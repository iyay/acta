---
parent: specs/2026-10-05-polish-as-plan-task
depth: minimal
id: PLN-0087
created: "2026-10-05 13:48:17"
hash: l6ezwmx
started: "2026-10-05 13:55:00"
---
# Polish as a plan task Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A review polish shows as an open task in the plan, and `dispatch send --round polish` hands that task over.

**Spec:** `.acta/specs/2026-10-05-polish-as-plan-task.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- The section heading is exactly `## Polish`; the task heading is `### Task N: Review polish`.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: dispatch picks the polish task

**Files:** Modify `internal/cli/dispatch_brief.go`, `internal/cli/dispatch_brief_test.go`, `internal/cli/dispatch_send_test.go`.
**verify:** For every round (first, fix-N, polish) the brief and checkpoint name exactly the tasks of that round's section and no other; a polish round on a plan with no `## Polish` section refuses with a message naming the section. List each round and plan shape checked.
- [x] Failing test: a plan with tasks, a `## Fix round 1` and a `## Polish` section; `--round polish` must name only the polish task, the first round must leave it out; fails because polish takes no tasks today.
- [x] Code: `scanPlan` also marks the `## Polish` line; `briefTasks` maps round polish to that section and ends the first and fix rounds before it; refuse when it is missing.
- [x] Commit: `dispatch: a polish round hands over the plan's polish task`.

### Task 2: review and dispatch skill text

**Files:** Modify `plugin/skills/review/SKILL.md`, `plugin/skills/build/dispatch.md`.
**verify:** Every place the skills tell an agent how to run a polish says to append the `## Polish` task first, and none still says the polish has no task; `internal/plugincheck` passes. List each place checked.
- [x] Failing test: `scripts/test ./internal/plugincheck` stays green before; check with grep that no polish line names the section yet.
- [x] Code: in `## After a CLEAN round` step 2, append the `## Polish` block from the spec before the commit; in dispatch.md, one line that `--round polish` hands over that task.
- [x] Commit: `skills: a review polish is a task in the plan`.

### Task 3: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.

## State

### Findings

Checked all 14 polish lines in both skills; each run-a-polish line names the task.
Step 2 appends the spec's Polish block before any commit.
Both --round polish sends hand over the Polish task.
No line says polish has no task or skips reviewers.
Byte caps raised on purpose: review 11494>11874, dispatch 6338>6371.
plugincheck fresh run green.
