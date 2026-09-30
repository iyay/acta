# Dispatch — the plan in another pane

Read this only when `acta:build` runs the `dispatch` executor. Build already made the worktree (`## Worktree` in SKILL.md) and picked this executor. This file says how to hand the plan to an omp agent in its own herdr tab, how to wait for it, and how to send a fix round back to it. Review and landing are build's `## Close`, `acta:review` and `acta:land`; only the steps a tab adds live here.

**Skill notation.** Every `acta:name` here is a skill from the `acta` plugin. `mattpocock-skills:*`, `/to-spec`, `/to-tickets`, `/implement`, `/code-review` are gone; a brief still naming them is stale — rewrite it before sending. "Ticket" below = one `acta:plan` task, mirrored in the tracker when the repo has one.

## Never wait for the recipient — HARD RULE

**After the comprehension checkpoint, Phase 1 is over. End your turn.**

Forbidden until the user's next message or the reply-back notification: `herdr agent wait`, polling loops, repeat pane reads, `sleep`, `Monitor` until-loops, any long-timeout call whose job is to sit.

Why: the recipient takes minutes to hours; blocking burns your context and freezes the user's session. Its completion signal is unreliable anyway (idle can mean stopped early), so you re-derive every number from git whenever you look. Waiting adds nothing; looking later costs nothing.

Two exceptions: the ~20s pause before the single checkpoint read, and the idle watcher below, which runs as a background job and never blocks the turn.

Phase 1 report, ADHD shape (first line = fact, last line = one next action): base SHA, worktree path, branch, slug + pane id, ticket ids handed over, checkpoint verdict, any call you locked in the brief, then the literal next step ("on reply-back: verify `<base>..HEAD` → `acta:review` → fix rounds → land, no further ask"). Then yield.

Phase 2 resumes on either trigger: the REPLY-BACK notification lands, or the user asks. **Neither needs the other.** A notification alone starts verify → review → route/land in that turn; do not wait for the user to confirm it.

## The loop at a glance

**Phase 1 — dispatch (one turn, hard stop):**

1. **build picked `dispatch` and made the worktree** (`## Worktree` in SKILL.md).
2. **Provision the tab**: look the agent up by slug and reuse; create its own tab only on a miss. Never a split beside you.
3. **Write the hand-off doc** (Step 0a). Pointer at plan + tickets plus the mechanics no skill supplies. Not a rewritten plan.
4. **Deliver**: `/new` on a new session/plan/worktree, `acta dispatch init` in the worktree, then always `/goal` carrying the brief path and the reply-back line.
5. **Comprehension checkpoint**: exactly one read ~20s later — todo list must name the ticket ids, else drift; correct and re-dispatch. Todo list not up yet: report "checkpoint unconfirmed" and yield.
6. **Start the watcher** (background, below), then stop.
7. **STOP. Report and yield.**

The recipient is the main agent in its pane: writes zero code itself, **fans every ticket out to implementer subagents per `acta:build` — as many in parallel as file ownership allows** (see "Spread the work"), commits per ticket, does NOT review, fires reply-back, never touches its tab.

**Phase 2 — verify, review, land (later turn):**

7. **Verify** from git. Verify ≠ review.
8. Tab: nothing to do. A stray pane split beside you → move it to its own tab once.
9. **Review** via `acta:review`, over the range and the plan path build's `## Close` sets, every round, no exceptions.
10. BLOCKERs or unfinished tickets → the one fix round in "Fix rounds" below, **stop and yield**. One turn per round. Round 3 still BLOCKED → stop, ask the user (see `acta:review`).
11. **Clean AND complete → build's `## Close` runs `acta:land`, then "Close the tab" below.** Never leave a reviewed branch parked.

## One tab per dispatch — HARD RULE

One worktree = one slug = one tab = one agent, for the life of the branch. A split costs the user half a screen for hours of idle agent; a tab costs nothing until they look.

Create as a tab (herdr):

```bash
MY_WS=${HERDR_PANE_ID%%:*}
TAB_JSON=$(herdr tab create --cwd <worktree-abs-path> --workspace "$MY_WS" --label <slug> --no-focus)
PANE=$(printf '%s' "$TAB_JSON" | jq -r '.result.root_pane.pane_id')
TAB=$(printf '%s'  "$TAB_JSON" | jq -r '.result.tab.tab_id')
```

