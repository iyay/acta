---
# guards: both from [code-reviewer.md](code-reviewer.md)
# Rejects a run that stops with no verdict.
type: regex
pattern: "CLEAN|BLOCKED"
target: last_message
---
