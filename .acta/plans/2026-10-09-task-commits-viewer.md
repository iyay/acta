---
parent: specs/2026-10-09-task-commits-viewer-design
depth: minimal
closes: [SCR-0048]
id: PLN-0127
created: "2026-10-09 20:47:04"
hash: broaksz
started: "2026-10-09 20:51:08"
---
# Task commits viewer

**Goal:** Every code commit names its task in a `Task:` trailer, `acta commits` lists a plan's or task's commits, and the TUI shows them in the detail pane and in a lazygit-style Commits screen with a colored diff.

**Spec:** `.acta/specs/2026-10-09-task-commits-viewer-design.md`

**Tests:** fast `scripts/test ./internal/commits ./internal/cli ./internal/tui ./internal/plugincheck`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Trailer form, verbatim: `Task: PLN-<hash>#<n>`, for example `Task: PLN-uyo6akt#3`. Plan by hash, never short id.
- git runs through the git binary with `exec`, the way `internal/gitc` does. No go-git, no new dependency.
- Refs searched: HEAD of the repo plus the branch of every `board.Tree`. Never `git log --all`.
- Chore commit = subject starts with `chore(`.
- No version bump.
- Tests build git repos under `t.TempDir()` and never run `rm -rf` or `rm -f`.

## Waves

- Wave 1: Task 01, Task 02, Task 03
- Wave 2: Task 04, Task 05, Task 06
- Wave 3: Task 07

### Task 01: commits package

**Files:**
- Create: `internal/commits/commits.go`
- Test: `internal/commits/commits_test.go`

**verify:** Every commit reachable from the given refs that carries a `Task:` trailer shows once under each task it names, and nothing else shows. List every case checked: no trailer, two trailers, a trailer-shaped line in the middle of the body, one sha seen from two refs (branch is the first ref, HEAD first), a chore subject, a bad trailer value.

- [x] **Failing test:** in a temp repo with a worktree branch, `Find(repo, []Ref{{Name:"HEAD"},{Name:"feat"}})` returns `map[string][]Commit` keyed `"<plan hash>#<n>"` (Commit holds Sha, Date, Subject, Branch, Chore), and `Diff(repo, sha)` returns `git show --stat -p --no-color --no-ext-diff` text; fails because the package does not exist.
- [x] **Change:** one `git log --grep='^Task: ' --format=<sha, author date, subject, branch, body>` per ref, parse trailers with an exported `Trailers(body string) []string`, dedupe by sha in ref order; `Diff` runs `git show`.
- [x] **Commit:** `feat(commits): find task commits by trailer`

### Task 02: Build flow asks for the trailer

**Files:**
- Modify: `plugin/skills/build/SKILL.md`, `plugin/skills/build/implementer-prompt.md`, `plugin/skills/build/dispatch.md`
- Modify: `internal/plugincheck/skill_build_test.go`
- Create: `plugin/evals/task-trailer/` (`case.yaml`, `prompt.md`, `scaffold.sh`, `graders/`)

**verify:** Every path that makes a code commit during a build (subagent implementer, omp dispatch, inline, fix round, polish) is told the exact `Task: PLN-<hash>#<n>` line, and planning `chore(` commits are told to carry none. List each path and the text that covers it.

- [x] **Failing test:** add `"Task: PLN-"` to the build skill's `Must` list in `skill_build_test.go`; fails because no build file says it yet.
- [x] **Change:** build hands the exact trailer line (plan hash from `acta show <plan>`) in each hand-off, fix-round and polish commits carry the trailers of the tasks they fix; add an eval case where a one-task minimal plan is built inline and a grader checks the code commit body has `Task: PLN-<hash>#1`; keep within the plugincheck caps (wiki: plugincheck-caps-per-folder) and embed any new plugin file the binary needs (wiki: plugin-ships-in-binary).
- [x] **Commit:** `feat(build): ask implementers for the Task trailer`

### Task 03: Diff coloring

**Files:**
- Create: `internal/tui/diffcolor.go`
- Test: `internal/tui/diffcolor_test.go`

**verify:** Every line of `git show --stat -p` output gets exactly one style and is never wider than the given width. List each line kind checked: `commit` header, `diff --git` header, `---`/`+++` file lines, `@@` hunk, `+`, `-`, context, stat, empty, a line wider than the width, a line with wide runes.

