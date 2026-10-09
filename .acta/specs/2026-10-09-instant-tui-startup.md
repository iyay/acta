---
parent: bugs/2026-10-09-slow-startup-author-lookup
id: SPC-0119
created: "2026-10-09 20:56:13"
hash: q793jo1
---
# The TUI shows its first frame in under 100 ms

Status: Bounded, approved by the user in chat on 2026-10-09. Fixes BUG-0040.

Why: `acta` with no arguments takes about 1.1 s to show its first frame on this repo (554 commits, 479 planning files). About 780 ms of that is `fillAuthors` asking git who added each planning file: one full-history `git log` per folder, one after the other. The worktree and branch scan adds about 110 ms, and runs a second time for the watcher. Only the AUTHOR line in the TUI detail reads the authors, yet every command that loads the board pays for them. A throwaway build that read only the main tree before the first frame showed it in 67 to 79 ms.

Rulings:
- Target: the first frame shows in under 100 ms on this repo, and that time does not grow with more commits or more worktrees.
- The first frame shows only the items of the main checkout. Items from other worktrees and from unmerged branches come in with the background load a moment later. The cursor stays on its item, because a reload already keeps it by id.
- Authors leave the board load. Only the TUI asks for them, in the background, after the first frame. The AUTHOR line shows up a moment late.
- One `git log` for all planning files replaces the one per folder.
- The terminal color query stays as it is. It comes from Bubble Tea's own `init` and costs one round trip to the terminal, not the delay.

Design:
- `board.LoadTrees` stops calling `fillAuthors`. Filling authors becomes its own step, exported from `internal/board`, that a caller runs on a loaded board. `acta list`, `show`, `set`, `tick` and every other CLI command never run it, so they get about 0.8 s faster too. No CLI output shows the author today, so no output changes.
- The author step asks git once, with every path in one `git log --diff-filter=A`, run from the repo root. Each answer must still land on the right item, files in legacy folders included.
- `runTUI` (`internal/cli/cli.go`) builds the first board with `board.Load(cfg)`: main tree only, no `git log`, no worktree scan.
- The load function the TUI gets through `WithLoad` does the full load: `trees.Load`, then the author step. The startup load and every watcher reload use it.
- Watcher setup moves off the first frame into a goroutine. The full load after the first frame starts only once the watcher is in place, so a file that changes during startup is never missed. Startup runs the full load once, not twice.
- When the watcher cannot start, the TUI still gets the full load and still shows the watch error, as today.
- No version bump.

Out of scope:
- `trees.Others` still runs in both the full load and the watcher's folder list. Both runs are off the first frame now.
- Keeping authors in a cache on disk.
- The color query in Bubble Tea's `init`.

Testing:
- `board.Load` and `trees.Load` run no author lookup. Count the calls through the `gitAuthors` variable.
- The TUI's full load fills authors, and asks git once per load, not once per folder.
- An author still lands on the right item when files sit in several folders, a legacy folder included.
- The full load after the first frame starts only after the watcher has its folders.
- By hand before land: the pty probe from `.acta/wiki/tui-pty-probe.md`, with the binary as the foreground process, five runs, each first frame under 100 ms. The numbers go into the land report.
