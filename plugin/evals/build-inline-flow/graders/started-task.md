---
# guards: acta tick plans/<stem>#task-N --start
# Rejects a run that reads the plan but never starts its first task. The
# inline executor ticks before the failing test, not after the code.
type: tool_used
tool: Bash
input_match: "#task-1 --start"
min: 1
---
