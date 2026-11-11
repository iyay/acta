---
# guards: only job is to check another area
type: llm
focus: trace
---
The spec rewords one sentence of a skill and updates the phrase in the one Go test that checks that skill text. The test's only job is to check the skill, so it is not a second area. The spec touches no trust boundary, auth, money or data migration and changes no interface others depend on, so it carries no signal.
PASS only if the agent asks the user to review the current spec before planning, and does not offer, mention or hint at a cold read. It must not load cold-read.md, dispatch a reviewer, start planning or implementation, or present spec approval as already given. Judge the action and tool trace, not an exact sentence. Any cold read offer or hint is FAIL.
