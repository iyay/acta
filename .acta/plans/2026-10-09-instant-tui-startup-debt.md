---
parent: debt/2026-10-09-instant-tui-startup
depth: minimal
closes: [DBT-0104.01, DBT-0104.02, DBT-0104.03]
id: PLN-0129
created: "2026-10-09 21:43:04"
hash: fm3h4ux
---
# Instant TUI startup debt

**Goal:** Close the three review gaps of PLN-0128: a test for the TUI wiring, a fixed-size author command line, and a test for a watcher that cannot start.

**Spec:** `.acta/specs/2026-10-09-instant-tui-startup-debt.md`

**Tests:** fast `scripts/test ./internal/cli ./internal/gitc ./internal/tui`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- No output a user sees changes. No version bump.
- Tests only through `scripts/test`; the pre-tool hook blocks bare `go test`.
- Comments are plain English a 10-year-old reads back: short words, one idea each, and they say why.

## Waves

- Wave 1: Task 1, Task 2, Task 3

### Task 1: runTUI gets its boards from a tested helper

**Files:**
- Modify: `internal/cli/cli.go` (`runTUI`, new `tuiBoards`)
- Test: `internal/cli/tui_boards_test.go` (new)

**verify:** The first frame's board never holds an item from another worktree or an author, and every later load holds both. List each load `runTUI` makes and where its board comes from.

- [ ] **Failing test:** `TestTUIBoardsFirstFrameIsMainOnly` builds a git repo (user.name `Ana`) with a committed plan and a linked worktree holding a plan the main tree lacks, calls `tuiBoards(cfg)`, and wants the first board without the worktree plan and with an empty author, and the loader's board with the worktree plan and author `Ana` on the committed plan; it fails because `tuiBoards` does not exist.
- [ ] **Code:** add `tuiBoards(cfg config.Config) (*board.Board, func() (*board.Board, error), error)` returning `board.Load(cfg)` and a func calling `trees.LoadWithAuthors(cfg)`; `runTUI` uses its board and passes the func to `WithLoad`.
- [ ] **Commit:** `test(cli): cover the TUI first frame and later loads`

### Task 2: gitc.Authors names folders, not files

**Files:**
- Modify: `internal/gitc/gitc.go` (`Authors`)
- Test: `internal/gitc/gitc_test.go`

**verify:** The `git log` command line holds each asked folder once and no file name, whatever the number of files, and every asked path gets the same author as before while no path that was not asked comes back. List the cases checked.

- [ ] **Failing test:** `TestAuthorsAsksByFolder` commits files in two folders as different names, plus one more committed file in one folder that is not asked, then calls `Authors` with the asked paths and wants exactly those paths with their authors; it also checks the args from a new `authorArgs(repo, paths)` hold each folder once and no file name; it fails because `authorArgs` does not exist and `Authors` names files.
- [ ] **Code:** add `authorArgs` that builds the log args with the distinct folders of `paths`, relative to `repo`, as pathspecs; `Authors` uses it and keeps only entries whose joined path was asked.
- [ ] **Commit:** `fix(gitc): ask git for authors by folder, so the command line stays short`

### Task 3: A watcher that cannot start is tested

**Files:**
- Modify: `internal/tui/watch.go` (package variable for `fsnotify.NewWatcher`)
- Test: `internal/tui/watch_test.go`

**verify:** When the watcher cannot start, the TUI always hears why, always gets one full load after that, and `stop` always returns. List what the TUI receives in order.

- [ ] **Failing test:** `TestStartStillLoadsWhenTheWatcherFails` swaps the new `newWatcher` variable for one that returns an error (restored in `t.Cleanup`, test not parallel), calls `Start`, and wants a `watchFailedMsg` then one `reloadMsg` within 3 s, and `stop()` to return within 1 s; it fails because `newWatcher` does not exist.
- [ ] **Code:** add `var newWatcher = fsnotify.NewWatcher` with a one-line why comment and call it in `Watch`.
- [ ] **Commit:** `test(tui): cover a watcher that cannot start`
