---
# Rejects a run that installs packages with any tool, such as pip, pip3 or
# poetry. The repo needs no package, so an install is wasted work.
type: tool_used
tool: Bash
input_match: "install"
min: 0
max: 0
---