- `--no-focus` on every tab/pane op. The user's focus is theirs.
- `--workspace "$MY_WS"` is mandatory: worktrees live outside the repo, so herdr would file the tab under another workspace (observed 2026-08-21).
- Address the agent by slug, never by remembered pane id — ids churn on cross-workspace moves.
- Tab persists across rounds; closed once, at landing.
- Brief carries zero tab/pane commands.
- Wrong workspace already? `herdr pane move <pane> --new-tab --workspace "$MY_WS" --label <slug> --no-focus`; read the new pane id from `.result.move_result.pane.pane_id`.

Full surface: the separate `herdr` skill.

## Operation map

| Operation | herdr |
| --- | --- |
| Own address | `$HERDR_PANE_ID` |
| Provision | `herdr agent get <slug>` → reuse, else `herdr tab create --cwd <worktree> --label <slug> --no-focus` + `herdr agent rename` |
| Launch harness | `herdr pane run <pane> "omp"`, detect, `herdr agent rename <pane> <slug>` |
| Deliver | `herdr agent prompt <slug> "<text>"` |
| `/new` + `/goal` | both via `herdr agent prompt <slug>` |
| Checkpoint | `herdr agent read <slug> --source recent-unwrapped --lines 60` |
| Completion (later turn) | `herdr agent get <slug>` once when you return |
| Reply-back | `acta reply-back` in the worktree, run by the build skill; the brief and the `/goal` only name it |
| Record | `acta dispatch init --pane $HERDR_PANE_ID --plan <plan> [--round <slug>]` in the worktree, before every `/goal` |
| Stray split pane | `herdr pane move <pane> --new-tab --no-focus --label <slug>` |

Map, not instructions. Flags and failure handling live in the reference file, [herdr-delivery.md](herdr-delivery.md).

## `/new` and `/goal` — every dispatch, every backend

- **`/new` first** on a new session, new plan, or new worktree. Skip only for a fix round on the same plan in the same worktree — context is the asset there. In doubt → send it; the brief is on disk.
- **`/goal` every turn, no exceptions**, including fix rounds and re-dispatch after drift. The goal survives the harness's own compaction; a plain prompt does not. One `/goal` per turn, one brief, no bare follow-up.
- `/goal` names the reply-back line: after the last commit the build skill runs `acta reply-back`. It carries no command to fill in.

**Write the record before the `/goal` — HARD RULE, every dispatch and every fix round.** In the worktree, so `acta reply-back` on the other side has a pane, a base and a plan to send:

```bash
acta dispatch init --pane $HERDR_PANE_ID --plan <plan> [--round <slug>]
```

`base` is the HEAD at that moment, so each fix round reviews its own range. Run it again before every fix round's `/goal`; `--round` defaults to the branch name.

**Order — HARD RULE, five steps, a check after each:**

1. Wait until omp is ready: the pane shows its empty input box (`herdr agent read <slug> --source visible`). A just-launched omp is not ready yet.
2. Send `/new` alone. Read the pane until it shows "New session started". Then rename the agent by pane id (`/new` drops the name).
3. Run `acta dispatch init` in the worktree, so the reply-back has a record to read.
4. Send `/goal` alone. Read the status bar until it shows `🎯 Goal`. `⏸ Goal` or no goal → see the held-goal block below and send it again.
5. fix round: `acta dispatch init` again, then `/goal` only — no `/new`; still check `🎯 Goal`.

`/new` and `/goal` go as two prompts, never in one prompt, and `/new` never goes before omp is ready: otherwise both land as one message and the goal never sets. These reads are part of delivery, not the comprehension checkpoint. Checkpoint after the `/goal`.

**A held goal silently refuses the next one — on omp `/new` does NOT clear it.** Symptom: prompt returns success, pane shows `Warning: Resume the current goal first, or drop it…`, status bar `⏸ Goal 0`, agent idle having never seen your objective. Clear it:

```bash
herdr agent prompt <slug> "/goal drop"
herdr agent send-keys <slug> enter
herdr agent send-keys <slug> enter
```

Then send the new `/goal`; confirm `🎯 Goal <n>K` in the status bar. Bites on every re-dispatch after drift or interruption; use `--source visible` to see the warning.

Harness with no `/goal` → plain prompt, and say so in the report.

## Reply-back — every brief, every backend

Every brief ends with a REPLY-BACK line, and every `/goal` repeats it. Belt and braces: it is the single most-skipped instruction.

One sentence, the same in both places:

