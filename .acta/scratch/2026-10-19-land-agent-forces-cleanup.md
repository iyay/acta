---
id: SCR-0076
hash: vjukzmc
title: Land runs force and rm -rf the skill never asks for
status: raw
created: "2026-10-19 06:50:17"
schema: "1"
finished: "2026-10-31 20:19:51"
---
# Land runs force and rm -rf the skill never asks for

## Words

### 2026-10-19

In eval runs on 2026-10-19 the land step did things the land skill never names:
- review-understanding-report: `git commit --amend` on the merge commit.
- state-resume (trace e-nMn6yY, call 26 and 28): `rm -rf src/__pycache__` and `git worktree remove --force wt`.

The skill says `git worktree remove <path>` and never delete a worktree with uncommitted work. Find why agents reach for force and amend here, and whether the skill should ban them by name.

## Context

## Log

### 2026-10-31

Shape started 2026-10-31 as Bounded with a written spec (it touches delete commands). Probe round 1, user answered "ok semua" (typed "oo semua"):
- Q1: cheap spike first: one state-resume eval run with --keep-temp, read the trace to find why the agent forces.
- Q2: two layers: skill text plus a pre-tool hook that blocks --force, rm -rf and commit --amend, the same way it blocks a bare go test.
- Q3: when git worktree remove refuses, the skill says stop, show git status of that worktree, and ask the user; never --force.
- Q4: amend of the merge commit is in scope too: land never amends.

### 2026-10-31

Spike 2026-10-31: one state-resume run with --keep-temp (score 1.00, kept temp /private/tmp/e-ZBrpEM). The trace shows the cause: after the task tests ran, git status listed `?? src/__pycache__/` (test byproduct, not ignored by the scaffold). The build step wanted a clean tree before committing the plan file, so the agent ran `rm -rf src/__pycache__`. git worktree remove refuses a tree with untracked files, which is the same push toward --force at land. Root cause: untracked test output plus a clean-tree rule with no word on what to do with it.

### 2026-10-31

Probe round 2, user answered "ok":
- Q5: untracked files a test or tool made are not work and are never deleted; show them to the user and offer a .gitignore line in the task commit, with the user's yes.
- Q6: the hook blocks in every repo that uses acta; a user who really wants the command runs it with ! in the prompt.
- Q7: amend is blocked only when HEAD is a merge commit; amending one's own unpushed plain commit stays allowed.

### 2026-10-31

Spec SPC-0186 written. Cold read found five gaps; the user approved every suggested fix with a thumbs up: amend check reads the -C or cd repo and skips git global options; the hook blocks only inside a git repo; core rule 5 changes to "never run it, let the user run it"; git clean -f is blocked and the eval grader checks the test folder survives; land offers three choices for test output (ignore line, user removes it, stop) and the omp-report wiki page is rewritten. Spec updated at 7c03757; waiting for the user's spec review.

### 2026-10-31

User replied "lanjut" (2026-10-31 20:23): spec SPC-0186 approved. Next: acta:slice.

### 2026-10-31

Sliced SPC-0186 into two minimal plans that share no file: PLN-0206 hook (DestructiveBlock, pre-tool wiring, core rule 5) and PLN-0207 skill text (build, land, tidy, omp-report wiki, state-resume grader). Waiting for the user's yes on the plans.

### 2026-10-31

User replied "yes pake subagent" (2026-10-31 20:24): plans approved; build both with the subagent executor.

## Open questions