- [x] **Failing test:** `colorDiff(text string, w int, pal palette) []string` over fixed diff text returns `+` lines green, `-` lines red, `@@` cyan, `diff --git` and `commit` bold, the rest plain FG, each cut to `w` cells; fails because the function does not exist.
- [x] **Change:** style by line prefix from the theme ANSI slots the TUI already uses, cut with the existing width helpers (wiki: xansi-wrap-overflows).
- [x] **Commit:** `feat(tui): color diff lines from the theme`

### Task 04: acta commits command

**Files:**
- Create: `internal/cli/commits.go`
- Modify: `internal/cli/cli.go` (dispatch and usage)
- Test: `internal/cli/commits_test.go`

**verify:** For any plan given by short id or hash, with or without a task, the command prints exactly the linked non-chore commits oldest first (all with `--all`), and every bad input exits 1. List every input checked: short id, hash, plan with no task, task, `--all`, `--json`, no linked commits (prints `no linked commits`, exit 0), unknown plan, task number not in the plan.

- [x] **Failing test:** `acta commits <plan> [task]` in a temp repo prints `sha  date  #task  subject  (branch)` lines, 7-char sha and `YYYY-MM-DD`; fails because the command does not exist.
- [x] **Change:** resolve the plan through the board, build refs from HEAD plus `trees.Others` branches, call `commits.Find`, filter, sort by date, print text or JSON.
- [x] **Commit:** `feat(cli): acta commits lists a task's commits`

### Task 05: tick warns on a missing trailer

**Files:**
- Modify: `internal/cli/tick.go`
- Test: `internal/cli/tick_test.go`

**verify:** tick's exit code and file writes are the same with or without a trailer; the warning shows only when HEAD lacks the trailer for the ticked task. List every case checked: trailer present, missing, trailer for another task, HEAD not readable or not a git repo (no warning, no failure).

- [x] **Failing test:** tick a task whose HEAD commit has no `Task:` trailer; expect stderr `warning: HEAD has no "Task: PLN-<hash>#<n>" trailer` and the usual exit; fails because tick says nothing.
- [x] **Change:** after a successful tick, read HEAD's body with `git log -1 --format=%B`, check `commits.Trailers`, print the warning to stderr when missing.
- [x] **Commit:** `feat(tick): warn when HEAD lacks the task trailer`

### Task 06: TUI loads commits and shows them in the detail pane

**Files:**
- Modify: `internal/tui/model.go` (`reloadMsg`, `reloadCmd`), `internal/tui/detail.go`
- Test: `internal/tui/commits_detail_test.go`

**verify:** The detail pane of every task and plan shows the commits from the same load as the board, chore ones never counted, and no other item kind shows a COMMITS section. List every case checked: task with commits, task with more than 5, plan, task with only chore commits, no commits, a bug or debt row, a load where git fails (board still shows).

- [x] **Failing test:** a model fed a `reloadMsg` with a commits map renders `COMMITS`, up to 5 `sha  subject` rows, then `+N more · d to open`, or `no linked commits`; fails because the section does not exist.
- [x] **Change:** a `var findCommits = commits.Find` seam next to `rounds`, called in `reloadCmd` with HEAD plus board tree branches, carried in `reloadMsg` and stored on the model; section drawn after `stepLines`.
- [x] **Commit:** `feat(tui): show a task's commits in the detail pane`

### Task 07: Commits screen

**Files:**
- Create: `internal/tui/commits.go`
- Modify: `internal/tui/model.go` (`d` key, key and mouse routing), `internal/tui/view.go` (`helpLines`, screen draw), `internal/tui/hints.go`
- Test: `internal/tui/commits_test.go`

**verify:** While the Commits screen is open every key and wheel event goes to it, and esc or q always returns to the board with the same cursor. List every key checked: d on task, d on plan, d on other rows, j/k, tab, ctrl+d/ctrl+u, g/G, wheel, c, o (ExecProcess then mouse back on, wiki: exec-process-drops-mouse), esc, q; and widths under and over 80 columns.

- [ ] **Failing test:** press `d` on a task row; the view shows the commit list left (about 35%) and the colored diff right, `tab` moves focus, `c` shows chore commits, `esc` restores the board and cursor, under 80 columns list and diff stack; fails because `d` does nothing.
- [ ] **Change:** screen state on the model; diff loads async per picked sha through a `var showDiff = commits.Diff` seam, cached per sha and cleared on reload, drawn with `colorDiff`; `o` runs `git show <sha>` via `tea.ExecProcess`; `d commits` added to help and hint bar, own hint bar on the screen.
- [ ] **Commit:** `feat(tui): Commits screen with list and diff`
