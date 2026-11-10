---
parent: scratch/2026-11-10-build-executor-ask-on-parallel-plans
id: SPC-0197
created: "2026-11-10 18:02:54"
hash: fy1de0h
started: "2026-11-10 18:19:57"
---
# Ask subagent or inline when plans can run at once

Architectural. Worked out with the user on WhatsApp on 2026-11-10 (SCR-0094, probe rounds 1 to 3, approach X, cold read fixes F1 to F5 with F2 option a).

## Problem

`build_executor` defaults to `inline`. When `plan_split` cuts a spec into plans that do not wait on each other, slice prints one `build PLN-xxx` line per plan and the user has to open a session for each. Nothing offers to build them at once from the session the user is already in.

## Goal

When two or more approved plans are ready at once, acta asks the user once per run: build them in parallel here with one subagent per plan, or keep inline. After three subagent picks it offers to make that the default for parallel plans only. Building a single plan never changes.

## Settings

Two settings, one meaning each:

- `build_executor` (unchanged): who writes the code of one plan. This work never sets it.
- `parallel_plans` (new, user config only): how ready plans run together. Only value: `subagent`. Unset means ask. Set with `acta config set --parallel-plans subagent`; `--repo` refuses it.

Also new in the user config: `subagent_picks` (a count) and `subagent_offer` (unset or `declined`).

## Design

### Part 1: the ask and the count

1. In `plugin/skills/slice/split.md`, the ask comes after the user's yes to all plans and after step 5 (the Build briefs). It fires when 2 or more plans are ready now (`acta ready`).
2. When `parallel_plans` is `subagent` or `build_executor` is `subagent`: no ask, go straight to Part 2. Otherwise ask "subagent (all ready plans built here, in parallel) or inline?". `dispatch` is not offered here; it stays reachable through config or `/build dispatch`.
3. Inline: print the `build PLN-xxx` lines as today. Not counted.
4. Subagent: run `acta config pick subagent`. It adds 1 to `subagent_picks` in the user config (`~/.acta/config.yaml`, never the repo layer) and prints the new count. Any other argument is an error. Then go to Part 2.
5. When `subagent_picks` is 3 or more and `subagent_offer` is not `declined`, the ask in step 2 becomes "You picked subagent 3 times for independent plans. Build parallel plans with subagents by default? Single-plan builds stay as they are." Yes: `acta config set --parallel-plans subagent`, then Part 2. No: `acta config set --subagent-offer declined`, then the plain ask from step 2. A declined offer never comes back.
6. `acta config show` (text and `--json`) prints `parallel_plans`, `subagent_picks` and `subagent_offer`.
7. Slice still never starts a build on its own. The one exception is Part 2 after this ask or after a `subagent` setting. The rule in `plugin/skills/slice/SKILL.md` ("A split never starts a build") and step 6 of `split.md` ("do not start a build yourself") both name that exception.

### Part 2: building several plans with subagents

A new file `plugin/skills/build/many.md`. `plugin/skills/build/SKILL.md` gets one line beside the executor table: read `many.md` (and not an executor file) when slice hands over two or more ready plans for a parallel build, or when `acta state` of a plan says it belongs to a parallel run. Building one plan does not change.

1. Each plan's `## Build brief` wins for its code repo, parent branch, land method and land order. The worktree for each plan is made where its brief says, the usual way.
2. Start one subagent per ready plan, all in one message, so they run at once. On omp the subagent runs as `agent()` with `agent="task"`.
3. The brief is `implementer-prompt.md`, with "every task of the plan, in wave order" in place of "one task". Everything else in it carries over: the `pwd` and `git rev-parse --show-toplevel` check, the `house-rules.md` path, red test first, one commit per task with the `Task:` trailer, `acta tick` per task, the `acta state set` step, and the first report line naming the model. The subagent never starts a subagent of its own and never touches untracked files it did not make. Its report is short: DONE with the list of shas, or BLOCKED with the task number.
4. When a plan reports DONE, the orchestrator runs the close from build `SKILL.md` for that plan (the `progress.total` and `progress.done` check, untracked files), then `acta:review` over that plan's range. A BLOCKER round goes to a new subagent for that plan, with the findings. Three rounds at most, as usual.
5. Lands go one at a time through `acta:land`. A land order set in a Build brief wins; otherwise lands go in the order plans finish. After each land, any plan whose `after:` list has all landed starts as in step 2.
6. A stuck plan (BLOCKED, three review rounds with a BLOCKER left, or a red land) keeps its worktree and is reported to the user. Plans that wait on it stop too. The rest go on.
7. After each step the orchestrator writes `acta state set <plan id> next` for every plan in the run. The text holds the parent spec id, the line `parallel run`, and each sibling plan with its status: running, waiting on which ids, review round n, landed, or stuck. A new session that runs `/build` on any plan of the run reads that state, opens `many.md`, and picks up every sibling from its status.

### Why one subagent per plan

The user ruled out one subagent per task: each subagent starts from nothing and reads the plan, files and rules again, so cost grows with the number of subagents, and every brief and report fills the orchestrator's context. One per plan keeps both small. Inline stays the default and the cheapest; the ask only shows when there is real parallel work.

## Testing

- Go: `acta config pick subagent` adds 1 and prints the count; a second call reaches 2; a bad argument fails; the value lands in the user file, not the repo file. `acta config set --parallel-plans subagent` and `--subagent-offer declined` save; any other value fails; `--repo` refuses both. `acta config show` text and `--json` print all three fields.
- `internal/plugincheck`: `many.md` exists under build, size caps hold, build `SKILL.md` names `many.md`, and `split.md` names `acta config pick subagent`.
- Eval cases on slice split with two ready approved plans: with `build_executor: inline` it asks; with `subagent_picks: 3` it offers the default; with `subagent_offer: declined` it asks but does not offer; with `parallel_plans: subagent` it does not ask.
- At land: full suite, and the slice and build eval cases.
