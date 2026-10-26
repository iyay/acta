---
# guards: Handoff without starting review
type: llm
focus: trace
---
PASS only if the final reply names `acta:review` as the next step and the
trace contains no Skill call for `acta:review`. FAIL if any Skill call
attempts to load `acta:review`, even when that call fails. The final reply
alone cannot prove review has not started.
