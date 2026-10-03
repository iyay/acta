# Dispatch: the plan in an omp tab

Read this only when `acta:build` runs the `dispatch` executor and `HERDR_ENV=1` is set. Without `HERDR_ENV=1` there is no pane to hand the plan to: `dispatch` runs as `subagent` (`## Executors` in SKILL.md). Build already made the worktree. `acta dispatch send` does the whole hand-off in one call. Review and landing stay `acta:review` and `acta:land`.

## Send

Run it in the worktree:

```bash
acta dispatch send --plan .acta/plans/<stem>.md --rules <abs path> [--note-file <path>|-]
```

`--rules` is the absolute path of this skill's base dir plus `../../references/house-rules.md`. Job facts the plan lacks (a locked call, a trap) go in a note file, or `-` for stdin.

It writes the dispatch record and the brief, finds or makes the omp tab for the branch, sets the goal, waits about 20 seconds and reads the pane once. It prints `slug:`, `pane:`, `base:`, `brief:`, `checkpoint:` and `watcher:`. Read those lines and the exit code:

- `0`: sent. `checkpoint: ok` means every task id is on the todo list. `unconfirmed` means no todo list yet: say so in the report. The reply-back is the real signal.
- `1`: bad input, nothing sent. Fix the input and run it again.
- `2`: delivery failed. Read herdr's reason. An agent still `working` is refused this way: an earlier round runs, so do not interleave. Report and stop.
- `3`: a git read failed (a detached main checkout counts). Nothing sent. Report and stop.
- `4`: drift. The pane text is printed. Read it first. Real drift: the list names work outside the plan or skips wave 1 tasks. Then press `herdr agent send-keys <slug> esc`, put the real task ids in a note ("there is NO `<X>` in this plan") and send again. A folded list (omp shows 8 rows) or an old round's card is no drift: yield.

The omp harness needs `acta:build` installed too, because skills do not travel with the worktree. Missing: stop and report.

## Never wait for the recipient — HARD RULE

After `send`, end your turn. Not allowed until the reply-back or the user's next message: a foreground `herdr agent wait`, polling loops, repeat pane reads, `sleep`, `Monitor` loops. Waiting burns your context.

The report is short. First line is a fact: slug, pane, base, tasks sent, checkpoint. Last line is one next action: "on reply-back: verify, then `acta:review`, then land".

## Idle watcher

Start the printed `watcher:` command (`herdr agent wait <slug> --until idle --until done`) with `run_in_background`, then yield. Its exit comes as a notification:

1. A reply-back already arrived: do nothing.
2. Every task of the plan is ticked: start `acta:review`, like a reply-back.
3. Tasks still open: send one plain prompt, `herdr agent prompt <slug> "Finish the open plan tasks, then run acta reply-back."`, and start the watcher again.
4. Idle again with no new commit since the nudge: tell the user and stop the loop for this plan.

`idle` can flash between two subagents, so step 2 reads git and the plan, never the agent status alone. `blocked`: `herdr agent read <slug> --source recent-unwrapped --lines 60`, then answer with `herdr agent send-keys <slug> <key>` or one more prompt.

## Verify the reply

A reply-back is a claim, not a fact. It starts this phase, and so does the user. Run `herdr agent get <slug>` once. `working`: one line, yield. Otherwise, with `<base>` from the `base:` line:

```bash
git log --oneline <base>..HEAD
git diff --stat <base>..HEAD
git diff --name-only <base>..HEAD      # only plan files
<the plan's fast test command> 2>&1 | tail -20
git diff --name-only <base>..HEAD -- .acta/bugs
```

Compare the commits with the task list. On security or destructive changes, re-run the decisive mutation yourself. Name each bug file of the last command under `Bugs found by recipient:` in the landing report, or write `none`.

## Review

`acta:review` runs on every round, over the range and plan path build's `## Close` sets. Its small-change self-review never applies to a dispatch: you did not write or watch the code. To have a different pane review, the message needs the literal skill name, a review keyword, and the range.

NOTEs are never written to memory: once the round is CLEAN, `acta:review`'s `## After a CLEAN round` sorts them into `[fix]`, `[debt]` or `[note]`.

## Fix rounds

BLOCKERs or open tasks go back in this same turn. A findings list with no `send` behind it is an unfinished turn. The round cap is `acta:review`'s: round 3 still BLOCKED, ask the user.

1. Append `## Fix round <n>` with its task and `verify:` line to the plan, as `acta:review` says.
2. Run `acta dispatch send --plan <plan> --rules <abs path> --round fix-<n>`. It reuses the same tab, so the agent keeps its context. Write the note as if it remembers nothing, because omp compacts.
3. Start the watcher, then stop. One turn per round.

The polish goes out the same way: `--round polish` with the `[fix]` NOTE list as the note. The review skill's `## After a CLEAN round` owns what happens on reply-back.

Clean and complete: run `## After a CLEAN round`; its `acta:land` is build's `## Close`. Close the tab first (below). Never leave a reviewed branch parked.

## Close the tab

`acta dispatch close`, run in the worktree, closes the tab. It takes the slug from the current branch, so run it once the review is CLEAN, right before `acta:land` removes the worktree and branch. `acta:land` has no pane step.

The landing report: merge SHA first, gate numbers, what was cleaned up, then two lines you owe: `Bugs found by recipient:` and `Harvested from omp: <note titles>`, or `omp memory: nothing to harvest`.

## Harvest omp memory

Omp recipient only, after the land. Read `~/.omp/agent/memories/--<worktree abs path with every / turned into ->--/learned.md`, for example `--home-me-code-app-worktree--`. For each point ask: does a fresh agent lose time or repeat a mistake without it? If yes and Claude memory lacks it, write one note. A fact for this project goes in `~/.claude/projects/<main checkout sanitized>/memory/`, a fact true everywhere in `~/.claude/memory/`, each with one index line in that folder's `MEMORY.md`. No question to the user.

## Advisor

Before security, auth, data-migration or money work, pair a reviewer model to omp's `advisor` role (`/advisor` in the pane shows it). One advisor note costs less than a review round.
