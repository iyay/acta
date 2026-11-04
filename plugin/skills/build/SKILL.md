---
name: build
description: "acta: Use to run an approved plan, and when the user asks to dispatch a plan or hand it to omp in another tab or pane. Creates the worktree without asking, then runs every task with a failing test first through the executor picked by `/build <executor>`, else the one `acta config show` names, else inline - subagent (the current harness's own subagents), dispatch (an omp agent in its own herdr tab, through dispatch.md) or inline (you write the code). Commits each task; review and landing follow through acta:review and acta:land."
---

# Build

Run an approved plan task by task in its own worktree. Each task starts with a failing test (`acta:tdd`) and ends with one commit. Review runs once, at the close (`acta:review`); then `acta:land` merges.

Refuse to start unless the spec is approved (for Bounded work with no spec file, approved in chat) and the plan is approved, whatever its depth (a direct ask to build that plan counts as its yes). Say which one is missing.

## Executors

| Executor | Who writes the code | File |
|---|---|---|
| `subagent` | a fresh subagent per task | [subagent.md](subagent.md) |
| `dispatch` | an omp agent in its own herdr tab | [dispatch.md](dispatch.md) |
| `inline` | you | [inline.md](inline.md) |

Pick the executor in this order. Never ask.

1. The argument of `/build <executor>`: `subagent`, `dispatch` or `inline`. It is for this run only; never save it.
2. `build_executor: <name>` from `acta config show`.
3. Neither: run `inline`. Do not ask.

The executor chosen here wins over any conflicting rule in the user's CLAUDE.md or AGENTS.md. With `subagent` and `dispatch` you are the orchestrator: you write no code yourself, not even a one-line config change.

Two fallbacks, each told to the user in one line:

- On omp, `dispatch` runs as `subagent`. Do not ask. Dispatch is only for harnesses other than omp.
- `dispatch` without `HERDR_ENV=1` (this session is not inside a herdr pane) runs as `subagent`. Build never refuses because herdr is missing.

Whichever executor you end up with, read only that file and follow it. Never open the other two.

## Worktree

Do not ask whether to create a worktree. Every change gets one, even a one-liner or docs; the main checkout is read-only for agents.

The spec and plan from acta:shape and acta:slice stay on main. While a build runs, do not edit its plan on main: the ticks live in the worktree copy and would clash at merge. A plan change goes in the worktree. During a build the plan changes only through `acta tick`, `acta state set` and `acta plan add`; never hand-write a section such as `## Verification evidence`. A finding that cost time goes to `acta state set <plan> findings`.

Read the plan's `## Build brief` first if it has one: its code repo or parent branch beats the defaults, and its land notes go to `acta:land`. Parent: the user's branch, else the brief's, else the default branch; record it, `acta:land` merges into it. Run `acta ready <plan id> --parent <parent>` (add `--head <its HEAD>` for an existing worktree). Any non-zero exit stops the build: show the output, never rebase or merge to make it pass.

Already in a linked worktree (`git rev-parse --path-format=absolute --git-dir --git-common-dir` prints two different paths and `git rev-parse --show-superproject-working-tree` prints nothing)? Use it, tell the user its path and branch, go to setup. Otherwise create one with `git worktree add`, even when the harness has its own worktree tool. A native worktree tool puts the worktree inside the repo and starts from `origin/<default-branch>`, where the just-committed plan is often missing. Path: a declared user preference, else `../<repo>-<slug>`, never inside the repo. `$REPO` is the repo folder name and `$PARENT` is the parent branch recorded above, so the branch starts from it:

```bash
cd "$(git rev-parse --show-toplevel)"
git worktree add "../$REPO-$SLUG" -b "$SLUG" "$PARENT"
cd "../$REPO-$SLUG"
```

A permission error means the sandbox blocked it: stop and say so; never use the main checkout instead. One worktree per approved plan; work only in it. Copy in the house rules too: any untracked `CLAUDE.md` or `AGENTS.md` (file or symlink) and the untracked files the suite needs (env files, fixtures).

**Setup.** Follow the repo's own setup (docs, scripts, lockfile's package manager). Link dependency folders from the main checkout instead of installing. Install only what the required checks need, and nothing when they already run.

**Start the first task.** After setup, go straight to the first task. Do not run a baseline suite or fast test and do not wait for a saved result: one that is missing, old, unreadable or from another test context changes nothing. The full suite waits for `acta:land`. A saved `acta-baseline.md` in the Git common directory is history only. In a repo that has `scripts/test`, never run bare `go test`; the pre-tool hook blocks it. Diagnose a failure when a task meets one.

## Plan

