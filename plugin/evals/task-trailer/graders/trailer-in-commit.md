---
# guards: the Task trailer line in the code commit of a build
# Rejects a run whose commit command never carries the line for this plan's
# hash and task 1. The hash is shoutq3, set in the scaffold. The text is
# matched inside the Bash input, so a commit that only mentions the task in
# its subject does not pass.
type: tool_used
tool: Bash
input_match: "Task: PLN-shoutq3#1"
min: 1
---
