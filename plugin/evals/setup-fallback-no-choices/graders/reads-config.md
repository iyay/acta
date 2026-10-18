---
# Rejects a run that never reads the config, so it cannot know which
# settings are missing or still on their default.
type: tool_used
tool: Bash
input_match: "config show"
min: 1
---
