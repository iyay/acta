---
parent: specs/2026-10-05-dispatch-false-drift
depth: minimal
id: PLN-0092
created: "2026-10-05 19:15:46"
hash: eea59pz
---
# Dispatch false drift Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** A dispatch whose todo card names no task ids yet reports `unconfirmed`, not drift.

**Spec:** `.acta/specs/2026-10-05-dispatch-false-drift.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: checkpoint verdict

**Files:** Modify `internal/cli/dispatch_herdr.go`, `internal/cli/dispatch_herdr_test.go`, and `internal/cli/dispatch_send_test.go` if a send test pins the old verdict.
**verify:** For every pane text (no card, card with no task ids, card with some ids, card with all ids) the verdict is exactly unconfirmed, unconfirmed, drift with the missing ids, ok; no card that names no task id ever gives drift. List each pane text checked.
- [ ] Failing test: pane text with a todo card of generic steps and no task ids; checkpoint returns drift today, so the test wanting unconfirmed fails.
- [ ] Code: count the ids found; zero found gives unconfirmed, some but not all gives drift, all gives ok.
- [ ] Commit: `dispatch: a todo card with no task ids is unconfirmed, not drift`.

### Task 2: dispatch skill text

**Files:** Modify `plugin/skills/build/dispatch.md`, and `internal/plugincheck/budget_test.go` only if the cap needs it.
**verify:** Every line in dispatch.md that explains exit 4 or `unconfirmed` matches the new verdicts; `internal/plugincheck` passes. List each line checked.
- [ ] Failing test: `scripts/test ./internal/plugincheck` green before; read the exit 4 lines and see they treat any missing id as drift.
- [ ] Code: say exit 4 means a card that names some task ids and skips others; a card with no ids comes back as unconfirmed.
- [ ] Commit: `skills: dispatch drift means a card that skips some task ids`.

### Task 3: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.
