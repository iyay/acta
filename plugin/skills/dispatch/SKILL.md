---
name: dispatch
description: Use when pm:build runs with the dispatch executor, or the user asks to hand an approved plan to omp in another tab. Provisions the worktree and a dedicated herdr tab, hands off the plan's tasks through a short brief and /goal, then verifies from git, reviews with pm:review, routes BLOCKERs back as one fix task per round (three rounds at most), lands with pm:land, and closes the tab. Refuses without herdr or without an approved plan.
---

# dispatch — approved work in another pane

Hand approved tickets to an agent in another pane, then run verify → review → route → land. Runs on herdr.

**Skill notation.** Every `pm:name` here is a skill from the `pm` plugin. `mattpocock-skills:*`, `/to-spec`, `/to-tickets`, `/implement`, `/code-review` are gone; a brief still naming them is stale — rewrite it before sending. "Ticket" below = one `pm:plan` task, mirrored in the tracker when the repo has one.

## Entry gate — no dispatch without an approved plan

Dispatch is the handoff step of `pm:build`. Before you touch a pane, all of these are true:

- `pm:brainstorm` produced the design, the user said yes, `pm:plan` produced the plan at `.pm/plans/`, and the user said yes to that too. Tracker present → every task is a ticket there.
- The plan carries confirmed files + symbols. The recipient gets facts, never "go look".
- Every ticket has a `verify: [check]` line.

Any missing → STOP, report which one, do not provision. "Small" is not an exemption.

## Step -2 — Detect herdr FIRST

```bash
printf 'herdr=%s\n' "${HERDR_ENV:-}"
```

- `HERDR_ENV=1` → read [herdr-delivery.md](herdr-delivery.md).
- Anything else → STOP: "dispatch requires herdr". Do not fall back to an in-process subagent — different execution model.

## Autonomous loop — HARD RULE, no user in the loop after plan approval

Plan approval is the one yes you need. From then on the loop runs itself, end to end:

1. **Reply-back lands → review starts in that same turn.** The notification IS the trigger. Do not report "the agent is done" and wait for the user to say "review it". Do not ask "shall I review?". Verify from git, then invoke `pm:review` immediately.
2. **Review not clean → route fixes back in that same turn.** BLOCKERs → ONE fix ticket via `pm:plan` → prompt the same agent by slug → yield. Rounds 1 and 2 only; a BLOCKER after round 3 stops the loop (see Autonomy). No "here are the findings, want me to send them back?". The user sees findings only as a one-line status while the fix round is already dispatched.
3. **Review clean + every ticket done → land in that same turn.** Merge `--no-ff`, re-gate, remove worktree, delete branch, close tab. No "ready to merge, shall I?". The approved plan already covers landing (`pm:land`).

Asking the user for any of these three is a violation, same class as skipping `pm:review`. The user reads the turn's report after the fact.

The only things that stop the loop and go back to the user are the blockers listed under "Autonomy" below — a human decision, a BLOCKER still open after review round 3, a Destructive-list command, a merge you cannot resolve, red post-merge gates. Everything else is yours.

## Autonomy — decide inside the phase, then yield

Inside a phase every call is yours: provision, brief, deliver, checkpoint, verify, review, route, re-review, land. No mid-phase options, no "which pane?". Never `git push`.

Autonomy is not a licence to hold the turn. Two phases, hard stop between them (see "Never wait").

Phase stops early on a blocker report, not a permission ask:

- No multiplexer.
- A finding needing a human decision — design, scope, tradeoff, plan conflict.
- **Round cap** — review round 3 still has a BLOCKER. `pm:review` allows three review rounds per plan; round 4 never starts on your own. STOP, report the BLOCKER, ask the user: land anyway, fix, or revert to `ROUND1_HEAD`. New BLOCKERs do not buy more rounds.
- **Delivery stuck** — delivery that will not submit after ~2 attempts.
- Drift at the comprehension checkpoint (correct, re-dispatch, then stop as usual).
- Destructive-list command needed (`rm -rf`, `--force`, `drop`, `truncate`, `git reset --hard`, mass delete) → plain sentences, wait for the user.

