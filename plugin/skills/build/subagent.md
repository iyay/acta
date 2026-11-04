# Subagent: a fresh worker per task

Read this only when `acta:build` runs the `subagent` executor. On omp, workers run as `agent()` with `agent="task"`. The task loop, close and reply-back rules are in SKILL.md.

## Models

**Subagent models.** Name no model for any subagent; the harness default picks it. On omp, workers run as `agent="task"` and its role config picks the model. The first line of every subagent report names its model; missing, discard the report and dispatch again.

## Worker hand-off

One task, one dispatch, built from [implementer-prompt.md](implementer-prompt.md): worktree path, branch and parent branch, plan path and task number, files, verify line, the trailer line verbatim, the `acta state set` step, the absolute path of `references/house-rules.md` (two folders up from this skill), and the absolute path of build's `SKILL.md`. The worker's first action: print `pwd` and `git rev-parse --show-toplevel`; both must equal the worktree path, else stop with zero writes. Workers never spawn subagents: no helpers, never a reviewer.

Inside a wave, dispatch every task in one message; serial only for a shared file or a real dependency. Reports are DONE, DONE_WITH_CONCERNS, BLOCKED or NEEDS_CONTEXT. DONE: check the commit and the verify line. Concerns about correctness or scope get answered before moving on. NEEDS_CONTEXT or BLOCKED: give the context, split the task, or use a stronger model; never retry the same dispatch unchanged, and never move on while the plan depends on it.

## Fix rounds

A fix round goes to a new implementer, dispatched the same way.
