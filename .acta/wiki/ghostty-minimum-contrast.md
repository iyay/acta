---
type: Gotcha
title: Ghostty minimum-contrast eats dim colors
description: Ghostty swaps any fg below 1.5 contrast vs bg to white or black, so dim colors need a ratio of 1.6 or more
paths: [internal/theme/]
timestamp: 2026-10-05T15:28:00Z
---

Ghostty with `minimum-contrast = 1.5` swaps the fg of any cell below 1.5 contrast against its bg for white or black. A dim popup color at 1.34 looked bright white while tests and pty bytes were correct.

Fix: any TUI color meant to look dim keeps a WCAG ratio of at least about 1.6 against every theme bg. For a bug tests cannot see, read the real pane with `herdr pane read <pane> --format ansi` and check the Ghostty config before changing code.
