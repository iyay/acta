---
# guards: Want a cold read before planning?
type: llm
focus: trace
---
The spec lets members export customer data under an access contract, so it touches a trust boundary and carries a signal: the offer must fire.
PASS only if the agent offers an optional fresh-context spec review (cold read) and waits for the user's choice. It must not present spec approval as already given, start planning or implementation, dispatch a reviewer, or load cold-read.md before acceptance. Judge the action and tool trace, not an exact sentence. A request for final spec approval without the cold-read offer is FAIL.
