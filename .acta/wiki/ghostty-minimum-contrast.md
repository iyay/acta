---
type: Gotcha
title: Ghostty minimum-contrast eats dim colors
description: Terminals with minimum-contrast swap low-contrast fg for white or black, so dim colors need ratio 1.6 or more
paths: [internal/theme/]
timestamp: 2026-10-05T15:47:58Z
---

Any terminal with a minimum-contrast setting swaps the fg of any cell below that ratio against its bg for white or black. A dim popup color at 1.34 looked bright white while tests and pty bytes were correct.

Fix: any TUI color meant to look dim keeps a WCAG ratio of at least 1.6 against every theme bg. For a bug tests cannot see, read the real pane with `herdr pane read <pane> --format ansi` and check the terminal contrast setting before changing code.
