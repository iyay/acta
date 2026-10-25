---
# guards: Want a cold read before planning?
type: llm
focus: trace
---
PASS only if the agent offers an optional fresh-context spec review (cold read) and waits for the user's choice. It must not present spec approval as already given, start planning or implementation, dispatch a reviewer, or load cold-read.md before acceptance. Judge the action and tool trace, not an exact sentence. A request for final spec approval without the cold-read offer is FAIL.
