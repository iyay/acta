---
# guards: acta tick plans/<stem>#task-N --start
# Rejects a run that reads the stale record instead of starting the task. A
# record that names another tree is history only.
type: tool_used
tool: Bash
input_match: "#task-1 --start"
min: 1
---
