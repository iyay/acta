---
parent: bugs/2026-10-09-slow-startup-author-lookup
depth: minimal
id: PLN-0128
created: "2026-10-09 21:02:33"
hash: lwl7n0v
started: "2026-10-09 21:06:08"
---
# Instant TUI startup Implementation Plan

**Goal:** The TUI shows its first frame in under 100 ms, and worktrees, branches and authors arrive after it in one background load.

**Spec:** `.acta/specs/2026-10-09-instant-tui-startup.md`

**Tests:** fast `scripts/test ./internal/board ./internal/tui ./internal/trees ./internal/cli`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Before the first frame, nothing may run git except the `git rev-parse` inside `config.Load`.
- No CLI output changes. No version bump.
- Run tests only through `scripts/test`; the pre-tool hook blocks bare `go test`.
- Comments are plain English a 10-year-old reads back: short words, one idea each, and they say why.
- Before land, the orchestrator times the branch binary with the foreground pty probe from the spec's Testing section. Implementers do not.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3

### Task 1: Authors leave the board load and are asked once per checkout

**Files:**
- Modify: `internal/board/board.go` (`Board`, `LoadTrees`)
- Modify: `internal/board/closed.go` (`fillAuthors` becomes `FillAuthors`)
- Test: `internal/board/board_test.go`
- Test: `internal/tui/detail_test.go` (`detailLines` calls `b.FillAuthors()` after `board.Load`, because `TestDetailShowsTheAuthorUnderTheStatus` reads the author)

**verify:** No board load (`Load`, `LoadTrees`) ever asks git for authors or `user.name`, and `FillAuthors` gives every on-disk item the author the old per-folder lookup gave, with one git question per checkout. List each kind of item checked: main tree, legacy folder, another worktree on disk, a worktree nested inside the main folder, a branch read only from git, a file never committed, a board outside git.

- [x] **Failing test:** add `TestLoadNeverAsksForAuthors` (stub `gitAuthors` and `gitUserName` to count, load with `boardWith`, want 0 calls of each); turn `TestAuthorIsAskedOncePerFolder` into `TestAuthorIsAskedOncePerCheckout` (its four files in three folders, plus one spec under the legacy `docs/superpowers/specs/` folder, give one call that names each file once, `user.name` read once; then a second checkout passed to `LoadTrees` as a `Tree` whose folder sits inside the first, with a file the main tree lacks, gets its own call holding only its own path); make the other author tests call `b.FillAuthors()` after loading; run `scripts/test ./internal/board ./internal/tui -run 'Author'` and watch it fail, because `Load` still asks git once per folder and `FillAuthors` does not exist.
- [x] **Code:** `LoadTrees` stops calling `fillAuthors` and keeps on the `Board` the root of each checkout it read from disk (main `RepoRoot` first, then each other tree whose `Files` is nil); `fillAuthors(root)` becomes `FillAuthors()`, which groups the on-disk paths by the checkout root they sit under (the longest matching root wins), calls `gitAuthors(root, paths)` once per group, and reads `gitUserName` from the main checkout at most once.
- [x] **Commit:** `perf(board): fill authors only on request, once per checkout`

### Task 2: The watcher starts off the first frame and sends the first full load

**Files:**
- Modify: `internal/tui/watch.go` (new `Start`)
- Modify: `internal/tui/model.go` (`Init`)
- Test: `internal/tui/watch_test.go`

**verify:** At startup the full load runs exactly once and never before the watcher has its folders, and `Init` starts no load. List each path (watcher starts, watcher fails, the user quits before setup ends) and what the TUI receives on it.

- [x] **Failing test:** add `TestStartLoadsOnceAfterTheWatcherIsReady` (`dirs` marks under a mutex that it ran, `reload` notes whether `dirs` ran first and counts calls, `send` feeds a channel; want one `reloadMsg` within 3 s, a count of 1, and `dirs` before `reload`) and `TestInitStartsNoLoad` (a model whose `load` counts calls, run `m.Init()` with `runNow`, want 0 loads); run `scripts/test ./internal/tui -run 'TestStart|TestInitStartsNoLoad'` and watch it fail, because `Start` does not exist and `Init` still loads.
- [x] **Code:** add `Start(dirs func() []string, reload func() tea.Msg, send func(tea.Msg)) (stop func())` to `watch.go`: a goroutine calls `Watch`; on error it sends `watchFailedMsg` and keeps a stop that does nothing; it puts the stop into a channel of size 1, then sends `reload()`; the returned stop takes from that channel and calls it. `Init` drops `m.reloadCmd()`, keeps the clock and toast timers, and its comment says the watcher's start sends the first full load.
- [x] **Commit:** `perf(tui): start the watcher off the first frame`

### Task 3: The first frame reads only the main tree

**Files:**
- Modify: `internal/trees/trees.go` (new `LoadWithAuthors`)
- Modify: `internal/cli/cli.go` (`runTUI`)
- Test: `internal/trees/trees_test.go`

**verify:** Before `p.Run` the TUI runs no `git log`, no worktree scan and no watcher setup, and every load after the first frame is the full one with authors. List every load the TUI makes (first frame, startup load, watcher reload, the `r` key, the reload after a write action) and the function each one calls.

- [ ] **Failing test:** add `TestLoadWithAuthorsFillsAuthors`: from `setup`, commit a new plan in the worktree as another name (`git -c user.name=Wen commit`); want `Load(cfg)` to leave `Author` empty on both the main plan and the worktree plan, and `LoadWithAuthors(cfg)` to give `test` and `Wen`; run `scripts/test ./internal/trees -run TestLoadWithAuthorsFillsAuthors` and watch it fail, because `LoadWithAuthors` does not exist.
- [ ] **Code:** `LoadWithAuthors(cfg)` calls `Load(cfg)` and, when that worked, `b.FillAuthors()`; in `runTUI` the first board comes from `board.Load(cfg)`, `WithLoad` gets `func() (*board.Board, error) { return trees.LoadWithAuthors(cfg) }`, and the `tui.Watch` block becomes `stop := tui.Start(dirs, load, p.Send)` with `defer stop()`.
- [ ] **Commit:** `perf(cli): show the first frame before any git scan`