```
REPLY-BACK: after the last commit the build skill runs `acta reply-back`.
```

- The recipient runs no command and fills no placeholder. `acta reply-back` reads the record `acta dispatch init` wrote, so the range, the plan, the round and the pane all come from the worktree instead of from a hand-edited line.
- A real blocker is `acta reply-back --blocked "<reason>"`, sent the same way.
- Arrival does not excuse verify. A recipient can fire it after 2 of 6 tickets.
- No self-close line. The tab is reused; you close it at landing.

## omp magic keywords — two free words in every `/goal`

The recipient is omp. Two standalone lowercase prose words change how it works that turn, at no cost:

- **`ultrathink`** — careful multi-step reasoning, and the highest thinking effort the active model supports. The user's omp roles are already pinned to `:xhigh`, so the gain over that pinned effort is small — keep it anyway, it costs nothing.
- **`orchestrate`** (`src/prompts/system/orchestrate-notice.md`) — the orchestrator contract: decompose the job, dispatch `task` subagents, enumerate the full surface into todo items, parallelize disjoint work in ONE message, keep every task self-contained (explicit target paths, acceptance criteria), verify, and never yield before the whole thing is closed. This is the word that matches the PARALLEL rules above.

Not `workflowz`: that keyword drives a keep-alive `workpool()` meant for research, review, migration and other open-ended work lists — each item goes to the least context-loaded idle worker, so context piles up across items instead of staying isolated per ticket, which the PARALLEL rules forbid. Its own notice says "Quick lookup/single edit: direct; no agents" — right tool for a survey or a review pass, wrong one for dispatching tickets.

Put **`ultrathink orchestrate`** in every non-trivial `/goal`.

**Matching is strict, and the trap is quoting.** Exact lowercase, standalone prose only. `orchestrate,` triggers; `orchestrated`, `orchestrate.ts`, `orchestrate()` and `foo::orchestrate` do not. **Anything inside a fenced block, an inline code span, or an HTML/XML tag is ignored** — so a keyword you wrap in backticks to look tidy is dead. The word applies only to the turn carrying it: repeat it every round, including fix rounds.

## omp `advisor` — a second model watching every turn

omp routes work across nine roles (`default`, `smol`, `slow`, `plan`, `commit`, `vision`, `task`, `advisor`, `tiny`). Pair a reviewer model to **`advisor`** and it reads every turn the doer takes on its own context and its own model, injecting an aside, a concern, or a hard blocker inline. The doer course-corrects mid-task or says why it won't.

This catches the failure class this loop is worst at: omp's own documented example is an agent scoping a `catch` to ENOENT instead of every error, with the advisor flagging that the narrowed fix **no longer meets the literal acceptance criterion**. That is the half-fix shape — caught during the work instead of two review rounds later.

Set it up before dispatching security, auth, data-migration or money work; check with `/advisor` inside the recipient pane. An advisor note is cheaper than a fix round by roughly the cost of a full `acta:review`.

Also under-used, and worth naming in briefs: **`lsp`** answers "every caller of X" mechanically (references, rename, diagnostics) where grep guesses; **`todo`** carries phase tracking; **`agent://<id>/findings.0.path`** pulls a typed field straight out of a subagent's result.

## Provision the tab

The worktree already exists (the worktree from `## Worktree` in SKILL.md). This step adds the tab and nothing else.

1. **Reuse the worktree's agent by slug on every turn.** `herdr agent get <slug>` is the first call on any turn touching the recipient. Resolves → reuse. Misses → create a tab, rename agent. Never a second tab for one worktree (two agents fight over one branch). Never scan for a "free" pane — `agent == ""` matches dev servers and tunnels; three mis-prompts in one session.

2. **Launch the harness in the worktree**; wait for its prompt.

The full tab surface — reuse vs create, the rename handle, the response shape, the `--yolo` layer, and failure handling — lives in [herdr-delivery.md](herdr-delivery.md).

## Step 0a — Write the hand-off doc

Instructions live in a file, never in the message. Multi-line pastes fragment in exactly the harnesses that need them; a file reads atomically, survives `/new` and compaction.

Path inside the repo the recipient works in: `.claude/dispatch/<slug>-brief.md` (git-excluded; check `git check-ignore -v .claude/`). Not `/tmp` — sandboxed harnesses cannot read outside their tree.

**What the brief is now.** With `acta:brainstorm` + `acta:plan` upstream and `acta:build` downstream, the brief no longer restates tasks, decisions, or method. It is:

