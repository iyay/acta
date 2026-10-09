---
id: SCR-0048
hash: k06jtp5
title: Commit list and changes viewer per task
status: brainstorming
created: "2026-10-06 10:14:54"
schema: "1"
started: "2026-10-09 20:35:49"
---
# Commit list and changes viewer per task

## Words

### 2026-10-06

User idea, 2026-10-06, raw: "commit list dan changes viewer-nya di setiap task" (a commit list and a changes viewer for every task).

Likely in the TUI: from a task row, see the commits that task made and open their diff.

Not yet decided: how a commit is tied to a task (commit message, tick timestamps, or a recorded sha), whether the diff opens inside the TUI or in an outside pager or tool, and whether plans and debt items get the same view.

## Context

Facts found 2026-10-09:
- No link between a commit and a task exists today. Task commits have conventional subjects, empty bodies, no plan or task id, no trailers. The implementer prompt gives no message format.
- acta tick records no sha or branch; .agents.json keeps agent + time per task id.
- tidy at land keeps one commit per code task but rewrites every sha. Shas in .acta files are remapped (except in the last commit); shas outside the tree (.agents.json) go stale.
- All git access runs the git binary via exec. Nothing shows git show/diff yet. The editor runs through tea.ExecProcess and turns the mouse back on after.

## Log

### 2026-10-09

Round 1 answers (2026-10-09), user took every recommendation:
- Scope: plan tasks, plus one combined view per plan. Bugs and debt items later.
- Diff opens in a pane inside the TUI, plus one key to open it in the outside pager.
- Code commits by default; chore planning commits hidden behind a toggle.
- Commits of tasks still in a worktree show too.
- A small CLI command shares the same logic as the TUI.

### 2026-10-09

Round 2 answers (2026-10-09), user took every recommendation:
- Link: a "Task:" trailer in the commit body, written by the implementer; build skill, implementer prompt and dispatch brief ask for it. Survives tidy since kept messages carry over. Folded fix-round changes riding in another task's commit is accepted.
- Layout: a Commits section in the task detail pane; enter opens the diff full screen, esc goes back. No new tab.
- Diff: the TUI colors plain git show --stat -p output with theme colors; the pager key runs plain git show so the user's own pager setup applies.

### 2026-10-09

Round 3 answers (2026-10-09), user took every recommendation:
- Trailer holds the stable hash form, "Task: PLN-<hash>#<n>"; TUI and CLI still accept short ids and resolve them.
- Search main HEAD plus the branch of every worktree the board knows; not git log --all. One row per sha.
- Old plans with no trailers show "no linked commits"; no backfill, no history rewrite.
- CLI: acta commits <plan> [task], one line per commit (sha, date, subject, branch), --json for agents. No diff subcommand; git show does that.
- Keys: enter opens diff, o outside pager, c toggles chore commits, esc back; check for clashes when writing the spec.

### 2026-10-09

Approach (2026-10-09): user picked B, a lazygit-style Commits screen. The detail pane only lists the linked commits, read only. d on a task or plan row opens a full screen: commit list on the left, diff of the picked commit on the right, tab moves focus, o outside pager, c toggles chore commits, esc back. Rejected: A (cursor inside the detail pane, breaks its scroll) and C (popup picker, too small, no side-by-side).

### 2026-10-09

Section 1 approved (2026-10-09): trailer "Task: PLN-<hash>#<n>", one line per task, several allowed. Build hands the exact line to the implementer (implementer-prompt.md, dispatch.md). Fix-round and polish commits carry the trailers of the tasks they fix; chore planning commits carry none. acta tick warns, never fails, when HEAD has no trailer for the task. One eval case checks the implementer writes it.

### 2026-10-09

Section 2 approved (2026-10-09): new internal/commits. Find runs one git log --grep '^Task: ' over HEAD plus every board tree branch, parses trailers into a map "<plan hash>#<n>" to commits (sha, date, subject, branch, chore flag), one row per sha, main first. Chore = subject starts with "chore(". Diff runs git show --stat -p --no-color --no-ext-diff. CLI acta commits <plan> [task]: short id or hash, oldest first, line "sha date #task subject (branch)", chore hidden unless --all, --json, "no linked commits" exits 0. Tests use a temp git repo with trailer, no trailer, a worktree branch and one sha seen from two refs.

### 2026-10-09

Section 3 approved (2026-10-09): reloadCmd runs commits.Find next to the board load and carries the map in reloadMsg; diffs load async on pick, cached per sha, cache cleared on reload. Detail pane of a task or plan gets a COMMITS section (up to 5 rows "sha subject", then "+N more · d to open", or "no linked commits"; chore not counted). d on a task or plan row opens the Commits screen over every pane: list left (~35%), diff right, tab moves focus, o runs git show via ExecProcess and turns the mouse back on, c toggles chore, esc or q back with the cursor kept; under ~80 columns list and diff stack. Diff colors from theme ANSI slots: + green, - red, @@ cyan, diff --git and commit headers bold; long lines cut, not wrapped. Help and hint bar gain "d commits"; the screen has its own hint bar. Tests: model tests for d, tab, c, esc, narrow screen; a coloring test on fixed diff text.

## Open questions
