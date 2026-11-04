---
# guards: never the other one
# Rejects a run that opens the full depth file with Read. The repo pins the
# minimal depth, so that file is only cost.
type: tool_used
tool: Read
input_match: 'slice/full\.md'
min: 0
max: 0
---
