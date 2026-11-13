---
# guards: ## Review tiers
# Match declared reviewer launches in actual tool inputs; resumes reuse an agent.
type: tool_used
tool: Agent
input_match: '^(?![\s\S]*"resume"\s*:\s*"[^"]+")[\s\S]*"prompt"\s*:\s*"ACTA_REVIEW_AXIS=STANDARDS(?:\\n|\\r\\n| )'
min: 1
max: 1
---
