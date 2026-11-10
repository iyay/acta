---
parent: scratch/2026-11-10-build-executor-ask-on-parallel-plans
id: SPC-0197
created: "2026-11-10 18:02:54"
hash: fy1de0h
---
# Ask subagent or inline when plans can run at once

Architectural. Worked out with the user on WhatsApp on 2026-11-10 (SCR-0094, probe rounds 1 to 3, approach X).

## Problem

`build_executor` defaults to `inline`. When `plan_split` cuts a spec into plans that do not wait on each other, slice prints one `build PLN-xxx` line per plan and the user has to open a session for each. Nothing offers to build them at once from the session the user is already in.

## Goal

When two or more plans are ready at once, acta asks the user once per run: build them in parallel here with subagents, or keep inline. After three subagent picks it offers to make subagent the default.

## Design

### Part 1: the ask and the count

1. In `plugin/skills/slice/split.md`, after the `acta ready` check: when 2 or more plans are ready now and `acta config show` does not say `build_executor: subagent`, ask "subagent (all ready plans built here, in parallel) or inline?". This covers an unset value and an explicit `inline`. `dispatch` is not offered here; it stays reachable through config or `/build dispatch`.
2. Inline: print the `build PLN-xxx` lines as today. Not counted.
3. Subagent: run `acta config pick subagent`. It adds 1 to `subagent_picks` in the user config (`~/.acta/config.yaml`, never the repo layer) and prints the new count. Any other argument is an error.
4. When `subagent_picks` is 3 or more and `subagent_offer` is not `declined`, the ask becomes "You picked subagent 3 times for independent plans. Make it the default?". Yes: `acta config set --executor subagent`, then go on as subagent. No: `acta config set --subagent-offer declined`, then ask subagent or inline as in step 1. A declined offer never comes back.
5. `acta config show` (text and `--json`) prints `subagent_picks` and `subagent_offer` so agents can read them.

### Part 2: building several plans with subagents

A new file `plugin/skills/build/many.md`, read only in this mode. Building one plan does not change.

1. Make a worktree per ready plan, then start one subagent per plan, all in one message so they run at once.
2. The subagent brief: do every task of the plan in wave order, red test first, one commit per task with the `Task:` trailer, `acta tick` per task. It never starts a subagent of its own. Its report is short: DONE with the list of shas, or BLOCKED with the task number. On omp the subagent runs as `agent()` with `agent="task"`.
3. When a plan reports DONE, the orchestrator runs `acta:review` over that plan's range. A BLOCKER round goes to a new subagent for that plan, with the findings. Three rounds at most, as usual.
4. Lands go one at a time through `acta:land`, in the order plans finish. After each land, any plan whose `after:` list has all landed starts the same way.
5. A stuck plan (BLOCKED, three review rounds with a BLOCKER left, or a red land) keeps its worktree and is reported to the user. Plans that wait on it stop too. The rest go on.
6. After each step the orchestrator writes `acta state set <plan id> next`, so a new session can pick up a dead one.

### Why one subagent per plan

The user ruled out one subagent per task: each subagent starts from nothing and reads the plan, files and rules again, so cost grows with the number of subagents, and every brief and report fills the orchestrator's context. One per plan keeps both small. Inline stays the default and the cheapest; the ask only shows when there is real parallel work.

## Testing

- Go: `acta config pick subagent` adds 1 and prints the count; a second call reaches 2; a bad argument fails; the value lands in the user file, not the repo file. `acta config set --subagent-offer declined` saves; any other value fails. `acta config show` text and `--json` print both fields.
- `internal/plugincheck`: `many.md` exists under build, size caps hold, and `split.md` names `acta config pick subagent`.
- Eval cases: slice split with two ready plans and `build_executor: inline` asks the question; with `subagent_picks: 3` it offers the default; with `subagent_offer: declined` it does not offer.
- At land: full suite, and the slice and build eval cases.