## Never wait for the recipient — HARD RULE

**After the comprehension checkpoint, Phase 1 is over. End your turn.**

Forbidden until the user's next message or the reply-back notification: `herdr agent wait`, polling loops, repeat pane reads, `sleep`, `Monitor` until-loops, any long-timeout call whose job is to sit.

Why: the recipient takes minutes to hours; blocking burns your context and freezes the user's session. Its completion signal is unreliable anyway (idle can mean stopped early), so you re-derive every number from git whenever you look. Waiting adds nothing; looking later costs nothing.

One exception: the ~20s pause before the single checkpoint read.

Phase 1 report, ADHD shape (first line = fact, last line = one next action): base SHA, worktree path, branch, slug + pane id, ticket ids handed over, checkpoint verdict, any call you locked in the brief, then the literal next step ("on reply-back: verify `<base>..HEAD` → `pm:review` → fix rounds → land, no further ask"). Then yield.

Phase 2 resumes on either trigger: the REPLY-BACK notification lands, or the user asks. **Neither needs the other.** A notification alone starts verify → review → route/land in that turn; do not wait for the user to confirm it.

## The loop at a glance

**Phase 1 — dispatch (one turn, hard stop):**

1. **Entry gate** passed (above).
2. **Provision** (Step -1): worktree outside the repo, then look the agent up by slug and reuse; create its own tab only on a miss. Never a split beside you.
3. **Write the hand-off doc** (Step 0a). Pointer at plan + tickets plus the mechanics no skill supplies. Not a rewritten plan.
4. **Deliver**: `/new` on a new session/plan/worktree, then always `/goal` carrying the brief path and the literal reply-back command.
5. **Comprehension checkpoint**: one read ~20s later — todo list must name the ticket ids, else drift; correct and re-dispatch.
6. **STOP. Report and yield.**

The recipient is the main agent in its pane: writes zero code itself, **fans every ticket out to implementer subagents per `pm:build` — as many in parallel as file ownership allows** (see "Spread the work"), commits per ticket, does NOT review, fires reply-back, never touches its tab.

**Phase 2 — verify, review, land (later turn):**

7. **Verify** from git. Verify ≠ review.
8. Tab: nothing to do. A stray pane split beside you → move it to its own tab once.
9. **Review** via `pm:review` + plan ref — every round, no exceptions, Spec axis + Standards axis. Range per round: round 1 `<base>..<head>`; round 2 `<ROUND1_HEAD>..<head>` plus the direct callers of every function the fix touched; round 3 `<ROUND2_HEAD>..<head>` only.
10. BLOCKERs or unfinished tickets → ONE fix ticket via `pm:plan` that holds all of them, route it to the same agent by slug, **stop and yield**. One turn per round. Round 3 still BLOCKED → stop, ask the user (see Autonomy).
11. **Clean AND complete → land** (merge `--no-ff`, re-gate, remove worktree, close tab). Never leave a reviewed branch parked.

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
| Between rounds | nothing |
| Stray split pane | `herdr pane move <pane> --new-tab --no-focus --label <slug>` |
| Reply-back | in brief AND inline in `/goal` (`herdr agent prompt $HERDR_PANE_ID …`) |

Map, not instructions. Flags and failure handling live in the reference file.

## `/new` and `/goal` — every dispatch, every backend

- **`/new` first** on a new session, new plan, or new worktree. Skip only for a fix round on the same plan in the same worktree — context is the asset there. In doubt → send it; the brief is on disk.
- **`/goal` every turn, no exceptions**, including fix rounds and re-dispatch after drift. The goal survives the harness's own compaction; a plain prompt does not. One `/goal` per turn, one brief, no bare follow-up.
- `/goal` carries the reply-back command inline.

Order: `/new` → `/goal <one-line summary> + read brief at <abs-path> + literal reply-back`. Back-to-back, no settle-wait. Checkpoint after the `/goal`.

