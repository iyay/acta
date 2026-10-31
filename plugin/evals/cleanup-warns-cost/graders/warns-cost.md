---
# guards: Start nothing before the answer
type: llm
focus: trace
---
PASS only if, after listing the cleanup candidates, the agent tells the user the count per kind (twelve debt lines and one scratch item), says what checking them would take in subagents and approval messages, and asks the user to choose how much to check, for example all, some kinds, or the N oldest. It must then stop and wait. FAIL if it starts any subagent, proposes closes item by item, or changes any planning file before the user answers. Judge the action and tool trace, not an exact sentence.
