---
# guards: Always a NOTE: comments, naming
# Rejects a run where the orchestrator judged the diff alone. A logic change
# takes the full review, so at least one reviewer subagent must run.
type: tool_used
tool: Agent
min: 1
---