**A held goal silently refuses the next one — on omp `/new` does NOT clear it.** Symptom: prompt returns success, pane shows `Warning: Resume the current goal first, or drop it…`, status bar `⏸ Goal 0`, agent idle having never seen your objective. Clear it:

```bash
herdr agent prompt <slug> "/goal drop"
herdr agent send-keys <slug> enter
herdr agent send-keys <slug> enter
```

Then send the new `/goal`; confirm `🎯 Goal <n>K` in the status bar. Bites on every re-dispatch after drift or interruption; use `--source visible` to see the warning.

Harness with no `/goal` → plain prompt, and say so in the report.

## Reply-back — every brief, every backend

Every brief ends with a REPLY-BACK line, and every `/goal` repeats it inline. Belt and braces: it is the single most-skipped instruction.

Literal, runnable, one line, address + range pre-filled. The payload is the review command itself, so the notification is the trigger:

```
REPLY-BACK — run this VERBATIM after the last commit. No summary, no test counts, no verdict:
herdr agent prompt wJ:p1 "/pm:review <base-sha>..<new-head-sha> — plan <plan-path-or-ticket-ref>, round <slug>, pane $HERDR_PANE_ID"
```

- Range first and bare. Notes after an em-dash: the plan ref (`requesting-code-review` needs the plan for its `{PLAN_OR_REQUIREMENTS}` slot, which is the Spec axis) and `pane $HERDR_PANE_ID` (expands in the recipient's shell → names the pane that produced it).
- Scope only, never numbers or verdict. Self-reported figures were wrong 5/5 in one session.
- No self-close line. The tab is reused; you close it at landing.
- Arrival does not excuse verify. A recipient can fire it after 2 of 6 tickets.
- Blocked → same command, blocker as prose in place of the range.

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

Set it up before dispatching security, auth, data-migration or money work; check with `/advisor` inside the recipient pane. An advisor note is cheaper than a fix round by roughly the cost of a full `pm:review`.

Also under-used, and worth naming in briefs: **`lsp`** answers "every caller of X" mechanically (references, rename, diagnostics) where grep guesses; **`todo`** carries phase tracking; **`agent://<id>/findings.0.path`** pulls a typed field straight out of a subagent's result.

## Step -1 — Provision worktree, then tab

Worktree first — the tab's pane is placed into it. `pm:build` worktree rules apply (no consent prompt); here you run the git fallback yourself because the worktree must exist before the recipient's pane does.

1. **Worktree, outside the repo, per `pm:build`:**

   ```bash
   REPO_ROOT=$(git rev-parse --show-toplevel)
   REPO=$(basename "$REPO_ROOT")
   git -C "$REPO_ROOT" worktree add "../${REPO}-<slug>" -b <slug> production   # other parent only if the user named it
   WT=$(cd "$REPO_ROOT/../${REPO}-<slug>" && pwd)                              # absolute path for pane commands
   ```

   Record the parent branch in the brief; landing needs it. Never in-tree (`.claude/worktrees/`, `worktrees/`): recursive scanners (Shopify CLI, Vite, jest globs, tsc, eslint, docker) find both copies and fail — real breakage 2026-08-21. `.gitignore` hides it from git only.

   Create, do not enter. You orchestrate from the main checkout.

   Inherited in-tree worktree → `git worktree move <old> <new>`, re-point brief, re-dispatch. Stop a mid-round agent first; expect the held-goal block.

2. **Reuse the worktree's agent by slug on every turn.** `herdr agent get <slug>` is the first call on any turn touching the recipient. Resolves → reuse. Misses → create a tab, rename agent. Never a second tab for one worktree (two agents fight over one branch). Never scan for a "free" pane — `agent == ""` matches dev servers and tunnels; three mis-prompts in one session.

3. **Launch the harness in the worktree**; wait for its prompt.

**Worktree prep the recipient cannot do.** Run all four; each has burned a dispatch.

```bash
ln -s <repo>/node_modules <worktree>/node_modules          # plus backend/.venv where one exists
cp <repo>/.env.* <worktree>/                                # git-ignored env files the suite needs
cp -R <repo>/.scratch/<feature> <worktree>/.scratch/        # gitignored tracker export, if the plan is not yet committed on the branch
cp <repo>/AGENTS.md <repo>/CLAUDE.md <worktree>/ 2>/dev/null # house rules, see below
```

The plan under `.pm/plans/` is a tracked file: commit it on the branch as the first commit so it clones with the worktree. The brief forbids `npm install`.

**House rules do not always reach the worktree, and omp reads them natively.** omp ingests `AGENTS.md` (and Cursor MDC, `.clinerules`, Copilot `applyTo`) in their native shape — so a worktree that has the file gets the rules for free, and one that doesn't gets nothing. Two ways the file goes missing:

- `AGENTS.md` is a **symlink to CLAUDE.md** and CLAUDE.md is untracked or excluded (`.gitignore`, `.git/info/exclude`) → the worktree gets **neither**. Observed 2026-09-03: a backend agent ran a whole ticket with zero house rules and shipped `hasattr` guards the repo's own ADRs had removed, plus an untyped helper in a strict-typing codebase.
- The repo keeps its rules only in an untracked local file.

Check before dispatching: `ls <worktree>/AGENTS.md <worktree>/CLAUDE.md`. Missing → copy them in. The brief's restates are a floor, not a substitute: a restate carries five lines, the file carries the whole standard.

## Step 0a — Write the hand-off doc

Instructions live in a file, never in the message. Multi-line pastes fragment in exactly the harnesses that need them; a file reads atomically, survives `/new` and compaction.

Path inside the repo the recipient works in: `.claude/dispatch/<slug>-brief.md` (git-excluded; check `git check-ignore -v .claude/`). Not `/tmp` — sandboxed harnesses cannot read outside their tree.

**What the brief is now.** With `pm:brainstorm` + `pm:plan` upstream and `pm:build` downstream, the brief no longer restates tasks, decisions, or method. It is:

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

**Blocking = a concrete input that produces a wrong result for a real user, today.** Same bar as `pm:review`: BLOCKER or NOTE, nothing between. Everything else is logged as one NOTE (memory) or one follow-up ticket, and the branch lands.

Not blocking, however tempting: test durability against hypothetical future edits, mock fidelity for props nothing passes, a substring that could be reworded around, naming, an assertion with no message. Those are real and they are cheap to log. Routing them as fix rounds is how a branch that was correct after round 2 reaches round 6.

Security, auth, data-migration and money paths keep the full bar — there, "a future edit could" often means "a caller today can".

## Spread the work — maximum subagents, every brief, every backend

**The recipient is an orchestrator, not a typist.** `pm:build` already forbids it from writing code and already says fresh subagent per task; this section makes it fan out as wide as the ticket list allows. A recipient that runs six tickets through one implementer subagent serially is the slowest and least reviewable shape available: one context accumulates every file and every failed attempt, and by ticket 5 it has compacted away ticket 1's constraints.

The PARALLEL rules live in `references/house-rules.md` now, read before the todo list — no need to retype them into the brief. The `/goal` still has to carry the push: the `orchestrate` keyword's contract (decompose, dispatch subagents, parallelize disjoint work in ONE message, verify, never yield before closure) says the same thing, so put it in every `/goal` — it is skipped as often as the reply-back line.

- **One implementer subagent per ticket, minimum.** The recipient never implements a ticket in its own context. Every subagent runs its own `pm:tdd` loop: red first, mutation-verify, surgical, commits its own ticket.
- **Parallel by default, serial by exception.** Independent tickets go out **in one message, several at once** — not one after another. Something must force serialization: the same file touched twice, a real dependency (B imports what A creates), or a migration that must land alone.
- **Partition by file ownership.** Two subagents in one wave must never hold the same file — the one collision that produces silent lost edits in a shared worktree. Group tickets into waves on that basis before dispatching anything.
- **Split further where the plan allows.** A ticket that names two disjoint files can go to two subagents (one per file, tests with the code they test). A ticket with an interface seam can split into "write the failing tests" and "make them green" only when the test author commits first — otherwise keep them together.
- **Read-only work fans out cheapest.** Codebase probes, API-surface checks, "where is X used", lookups the plan did not already answer — dispatch those in parallel even when implementation must be serial (`pm:build`). Never block a wave on a probe one subagent could run beside it.
- **Waves declared up front** in the todo list ("wave 1: T-1, T-3, T-4 parallel · wave 2: T-2 — depends on T-1"), so drift is visible at the comprehension checkpoint.
- **Parallelism never relaxes the gates.** No batching commits at the end of a wave, no "test it all together", no shared scratch files between subagents.
- **Skills are HOW, not WHO.** Subagents are plain `general-purpose` implementers; `pm:tdd` runs inside them. Never a skill name in a `subagent_type` slot.
- **OVERRIDE on the skill:** the recipient skips `pm:build`'s per-task reviewer and its "fix round R of 5" loop. Review is yours, once per round, from your side (`pm:review`).

You require this; you do not police the exact wave count. A genuinely dependent chain run serially is correct. Six independent tickets run serially is a finding — say so in the fix round, and re-dispatch the remaining tickets as a parallel wave.

## Comprehension checkpoint

~20s after the `/goal`, one read. Two things must be true in the todo list:

1. It names the ticket ids from the brief. Invented modules/phases/endpoints = it confabulated before reading → interrupt key, re-dispatch with a corrective preamble ("there is NO `<X>`, NO `<Y>` — writing those = drift").
2. It declares waves, and wave 1 holds every ticket that has no dependency. A flat serial list of independent tickets = fan-out drift → interrupt, re-dispatch with the wave plan spelled out ("wave 1: T-1, T-3, T-4 in ONE message").

Never skip for destructive work or weaker harnesses; both kinds of drift track model tier.

## Hand-off brief format

```
<one-line summary of the job>

PLAN: <.pm/plans/…md> (design: <.pm/specs/…md>) — read FIRST, before any todo list. Its tasks are the ONLY tickets; no decomposition of your own.
TICKETS (anchor for your todo list — exactly these N):
  <TICKET-1> — <title> → verify: <the PROPERTY that must hold, over every path>
  <TICKET-2> — <title> → verify: <property>
WORKTREE: <abs path> (branch <slug>, parent <production>, base <base-sha>) — cd there FIRST, work ONLY there. Main checkout stays clean. No git checkout, no cd out, NO git push.
FILES: the plan names the area; find the exact lines yourself with lsp and grep. Any line number here is a hint, never the boundary — the defect may sit beside it. Surgical: every changed line traces to a ticket.
HOUSE RULES: before the todo list, read `references/house-rules.md` of the pm plugin (two folders up from this skill) and AGENTS.md in this worktree. Write that file's absolute path into the brief, because the recipient cannot resolve a relative path. The brief gives the job facts; those files give the rules.
MEMORY: before the todo list, read ~/.claude/memory/MEMORY.md and <project memory>/MEMORY.md (<project memory> = ~/.claude/projects/<main checkout abs path with every / and . turned into ->/memory, the MAIN checkout, never the worktree). They are indexes: open a linked note only when its hook fits a ticket. Read-only — never write there; your own omp memory keeps what you learn.
GATES (from the worktree): <one-shot test runner>; typecheck; git diff --stat vs <base-sha> shows only plan files.
REPLY-BACK — run VERBATIM after the last commit. Scope only. Blocked? same command, blocker as prose.
  herdr agent prompt <orchestrator-pane> "/pm:review <base-sha>..<new-head-sha> — plan <path>, round <slug>, pane $HERDR_PANE_ID"
```

Include only sections that apply, keep the order. REPLY-BACK stays one line.

Why the rules live in a separate file instead of the brief: the recipient may never load the repo rules (omp does not by default), so the rules still have to reach it — but from one file, read once, instead of being retyped into every brief. If the recipient runs in a sandbox that cannot read outside its worktree, copy `references/house-rules.md` next to the brief and point the HOUSE RULES line there instead.

## Verify the reply — a claim, not a fact

Phase 2 starts because the user asked or the notification landed. Read lifecycle once (`herdr agent get <slug>`). `working` → one line, yield. Otherwise:

```bash
git log --oneline <base>..<head>
git diff --stat <base>..<head>
git diff --name-only <base>..<head>          # only plan files
<one-shot test runner> 2>&1 | tail -20       # real counts, shown
```

Compare commits against the ticket list. Re-run the decisive mutation yourself on security/destructive changes. This is `pm:land` applied to someone else's claim: evidence in this turn, or it did not happen.

- List the bugs the recipient recorded: `git diff --name-only <base>..<head> -- .pm/bugs`. Name each in the landing report under `Bugs found by recipient:`, or write `none`.

**Verify ≠ review.** Verify asks "do tests catch the defects we thought of?" Review asks "what did we not think of?" Four clean mutation rounds in one session missed a symlink path-containment bypass, a swallowed audit record, a preview/cleanup divergence, and a scope bug. Re-review every round.

## Review — `pm:review`, your side, every round

The notification's only correct response: verify, then `pm:review` over this round's range (below) with the plan path in `{PLAN_OR_REQUIREMENTS}`. Runs in **your** session. Per `pm:review` you dispatch TWO read-only reviewer subagents in parallel (`pm:build`): Spec axis (faithful to the approved design + plan?) and Standards axis (repo standards + Fowler smells on the changed lines). Never prompt the implementor to review itself.

**`pm:review`'s small-change self-review exception does NOT apply here.** You did not write or watch the code; "a few lines" is what a drifted diff looks like from outside. Skill, every round, any diffstat.

- Name the skill literally, with the full prefix.
- Range explicit, per round. Round 1: `<base>..<head>`. Round 2: `<ROUND1_HEAD>..<head>` plus the direct callers of every function the fix touched; fixed code on a trust boundary, auth, money, migration or delete path gets the full deep lens again. Round 3: `<ROUND2_HEAD>..<head>` only. `ROUND1_HEAD` / `ROUND2_HEAD` = the head you reviewed in that round. No range → the skill defaults to `HEAD~1..HEAD` and reviews the wrong span.
- Pass the plan path after an em-dash.
- Findings only; reviewers never edit. Fixes = new tickets.
- Findings come back through `pm:review`: verify each against the code before routing, push back on wrong ones with reasoning — a bad finding routed as a ticket is a wasted round.

User explicitly wants a *different* pane to review → that message needs: literal skill name, a review keyword, and the range.

## After a review — route as one fix ticket, three rounds max

**Clean** = Spec axis matches AND zero BLOCKERs (the skill's Critical with a reproducible scenario; Important/Minor = NOTE). Clean AND every ticket done → "Landing", same turn, no ask.

Not clean after round 1 or 2 → the fix round goes out **in this same turn, automatically**. A findings list handed to the user with no dispatch behind it is an incomplete turn. Steps:

1. **BLOCKERs → ONE fix ticket via `pm:plan`.** Append a `## Fix round <n>` section to the same plan file with ONE task that lists every BLOCKER: `file:line`, wrong vs expected, the mutation that proves each fix, one `verify:` line per BLOCKER. The whole task lands as one commit. NOTEs go to memory, one line each, never a ticket. Mirror the task to the tracker like any other ticket. No inline findings list in the prompt — an inline list is what makes a weak harness freestyle.
   Exception: 1–2 findings, clear one-liners, no new context → inline `/goal`, still naming the implementer-subagent rule. 3+ or reasoning needed → the fix task in the plan.
2. **Reuse the same agent** — slug lookup first, prompt where it lives, no move. Fix round on the same plan in the same worktree → `/goal` only, **no `/new`**. Write the fix brief as if it remembers nothing (omp compacts); reuse is an optimisation.
3. Fix brief = same format, TICKETS = the new ticket ids, REPLY-BACK over `<fixed-from>..<new-head>`.
4. **Dispatch, then STOP and yield.** One turn per round.
5. Return → verify → `pm:review` again over the fix range (round 2: `<ROUND1_HEAD>..<head>` plus direct callers; round 3: `<ROUND2_HEAD>..<head>`). Fixes introduce defects at draft rate; in one session three rounds each added one, twice because the dispatch instruction itself was wrong. A green suite cannot tell you that.

**Three review rounds per plan, hard cap (`pm:review`).** Rounds 1 and 2 are the same three moves in one turn: verify → review → (clean ? land : dispatch the one fix ticket). Never hand the user a findings list and stop after round 1 or 2; never ask "another round?". Round 3: clean → land; BLOCKER → STOP, report it, ask the user: land anyway, fix, or revert to `ROUND1_HEAD`. Round 4 never starts on your own. A round with zero BLOCKERs does not start.

Escalate before the cap only when a finding needs a human decision. New findings do not reset or extend the round count.

## Landing — WORKTREE LANDING, from the main checkout

Clean and complete → you merge and tear down **in the same turn the clean verdict arrives**. Never ask; never report "ready to merge". The approved plan covers it (`pm:land`: "land without asking" — the options menu is overridden, its test gate and cleanup steps still apply).

1. **Preconditions, all:** every ticket done · nothing uncommitted in the worktree · typecheck + full suite green, output shown · verdict explicitly clean. Any missing → report blocker, no merge.
2. **Contamination check:**
   ```bash
   git diff --name-only <base>..<head> | grep -E 'node_modules|\.venv|\.env'   # must be empty
   git ls-tree -r <head> | awk '$1=="120000"{print $4}'                        # symlinks: must be empty
   ```
   Non-empty → do not merge; strip from the branch first.
3. **Parent** = branch recorded at worktree creation (or `git rev-parse --abbrev-ref <branch>@{u}` / merge-base). Unsure → STOP, ask. Main checkout: `git status --porcelain` empty, `git rev-parse --abbrev-ref HEAD` = parent.
4. **Merge:** `git -C <main-checkout> merge --no-ff <slug> -m "<what landed>. Verified at merge: <gate numbers>"`. Conflict → resolve hunk by hunk, by intent from each side's primary source, finish the merge (no pm skill for this). Never `--abort`, never discard a side. Unresolvable → STOP, worktree intact, report.
5. **Re-run the gates on the merge result.** A merge can break what both sides passed alone. Red → say so plainly, leave the merge for the user.
6. **Tear down:** `git worktree remove <path>`; `git branch -d <slug>` (plain `-d`: refusal = not fully merged = STOP). Then `herdr pane close <pane-id>` (resolve from slug, not memory). The one place a dispatch tab is closed.
7. **NEVER `git push`.**

The only cases that stop landing and go to the user: a finding needing a human decision; the parent is one the user has called protected; the merge conflict cannot be finished by intent; post-merge gates red. An open BLOCKER after round 1 or 2 is not one of these — it goes back to the agent as the fix ticket. After round 3 it goes to the user (see "After a review").

Report after landing, ADHD shape: merge SHA first, post-merge gate numbers you ran, what was cleaned up, bugs recorded from the recipient, out-of-scope follow-ups the review surfaced, one next action last.

## Memory sweep before "done"

Memory: before reporting done, one sweep — a gotcha that burned time, a dispatch mechanic that failed, an agreed convention → write it now. Routing: domain term → `CONTEXT.md`, decision → ADR, runbook → `.okf/`, session-scoped → memory. Diff play-by-play is noise.

**Harvest omp's memory (omp recipient only), after the pane is closed.** Read `~/.omp/agent/memories/--<worktree abs path with every / turned into ->--/learned.md`, for example `--Users-iyay-Nayakatara-PMIS-codes-febe-be-pmis-.claude-worktrees-qty-fix--`. For each point, apply the memory test ("fresh agent opens this tomorrow — does missing this fact cost time or repeat a mistake?"). Passes and not already in Claude memory → write it as one note: project-only fact in `~/.claude/projects/<main checkout sanitized>/memory/`, fact true in every project in `~/.claude/memory/`, plus one index line in that folder's `MEMORY.md`. Already covered → skip. No user question; this is automatic. Landing report gets one line: `Harvested from omp: <note titles>`, or `omp memory: nothing to harvest` when the file is missing or adds nothing.
