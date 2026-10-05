---
parent: specs/2026-10-06-tui-review-rounds-once-per-load
closes: [SPC-0088]
depth: minimal
id: PLN-0098
created: "2026-10-06 05:34:45"
hash: wcciscq
started: "2026-10-06 05:37:45"
finished: "2026-10-06 06:04:20"
---
# Review rounds once per load and dispatch agent name Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The TUI never runs `board.PlanStates` during a render, and every tick in a dispatched worktree records the agent the plan was sent to.

**Spec:** `.acta/specs/2026-10-06-tui-review-rounds-once-per-load.md` and `.acta/specs/2026-10-06-dispatch-default-agent-name.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Comments in plain English a 10-year-old reads back without stopping.
- `.agents.json` writes keep the existing lock and temp-file rename in `internal/write/agents.go`; no second writer path without the lock.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3
- Wave 3: Task 4

### Task 1: review rounds read once per board load

**Files:** Modify `internal/tui/model.go`, `internal/tui/detail.go`; test in `internal/tui`.
**verify:** No render path calls the `rounds` hook: it runs only inside the load command (start, fsnotify reload, reload after a write), off the UI goroutine. List every caller of `rounds` and every place `roundText` gets its data; a plan in review still shows its round, and `(none)` before the first load.
- [x] Failing test: a counting fake `rounds` hook sees more than one call when the detail of a plan in review is rendered several times after one load.
- [x] Code: `reloadCmd` (and the first load) call `rounds(cfg)` next to `load()` and carry the result in `reloadMsg`; `Model` keeps it; `roundText` reads the stored list.
- [x] Commit: `fix(tui): read review rounds once per board load (BUG-0032)`.

### Task 2: default agent name in .agents.json

**Files:** Modify `internal/write/agents.go`, `internal/board/agents.go`; tests in `internal/write`, `internal/board`.
**verify:** A tick with no flag and no `AI_AGENT` records the `"*"` name, unless the task already has a name; a flag or `AI_AGENT` always wins; no reader shows `"*"` as a task; a file with no `"*"` key behaves as before. List every reader of `.agents.json` and every way `RecordAgent` picks a name.
- [x] Failing test: after writing a default `omp`, `RecordAgent(root, id, "", now, true)` records `""`, and the board lists a task for key `"*"`.
- [x] Code: add `SetDefaultAgent(root, agent string) error` in `internal/write/agents.go` under the same lock and rename; `RecordAgent` falls back to `recs["*"].Agent` after `prev.Agent`; `readAgents` skips `"*"`.
- [x] Commit: `fix(write): ticks fall back to the worktree default agent (BUG-0033)`.

### Task 3: dispatch send records the recipient

**Files:** Modify `internal/cli/dispatch_send.go`; test in `internal/cli/dispatch_send_test.go`.
**verify:** Every successful `acta dispatch send` leaves `"*": omp` in the worktree's `.acta/.agents.json`, and a failed send writes nothing new. List each exit path of `cmdDispatchSend` and what it writes.
- [x] Failing test: a send through the existing fake herdr leaves no `"*"` key in the worktree `.agents.json`.
- [x] Code: after delivery succeeds, call `write.SetDefaultAgent(<worktree acta root>, "omp")`; an error there is reported on stderr and does not undo the send.
- [x] Commit: `fix(dispatch): send names omp as the worktree default agent (BUG-0033)`.

### Task 4: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.

## Polish

### Task 5: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] Add one plain comment in `Init` (`internal/tui/model.go`) saying why the board loads again there: the second load fills the review rounds off the UI goroutine.
- [x] In `internal/cli/dispatch_send.go`, print the error from `hook.EnsureGitignore` on stderr the same way the `SetDefaultAgent` error is printed; the send still counts.
- [x] Commit: `polish: review notes for PLN-0098`

## Review notes

- Init starts its own board load on top of the one runTUI does, so the TUI loads the board twice at start; the second load is what fills the review rounds.
- Reload reads m.load and m.cfg when called; a later WithLoad on another copy would not reach the watcher. No caller does that today.
- SetDefaultAgent stamps time.Now() directly, while RecordAgent takes now as a parameter, so tests cannot pin the time on the "*" record.
- The formatting-only commit omp made in round 1 was folded into its task commit with a rebase before landing.