- the pointer: design doc + plan path + ticket refs, worktree, base SHA, parent branch;
- a pointer to `references/house-rules.md`, which carries the fixed rules every hand-off requires (TDD, ponytail, COMMENTS, PARALLEL); `ticket → verify` stays job-specific, in TICKETS;
- the mechanics no skill supplies (gates, reply-back, stay-put).

Locked decisions belong in the design doc or plan. If one is missing there, fix the plan first — do not patch it into the brief.

## Every criterion is a PROPERTY, never the instance you saw — HARD RULE

**This is the highest-leverage line in the skill.** An agent given a narrow target returns a narrow fix; the review then finds the part of the property you never named, and that is your next round. Six rounds in one session (2026-09-03) traced to this and nothing else.

Write each `verify:` and each acceptance box as a claim that must hold **over every path**, and make the agent enumerate them:

| Instance (breeds rounds) | Property (closes it) |
| --- | --- |
| "Return null when the row cannot be read" | "A consenting merchant must never read as not-consented by **any** failure path. Enumerate every path that can produce a false or missing answer, and name what each returns." |
| "The card makes no present-tense claim" + a line range | "**No** surface a merchant sees claims this happens today. List every surface you checked." |
| "Guard the old heading string" | "Old copy cannot silently return **in any form** — reworded, partial, or added beside the new copy." |

Three habits that go with it:

1. **Never hand a line range as the target.** A range is an instance. Name the rule and let the agent find the lines — it has `lsp` and `grep`. Giving `:136-140` while the defect also sat at `:133` cost a full round.
2. **Say "enumerate every caller / path / surface, and report the list."** The list is the deliverable; a fix without it is unverified.
3. **When your own fix widens a boundary, name its neighbours in the same ticket.** Widening one write and not the one beside it created a request that stores half its work and reports failure.

Reviewers get the same treatment: ask them to test a property, not to hunt. "Find a case I did not test" always succeeds and always yields another round.

## Stopping rule — what blocks, what gets logged

**Blocking = a concrete input that produces a wrong result for a real user, today.** Same bar as `acta:review`: BLOCKER or NOTE, nothing between. Everything else is logged as one NOTE (memory) or one follow-up ticket, and the branch lands.

Not blocking, however tempting: test durability against hypothetical future edits, mock fidelity for props nothing passes, a substring that could be reworded around, naming, an assertion with no message. Those are real and they are cheap to log. Routing them as fix rounds is how a branch that was correct after round 2 reaches round 6.

Security, auth, data-migration and money paths keep the full bar — there, "a future edit could" often means "a caller today can".

## Spread the work — maximum subagents, every brief, every backend

**The recipient is an orchestrator, not a typist.** `acta:build` already forbids it from writing code and already says fresh subagent per task; this section makes it fan out as wide as the ticket list allows. A recipient that runs six tickets through one implementer subagent serially is the slowest and least reviewable shape available: one context accumulates every file and every failed attempt, and by ticket 5 it has compacted away ticket 1's constraints.

The PARALLEL rules live in `references/house-rules.md` now, read before the todo list — no need to retype them into the brief. The `/goal` still has to carry the push: the `orchestrate` keyword's contract (decompose, dispatch subagents, parallelize disjoint work in ONE message, verify, never yield before closure) says the same thing, so put it in every `/goal` — it is skipped as often as the reply-back line.

- **One implementer subagent per ticket, minimum.** The recipient never implements a ticket in its own context. Every subagent runs its own `acta:tdd` loop: red first, mutation-verify, surgical, commits its own ticket.
- **Parallel by default, serial by exception.** Independent tickets go out **in one message, several at once** — not one after another. Something must force serialization: the same file touched twice, a real dependency (B imports what A creates), or a migration that must land alone.
- **Partition by file ownership.** Two subagents in one wave must never hold the same file — the one collision that produces silent lost edits in a shared worktree. Group tickets into waves on that basis before dispatching anything.
- **Split further where the plan allows.** A ticket that names two disjoint files can go to two subagents (one per file, tests with the code they test). A ticket with an interface seam can split into "write the failing tests" and "make them green" only when the test author commits first — otherwise keep them together.
- **Read-only work fans out cheapest.** Codebase probes, API-surface checks, "where is X used", lookups the plan did not already answer — dispatch those in parallel even when implementation must be serial (`acta:build`). Never block a wave on a probe one subagent could run beside it.
- **Waves declared up front** in the todo list ("wave 1: T-1, T-3, T-4 parallel · wave 2: T-2 — depends on T-1"), so drift is visible at the comprehension checkpoint.
- **Parallelism never relaxes the gates.** No batching commits at the end of a wave, no "test it all together", no shared scratch files between subagents.
- **Skills are HOW, not WHO.** Subagents are plain `general-purpose` implementers; `acta:tdd` runs inside them. Never a skill name in a `subagent_type` slot.

