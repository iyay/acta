---
id: SCR-0068
hash: yv49k55
title: Setup asks whether to install the TUI
status: raw
created: "2026-10-15 09:09:17"
schema: "1"
started: "2026-10-15 09:09:19"
finished: "2026-10-15 09:16:52"
---
# Setup asks whether to install the TUI

## Words

### 2026-10-15

Setup asks whether to install the TUI. Users who only want the skills and hooks get a CLI-only acta binary (no Bubble Tea), since every skill needs the CLI but none needs the TUI. Today acta is one 22.6M binary: no args opens the TUI, subcommands run the CLI.

## Context

## Log

### 2026-10-15

Q1 why no TUI: binary size. User wants a smaller binary (option 1).

### 2026-10-15

Measured 2026-10-15: release build has no -s -w. Plain build 24M, with -s -w 17M (-7M). TUI deps (bubbletea, lipgloss, fsnotify) add about 1.5M stripped, plus internal/tui code (21k lines), so a no-TUI build saves roughly 2-3M only.

### 2026-10-15

Decision 2026-10-15: strip the release binary (-s -w) as a separate Bounded change. The no-TUI build is on hold; reopen only if 17M is still too big.

## Open questions
