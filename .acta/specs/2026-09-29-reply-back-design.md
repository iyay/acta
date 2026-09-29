---
parent: scratch/2026-09-28-dispatch-brief-loads-build
id: SPC-0013
hash: xbaq5tx
---

# Dispatch Reply-Back Design

## Problem

When a plan runs through the `dispatch` executor, the omp recipient often does not tell the orchestrator it is done. Seen during PLAN-18, PLAN-20 and PLAN-21:

- it goes idle and says nothing;
- it asks where to reply, although the brief and the `/goal` carry the command;
- it never loads the build skill, so it skips `acta tick --start` and the Active pane stays empty.

Causes found in the repo:

- `acta:build` has no reply-back step. Reply-back lives only as text in the brief and the `/goal` tail (`plugin/skills/dispatch/SKILL.md`, "Reply-back").
- The reply-back command holds a placeholder (`<new-head-sha>`) the recipient must fill, and a `$HERDR_PANE_ID` that needs careful escaping.
- Briefs listed "allowed" acta commands, which overrode the build skill.
- herdr has no push to a pane when an agent goes idle, so the orchestrator only learns about a silent stop when the user says so.

## Goal

The orchestrator always learns that a dispatched plan is finished, blocked or stalled, without the user stepping in and without the orchestrator blocking its own turn.

## Global rules

- Implement ponytail-lazy: YAGNI, then existing code, then stdlib, then native, then an installed dependency, then one line, then the minimum; never cut validation, security or accessibility.
- The record file is untrusted input. No value from it reaches a shell.
- Comments in plain English a 10-year-old can read; say why, not what.

## 1. Dispatch record

File: `<acta root>/.dispatch.json` in the worktree (default `.acta/.dispatch.json`). It is added to the acta root's `.gitignore` with `hook.EnsureGitignore`, the same way `.agents.json` is.

Fields:

- `pane`: the orchestrator's herdr pane id, for example `wM:pH`.
- `base`: the full commit sha the current review round starts from.
- `plan`: the plan path relative to the repo, for example `.acta/plans/2026-09-29-reply-back.md`.
- `round`: a short slug shown in the review request.

## 2. `acta dispatch init --pane <id> --plan <path> [--round <slug>]`

The orchestrator runs it in the worktree before it sends the `/goal`, and again before each fix round.

- `base` is `git rev-parse HEAD` at that moment, so each fix round's range starts at the previous round's head.
- `round` defaults to the worktree's branch name.
- It overwrites an existing record.
- Exit 1 with one line on stderr: not in a git repo; empty `--pane` or `--plan`; `--pane` does not match the pane pattern (section 4); the plan is not on the board ("no plan <path>").
- It does not commit anything.

## 3. `acta reply-back [--blocked "<reason>"]`

The recipient runs it as the last step of `acta:build` (section 5).

- It reads and checks the record (section 4). A missing or broken record gives exit 1.
- It reads the plan's progress with the same engine as `acta show`.
- Without `--blocked`:
  - `done` less than `total`: exit 1, and stderr lists every open task id (`plans/<stem>#task-N`). Nothing is sent.
  - Complete: it runs `herdr agent prompt <pane> "/acta:review <base>..<head> - plan <plan>, round <round>, pane <own pane>"`. `<head>` is `git rev-parse HEAD`. `<own pane>` is `$HERDR_PANE_ID`, or `unknown` when it is not set.
- With `--blocked "<reason>"`: it sends `"<round> blocked: <reason> - pane <own pane>"` to the orchestrator pane, whatever the progress. An empty reason gives exit 1.
- `herdr` missing from `PATH`, or `herdr` exiting non-zero: exit 3, with herdr's stderr shown.
- Success: exit 0 and one line on stdout naming what was sent.

## 4. Record checks

Before any value is used:

- `pane` must match `^[A-Za-z0-9]+:[A-Za-z0-9]+$`.
- `base` must be 40 lowercase hex characters.
- `plan` must be a relative path that, once joined to the repo root and cleaned, stays inside the acta root and names a plan on the board.
- `round` must match `^[a-z0-9][a-z0-9-]{0,63}$`.
- `herdr` is run with `exec.Command` and separate arguments, never through a shell.

Any failed check gives exit 1 and sends nothing.

## 5. Skill changes

`plugin/skills/build/SKILL.md`:

- A new last step, after every task is committed: when `<acta root>/.dispatch.json` exists, run `acta reply-back`. Exit 1 means finish the open tasks it lists and run it again. A real blocker means `acta reply-back --blocked "<reason>"`.
- A dispatch recipient does not stop before `acta reply-back` exits 0.

`plugin/skills/dispatch/SKILL.md`:

- The brief template gets a required line: `SKILL: load build (omp: build) and tdd before the todo list, and follow them for every task.`
- Briefs never list "allowed" acta commands. They may only forbid acta write commands the build skill does not call.
- The REPLY-BACK line of the brief, and the reply-back tail of the `/goal`, become one sentence: after the last commit the build skill runs `acta reply-back`.
- Delivery adds `acta dispatch init --pane $HERDR_PANE_ID --plan <plan> [--round <slug>]` in the worktree before the `/goal`, and before each fix round's `/goal`.
- "Never wait" gets one exception: the watcher (section 6). Foreground waits, polling loops and `sleep` stay banned.

`plugin/skills/dispatch/herdr-delivery.md`:

- It drops the `--start` and tick restates; the build skill owns them.
- It adds the `acta dispatch init` step and the watcher step.

## 6. Orchestrator watcher

Right after the comprehension checkpoint, the orchestrator runs `herdr agent wait <slug> --until idle --until done` as a background job (Claude Code: `run_in_background`). The job's exit arrives as a notification and does not block the turn.

On that notification:

1. A reply-back already arrived: do nothing.
2. Every task of the plan is ticked: start `acta:review`, the same as a reply-back.
3. Tasks are still open: send one short `/goal` (finish the open tasks, then the build skill runs `acta reply-back`), and start the watcher again.
4. The agent is idle again and no new commit landed since the nudge: tell the user, and stop the loop for this plan.

`idle` can flash between two subagents. That is why step 2 checks git and the plan, not the agent status.

## 7. Guards

`internal/plugincheck` text tests fail when:

- the dispatch brief template loses the `SKILL: load build` line;
- a skill file lists allowed acta commands again (the phrase "only allowed");
- `build/SKILL.md` loses the `acta reply-back` step.

## Testing

Red first for each point:

- `dispatch init`: writes the current HEAD and the gitignore line; a second init after a new commit moves `base`; each bad input gives exit 1.
- `reply-back`:
  - open tasks: exit 1, stderr names them, nothing sent;
  - complete: a fake `herdr` on `PATH` receives exactly `agent prompt <pane> <text>`;
  - `--blocked`: the fake `herdr` receives the prose form;
  - missing record, broken JSON, bad pane, bad sha, plan outside the root: exit 1, nothing sent;
  - `herdr` missing or failing: exit 3.
- The three plugincheck guards.

The watcher is orchestrator behaviour written in a skill, so it has no automated test. The next real dispatch checks it.

## Out of scope

- Push notifications from herdr.
- Harnesses other than omp as recipient.
- Changing how `acta:review` itself runs.
