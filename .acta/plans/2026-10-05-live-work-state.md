---
parent: specs/2026-10-05-live-work-state-design
depth: minimal
id: PLN-0082
created: "2026-10-05 08:18:43"
hash: j3kgvms
started: "2026-10-05 08:32:15"
finished: "2026-10-05 10:31:11"
---
# Live work state Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Keep a `## State` section in each plan file and show running plans at session start, so a fresh session goes on without the user re-explaining.

**Spec:** `.acta/specs/2026-10-05-live-work-state-design.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Subsections are exactly `### Next`, `### Findings`, `### Open rulings`, in that order, under `## State`; at most 10 lines each.
- Never write the current task, last commit, worktree path or review round into the file; compute them on read.
- The session start hook never blocks and never fails the session; on any error it adds nothing.
- Write commands auto-commit: tests run them in temp clones only.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3, Task 4
- Wave 4: Task 5

### Task 1: acta state set

**Files:** Create `internal/write/state.go`, `internal/write/state_test.go`; modify `internal/cli/cli.go` (new `state` case), create `internal/cli/state.go`, `internal/cli/state_test.go`.
**verify:** No call to `acta state set` ever changes any byte of the plan outside the one named subsection, and no body over 10 lines is ever written. List every path checked: missing `## State`, missing subsection, existing subsection, empty body, over-cap body, unknown subsection name, unknown plan.
- [x] Failing test: setting `next` on a plan with no `## State` creates it with only `### Next`; it fails because `write.SetState` does not exist.
- [x] Code: `write.SetState(cfg, b, planID, part string, body []byte) (Outcome, error)` replaces or creates the subsection, empty body clears it, over 10 lines errors with the limit named; the CLI reads stdin and commits `acta: state <plan id>`.
- [x] Commit: `state: acta state set writes one State subsection in a plan`.

### Task 2: acta state view

**Files:** Modify `internal/cli/state.go`, `internal/cli/state_test.go`; create `internal/board/state.go`, `internal/board/state_test.go`.
**verify:** A plan counts as running only when it has `started`, no `finished`, and a worktree from `internal/trees`; every derived fact (current task, last commit, worktree, review round) comes from the plan ticks and git, never from the file text. List each fact and its source, plus the case where none exists.
- [x] Failing test: `acta state <plan>` in a temp clone with a worktree prints the first unticked task, the branch's last commit, the worktree path, `round: none`, then the three subsections; it fails because the view does not exist.
- [x] Code: `board.Running(cfg) []RunState` and `RunState.Full() string` and `RunState.Short() []string` (max 3 lines); review round from the newest `acta: tick fix round N` commit on the branch; `acta state` alone prints one line per running plan.
- [x] Commit: `state: acta state shows running plans with derived facts`.

### Task 3: session start summary

**Files:** Modify `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/cli/hook.go`, `internal/cli/hook_session_test.go`.
**verify:** The session start text never shows more than 3 plans or more than 3 lines per plan, never shows a plan from another repo, and on every error path adds nothing. List each error path checked.
- [x] Failing test: with 4 running plans, `hook.SessionStart` output holds 3 short blocks, one `+1 more, run acta state` line and the rule line; it fails because `Input` has no running plans field.
- [x] Code: `Input.Running []string` blocks filled in `cmdHook` session-start from `board.Running` for the session repo and its worktrees, errors dropped; rule line: before you stop, or when context runs low, run `acta state set <plan> next`.
- [x] Commit: `hook: session start lists running plans and their Next`.

### Task 4: skills and dispatch brief

**Files:** Modify `plugin/skills/build/SKILL.md`, `plugin/skills/land/SKILL.md`, `internal/cli/dispatch_brief.go`, `internal/cli/dispatch_brief_test.go`.
**verify:** Every brief round tells the recipient to read `acta state <plan>` first, and the build and land skills each name when State is written or read; `internal/plugincheck` stays green. List every brief round kind checked.
- [x] Failing test: a built brief for each round kind contains `acta state`; it fails because the brief has no such line.
- [x] Code: one brief line; build: update Next after each task commit, Findings for a finding that cost time, Open rulings for a question waiting on the user; land: before the merge read Findings and move each lasting one to a wiki page or a debt item.
- [x] Commit: `skills: build writes plan State, land drains Findings, brief reads it`.

### Task 5: resume eval and version

**Files:** Create `plugin/evals/state-resume/` (`case.yaml`, `prompt.md`, `scaffold.sh`, `graders/`); modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The eval passes only when the agent's first work step follows Next and fails when it asks the user where things stand; the three version files agree on 0.1.6.
- [x] Failing test: the eval case layout check in `internal/plugincheck` and the eval itself on a scaffold with a half done plan and a Next line; red before the case exists.
- [x] Code: scaffold a repo with a running plan, worktree and Next; graders check the agent acts on Next without a question; bump the patch in all three files.
- [x] Commit: `eval: fresh session resumes from plan State; version 0.1.6`.

## Review notes

- runningBlock applies the 3-line cap before blank lines are dropped, so a block with a blank line in the middle shows fewer lines.
- planStates takes index 0 of `git worktree list` as the main checkout; it leans on git's output order.
- The dispatch brief line names `acta state <plan>`; nothing checks the plan argument form.
- The spec says the worktree comes from internal/trees; the code uses gitc.Worktrees with the same effect.
- `acta state <plan>` also shows a finished plan still in a worktree, on purpose, so land can read Findings.
- sameDir in internal/board/state.go is used only by tests.
- The build skill line is looser in wording than the spec's three triggers, but names all three.
- The help test splits the help text on ", ", so a change to a newline list breaks it.

## Fix round 1

### Task 6: state-resume graders read the worktree

**Files:** Modify `plugin/evals/state-resume/graders/followed-next.md`, `plugin/evals/state-resume/graders/read-the-price.md`, and `plugin/evals/state-resume/scaffold.sh` only if a comment there names the graded path.
**verify:** The state-resume case passes when the agent makes the Next change in the plan's worktree (`wt/src/invoice.py`) whether or not the merge to main happens, and still fails when the agent only reads the state and changes no file. List the graded paths and both outcomes checked.
- [x] Failing test: `scripts/eval --case state-resume` with the branch binary on PATH scores 0.50; the run trace shows the agent wrote `wt/src/invoice.py` and the sandbox blocked the merge, so `src/invoice.py` stayed unchanged.
- [x] Code: point both file graders at `wt/src/invoice.py`; keep their patterns as they are.
- [x] Commit: `eval: state-resume grades the worktree file the build changes`.
- Fix round 1: a run that edits only main's src/invoice.py and never the worktree now fails the state-resume file graders, on purpose.
- The state-resume grader files end with no newline, as before.
