---
# guards: acta commit <path> -m "<message>"
# Rejects a run that never calls acta commit: one that edits the spec and
# stops, or that commits some other way.
type: tool_used
tool: Bash
input_match: "acta commit"
min: 1
---
