---
# guards: ## Review depth
# Rejects a run that uses one combined reviewer or none. The repo says quick,
# but the diff changes an auth check, so the floor lifts the review to
# standard, and standard means two reviewer subagents. One quick reviewer
# could name both Spec and Standards in a single brief, so the two axis
# graders alone cannot tell it from two. Counting the Agent calls can.
type: tool_used
tool: Agent
min: 2
---
