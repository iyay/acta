---
type: Runbook
title: Drive the TUI from a pty probe
description: A Python pty plus key and SGR wheel escapes plus ACTA_TUI_TRACE measures the TUI with no user tab
paths: [internal/tui/]
timestamp: 2026-10-09T14:47:20Z
---

Build the binary, open a pty sized with TIOCSWINSZ, start it with stdin/stdout/stderr on the slave, cwd at the repo, env `ACTA_TUI_TRACE=<log>` and `TERM=xterm-256color`, and drain the master in a thread. Wait 3 s for startup, send keys, then wheel notches as SGR escapes `\x1b[<65;X;YM` (down) or `\x1b[<64;X;YM` (up), 1-based X/Y. Read `tail -1` of the log.

Gotchas: on the Plans tab the open list may be empty, so select with `5`, tab, `G`, tab. An unscrolled pane logs `frames=0`. A pty has no emulator, so herdr or Ghostty paint cost is not in the numbers.

To time startup, start the binary with `pty.fork`, so it is the foreground process of the pty. Answer the color query `\x1b]11;?` and the cursor query `\x1b[6n` the moment they show, and take the time of the first `\x1b[?1049h`. A child in its own session skips Bubble Tea's color query, and an unanswered query waits up to 5 s, so both read wrong. See /tui-first-frame.md for the budget.
