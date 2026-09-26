# herdr delivery

Every `acta:name` here is a skill from the `acta` plugin, same as SKILL.md. `mattpocock-skills:*` is gone.

## Completion signal — a reply-back push plus a one-shot lifecycle read, never a wait

Two independent channels. Use both; they cover each other's failure mode.

**Channel 1 — the reply-back push (required in every brief).** The recipient's final action fires the review command straight into your session:

```bash
herdr agent prompt <your-pane-id> "/acta:review <base-sha>..<new-head-sha> — plan <path>, round <slug>, pane $HERDR_PANE_ID"
```

Your pane id is `$HERDR_PANE_ID` — resolve it at dispatch time and embed it **literally** in the brief; the recipient cannot look it up.

**`$HERDR_PANE_ID` appears twice in that one line and means two different panes.** In the *address* it must arrive already expanded to **your** id (`wJ:p1`). In the *payload* it must arrive **unexpanded**, so it resolves in the recipient's shell to **its own** id. In a brief file on disk that is automatic. When you build the `/goal` in your own shell it is not: write the address expanded and escape the payload one (`\$HERDR_PANE_ID`). Verified 2026-08-06: this surfaces in the orchestrator's Claude Code session as a mid-turn message, and a slash command delivered this way executes, so the review is self-triggering.

Payload rules: the range **first and bare**, then after an em-dash the plan path (the `{PLAN_OR_REQUIREMENTS}` slot of `requesting-code-review` — the Spec axis needs it) and the pane note. Nothing else — no test counts, no summary, no verdict. Still run the verify step when it lands.

Its failure mode: the reply-back line is the single most-skipped instruction in a dispatch. Its silence proves nothing.

**Channel 2 — the lifecycle pull.** Read it once, on a later turn:

```bash
herdr agent get <slug>          # .result.agent.agent_status → working | idle | blocked
```

**Do NOT use `herdr agent wait`.** It blocks the orchestrator's turn for as long as the recipient works. See "Never wait for the recipient" in SKILL.md.

`idle` can also mean the recipient stopped early, and `unknown` never proves completion. Compare `git log` and `git diff --stat` against the brief's ticket list before reviewing. Because the lifecycle proves so little, **when you read it does not matter** — which is why waiting on it is wasted.

`working` → say so in one line and yield; `idle`/`done` → verify from git, then `acta:review`, then route or land — all in this turn, no ask; `blocked` → see Failure handling.

`idle` means ready for input and its tab was seen in the focused UI; `done` is the same idle state after unseen background work. CLI reads do not mark a tab seen.

## Provision the tab — ONE named tab per worktree, REUSED every round, never self-closed

**The worktree must already exist** (SKILL.md Step -1).

The dispatch agent is named after the worktree slug and lives in its OWN TAB. Look it up by name first and reuse. Create a tab only when the name does not resolve — never split beside the orchestrator.

Why: a reused agent keeps the plan it read, the code it navigated, the gates it ran. Throwing the pane away each round measurably degraded output (observed 2026-08-06, six rounds on one branch).

### 1. Look it up by name

```bash
SLUG=<worktree-dir-name>        # the SAME slug every round
herdr agent get "$SLUG" 2>/dev/null | jq -r '.result.agent | "\(.pane_id) \(.agent) \(.agent_status) \(.cwd)"'
```

Name lookup resolves the same from any tab (verified 2026-08-06). Three outcomes:

- **Resolves, `agent` is `omp`, `cwd` is the worktree** → reuse in place. `/new` if this round is a new plan, then the `/goal` pointer. No move, no relaunch.
  ```bash
  PANE=$(herdr agent get "$SLUG" | jq -r '.result.agent.pane_id')
  ```
  Only move you ever make: evicting a legacy pane split beside you — `herdr pane move "$PANE" --new-tab --no-focus --label "$SLUG"`.
- **Resolves but `agent_status` is `working`** → a previous round still runs. Do not prompt; you would interleave two briefs. Report and stop.
- **Does not resolve** → create it, once, per step 2.

**Never scan `herdr pane list` for a "free" pane.** `agent == ""` matches a dev server, `uvicorn`, a `cloudflared` tunnel, a log tail — prompting it types into that process's stdin (three times in one session, 2026-08-06).

### 2. Create it only when the name does not resolve — as a TAB, never a split

