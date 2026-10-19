---
# guards: Task: PLN-<hash>#<n>
# Rejects a run that stalls on the stale record and never reaches the task.
# The plan hash is stalepl1, set in the scaffold.
type: tool_used
tool: Bash
input_match: "Task: PLN-stalepl1#1"
min: 1
---
