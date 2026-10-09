---
parent: scratch/2026-10-06-task-commits-viewer
id: SPC-0118
created: "2026-10-09 20:43:50"
hash: tv4q2yu
---
# Commit list and diff viewer per task

Status: Architectural, approved by the user in chat on 2026-10-09, section by section. Full answers are in the Log of SCR-0048.

## Why

From a task row in the TUI there is no way to see which commits that task made or what they changed. Today nothing ties a commit to a task at all: task commits carry a conventional subject, an empty body and no plan or task id, and `acta tick` records no sha. This spec adds the link and a lazygit-style screen to read it.

## Rulings

- Scope: plan tasks, plus one combined view per plan. Bugs and debt items come later.
- The link is a commit trailer written by the implementer. Not a sha stored by tick (tidy rewrites every sha at land), not a guess from timestamps.
- The trailer names the plan by hash, not short id: `acta id --fix-duplicates` can renumber short ids after land; hashes never change.
- Search HEAD plus the branch of every tree the board already reads, never `git log --all` (that pulls in backup branches, stashes and `refs/acta/tidy/*`, so a commit shows twice).
- Old plans with no trailers show `no linked commits`. No backfill, no history rewrite.
- Code commits show by default; `chore(` commits hide behind a toggle.
- Commits of work still in a worktree or an unmerged branch show too.
- The diff opens inside the TUI, with one key to open it in the outside pager.

## 1. Trailer and the build flow

- Form: `Task: PLN-<hash>#<n>`, for example `Task: PLN-uyo6akt#3`. One line per task in the commit body; a commit may carry several.
- acta:build puts the exact line in every hand-off: `plugin/skills/build/implementer-prompt.md` for subagents, `plugin/skills/build/dispatch.md` for omp. The implementer copies it; it never looks up the hash.
- Fix-round and polish commits carry the trailer of each task they fix. Planning `chore(` commits carry none.
- `acta tick <task>` checks whether HEAD carries a trailer for that task. When it does not, tick still ticks and prints a warning to stderr naming the task and the line it expected. It never fails on this.
- One eval case checks that an implementer run writes the trailer.

## 2. `internal/commits` and the CLI

- New package `internal/commits`.
- `Find(cfg, refs)` runs one `git log --grep='^Task: '` over the given refs: HEAD of the repo plus the branch of every `board.Tree` (other worktrees and unmerged local branches, the same sources the board reads). It parses the trailers from each body into a map from `"<plan hash>#<n>"` to commits. Each commit holds sha, date, subject, branch and a chore flag.
- A sha seen from several refs appears once. Its branch is the first ref that holds it, with the main checkout's HEAD first.
- Chore means the subject starts with `chore(`.
- `Diff(repo, sha)` runs `git show --stat -p --no-color --no-ext-diff <sha>` and returns plain text.
- CLI `acta commits <plan> [task]`:
  - `<plan>` takes a short id or a hash, resolved through the board.
  - With no task: every commit of the plan, each line labeled with its task.
  - One line per commit, oldest first: `sha  date  #task  subject  (branch)`, with a 7-char sha and the author date as YYYY-MM-DD.
  - Chore commits hidden unless `--all`. `--json` prints the same rows as JSON.
  - No linked commits: print `no linked commits`, exit 0. Unknown plan or task: error, exit 1.
- Tests build a temp git repo with: a commit with a trailer, one with none, one with two trailers, a worktree branch, and one sha reachable from two refs.

## 3. TUI

- Load: `reloadCmd` calls `commits.Find` on the same goroutine as the board load, and `reloadMsg` carries the map, so board and commits always come from one load. Diffs load async when a commit is picked, cached per sha; a reload clears the cache.
- Detail pane of a task or plan: a new `COMMITS` section below the steps. Up to 5 rows of `sha  subject`, then `+N more · d to open`. Empty: `no linked commits`. Chore commits are not counted.
- `d` on a task or plan row, with focus on a list or on the detail pane, opens the Commits screen for that item: the task's commits, or the whole plan's with a `#n` label per row. `d` on any other row does nothing.
- The Commits screen replaces every pane:
  - Left: commit list, about 35% of the width. j/k pick a commit; the diff follows the pick.
  - Right: the diff. `tab` moves focus between the two. Diff scroll: j/k, ctrl+d, ctrl+u, g, G and the mouse wheel.
  - `o`: run `git show <sha>` through `tea.ExecProcess`, then turn the mouse back on, the same as the editor flow (wiki: exec-process-drops-mouse). `o` keeps its sort meaning on the board; on this screen it opens the pager.
  - `c`: show or hide chore commits.
  - `esc` or `q`: back to the board, cursor where it was.
  - Under about 80 columns, list and diff stack top and bottom.
- Diff colors come from the theme's ANSI slots: `+` lines green, `-` lines red, `@@` lines cyan, `diff --git` and `commit` headers bold, the stat in plain FG. Long lines are cut, not wrapped, so diff columns stay straight.
- Help (`?`) and the hint bar gain `d commits`. The Commits screen has its own hint bar.
- Tests: model tests in the style of the existing ones for `d`, `tab`, `c`, `esc` and a narrow screen; a coloring test over fixed diff text.

## Known limits

- tidy folds fix-round commits into the kept commit before them, so a fix for task 2 can show inside a task 3 commit. Accepted.
- Commits made before this lands have no trailer and stay unlinked.

## Wiki

Approved by the user on 2026-10-09; acta:build writes it: a `Convention` page `task-trailer.md` saying every code commit carries `Task: PLN-<hash>#<n>`, why the hash and not the short id, and that tick only warns.
