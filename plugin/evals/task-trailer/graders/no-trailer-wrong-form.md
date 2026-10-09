---
# guards: the trailer form, plan by hash and task with no leading zero
# Rejects a run that writes the plan's short id or a zero-padded task number
# in the trailer, since the board and acta commits match the hash form only.
type: tool_used
tool: Bash
input_match: "Task: PLN-0001"
min: 0
max: 0
---
