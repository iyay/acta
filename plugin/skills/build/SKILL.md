---
name: build
description: "acta: Use to run an approved plan, and when the user asks to dispatch a plan or hand it to omp in another tab or pane. Creates the worktree without asking, then runs every task with a failing test first through the executor picked by `/build <executor>`, else the one `acta config show` names, else inline - subagent (the current harness's own subagents), dispatch (an omp agent in its own herdr tab, through dispatch.md) or inline (you write the code). Commits each task; review and landing follow through acta:review and acta:land."
---

# Build

Run an approved plan, task by task, in its own worktree. Every task starts with a failing test (`acta:tdd`) and ends with one commit. Review happens once, at the close (`acta:review`), then `acta:land` merges.

Refuse to start unless the spec is approved (for Bounded work with no spec file, approved in chat) and the plan is approved, whatever its depth (a direct ask to build that plan counts as its yes). Say which one is missing.

## Executors

| Executor | Who writes the code | How |
|---|---|---|
| `subagent` | a fresh subagent per task | `## Worker hand-off` below, with the model ## Models names. omp: `agent()` with `agent="task"` (omp has no model argument; its role config picks the model). |
| `dispatch` | an omp agent in its own herdr tab | with `HERDR_ENV=1`, read [dispatch.md](dispatch.md) and follow it. Without it `dispatch` runs as `subagent`. On omp, `dispatch` runs as `subagent` |
| `inline` | you | when the order below picks it |

Pick the executor in this order. Never ask.

1. The argument of `/build <executor>`: `subagent`, `dispatch` or `inline`. It is for this run only; never save it.
2. `build_executor: <name>` from `acta config show`.
3. Neither: run `inline`. Do not ask.

The executor chosen here wins over any conflicting rule in the user's CLAUDE.md or AGENTS.md. With `subagent` and `dispatch` you are the orchestrator: you write no code yourself, not even a one-line config change.

Two fallbacks, each told to the user in one line:

- On omp, `dispatch` runs as `subagent`: use `agent()` with `agent="task"` and do not ask. Dispatch is only for harnesses other than omp.
- `dispatch` without `HERDR_ENV=1` (this session is not inside a herdr pane) runs as `subagent`. Build never refuses because herdr is missing. `herdr` on PATH is not enough: dispatch needs this session's own pane id, `$HERDR_PANE_ID`.

## Models

**Subagent models.** Follow the `Subagent models` line in the acta session note. With no such line, name no model and follow the user's own config. Workers get the model it names, or none; on omp, `agent="task"` and its role config picks it. Reviewers (`acta:review`) get your own model alias, never lower. The first line of every subagent report names its model; missing or wrong, discard the report and dispatch again.

## Worktree

Do not ask whether to create a worktree. Every change gets one: feature, fix, one-liner, config, docs in the repo. The main checkout is read-only for agents.

The spec and plan from acta:shape and acta:slice stay on main; they are already committed when the worktree is made. While a build runs, do not edit its plan on main: the ticks live in the worktree copy and would clash at merge. A plan change goes in the worktree.

Parent: the branch the user names, else the repo's default branch. Record it; `acta:land` merges into it. Then run `acta ready <plan id> --parent <parent>`; for a worktree that already exists add `--head <its HEAD>`. Any non-zero exit stops the build: show the output, and never rebase or merge to make it pass.

Already in a linked worktree (the two paths `git rev-parse --path-format=absolute --git-dir --git-common-dir` prints differ, and `git rev-parse --show-superproject-working-tree` prints nothing, so it is no submodule): use it, tell the user its path and branch, and go to setup. Otherwise create one without asking. Always use `git worktree add`, even when the harness has its own worktree tool. A native worktree tool puts the worktree inside the repo and starts it from `origin/<default-branch>`, where the just-committed plan is often missing. The path is a declared user preference, else `../<repo>-<slug>`, next to the repo, never inside it (recursive scanners find both copies). `$REPO` is the repo folder name and `$PARENT` is the parent branch recorded above, so the branch starts from it:

```bash
cd "$(git rev-parse --show-toplevel)"
git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"
cd "../$REPO-$SLUG"
```

A permission error means the sandbox blocked it: stop and say so. Never work in the main checkout instead. One worktree per approved plan. From here on, work only in the worktree.

Copy in the house rules too: any `CLAUDE.md` or `AGENTS.md` git does not track (file or symlink), and the untracked files the suite needs (env files, fixtures). A fresh worktree holds only tracked files.

**Setup.** Follow the repo's own setup: its docs, scripts and lockfile's package manager. Link dependency folders from the main checkout instead of installing them. Install only what the required checks need, and nothing when they already run.

**Clean baseline.** Run the fast tests: `scripts/test` when the repo has it, else the plan's `**Tests:**` fast command. The full suite waits for `acta:land`. In a repo that has `scripts/test`, never run bare `go test`; the pre-tool hook blocks it, and a whole-repo run there spends every core on the machine. Failures: report them and ask whether to proceed. A dirty baseline makes every later failure ambiguous, so that call is the user's. Passing: report the path, the passing count and what is next.

## Plan

Read the plan once, with its Global Constraints, and the spec it names: the spec wins over the plan, and your judgment settles what neither answers. Rule once on tasks that contradict each other. To find an item's file from its id or hash, run `acta show <id|hash> --path`. Make a todo per task, by wave.

