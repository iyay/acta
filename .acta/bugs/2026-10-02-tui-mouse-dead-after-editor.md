---
id: BUG-0027
hash: btxr2zk
priority: high
started: "2026-10-02 05:51:34"
fixed_in: c3b3623
finished: "2026-10-02 07:42:45"
---
# TUI mouse stops working after the editor runs, until the TUI restarts

## Symptom

The TUI cannot be clicked at all: no row selects, no tab opens, no link
follows, and the wheel does nothing. Every key still works, so the board is
driven from the keyboard until acta is restarted. Reported as "sometimes the
TUI cannot be clicked, for example in tab 2" (tab 2 is where the reader
happened to be, not a cause).

## Root cause

`tea.ExecProcess` puts the terminal back the way it was before the program
ran, and Bubble Tea never puts the mouse back.

- `internal/tui/model.go:1070` (`e`, edit the row) and `internal/tui/model.go:954`
  (`n`, new bug) both run the editor with `tea.ExecProcess`.
- `exec.go:102` `Program.exec` calls `ReleaseTerminal`, which reaches
  `restoreTerminalState` (`tty.go:41`), and that calls `p.disableMouse()`
  (`tea.go:374`): the program writes `ESC[?1002l`, `ESC[?1003l` and
  `ESC[?1006l`, so the terminal stops reporting the mouse.
- `RestoreTerminal` (`tea.go:885`) puts back the alt screen, the renderer,
  bracketed paste and focus reporting, but not mouse mode. Nothing in acta
  sends it again, and `tea.WithMouseCellMotion()` (`internal/cli/cli.go:55`)
  only runs once at start.

So the first `e` or `n` kills the mouse for the rest of the session. Keys are
untouched, which is why the keyboard still works.

## Repro

1. Run `acta` in a terminal.
2. Press `2` (any tab with a row), then `e`, then quit the editor.
3. Click a row, a tab name, the status line, and turn the wheel: nothing
   happens. `j`, `k`, `1`-`6` still work.
4. Quit and run `acta` again: the mouse works until the next `e` or `n`.

## Fix direction

Re-enable the mouse when the editor comes back, in the one place both editor
paths land: the `editorDoneMsg` case in `update` (`internal/tui/model.go:364`).
`tea.EnableMouseCellMotion` is a command the event loop already handles, so
batching it with the return keeps the change in one line and needs no new
plumbing. Check that it does not double-enable when no editor ran: the
message only arrives after one did.
