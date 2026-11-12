---
# guards: both from [code-reviewer.md](code-reviewer.md)
# A reviewer that says CLEAN must show what it read and ran. The reviewer
# brief asks for that under a Coverage heading, so some Agent call must carry
# that heading in its prompt. The match is a case-sensitive substring.
type: tool_used
tool: Agent
input_match: "### Coverage"
min: 1
---
