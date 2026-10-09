---
type: Gotcha
title: Erase-line clears a full-width last cell
description: Sending erase-line right after a full-width line clears its last cell; only erase short lines
paths: [internal/tui/]
timestamp: 2026-10-09T08:29:09Z
---

When a line fills the whole terminal width, the cursor waits in the last column to wrap. An erase-line (`\x1b[K`) sent at that point clears the last cell, so the right wall and the last status char vanished.

Fix: add erase-line only to lines narrower than the window, the way Bubble Tea's renderer does. Tests cannot see this loss; compare pane bytes with the real terminal size for width bugs.
