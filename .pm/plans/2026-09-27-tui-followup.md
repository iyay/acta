---
id: PLAN-13
hash: bmx1
---
# TUI follow-up Implementation Plan

> **For agentic workers:** run this plan with pm:build, task by task. Steps use checkbox (`- [ ]`) syntax, and pmb reads those boxes as task progress.

**Goal:** The TUI bottom line shows only `? help` on the left and `<project> · live · <date time> | Donate  Feedback  <version>` on the right; `pmb tick <id> --start` marks a task started so it shows `doing` before any box is ticked; pane [1] lists in-progress items first in the accent color, a dim divider, then not-started items; `enter` focuses the detail pane and `e` edits.

**Architecture:** `--start` writes a `started` flag into the existing `.agents.json` record; the board reads it and derives `doing` for a 0-ticked task. The bottom line reads two new config keys (`links.donate`, `links.feedback`) and the Go build version; clicks on its links go through an injectable opener. Pane [1] rows gain a divider row type the cursor skips.

**Tech Stack:** Go 1.27, Bubble Tea v1.3.10, lipgloss, stdlib `runtime/debug`, `os/exec`.

**Spec:** `.pm/specs/2026-09-26-tui-followup-design.md`

**Worktree:** `/Users/iyay/Nayakatara/pm-board-tui-followup`, branch `tui-followup`, parent `main` (main merged in at 7aca6a9, PLAN-12 code present).

## Global Constraints

- Bottom line right side, verbatim: `<project> · live|paused · <YYYY-MM-DD HH:MM> | Donate  Feedback  <version>`; no `pmb` word; `Donate` only when `links.donate` is set; `links.feedback` default `https://github.com/iyay/acta/issues`; version from build info, `dev` for local builds; narrow window drops project then status first, always keeps date and time.
- `--start` ticks no box, cannot combine with `--step`/`--all` (exit 1).
- In-progress statuses: `doing`, `in-progress`, `fixing`, or a task with a `started` record; not started: `todo`, `draft`, `approved`, `open`. No section labels; divider hidden when a group is empty; cursor and clicks skip it.
- `enter` on a row focuses pane [3]; `enter` on the legacy group row still toggles; `e` edits from any pane; `esc` in [3] returns to the list pane.
- Gate before every commit: `test -z "$(gofmt -l .)" && go vet ./... && go test -count=1 ./... && (cd plugin && bun test)`.
- Edit only the files each task names. Stage by path. Never push. Never commit the plan file. gofmt inside the task commit. Never edit fixtures to pass. A defect in this branch's code is fixed here, never filed in `.pm/bugs/`. `internal/tui/model.go` is already 807 lines: any task that grows it must first move code out (for example the title/tab helpers into a new `internal/tui/title.go`) so it ends under 800.
- Comments in plain English a ten-year-old can read, saying why. No marker tags.
- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.

## File map

- Task 1: `cmd/pmb/tick.go`, `cmd/pmb/tick_test.go`, `internal/write/agents.go`, `internal/write/agents_test.go`, `internal/board/agents.go`, `internal/board/agents_test.go`, `internal/board/board.go` (status derivation only).
- Task 2: `plugin/skills/build/implementer-prompt.md`, `plugin/skills/build/SKILL.md`, `plugin/skills/dispatch/herdr-delivery.md`, `internal/plugincheck/skill_build_test.go`, `internal/plugincheck/skill_dispatch_test.go`.
- Task 3: `internal/tui/view.go` (status line), `internal/tui/title.go` (new, moved helpers), `internal/tui/model.go` (link clicks, opener, version field), `internal/tui/view_test.go`, `internal/tui/model_test.go`, `internal/config/config.go`, `internal/config/config_test.go`, `cmd/pmb/main.go` (pass version).
- Task 4: `internal/tui/model.go`, `internal/tui/view.go`, `internal/tui/model_test.go`, `internal/tui/view_test.go`.

## Waves

- Wave 1: Tasks 1, 2 and 3 (no shared files).
- Wave 2: Task 4 (needs Task 1's started status and Task 3's model.go split).

---

### Task 1: `pmb tick --start` marks a task started

**verify:** Every `pmb tick <id> --start` writes a started record (agent, time, `started: true`) and changes no box; combining `--start` with `--step` or `--all` exits 1 and writes nothing; a task with a started record and no ticked box derives `doing`, and its plan and spec derive in progress; a later normal tick keeps the record; a done task shows no agent; records from other worktrees count. List every flag combination and status path checked.

**Interfaces:** Produces `agentRec.Started bool` (json `started,omitempty`) in both `internal/write/agents.go` and `internal/board/agents.go`; `write.RecordAgent(root, taskID, agent string, now time.Time, started bool) error` (update both callers); `Item.Started bool` on tasks.

