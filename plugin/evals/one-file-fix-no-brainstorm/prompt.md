---
name: one-file-fix-no-brainstorm
description: A one-word typo fix in one file takes the light path, not the Architectural path.
tags: [brainstorm]
runs: 1
max_turns: 2
timeout_seconds: 120
allowed_tools: []
---

One line in src/greet.py is wrong. It returns 'Helo, ' + name, so the
greeting comes out misspelled. The fix is one word in that one line.

Does this change need the full Architectural design process, or is it the
light path? Answer in a few lines and stop. Do not change any file.
