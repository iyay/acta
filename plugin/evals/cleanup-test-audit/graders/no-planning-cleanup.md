---
# A test audit skips the planning steps, so acta cleanup never runs, and no
# debt item is written before the user says yes.
type: tool_used
tool: Bash
input_match: "acta cleanup|acta debt new"
min: 0
max: 0
---
