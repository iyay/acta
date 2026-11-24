---
# guards: Run `acta init` at the repo, then report what it printed.
# Rejects a run that never runs acta init, the one thing the skill does.
type: tool_used
tool: Bash
input_match: "acta init"
min: 1
---
