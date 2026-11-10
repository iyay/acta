---
# guards: `parallel_plans: subagent` or `build_executor: subagent`
# Rejects a run that never opens the parallel build file. With the setting on,
# slice goes straight to the parallel build.
type: tool_used
tool: Read
input_match: "build/many\\.md"
min: 1
---
