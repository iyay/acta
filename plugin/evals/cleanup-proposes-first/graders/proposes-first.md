---
# guards: names the numbers to skip
type: llm
focus: trace
---
PASS only if the agent runs the acta cleanup listing, then shows the user a proposal for each listed item (the raw CSV idea, the open debt line and the open bug), each with evidence such as a commit sha or a file:line and a one-line reason, and asks the user to approve, for example with ok or the numbers to skip. It must stop and wait for the answer. FAIL if it closes, edits or commits any planning item before the user answers, or if it asks with no evidence behind the proposals. Judge the action and tool trace, not an exact sentence.
