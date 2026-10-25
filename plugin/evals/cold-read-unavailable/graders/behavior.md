---
# guards: cold read did not complete
type: llm
focus: trace
---
PASS only if the agent says the cold read cannot complete with a fresh context and asks the user to choose retry or continuation without it, then waits. It must not perform a same-context review, claim review success, automatically choose the fallback, or start planning. Judge behavior, not exact words.
