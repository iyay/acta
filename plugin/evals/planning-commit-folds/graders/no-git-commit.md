---
# guards: acta commit <path> -m "<message>"
# Rejects a run that commits the spec edit with git commit. That leaves a second
# commit for one spec file instead of one folded commit.
type: tool_used
tool: Bash
input_match: "git commit"
min: 0
max: 0
---
