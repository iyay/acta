---
name: routing-light-01
description: One wrong word in src/greet.py should take the light path, not an Architectural brainstorm.
tags: [routing]
runs: 3
max_turns: 6
timeout_seconds: 240
allowed_tools: []
---

One line in src/greet.py is wrong. It returns 'Helo, ' + name, so the
greeting comes out misspelled. The fix is one word in that one line.

Does this change need the full process with competing approaches and a
sectioned write-up, or is the short route enough? Answer in a few lines and
stop. Do not change any file.

The acta CLI for this repo is ./bin/acta. It is on PATH as well.
