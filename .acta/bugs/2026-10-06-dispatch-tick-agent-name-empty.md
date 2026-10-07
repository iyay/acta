---
id: BUG-0033
hash: gf2xwrb
fixed_in: 9e13188
finished: "2026-10-06 06:19:07"
---
# Tasks run by a dispatched omp agent show no agent name

## Symptom
In the TUI, tasks and subtasks of a plan built through dispatch show no agent tag. `.acta/.agents.json` in the worktree records every task with `"agent": ""` and `"started": true` (seen on PLN-0097, all four tasks).

## Root cause
`acta tick` takes the name from `--agent`, else from the `AI_AGENT` variable (`internal/write/agents.go:22-26`). omp sets no `AI_AGENT`, and its ticks ran without `--agent omp`. The only place that asks for the flag is a conditional note in the build skill ("omp: add `--agent omp`", `plugin/skills/build/SKILL.md:145`, `implementer-prompt.md:41`); since 0ef7770 the dispatch brief does not repeat it, and omp's task subagents skip the note.

## Repro
1. `acta dispatch send --plan <plan>` from a worktree to an omp tab.
2. Let omp run one task.
3. `cat .acta/.agents.json` in the worktree: `"agent": ""`; the TUI row shows no agent.

## Found in
setup-block-refresh worktree (PLN-0097), 2026-10-06; user noticed in the TUI, confirmed by reading .agents.json.