You require this; you do not police the exact wave count. A genuinely dependent chain run serially is correct. Six independent tickets run serially is a finding — say so in the fix round, and re-dispatch the remaining tickets as a parallel wave.

## Comprehension checkpoint

About 20s after the `/goal`, exactly one read (`herdr agent read <slug> --source recent-unwrapped --lines 60`). No loops, no sleeps, no second read. If the todo list is not on screen yet, report "checkpoint unconfirmed" in the Phase 1 report and yield anyway: the reply-back is the real signal. A re-dispatch after drift gets its own one read, under the same rule. Reads to find out why something failed, after a reply-back or a stuck report, are outside this rule. When the todo list is on screen, two things must be true:

1. It names the ticket ids from the brief. Invented modules/phases/endpoints = it confabulated before reading → interrupt key, re-dispatch with a corrective preamble ("there is NO `<X>`, NO `<Y>` — writing those = drift").
2. It declares waves, and wave 1 holds every ticket that has no dependency. A flat serial list of independent tickets = fan-out drift → interrupt, re-dispatch with the wave plan spelled out ("wave 1: T-1, T-3, T-4 in ONE message").

Never skip for destructive work or weaker harnesses; both kinds of drift track model tier.

## Idle watcher — the one background job, right after the checkpoint

Right after the comprehension checkpoint, start this as a background job (Claude Code: `run_in_background`), then yield:

```bash
herdr agent wait <slug> --until idle --until done
```

Its exit arrives as a notification, so it never blocks your turn. This is the one exception to "Never wait": foreground waits, polling loops and `sleep` stay banned.

On that notification:

1. A reply-back already arrived → do nothing.
2. Every task of the plan is ticked → start `acta:review`, the same as a reply-back.
3. Tasks still open → send one short `/goal` (finish the open tasks, then the build skill runs `acta reply-back`) and start the watcher again.
4. Idle again with no new commit since the nudge → tell the user and stop the loop for this plan. `idle` can flash between two subagents, which is why step 2 reads git and the plan, not the agent status.

## Hand-off brief format

```
<one-line summary of the job>

PLAN: <.acta/plans/…md> (design: <.acta/specs/…md>) — read FIRST, before any todo list. Its tasks are the ONLY tickets; no decomposition of your own.
SKILL: load build (omp: build) and tdd before the todo list, and follow them for every task.
TICKETS (anchor for your todo list — exactly these N):
  <TICKET-1> — <title> → verify: <the PROPERTY that must hold, over every path>
  <TICKET-2> — <title> → verify: <property>
WORKTREE: <abs path> (branch <slug>, parent <production>, base <base-sha>) — cd there FIRST, work ONLY there. Main checkout stays clean. No git checkout, no cd out, NO git push.
FILES: the plan names the area; find the exact lines yourself with lsp and grep. Any line number here is a hint, never the boundary — the defect may sit beside it. Surgical: every changed line traces to a ticket. NOTES never write NOTEs to memory — the orchestrator files them with `acta debt new` once the round is CLEAN.
HOUSE RULES: before the todo list, read `references/house-rules.md` of the acta plugin (two folders up from SKILL.md) and AGENTS.md in this worktree. Write that file's absolute path into the brief, because the recipient cannot resolve a relative path. The brief gives the job facts; those files give the rules.
MEMORY: before the todo list, read ~/.claude/memory/MEMORY.md and <project memory>/MEMORY.md (<project memory> = ~/.claude/projects/<main checkout abs path with every / and . turned into ->/memory, the MAIN checkout, never the worktree). They are indexes: open a linked note only when its hook fits a ticket. Read-only — never write there; your own omp memory keeps what you learn.
GATES (from the worktree): <the plan's fast test command>; typecheck; git diff --stat vs <base-sha> shows only plan files.
REPLY-BACK: after the last commit the build skill runs `acta reply-back`. Nothing else to hand-fill.
```

