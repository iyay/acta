---
type: Gotcha
title: ExecProcess drops mouse mode
description: Bubble Tea v1.3.10 turns the mouse off when it runs an external program and never turns it back on
paths: [internal/tui/]
timestamp: 2026-10-09T14:31:19Z
---

`tea.ExecProcess` releases the terminal, and that turns mouse mode off. `RestoreTerminal` brings back the alt screen, paste and focus, but not the mouse. `WithMouseCellMotion` only works at startup. So after the editor closes, clicks and wheel are dead for the rest of the session, while keys still work (BUG-0027).

Fix: when the child program returns, batch `tea.EnableMouseCellMotion` with the reply (see the `editorDoneMsg` case in `internal/tui/model.go`). Any new `tea.ExecProcess` call needs the same. Model tests cannot see this; check the raw escape bytes in a real terminal (see /tui-pty-probe.md).