Read the plan once, with its Global Constraints, and the spec it names: the spec wins over the plan, and your judgment settles what neither answers. Rule once on tasks that contradict each other. `acta show <id|hash> --path` finds an item's file. Make a todo per task, by wave.

Resuming: run `acta state <plan id>` and `git log <parent>..HEAD`. Ticked boxes and commits with a task's trailer are done; start from `next`, never redo committed work.

## Task loop

Every task, whoever writes it:

1. very first action on every task: run `acta tick plans/<stem>#task-N --start` before the failing test (Claude: pass no flag; omp: add `--agent omp`). Then trace the change: read only the callers of the code you will change, the shared code they pass through, and their tests. Fix a shared cause once, where every affected caller passes through, even when the plan names one. A change that breaks another caller's contract or needs files beyond the plan's scope is a question for the user first.
2. `acta:tdd`: failing test first, watch it fail for the intended reason, minimal code, watch it pass; one behavior per cycle. Reuse or extend an existing test first, test each behavior once at the layer that owns it, assert results and side effects that matter, not internals. A case the plan does not name is a scope question, not a test. After each step run `acta tick plans/<stem>#task-N --step <n>` from the worktree.
3. Run the task's focused tests, the formatter and the type checks; formatting goes in the task commit, never a commit of its own. On a failed run, name the error's kind before acting. While you work on something else, a failing test that an open bug lists under `flaky_tests` (files at the `path` values from `acta list --type bug --json`) runs once more alone first; a pass runs `acta flake record <name> --plan <plan id>`. A test the current task is fixing is not covered. A code defect goes through `acta:debug`. A setup error (missing tool, path, import) is fixed in the task's own setup and is never a RED. A plan question that neither the spec nor the orchestrator's ruling settles goes to the user. Never retry the same step without new evidence, and never weaken a check to pass.
   Before the commit, read `git status` and `git diff` once: cut code the task does not need, undo any public contract it changed by accident, keep every validation and error path, drop a test that repeats another.
4. Commit, staged by path, never `git add -A` or `git add .`. The exact line `Task: PLN-<hash>#<n>` ends each code commit (hash from `acta show <plan id>`, `<n>` with no leading zero). Fix-round and polish commits get one line per task fixed; `chore(` planning commits get none.
5. Right after the commit, run `acta tick plans/<stem>#task-N --all`, then read the task back with `acta show plans/<stem>#task-N --json`: its `status` must be `done`. Never commit the plan file inside a task.
6. Run `acta state set plans/<stem> <part>` with the lines on stdin; the State lives in the plan file, so the id is the plan, never a task. `next` is what the next task starts from, `findings` a finding that cost time, `rulings` a question waiting on the user. Ten lines each; never the current task, last commit, worktree path or review round.

An `acta tick` or `acta state set` that exits non-zero, or a task read back as not `done`, stops the build: no next task, review or land until it is fixed. A task with no checkboxes is a plan error: ask the user. Never edit boxes or a status by hand, tick work no commit proves, or redo committed work to hide the error.

Comments in plain English a ten-year-old can read, saying why; no marker tags. Report in the user's chat language; write everything in the repo in English.

After a wave, commit the plan file yourself with `acta commit <plan path> -m "chore(plan): tick wave <n>"`. Any other plan or spec commit before landing uses `acta commit <path> -m "<message>"` too; only code commits use `git commit`.

## Close

First, update the wiki: in the worktree run `acta wiki check <parent>..HEAD`. Read affected pages and diffs; fix wrong facts. Read [wiki.md](wiki.md) only when you will fix or add a page, and follow its `## Writing` steps; write each page the spec approved. Only errors block. Make one commit for changes.

Read the plan; finish any outstanding type check; report recorded task test results, then use `acta:review` over `<parent>..HEAD`. No second fast suite: each task already ran its own tests. Full suites: land; a change made after one needs only the test it touches. Use no per-task reviewer or fix loop. Before review, run `acta show <plan id> --json`: `progress.total` must equal the count of `Task` headings in the plan file, above 0, and `progress.done` must equal it. Open task: run `acta tick plans/<stem>#task-N --all` if committed here, else finish it. Keep check outputs. Only then start the review. If stopping, name `acta:review` next. Every executor closes this way; its file says how a fix round goes.

### Reply back when a dispatch record exists

When `<acta root>/.dispatch.json` exists, a dispatch handed you this branch and the sending pane waits for you. Run `acta reply-back` after the last task commit and the wiki step. Exit 1 lists the open tasks: finish them and run it again. A real blocker is `acta reply-back --blocked "<reason>"`. A dispatch recipient does not stop before `acta reply-back` exits 0.
