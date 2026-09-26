# Dispatch house rules

These apply to any agent started from a pm:dispatch brief. Read this before writing the todo list. The brief carries the job facts (plan, tickets, worktree, gates); this file carries the rules that hold for every dispatch.

SKILL: pm:build — one implementer subagent per ticket. You are the main agent in this pane: you write ZERO code yourself. Review happens on the orchestrator's side, never yours. pm:tdd at every seam — failing test first, no code before red. Red test or error you cannot explain → pm:debug before any edit. Before you claim any ticket done → pm:land: run the gate, show output.
PROPERTIES, NOT INSTANCES: each verify line is a claim that must hold on EVERY path, not the one case that was reported. Before you fix anything, enumerate every path, caller and surface that could break the claim, and report that list with your commit. A fix that closes only the reported case is not done. If your change widens a boundary, check the neighbours that share it.
BUGS YOU CONFIRM OUTSIDE YOUR TICKETS: do not fix them (surgical). Record each one with pmb bug new (see pm:bug); in this worktree the bug file commits on your branch and lands with it. Only confirmed bugs with a repro; no guesses.
PARALLEL — spawn as many subagents as the tickets allow:
  - One implementer subagent per ticket, MINIMUM. You never implement a ticket in your own context.
  - Group tickets into waves by FILE OWNERSHIP. Disjoint files + no dependency = same wave, dispatched TOGETHER in ONE message, not one after another.
  - Two subagents never hold the same file in one wave. Same file or real dependency (B imports what A creates) → next wave.
  - Split further where the ticket allows: one subagent per file, read-only probes in parallel beside the wave (pm:build).
  - Declare the waves in your todo list BEFORE dispatching: "wave 1: T-1, T-3, T-4 parallel · wave 2: T-2 (needs T-1)".
  - Serial needs a stated reason. Six independent tickets run serially = you failed this brief.
  - Every subagent still does full pm:tdd + mutation-verify on its own slice and commits its own ticket. Parallel is not a licence to batch commits.
PONYTAIL: YAGNI → stdlib → native → dep → one line → minimum; never cut validation/security/accessibility. Deferrals in chat or ticket, no markers in files.
COMMENTS: plain English a 10-year-old reads back. WHY not what. No marker tags, no Latin, no emoji. Real names verbatim.
MUTATION-VERIFY every test: revert the fix, re-run, confirm FAILS, git checkout --, confirm green. Assert the effect, never just a status or call count.
COMMIT per ticket, todo marked done the moment it commits.
PROGRESS: right after each step of a ticket, run pmb tick plans/<stem>#task-N --step <n> from the worktree with the ticket id written out in full (for example pmb tick plans/2026-09-26-tick-fixes#task-3 --step 2), so the board shows live progress. Never commit the plan file yourself; the orchestrator commits it once after each wave.
NO repo-wide formatter, NO npm install, NO git push, NO database migration.
DO NOT REVIEW YOUR OWN WORK — no pm:review, no findings, no verdict. Commit the last ticket, fire REPLY-BACK.
STAY WHERE YOU ARE — you have your own tab. Do NOT move, park, close, or create any pane, tab, or workspace.
MEMORY: the Claude memory paths named in the brief are read-only. Never write there — your own omp memory keeps what you learn.
