---
id: SCR-0017
hash: s4h6zmf
title: Dim main window text when popup opens
status: raw
created: "2026-09-29"
---
ketika popup muncul, text color di main window harusnya redup. bukan cuman border


---
Agent notes (2026-09-29, found by reading, not by running the TUI):
- The code already tries to dim everything. `cover` in internal/tui/view.go:602 strips the colors from every line behind the popup and repaints each one with `dim`. The status line gets the same treatment at view.go:128-132.
- `dim` at view.go:68 is `Faint(true)` plus foreground 250 (light) / 240 (dark). The user says only the border looks dimmed, so the text most likely ends up looking about the same as before. Not proven yet. Maybe the terminal ignores faint, or 240 is close to the normal text color. Check with acta:debug before changing anything.
- Popups: help `?`, the picker, and the others drawn by `popupBox` (view.go:555).
- Open question: how dim is "redup"? A darker grey, or a fixed low-contrast color from the theme work (SCRATCH-2 / SPEC-20)?