- [ ] **Step 1: Failing tests:** `tick_test.go` (`--start` writes `started:true`, file content unchanged, exit 0; `--start --all` and `--start --step 1` exit 1 with no write); `agents_test.go` in write (started kept on a later record); `agents_test.go` in board (started + 0 ticked gives `doing`, parent in progress; done task has no agent).
- [ ] **Step 2: Run them to see them fail.**
- [ ] **Step 3: Implement.** Add `--start` to `cmdTick`; validate the three-way exclusion; on `--start` skip `write.Tick`, still run `EnsureGitignore` and `RecordAgent(..., true)` (with no agent name, record `started` anyway with an empty agent). In the board, `fillAgents` runs after `derive` today; read the agent records before `derive` (keep the fill step after it), set `Started` on tasks, and make the task status `doing` for `done == 0 && started`, so `derive` lifts the parent as today.
- [ ] **Step 4: Run to see them pass.**
- [ ] **Step 5: Gate, commit** (`feat(tick): --start marks a task started`), tick `#task-1 --all`.

---

### Task 2: Skills start each task with `--start`

**verify:** Every place in the build skill, the implementer prompt and the dispatch delivery that tells an implementer how to tick names `pmb tick <id> --start` as the very first action of a task, before the failing test, with the same `--agent` rule as other ticks (Claude: nothing extra; omp: `--agent omp`); each file has its own test that goes red if its line is removed. List every place checked.

- [ ] **Step 1: Failing tests** per file in plugincheck (`--start` present in each of the three files; the implementer prompt puts it before the failing-test step).
- [ ] **Step 2: Run to see them fail.**
- [ ] **Step 3: Write the text** in the three files next to the existing tick instructions.
- [ ] **Step 4: Run to see them pass** (MaxLines still hold; dispatch folder is near 620, keep the addition to one line per file).
- [ ] **Step 5: Gate, commit** (`feat(plugin): implementers mark tasks started first`), tick `#task-2 --all`.

---

### Task 3: Bottom line: `? help`, date and time, links, version

**verify:** At every width from 30 to 200 the bottom line never exceeds the width; the left shows only `? help` (or the search box or a status message while active); the right shows `<project> · live|paused · YYYY-MM-DD HH:MM | [Donate  ]Feedback  <version>` with no `pmb` word; `Donate` appears only when `links.donate` is set; narrow widths drop project, then status, never the date and time; a click on `Donate` or `Feedback` calls the opener with the configured URL and nothing else; version is the build info main version, `dev` when empty or `(devel)`; `model.go` ends under 800 lines. List every width and element checked.

**Interfaces:** `config.Config.Links struct{ Donate, Feedback string }` (yaml `links: {donate:, feedback:}`, feedback default); `tui.New(..., version string)` or `Model.WithVersion(v string)`; `Model.open func(url string) error` (default runs `open` on darwin, `xdg-open` elsewhere); links written with OSC 8 around the text.

- [ ] **Step 1: Failing tests:** `config_test.go` (links read from `.pm.yaml`, feedback default); `view_test.go` (line shapes at several widths with a fixed `m.now`, Donate hidden/shown, no `pmb`); `model_test.go` (click on each link calls a fake opener with the right URL; click elsewhere on the line calls nothing); version helper table (`v0.3.0`, empty, `(devel)`).
- [ ] **Step 2: Run to see them fail.**
- [ ] **Step 3: Implement.** First move `titlePiece`/`titlePieces`/`tabX` into `internal/tui/title.go` (no behavior change, tests stay green), then build the status line in `view.go` from pieces whose x positions the model also uses for clicks (same rule as tab boxes: one helper for text and boxes). `cmd/pmb/main.go` reads `debug.ReadBuildInfo()` and passes the version.
- [ ] **Step 4: Run to see them pass.**
- [ ] **Step 5: Gate, commit** (`feat(tui): bottom line with date, links and version`), tick `#task-3 --all`.

---

### Task 4: In-progress first, divider, `enter` focuses detail, `e` edits

**verify:** In every tab of pane [1], every in-progress item (including a task with only a started record) sorts before every not-started item, in-progress rows render in the accent text color and the others in normal text, one dim divider sits between the groups and is absent when either group is empty, and `j k g G` plus mouse clicks never select the divider; `enter` on any row of [1] or [2] focuses [3] on that item and opens no editor; `enter` on the legacy group row toggles it; `e` opens the editor from [1], [2] and [3]; `esc` in [3] returns focus to the list pane it came from; the `?` popup lists the new `enter`, `e`, `esc`. List every tab, key and click case checked.

**Interfaces:** Consumes `Item.Started` (Task 1) and the `title.go` split (Task 3). `row` gains `divider bool`.

- [ ] **Step 1: Failing tests** in `model_test.go` and `view_test.go` for each verify clause (rewrite only the existing tests that pinned `enter` opening the editor).
- [ ] **Step 2: Run to see them fail.**
- [ ] **Step 3: Implement** in `model.go` (`openRows` ordering and divider row, cursor/click skip, key changes, remembered list pane) and `view.go` (colors, divider line, help text). Keep `model.go` under 800 lines.
- [ ] **Step 4: Run to see them pass.**
- [ ] **Step 5: Gate, commit** (`feat(tui): in-progress first, enter focuses detail, e edits`), tick `#task-4 --all`.
