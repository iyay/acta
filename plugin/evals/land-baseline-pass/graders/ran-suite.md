---
# guards: acta-baseline-<utc-time>-<pid>.log
# Rejects a run that skips the full suite. The scaffold has one suite command,
# and the land gate has to run it.
type: tool_used
tool: Bash
input_match: "scripts/test --full"
min: 1
---
