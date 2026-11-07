---
parent: debt/2026-11-07-stray-worktrees-command
closes: [DBT-0157, DBT-0158]
id: SPC-0193
created: "2026-11-07 19:47:03"
hash: zytjaxb
---
# Worktree debt

## Problem

Three landed plans left eight debt items (DBT-0156.01 to .05, DBT-0157.01, DBT-0158.01 and .02). Four are about how `acta worktrees` and `acta doctor` judge worktrees, one is about `acta state`, and three are about skill text and evals for test output and temp worktrees.

## Goal

Close all eight items. Nothing new is added beyond what each item names.

## Design

### Plan A: Go code

1. DBT-0156.01. The doctor `worktrees` check splits strays by state. The message counts clean, missing and dirty strays apart. The fix line says `acta worktrees --clean` removes the clean and missing ones, and that dirty ones need the user. With only dirty strays, the fix line names no command.
2. DBT-0156.02. `acta setup` no longer runs the worktree check, so it never runs `git status` in any worktree. `acta doctor` still runs it in every idle linked worktree. That cost is accepted: doctor runs only when someone asks for it.
3. DBT-0156.03. When a worktree's admin folder exists but none of `HEAD`, `index` or `logs/HEAD` can be read, its age is unknown. It is protected with the reason "cannot read activity times", the same as a missing admin folder.
4. DBT-0156.04. A missing worktree whose branch ref no longer exists is checked like a detached one. Git still prints the `branch` line but gives `HEAD 0000...`, so the ref is checked with `git show-ref --verify refs/heads/<branch>`, and the commit comes from the new id on the last line of the worktree's `<admin>/logs/HEAD`. That commit must be held by a branch, tag or remote, or the worktree stays protected. When that file cannot be read, the reason is "cannot read its commits". A missing worktree whose branch ref exists keeps today's rule.
5. DBT-0156.05. Cause found: `planStates` (`internal/board/state.go`) counts a plan only when the branch's commits touched its file, so a plan whose ticks are not committed yet reads as "no running plan". Only `acta state <plan>` and `acta state set` change. When the plan they name is not found the usual way, they also look in the linked worktree the command runs in, and take the plan when `git status` there shows its file as modified or untracked. The main checkout never counts. The other callers of `planStates` (session hook, TUI, cleanup, the main-checkout guard) do not change, since a plan file git copied into a new worktree must not count as building.

### Plan B: skill text and evals

6. DBT-0157.01. When `git worktree remove` refuses the tidy temp gate worktree, the land stops and asks. The agent shows the untracked files (`git status --porcelain --ignored`, 10 lines at most) and gives two choices: (a) the user deletes the listed files, then the agent runs the plain remove again and the land goes on with step 4; (b) the temp worktree stays, the land goes on with step 4, and `acta worktrees` reports it later. The agent never forces and never deletes the files itself.
7. DBT-0158.01. `plugin/skills/build/implementer-prompt.md` tells the implementer to list untracked files a test or tool made in its report, and never to delete them. Under the subagent executor the task commit exists before the orchestrator reads that report, so the `.gitignore` line goes in a follow-up commit with the same `Task:` trailer, after the user's yes. Build step 3 says so.
8. DBT-0158.02. The `state-resume` eval never went red because its agent ran no tests and the scaffold made the cache folder up front. The scaffold's plan gets a real test command that writes the cache folder, the step the agent resumes says to run it, and the scaffold stops making the folder itself. The case is shown red with `plugin/skills/build/` taken from the commit before PLN-0207 landed, and green on today's text.

## Testing

- Plan A: Go tests at the layer that owns each rule: `internal/worktrees` for items 3 and 4, `internal/doctor` and the CLI wiring for items 1 and 2, `internal/board` and the CLI for item 5 (one case with an uncommitted plan file in the current worktree, one with a plan only in another worktree's uncommitted files, which must stay hidden).
- Plan B: the `land-tidy-temp-worktree` case gets a grader or a sibling case where the remove refuses and the agent stops, shows the files and does not force. A plugincheck test pins the implementer line. Item 8 is its own red then green eval.
- Plan B waits for nothing from Plan A. The two plans touch different files.
