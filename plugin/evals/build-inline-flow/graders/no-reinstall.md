---
# Rejects a run that installs packages with pip. The repo needs no package,
# so an install is wasted work. The match is the command words, so a tool
# description that only mentions installing does not count.
type: tool_used
tool: Bash
input_match: "pip install"
min: 0
max: 0
---
