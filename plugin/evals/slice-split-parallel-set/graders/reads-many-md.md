---
# guards: `parallel_plans: subagent` or `build_executor: subagent`
# Rejects a run that never opens the parallel build file. With the setting on,
# slice goes straight to the parallel build. Agents open the file with Read or
# with cat in Bash, so this looks for its path anywhere in the trace. No skill
# text holds the path skills/build/many.md, so only a real open puts it there.
type: regex
pattern: "skills/build/many\\.md"
target: trace
---
