---
parent: specs/2026-10-05-polish-as-plan-task
depth: minimal
id: PLN-0087
created: "2026-10-05 13:48:17"
hash: l6ezwmx
started: "2026-10-05 13:55:00"
finished: "2026-10-05 14:04:15"
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
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.

## State

### Findings

All three files agree on 0.1.11, one patch above main 0.1.10.
plugincheck green before and after the bump.
Polish tasks 1-2 already landed on this branch.
Commit holds only the three version files.
Plan file ticks stay uncommitted per contract.

## Fix round 1

### Task 4: polish round stops at the next section

**Files:** Modify `internal/cli/dispatch_brief.go`, `internal/cli/dispatch_brief_test.go`, `plugin/skills/review/SKILL.md`.
**verify:** For every plan shape (fix rounds before polish, polish before a fix round, polish followed by `## State` or `## Review notes`), `--round polish` names only the tasks under `## Polish` up to the next `## ` heading; and the review skill says what happens to the `## Polish` task when the polish commit is reverted. List each shape checked.
- [ ] Failing test: plan `### Task 1`, `## Polish` with `### Task 3: p`, `## Fix round 1` with `### Task 2: f`; `briefTasks(src, "polish")` returns p and f, so it fails.
- [ ] Code: record where the polish section ends (next `## ` heading) in `scanPlan` and use it as the upper bound for the polish round; remove the clamp in the fix-round branch at dispatch_brief.go:121-124 if it can never change the result, else say in its comment it is only a safety net; in review SKILL.md step 2 or 3, one short sentence: on revert, tick the `## Polish` task boxes and note that its items moved to debt.
- [ ] Commit: `dispatch: polish round stops at the next section; review says what a revert does to the polish task`.
