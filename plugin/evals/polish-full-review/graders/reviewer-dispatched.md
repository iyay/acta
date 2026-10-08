---
# guards: ## Review tiers
# Rejects a run that reviews the polish alone. The commit message says polish
# but one logic line changed, so the diff picks the full tier, and the full
# review is two reviewer subagents. Every reviewer brief names its axis as
# Spec or Standards. The match is a case-sensitive substring, so this one
# counts the Spec reviewer. The Standards reviewer has its own grader.
type: tool_used
tool: Agent
input_match: "Spec"
min: 1
---
