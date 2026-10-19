---
# guards: `git merge --no-ff <branch>`
# Rejects a run that saved the record but never landed the branch, whatever
# merge flags it used.
type: tool_used
tool: Bash
input_match: " merge (--[a-z-]+ )*land\\b"
min: 1
---
