---
parent: scratch/2026-10-04-live-work-state
id: SPC-0072
created: "2026-10-05 08:14:43"
hash: pu9pr01
---
# Live work state for running plans

Status: design approved by the user in chat on 2026-10-05, section by section. Architectural: it adds a plan section, a CLI command, a hook path and skill steps. Answers and rulings live in the SCR-0041 Log.

Why: when context runs out or a session dies in the middle of a plan, the next session (in any harness) has to be told by the user where the work stands. The plan ticks show which tasks are done, but not what the agent learned, what it was about to do, or what it was waiting to ask. This design keeps that in the plan file and shows it to every new session.

Out of scope: hand-off briefs for dispatch beyond one line, agents running side by side seeing each other, and work that has no plan (debug, shape).

## 1. The State section

- A plan file can hold a `## State` section with three subsections, in this order: `### Next`, `### Findings`, `### Open rulings`.
- The agent writes them with `acta state set <plan> next|findings|rulings`. The body comes from stdin. The command replaces that one subsection and leaves the rest of the file as it is. It creates `## State` or the subsection when missing. Then it commits with the message `acta: state <plan id>`.
- An empty body clears the subsection.
- Each subsection holds at most 10 lines. The command refuses a longer body with an error that names the limit.
- The agent writes in the build worktree, so the section merges with the branch at land.
- Facts that git and the plan already hold are never written: the current task, the last commit, the worktree path and the review round. acta works them out each time it reads (section 2), so they cannot go stale.
- After land the section stays as the last record.

## 2. Reading the state

- A running plan is one that has `started`, has no `finished`, and has a worktree that `internal/trees` finds.
- `acta state <plan>` prints the full view:
  - the current task: the first task with no tick;
  - the last commit on the plan's branch (short hash and subject);
  - the worktree path;
  - the last review round, read from the newest `acta: tick fix round N` commit, or none;
  - then the three subsections as written.
- `acta state` with no plan prints one line per running plan.
- The SessionStart hook adds a summary of the running plans in the session's repo and its worktrees only:
  - at most 3 lines per plan: id and title with the current task; worktree and last commit; the first line of Next;
  - at most 3 plans; the rest show as one line `+N more, run acta state`;
  - one rule line: before you stop, or when context runs low, run `acta state set <plan> next`.
- When the hook fails or `acta` is missing, it adds nothing and never blocks the session.

## 3. Skills, omp and tests

- `acta:build`: after each task commit, update Next. A finding that cost time goes in Findings. A question that waits for the user goes in Open rulings.
- `acta:land`: before the merge, read Findings. Move each lasting one to a wiki page or a debt item.
- omp: the extension calls the same `acta hook` session start path, so it gets the same summary. The dispatch brief gets one line: read `acta state <plan>` first.
- Go tests: `acta state set` (replace one subsection, create when missing, empty body clears, 10 line cap, commit message); the view (current task, last commit, review round, no round); the hook (3 plans by 3 lines cap, `+N more` line, silent on failure, other repos left out).
- One eval: a fresh session on a half done plan with Next written goes on from Next without asking the user where things stand.
- The last task adds 1 to the patch version in the three plugin files.