Include only sections that apply, keep the order. REPLY-BACK stays one line.

Briefs never list allowed acta commands; they may only forbid acta write commands the build skill does not call. A whitelist overrides the build skill and breaks it.

Why the rules live in a separate file instead of the brief: the recipient may never load the repo rules (omp does not by default), so the rules still have to reach it — but from one file, read once, instead of being retyped into every brief. If the recipient runs in a sandbox that cannot read outside its worktree, copy `references/house-rules.md` next to the brief and point the HOUSE RULES line there instead.

## Verify the reply — a claim, not a fact

Phase 2 starts because the user asked or the notification landed. Read lifecycle once (`herdr agent get <slug>`). `working` → one line, yield. Otherwise:

```bash
git log --oneline <base>..<head>
git diff --stat <base>..<head>
git diff --name-only <base>..<head>          # only plan files
<the plan's fast test command> 2>&1 | tail -20   # real counts, shown
```

Compare commits against the ticket list. Re-run the decisive mutation yourself on security/destructive changes. This is `acta:land` applied to someone else's claim: evidence in this turn, or it did not happen.

- List the bugs the recipient recorded: `git diff --name-only <base>..<head> -- .acta/bugs`. Name each in the landing report under `Bugs found by recipient:`, or write `none`.

**Verify ≠ review.** Verify asks "do tests catch the defects we thought of?" Review asks "what did we not think of?" Four clean mutation rounds in one session missed a symlink path-containment bypass, a swallowed audit record, a preview/cleanup divergence, and a scope bug. Re-review every round.

## Review — the self-review exception never applies here

**`acta:review`'s small-change self-review exception does NOT apply to a dispatch.** You did not write or watch the code; "a few lines" is what a drifted diff looks like from outside. Skill, every round, any diffstat. Everything else about the round — the two reviewers, the range, the plan path, the finding bar and the round cap — is `acta:review`.

## Fix rounds

A round that comes back with BLOCKERs goes out **in this same turn, automatically**. A findings list handed to the user with no dispatch behind it is an incomplete turn. The fix task itself, the round count and what happens after round 3 are `acta:review`; only the delivery is here:

1. **Exception — 1–2 findings, clear one-liners, no new context → inline `/goal`**, still naming the implementer-subagent rule. 3+ or reasoning needed → the fix task `acta:review` already defined.
2. **Reuse the same agent** — slug lookup first, prompt where it lives, no move. Fix round on the same plan in the same worktree → `/goal` only, **no `/new`**. Write the fix brief as if it remembers nothing (omp compacts); reuse is an optimisation.
3. Fix brief = same format, TICKETS = the new ticket ids, run `acta dispatch init` again first, and the REPLY-BACK line is the same one sentence as in the brief. No range in it: the recipient reads the head off the branch.
4. **Dispatch, then STOP and yield.** One turn per round.

## Close the tab

`acta:land` did the merge and removed the worktree. One step a dispatch adds, and only now: `herdr pane close <pane-id>`, with the pane id resolved from the slug and never from memory. `acta:land` has no pane step.

Report after landing, ADHD shape: merge SHA first, post-merge gate numbers, what was cleaned up, then two lines you owe the reader: `Bugs found by recipient:` naming every bug the recipient recorded under `.acta/bugs` in the diff, or `none`; and `Harvested from omp: <note titles>`, or `omp memory: nothing to harvest` when the file is missing or adds nothing. Then out-of-scope follow-ups the review surfaced, one next action last.

**Harvest omp's memory (omp recipient only), after the pane is closed.** Read `~/.omp/agent/memories/--<worktree abs path with every / turned into ->--/learned.md`, for example `--home-me-code-app-worktree--`. For each point, apply the memory test ("fresh agent opens this tomorrow — does missing this fact cost time or repeat a mistake?"). Passes and not already in Claude memory → write it as one note: project-only fact in `~/.claude/projects/<main checkout sanitized>/memory/`, fact true in every project in `~/.claude/memory/`, plus one index line in that folder's `MEMORY.md`. Already covered → skip. No user question; this is automatic.

Before you report done, sweep your own memory once: a gotcha that burned time or a dispatch mechanic that failed is a NOTE for `acta debt new` (`acta:review` owns that routing), an agreed convention or a decision belongs in `CONTEXT.md` or an ADR (`acta:brainstorm` owns those), a runbook goes to `.okf/`. Diff play-by-play is noise.