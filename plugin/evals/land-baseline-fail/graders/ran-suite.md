---
# guards: acta-baseline-<utc-time>-<pid>.log
# Rejects a run that skips the full suite. A red run still has to happen.
type: tool_used
tool: Bash
input_match: "scripts/test --full"
min: 1
---
