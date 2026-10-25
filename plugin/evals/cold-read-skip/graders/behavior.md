---
# guards: Skip goes to final spec approval
type: llm
focus: trace
---
PASS only if the agent honors the declined cold read and requests final approval of the current spec before planning. It must not offer cold read again, load cold-read.md, dispatch, apply edits, invoke slice, or create a plan. Judge actions and trace, not exact wording.
