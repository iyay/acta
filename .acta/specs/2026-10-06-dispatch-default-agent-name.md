---
parent: bugs/2026-10-06-dispatch-tick-agent-name-empty
id: SPC-0088
created: "2026-10-06 05:33:52"
hash: doa6fo8
---
# Ticks in a dispatched worktree name the agent the plan was sent to

Status: Bounded, approved by the user in chat on 2026-10-06. Fixes BUG-0033.

Why: `acta tick` takes the agent name from `--agent`, else `AI_AGENT` (`internal/write/agents.go`, `AgentName`). omp sets no `AI_AGENT`, and its task subagents skip the `--agent omp` note in the build skill, so every task of a dispatched plan is recorded with `"agent": ""` and the TUI shows no agent.

Design:
- `acta dispatch send` writes the recipient's name (`omp`) into the worktree's `.acta/.agents.json` under a reserved key `"*"`, with the same lock and temp-file rename `RecordAgent` uses.
- When a tick has no name from the flag or `AI_AGENT`, `RecordAgent` uses the `"*"` name, unless the task already holds a name from an earlier tick.
- Readers of `.agents.json` (board, TUI) skip the `"*"` key, so it never shows as a task.
- A worktree with no `"*"` key behaves as today.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: setting `AI_AGENT` inside omp; changing the build skill text.

Tests: after a dispatch-style default is written, a tick with no flag and no `AI_AGENT` records `omp`; a tick with `--agent x` still records `x`; the board loads no task named `"*"`.
