---
name: build
description: Use to run an approved plan. Creates the worktree without asking, then runs every task with a failing test first through one of three executors - subagent (default, the current harness's own subagents), dispatch (an omp agent in its own herdr tab, through pm:dispatch) or inline (you write the code). Commits each task; review and landing follow through pm:review and pm:land.
---

# Build

Run an approved plan, task by task, in its own worktree. Every task starts with a failing test (`pm:tdd`) and ends with one commit. Review happens once, at the close (`pm:review`), then `pm:land` merges.

Refuse to start without an approved spec and an approved plan. Say which one is missing.

## Executors

| Executor | Who writes the code | How |
|---|---|---|
| `subagent` (default) | a fresh subagent per task | Claude Code: the Agent tool with `model: "sonnet"`. omp: `agent()` with `agent="task"` (omp has no model argument; its role config picks the model). |
| `dispatch` | an omp agent in its own herdr tab | follow `pm:dispatch`; it refuses without herdr |
| `inline` | you | only when the user says "inline" |

With `subagent` and `dispatch` you are the orchestrator: you write no code yourself, not even a one-line config change. With every executor the rest holds: worktree first, failing test first, one commit per task, review once at the close.

## Models

- Implementers and other workers: `sonnet` in Claude Code, `agent="task"` in omp. Set it on every dispatch; a dispatch with no model is a mistake even when the default matches.
- Reviewers (`pm:review`): the orchestrator's own model, never lower.
- The first line of every subagent report names the model it ran on. Missing or wrong: discard the report and dispatch again.

## Worktree

Do not ask whether to create a worktree. Every change gets one: feature, fix, one-liner, config, docs in the repo. The main checkout is read-only for agents.

Path: `../<repo>-<slug>`, next to the repo, never inside it (recursive scanners find both copies). Parent: the branch the user names, else the repo's default branch. Record the parent; `pm:land` merges into it. One worktree per approved plan. Copy in the untracked files the suite needs (env files, fixtures) and link dependency folders instead of installing them.

### Step 0: Detect Existing Isolation

Before creating anything, check if you are already in an isolated workspace.

```bash
GIT_DIR=$(cd "$(git rev-parse --git-dir)" 2>/dev/null && pwd -P)
GIT_COMMON=$(cd "$(git rev-parse --git-common-dir)" 2>/dev/null && pwd -P)
BRANCH=$(git branch --show-current)
```

Submodule guard: `GIT_DIR != GIT_COMMON` is also true inside git submodules. Before concluding "already in a worktree," verify you are not in a submodule:

```bash
# If this returns a path, you're in a submodule, not a worktree — treat as normal repo
git rev-parse --show-superproject-working-tree 2>/dev/null
```

If `GIT_DIR != GIT_COMMON` (and not a submodule): you are already in a linked worktree. Skip to Step 2 (Project Setup). Do NOT create another worktree.

Report with branch state:
- On a branch: "Already in isolated workspace at `<path>` on branch `<name>`."
- Detached HEAD: "Already in isolated workspace at `<path>` (detached HEAD, externally managed). Branch creation needed at finish time."

If `GIT_DIR == GIT_COMMON` (or in a submodule): you are in a normal repo checkout. Create the worktree below without asking.

### Step 1: Create Isolated Workspace

You have two mechanisms. Try them in this order.

#### 1a. Native Worktree Tools (preferred)

The user has asked for an isolated workspace (Step 0 consent). Do you already have a way to create a worktree? It might be a tool with a name like `EnterWorktree`, `WorktreeCreate`, a `/worktree` command, or a `--worktree` flag. If you do, use it and skip to Step 2.

Native tools handle directory placement, branch creation, and cleanup automatically. Using `git worktree add` when you have a native tool creates phantom state your harness can't see or manage.

Only proceed to Step 1b if you have no native worktree tool available.

#### 1b. Git Worktree Fallback

Only use this if Step 1a does not apply — you have no native worktree tool available. Create a worktree manually using git.

Directory selection, in priority order. Explicit user preference always beats observed filesystem state.

1. Check your instructions for a declared worktree directory preference. If the user has already specified one, use it without asking.
2. Check for an existing project-local worktree directory (`ls -d .worktrees`, else `ls -d worktrees`). If found, use it. If both exist, `.worktrees` wins.
3. With no other guidance, default to `.worktrees/` at the project root.

Safety check for project-local directories only: run `git check-ignore -q` on the directory before creating the worktree. If it is not ignored, add it to .gitignore, commit that change, then proceed. An unignored worktree directory commits the whole tree into the repo.

Create it:

```bash
git worktree add "$LOCATION/$BRANCH_NAME" -b "$BRANCH_NAME"
cd "$LOCATION/$BRANCH_NAME"
```

Sandbox fallback: if `git worktree add` fails with a permission error, tell the user the sandbox blocked worktree creation and you are working in the current directory instead. Then run setup and baseline tests in place.

### Step 2: Project Setup

Auto-detect and run the setup the project needs:

```bash
# Node.js
if [ -f package.json ]; then npm install; fi

# Rust
if [ -f Cargo.toml ]; then cargo build; fi

# Python
if [ -f requirements.txt ]; then pip install -r requirements.txt; fi
if [ -f pyproject.toml ]; then poetry install; fi

# Go
if [ -f go.mod ]; then go mod download; fi
```

### Step 3: Verify Clean Baseline

Run the tests to make sure the workspace starts clean (`npm test`, `cargo test`, `pytest`, or `go test ./...`, whichever the project uses).

If tests fail: report the failures, ask whether to proceed or investigate. If tests pass, report ready: worktree path, passing count, and what is next. A dirty baseline makes every later failure ambiguous, so proceeding past failures is your human partner's call.

## The Process

Work the plan wave by wave (see ## Waves). For each task: dispatch the implementer with the hand-off below, handle its report, record the commit, move on.

## Setup

Work happens only in the worktree from ## Worktree. Never start implementation on the main branch without your human partner's explicit consent.

Read the plan once, note its context and Global Constraints, and make a todo per task. If the plan names a spec, read that too: the spec is the authority the plan argues from, and conflicts inside the plan resolve against it.

Before dispatching, scan the plan once for tasks that contradict each other or the Global Constraints, and rule on what you find against the plan text. The spec wins over the plan; your judgment settles what neither answers.

## The Task Loop

### 1. Dispatch the implementer

Every implementer hand-off, built from [implementer-prompt.md](implementer-prompt.md), says:
- the worktree path, branch and parent branch, the plan path and task number, the files, and the task's verify line;
- first action: print `pwd` and `git rev-parse --show-toplevel`; both must equal the worktree path, else stop with zero writes;
- `pm:tdd` at every step: failing test first, watch it fail, minimal code, watch it pass;
- stage by path, never `git add -A` or `git add .`;
- run the repo's formatter and type checks before the commit; formatting goes in the task commit, never in a commit of its own;
- comments in plain English a ten-year-old can read, saying why; no marker tags;
- report in the user's chat language (see the session rules); write everything in the repo in the repo language;
- right after each step, run `pmb tick <task-id> --step <n>` from the worktree so the board shows live progress; never commit the plan file.

The standing rules for every worker are in `references/house-rules.md` (two folders up from this skill); point the implementer at its absolute path.

One task, one dispatch: the brief is the plan path plus the task number, with the exact values to use verbatim. Never make a worker read the whole plan file. The implementer never dispatches subagents of its own — not helpers, and never a reviewer.

### 2. Handle the report

Implementers report DONE, DONE_WITH_CONCERNS, BLOCKED or NEEDS_CONTEXT. DONE: check the commit exists and the verify line passes, then move on. DONE_WITH_CONCERNS: read the concerns first; correctness or scope doubts get answered before moving on, plain observations are noted. NEEDS_CONTEXT or BLOCKED: give the missing context, break the task smaller, or use a more capable model — never force the same dispatch to retry unchanged.

Never ignore an escalation. If the implementer asks questions, before starting or mid-task, answer clearly and completely instead of rushing it into code.

### 5. Complete the task

The commit is in and the verify line passes: mark the todo complete and move on. Never move to the next task while this one is still BLOCKED on something the plan depends on.

## Common Rationalizations

| Excuse | Reality |
|--------|---------|
| "I'll fix it myself, dispatching is overhead" | Orchestrator fixes skip review and pollute your context. Dispatch the implementer. |
| "The implementer spawned its own reviewer — free extra assurance" | A worker-spawned reviewer duplicates the closing review at full cost and its approval counts for nothing. Tell implementers never to spawn reviewers. |
| "Review each task as it lands, to be safe" | No per-task reviewer. One review at the close costs less and catches the same breakage. |
| "Small task, skip the failing test" | No exemptions: `pm:tdd` on every task, even a one-liner. |
| "Let the implementer commit the plan file with its tick" | Several implementers share the plan file, so none of them commits it. You commit it per wave. |

## Waves

Run the plan's waves in order. Inside a wave, dispatch every task's implementer in one message, several at once; serial only for a shared file or a real dependency. Declare the waves in your todo list before dispatching. Read-only probes can run beside a wave.

Implementers tick their own boxes with `pmb tick`, and several of them share the plan file, so none of them commits it. When a wave is done, commit the plan file yourself, staged by path: `git commit -m "pm: tick wave <n>" -- <plan path>`.

### 1. Identify Independent Domains

Group the wave's tasks by what they touch. Tasks in different files with no shared state run together; tasks sharing a file or a real dependency run serially. Each domain is independent: finishing one task never rewrites another task's files.

### Agent Prompt Structure

Good hand-offs are:
1. Focused - one task, its files, its verify line.
2. Self-contained - worktree path, branch, parent branch, plan path and task number.
3. Specific about output - commits made, test evidence, concerns.

Too broad ("implement the plan") leaves the worker lost. No context (paths, verify line) leaves it guessing. No constraints (stage by path, never commit the plan file) lets it trample sibling tasks. Vague output ("done") tells you nothing; demand the short status contract from the template.

### Common Mistakes

**Too broad:** "Implement the wave" - the worker gets lost.
**Focused:** one task, its files, its verify line.

**No context:** "Fix the failing test" - the worker doesn't know where.
**Context:** worktree path, branch, plan path, task number, verify line.

**No constraints:** the worker refactors everything and commits the plan file.
**Constraints:** stage by path, `pmb tick` after each step, never commit the plan file.

**Vague output:** "done" - you learn nothing.
**Specific:** the short status contract from the template: status, commits, test summary, concerns.

**Shared files in parallel:** two implementers writing the same file collide. That pair runs serially.

## Close

When every task is committed: run the full test suite and the type checks, show the output, then use `pm:review` over `<parent>..HEAD`. There is no per-task reviewer and no per-task fix loop.