```bash
MY_WS=${HERDR_PANE_ID%%:*}   # worktrees live outside the repo, so pin the workspace
TAB_JSON=$(herdr tab create --cwd <worktree-abs-path> --workspace "$MY_WS" --label "$SLUG" --no-focus)
PANE=$(printf '%s' "$TAB_JSON" | jq -r '.result.root_pane.pane_id')
TAB=$(printf '%s'  "$TAB_JSON" | jq -r '.result.tab.tab_id')
herdr pane run "$PANE" "omp"
# wait for the harness prompt, then:
herdr agent rename "$PANE" "$SLUG"
```

Response shape verified 2026-08-18: `.result.root_pane.pane_id` and `.result.tab.tab_id`. Always `--cwd <worktree>`, `--no-focus`, `--label "$SLUG"`. Rename immediately — the name is the reuse handle.

**Never `herdr pane split`.**

### 3. Between rounds: nothing happens

The tab stays for the life of the worktree. No parking, no moves. The separate `herdr` skill covers the full surface.

### 4. The recipient touches NO pane, tab, or workspace

The brief carries zero pane/tab commands; the recipient's only herdr command is the reply-back `herdr agent prompt`. You close the pane yourself in "Landing" (SKILL.md), after merge and worktree removal:

```bash
herdr pane close "$PANE_ID"   # positional argument; the emptied tab auto-closes
```

### A different worktree means a different agent

Never re-point an existing dispatch pane at a different worktree: its context refers to the old tree.

### Reuse does not make the brief optional

`omp` compacts on long sessions. The brief, the plan and the tickets stay authoritative and self-contained; reuse is an optimisation.

## Launch the recipient harness

Default kind is **omp**, through the interactive shell so the alias expands (`omp` = `headroom wrap omp --yolo`):

```bash
herdr pane run <pane-id> "omp"
```

Poll detection up to 30 seconds, then name:

```bash
herdr agent get <pane-id>          # repeat until it returns an agent
herdr agent rename <pane-id> <slug>
```

**Do not** default to `herdr agent start --kind omp`: it bypasses the alias and drops `headroom wrap` and `--yolo`. Without `--yolo` an approval prompt parks the recipient in `blocked`.

Fallback only if detection misses within 30s:

```bash
herdr agent start <slug> --kind omp --pane <pane-id> -- --yolo
```

Report that the `headroom wrap` layer was lost. Other kinds only when the user names one. Names match `[a-z][a-z0-9_-]{0,31}`.

**acta must be installed in the recipient harness too.** Skills do not travel with the worktree; each harness installs its own copy. Before the first dispatch on a machine: check the harness lists `acta:dispatch`. Missing ‒ the recipient will freestyle every `acta:` line in the brief — STOP and report, do not dispatch.

## Agent lifecycle across rounds

- Reuse the worktree's agent when the slug resolves; else fresh tab.
- Fix rounds re-prompt the same named agent.
- `/new` on a new session, plan, or worktree; `/goal` every turn. A fix round on the same plan in the same worktree gets no `/new` but still a `/goal`.

## Deliver the pointer message

New session, plan, or worktree — HARD RULE, four steps, a check after each (same as SKILL.md):

```bash
# 1. omp ready: its empty input box is on screen
herdr agent read <slug> --source visible --lines 10
# 2. /new alone, then confirm "New session started", then rename by pane id
herdr agent prompt <slug> "/new"
herdr agent read <slug> --source visible --lines 10    # must show: New session started
herdr agent rename <pane-id> <slug>
# 3. /goal alone, then confirm 🎯 Goal in the status bar
herdr agent prompt <slug> "/goal ultrathink orchestrate <one-line summary>. FIRST read the hand-off at <abs-brief-path> and obey every line — it names the plan and ticket ids, the worktree, the gates. You are the main agent here: write ZERO code yourself. Run acta:build with as MANY implementer subagents as the tickets allow: one per ticket MINIMUM, each driving acta:tdd — failing test first. Review happens on my side, not yours. Group tickets into waves by file ownership and dispatch every wave in ONE message, several subagents at once; serial only for a shared file or a real dependency. Declare the waves in your todo list before dispatching. Do NOT move, park, close, or create any pane or tab. When your last ticket is committed, run this VERBATIM: herdr agent …
herdr agent read <slug> --source visible --lines 6     # status bar must show: 🎯 Goal
```

`/new` and `/goal` are never in one prompt, and `/new` never goes to an omp that is still starting.

**The very first action of every task, before the failing test:** `acta tick plans/<stem>#task-N --start --agent omp` (the brief names omp because the recipient harness sets no agent variable).

**The brief's tick rule carries the name.** The recipient is an omp pane and its harness sets no agent variable, so every tick command in the brief names it:

