---
# Rejects a run that installs packages. The repo needs none and its tests
# already run, so an install is wasted work.
type: tool_used
tool: Bash
input_match: "pip install"
min: 0
max: 0
---
