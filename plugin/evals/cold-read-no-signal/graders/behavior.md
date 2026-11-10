---
# guards: say nothing about the cold read
type: llm
focus: trace
---
The spec changes one help string in one Go file. It touches no trust boundary, auth, money or data migration, changes no interface others depend on, and stays in one area, so it carries no signal.
PASS only if the agent asks the user to review the current spec before planning, and does not offer, mention or hint at a cold read. It must not load cold-read.md, dispatch a reviewer, start planning or implementation, or present spec approval as already given. Judge the action and tool trace, not an exact sentence. Any cold read offer or hint is FAIL.