```bash
acta tick plans/<stem>#task-N --step <n> --agent omp
acta tick plans/<stem>#task-N --all --agent omp   # right after the ticket's commit
```

`acta tick` writes that name into the worktree's git-ignored `.agents.json`, which is what the board reads to show who works on what. Without the flag the name is missing and the row stays blank.

Step 4, fix round: `/goal` only, on the plan the agent already holds, same inline tail; still confirm 🎯 Goal:
```bash
herdr agent prompt <slug> "/goal ultrathink orchestrate <one-line summary>. FIRST read <abs-brief-path>. Fix the PROPERTY, not the reported case: enumerate every path that could break it. Tickets <ids> → one implementer subagent each (acta:build, acta:tdd inside), all independent ones in ONE message; declare waves first. When your last ticket is committed, run VERBATIM: herdr agent prompt $HERDR_PANE_ID \"/acta:review <fixed-from>..<new-head> — plan <path>, round <slug>, pane \$HERDR_PANE_ID\""
herdr agent read <slug> --source visible --lines 6     # status bar must show: 🎯 Goal
```

**Quoting**: `$HERDR_PANE_ID` in the *address* expands in your shell (your literal pane id); `\$HERDR_PANE_ID` in the payload stays unexpanded for the recipient.

**Backticks in an unquoted heredoc are command substitution.** `MSG=$(cat <<EOF … EOF)` executes every backtick pair — markdown `` `code` `` in your prompt gets replaced by empty output. Symptom: `command not found: <word>` in the tool result while the prompt still delivers. Quote the delimiter (`<<'EOF'`) or use plain quotes in prose. Survivable because the brief is on disk.

Never a bare pointer with no `/goal` in front — the goal survives compaction; a plain prompt does not.

`agent prompt` submits text and Enter atomically, so the pointer may be multi-line. Do not pass `--wait`.

## Magic keywords survive only in prose

`ultrathink` and `orchestrate` are prepended to the `/goal` text above deliberately. omp ignores a magic keyword inside a fenced block, an inline code span, or an HTML/XML tag, and ignores it when letters, digits, hyphens, slashes, dots or call syntax touch it. So `ultrathink` in the prompt string works; wrapping it in backticks to look tidy kills it, and so does `orchestrate()` or `orchestrated`.

They apply to one turn only. Repeat them in every round, fix rounds included.

## Check the advisor before dispatching risky work

`herdr agent prompt <slug> "/advisor"` reports whether a reviewer model is paired to omp's `advisor` role. When one is, it reads every turn on its own context and injects concerns or hard blockers inline, so a narrowed fix gets challenged during the work rather than two review rounds later. Unset on a security, auth, data-migration or money dispatch → say so in the Phase 1 report; one advisor note costs less than a `acta:review` round.

## Comprehension checkpoint

About 20 seconds after the prompt, exactly one read. No loops, no sleeps, no second read:

```bash
herdr agent read <slug> --source recent-unwrapped --lines 60
```

If the todo list is not on screen yet, report "checkpoint unconfirmed" in the Phase 1 report and yield anyway: the reply-back is the real signal.

When the todo list is on screen, it must name the brief's ticket ids. Invented modules, phases, or endpoints = confabulation:

```bash
herdr agent send-keys <slug> esc
```

Re-dispatch with a corrective preamble ("there is NO `<X>`, NO `<Y>` — writing those = drift").

## Failure handling

- **`blocked`** → `herdr agent read <slug>`, decide, answer with `herdr agent send-keys <slug> <key>` or a further prompt.
- **`agent_prompt_stalled`** → read the pane before retrying. Do not blindly re-prompt.
- **Read shows nothing even at higher `--lines`** → agent is on the alternate screen. Ask it to write its response as Markdown to a temp file and reply with the path. Fallback only.
- Read sources: `visible`, `recent`, `recent-unwrapped` (preferred), `detection`.
- Server errors: JSON on stderr, exit 1; CLI syntax errors exit 2.

## Worktree provisioning

Default: plain `git worktree add ../<repo>-<slug> -b <slug> production` (SKILL.md Step -1) plus the reuse-else-new-tab flow. This is the `acta:build` git fallback run by hand — no consent prompt.

`herdr worktree create` also creates a new **workspace** — a context switch. Probed 2026-08-06: it produced two workspaces and added no pane to the caller's workspace, so it cannot replace the default flow. Use only when the user explicitly asks for an isolated workspace. Clean-up: `herdr worktree remove --workspace <id>` and `herdr workspace close <id>`.
