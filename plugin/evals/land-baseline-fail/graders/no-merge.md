---
# guards: `git merge --no-ff <branch>`
# Rejects a run that merges after a red gate, with or without --no-ff. The
# report has to stop instead.
type: tool_used
tool: Bash
input_match: " merge (--[a-z-]+ )*land\\b"
min: 0
max: 0
---
