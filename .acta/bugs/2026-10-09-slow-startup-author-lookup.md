---
id: BUG-0040
hash: nywhcia
---
# acta takes over a second to open, and every board load is slow

## Symptom
Running `acta` with no arguments shows the first frame after about 1.1 s on this repo (554 commits, 479 planning files). `acta list` takes about 0.93 s. Every command that loads the board pays the same cost, and so does every TUI reload after a file change. The delay grows with every commit.

## Root cause
1. `internal/board/board.go:215` calls `fillAuthors` on every load. `internal/board/closed.go:35` asks git once per planning folder, one after the other. Each ask is `gitc.Authors` (`internal/gitc/gitc.go:396`), a `git log --diff-filter=A --name-only` over the whole history. Five folders cost about 780 ms of the 823 ms board load. Only the AUTHOR line in the TUI detail (`internal/tui/detail.go:120`) reads the result.
2. `internal/cli/cli.go:38` (`trees.Load`, `internal/trees/trees.go:125`) and `internal/cli/cli.go:54` (`trees.WatchDirs`, `internal/trees/trees.go:133`) both call `trees.Others`, so the worktree and branch scan runs twice before the first frame, about 110 ms each.
3. The watcher is set up before `p.Run`, so its scan also sits in front of the first frame.

## Repro
Measured with a pty probe that times from spawn to the first frame, on main at e99c596:
- `acta` as built: about 1.10 s to the first frame.
- Same binary with `gitAuthors` stubbed to return nothing: about 0.31 s.
- `board.Load` on this repo: 823 ms with authors, 42 ms without.
- `git log --no-renames --diff-filter=A --name-only --format=%x00%an -- <files>` per folder: specs 204 ms, plans 278 ms, bugs 138 ms, debt 175 ms, scratch 141 ms.

## Found in
main, by acta:debug while chasing a slow TUI startup.