Resuming: run `acta state <plan id>` and `git log <parent>..HEAD`. Ticked boxes and commits with a task's trailer are done; start from `next`, never redo committed work.

## Task loop

Every task, whoever writes it:

1. very first action on every task: run `acta tick plans/<stem>#task-N --start` before the failing test (Claude: pass no flag; omp: add `--agent omp`). Then trace the change: read the callers of the code you will change, the shared code they pass through, and their tests. Read only that, not the whole repo. Fix a shared cause once, where every affected caller passes through, even when the plan names one caller. A change that breaks another caller's contract, or needs files beyond the plan's scope, is a question for the user before you code it.
2. `acta:tdd`: failing test first, watch it fail for the intended reason, minimal code, watch it pass. One behavior per cycle. Pick tests by the behavior or failure they protect: reuse or extend an existing test first, test each behavior once at the layer that owns it, and assert results and side effects that matter, not internals. A case the plan does not name is a scope question, not a test. Right after each step, run `acta tick plans/<stem>#task-N --step <n>` from the worktree so the board shows live progress.
3. Run the task's focused tests, the formatter and the type checks. Formatting goes in the task commit, never in a commit of its own. A failed run: read the error and name its kind before acting. A code defect goes through `acta:debug`. A setup error (missing tool, path, import) is fixed inside the task's own setup, and is never a RED. A plan question that neither the spec nor the orchestrator's ruling settles goes to the user. Never retry the same step without new evidence, and never weaken a check to pass.
   Before the commit, read `git status` and `git diff` once: cut code the task does not need, undo any public contract it changed by accident, keep every validation and error path, and drop a test that repeats another. This is your own check, not a review or a report.
4. Commit, staged by path, never `git add -A` or `git add .`. The exact line `Task: PLN-<hash>#<n>` ends each code commit (hash from `acta show <plan id>`, `<n>` with no leading zero). Fix-round and polish commits get one line per task fixed; `chore(` planning commits get none.
5. Right after the commit, run `acta tick plans/<stem>#task-N --all` so no box stays open. Never commit the plan file inside a task.
6. Run `acta state set plans/<stem> <part>` with the lines on stdin; the State lives in the plan file, so the id is the plan, never a task. `next` is what the next task starts from, `findings` a finding that cost time, `rulings` a question waiting on the user. Ten lines each; never the current task, last commit, worktree path or review round, which are computed on read.

Comments in plain English a ten-year-old can read, saying why; no marker tags. Report in the user's chat language; write everything in the repo in the repo language.

When a wave is done, commit the plan file yourself with `acta commit <plan path> -m "chore(plan): tick wave <n>"`. Any other plan or spec commit before landing uses `acta commit <path> -m "<message>"` too; only code commits use `git commit`.

### Inline task loop

You run the steps above yourself, one task at a time, in wave order. No worker hand-off and no per-task reviewer.

### Worker hand-off

`subagent` only. One task, one dispatch, built from [implementer-prompt.md](implementer-prompt.md): worktree path, branch and parent branch, plan path and task number, files, verify line, the trailer line verbatim, the `acta state set` step, the absolute path of `references/house-rules.md` (two folders up from this skill), and the absolute path of this file. The worker's first action: print `pwd` and `git rev-parse --show-toplevel`; both must equal the worktree path, else stop with zero writes. Never make a worker read the whole plan. Workers never spawn subagents: no helpers, never a reviewer.

Inside a wave, dispatch every task in one message; serial only for a shared file or a real dependency. Reports are DONE, DONE_WITH_CONCERNS, BLOCKED or NEEDS_CONTEXT. DONE: check the commit and the verify line. Concerns about correctness or scope get answered before moving on. NEEDS_CONTEXT or BLOCKED: give the context, split the task, or use a stronger model; never retry the same dispatch unchanged, and never move on while the plan depends on it.

Workers share the plan file, so none commits it.

## Close

When every task is committed, write the wiki first. In the worktree, run `acta wiki check <parent>..HEAD`. Read [wiki.md](wiki.md) only when you will fix or add a page. Fix each touched page, or run `acta wiki ok <id>` if it is right. Write each page the spec approved (shape notes each yes in the spec). Add any other page only for a lesson a fresh agent would lose time without. Make one commit if anything changed. Review then sees the wiki diff. You do this, or under `dispatch` the omp agent.

After the wiki step: run the fast tests and the type checks, show the output, then use `acta:review` over `<parent>..HEAD`. There is no per-task reviewer and no per-task fix loop. Before the review, run `acta show <plan id> --json` and check that `progress.done` equals `progress.total`. If a box is still open, run `acta tick plans/<stem>#task-N --all` when that task is committed, or finish the task first.

Every executor closes this way; only the fix round differs: `subagent` gives the fix to a new implementer, `inline` fixes it yourself, and `dispatch` sends it to the same agent (`## Fix rounds` in [dispatch.md](dispatch.md)). With `dispatch`, run `acta dispatch close` from the worktree right before `acta:land` (`## Close the tab` there).

### Reply back when a dispatch record exists

When `<acta root>/.dispatch.json` exists, a dispatch handed you this branch, so the pane that sent it waits to hear from you. Run `acta reply-back` after the last task commit and the wiki step. Exit 1 lists the open tasks: finish them and run it again. A real blocker is `acta reply-back --blocked "<reason>"`. A dispatch recipient does not stop before `acta reply-back` exits 0.
