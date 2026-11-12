---
# guards: ## Review depth
# The diff is one line, but it changes a condition and no test covers it.
# Size is no reason for a lower level: a logic change with no test that went
# red first must get at least one reviewer subagent. A run that reviews it
# alone makes no Agent call and fails here.
type: tool_used
tool: Agent
min: 1
---
