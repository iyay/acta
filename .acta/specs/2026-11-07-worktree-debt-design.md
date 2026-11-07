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
2. DBT-0156.02. `acta setup` no longer runs the worktree check, so it never runs `git status` in any worktree. `acta doctor` still runs it.
3. DBT-0156.03. When a worktree's admin folder exists but none of `HEAD`, `index` or `logs/HEAD` can be read, its age is unknown. It is protected with the reason "cannot read activity times", the same as a missing admin folder.
4. DBT-0156.04. A missing worktree whose branch ref no longer exists is checked like a detached one: its HEAD commit must be held by a branch, tag or remote, or it stays protected. A missing worktree whose branch ref exists keeps today's rule.
5. DBT-0156.05. `acta state <plan>` and `acta state set` find a plan by name in the worktree the command runs in, even before the branch has committed the plan file. Cause found: `planStates` counts a plan only when the branch's commits touched its file, so a plan whose ticks are not committed yet reads as "no running plan". The fix goes where every caller passes, and a plan in another worktree still needs a branch commit to count.

### Plan B: skill text and evals

6. DBT-0157.01. When `git worktree remove` refuses the tidy temp gate worktree, the land still stops and asks. The agent shows the untracked files (`git status --porcelain --ignored`, 10 lines at most) and the exact command the user can run. It never forces and never deletes the files itself.
7. DBT-0158.01. `plugin/skills/build/implementer-prompt.md` tells the implementer to list untracked files a test or tool made in its report, and never to delete them, so the orchestrator can make the `.gitignore` offer that build step 3 describes.
8. DBT-0158.02. The `state-resume` eval scaffold makes its test command leave the cache folder, so a run that cleans it is caught. The case is shown red on the build text from before PLN-0207 and green on today's text.

## Testing

- Plan A: Go tests at the layer that owns each rule: `internal/worktrees` for items 3 and 4, `internal/doctor` and the CLI wiring for items 1 and 2, `internal/board` and the CLI for item 5 (one case with an uncommitted plan file in the current worktree, one with a plan only in another worktree's uncommitted files, which must stay hidden).
- Plan B: the `land-tidy-temp-worktree` case gets a grader or a sibling case where the remove refuses and the agent stops, shows the files and does not force. A plugincheck test pins the implementer line. Item 8 is its own red then green eval.
- Plan B waits for nothing from Plan A. The two plans touch different files.
