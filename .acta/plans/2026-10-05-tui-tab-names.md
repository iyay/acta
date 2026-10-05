---
parent: specs/2026-10-05-tui-tab-names
depth: minimal
id: PLN-0084
created: "2026-10-05 10:14:33"
hash: lq4fw29
---
# TUI tab names Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The TUI top tabs read Scratchpad, Bugs, Debt, Specs, Plans, Activity.

**Spec:** `.acta/specs/2026-10-05-tui-tab-names.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Tab order, keys and kinds do not change. CLI words and folder names do not change.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: rename the three tabs

**Files:** Modify `internal/tui/sidebar.go`, `internal/tui/model.go`, `internal/tui/styles.go`, and the tests in `internal/tui/` that name the old labels (`model_test.go`, `scroll_test.go`, `detail_test.go`, `priority_key_test.go`, `frame_test.go`, `styles_test.go`, `view_test.go`, `priority_test.go`, `sidebar_test.go`).
**verify:** No screen the TUI draws (tab bar, sidebar, help popup, detail, status line) ever shows `Scratches`, `Debts` or `Activities`, and the tab bar still fits at every width the tests check. List every screen checked and the narrowest width.
- [ ] Failing test: a rendered frame at the default size and with the help popup open contains `Scratchpad`, `Debt` and `Activity` and none of the old words; fails because the old names are still in `topTabs`.
- [ ] Code: change the three names in `topTabs` and every TUI string that repeats them; update the old labels in the tests.
- [ ] Commit: `tui: rename top tabs to Scratchpad, Debt, Activity`.

### Task 2: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [ ] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [ ] Code: add 1 to the patch in all three files.
- [ ] Commit: `plugin: bump patch version`.
