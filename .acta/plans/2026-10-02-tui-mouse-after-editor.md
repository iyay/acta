---
parent: bugs/2026-10-02-tui-mouse-dead-after-editor
depth: minimal
id: PLN-0076
created: "2026-10-02 05:50:16"
hash: q7mdmi7
started: "2026-10-02 05:51:34"
finished: "2026-10-02 05:54:37"
---
# TUI Mouse After The Editor Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** The mouse works again after `e` or `n` opens the editor and brings the TUI back (fixes `bugs/2026-10-02-tui-mouse-dead-after-editor`).

**Spec:** none (Bounded, approved in chat on 2026-10-02)

**Tests:** fast `scripts/test ./internal/tui`, full `scripts/test --full`

## Global Constraints

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- Every run step uses `scripts/test ./internal/tui -run <Name>`, never bare `go test` and never `./...`.
- The change lands in the one case both editor paths reach, `editorDoneMsg` in `update` (`internal/tui/model.go`), so neither `e` nor `n` needs its own fix.
- Comments are plain English a 10-year-old can read. They say why, not what.
- No new dependency, no new file outside `internal/tui`, and no change to `internal/cli/cli.go`.

## Waves

- Wave 1: Task 1.

### Task 1: The mouse comes back after the editor

**Files:**
- Modify: `internal/tui/model.go` (the `editorDoneMsg` case of `update`)
- Test: `internal/tui/model_test.go`

**verify:** Every path that runs the editor and returns to the TUI leaves mouse reporting on, and a model that never ran the editor is untouched. List every path checked: `e` on a row, `n` for a new bug, an editor that exits with an error, and a plain key press that sends no `editorDoneMsg`.

- [x] Failing test: a test in `internal/tui/model_test.go` that sends `editorDoneMsg` through `Update`, runs the returned command and wants one of the messages to be the one `tea.EnableMouseCellMotion` produces; it fails because the case returns only what `afterEditor` returns, which carries no mouse message.
- [x] Code: in the `editorDoneMsg` case of `update`, batch `tea.EnableMouseCellMotion` with the command `afterEditor` returns, with a comment saying the editor left the terminal without mouse reporting; every other case of `update` stays as it is.
- [x] Commit: `fix(tui): turn the mouse back on after the editor`
