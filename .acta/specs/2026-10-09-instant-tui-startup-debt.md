---
parent: debt/2026-10-09-instant-tui-startup
id: SPC-0120
created: "2026-10-09 21:40:11"
hash: mjroctr
---
# Close the review debt of instant TUI startup

Status: Bounded, approved by the user in chat on 2026-10-09. Works DBT-0104 (all three items).

Why: the review of PLN-0128 left three gaps. The TUI wiring in `runTUI` has no test, so a change back to the slow first frame or to a loader without authors stays green. `gitc.Authors` names every file of a checkout on one command line, which grows with every file and nears the 32 KB Windows cap. The path where the watcher cannot start inside `tui.Start` has no test.

Design:
- `internal/cli/cli.go`: the two loads move out of `runTUI` into `tuiBoards(cfg config.Config) (*board.Board, func() (*board.Board, error), error)`. The board is the first frame's, from `board.Load`. The func is the loader every later load uses, `trees.LoadWithAuthors`. `runTUI` calls it and behaves as today.
- A test in `internal/cli` builds a real git repo with a committed plan and a linked worktree holding a plan the main tree lacks. The first board must hold no worktree item and no author. The loader's board must hold the worktree item and the committed plan's author.
- `internal/gitc/gitc.go` `Authors`: the `git log` pathspecs are the distinct folders of the asked paths, relative to `repo`, not the paths themselves. The answer keeps only the paths that were asked. So the command line stays the same size however many files a folder holds. Every asked path still gets the same author as before.
- `internal/tui/watch.go`: the call to `fsnotify.NewWatcher` goes through a package variable, so a test can make it fail. A test checks that a failed start sends `watchFailedMsg`, then still sends one full load, and that `stop` returns.
- No output a user sees changes. No version bump.

Testing:
- Each new test goes red when its change is reverted: the first board from `trees.Load`, the loader as plain `trees.Load`, the pathspecs as files again, and `Start` skipping the load after a failed watcher.
- `gitc.Authors` with paths in two folders, plus a committed file in one of them that was not asked, returns exactly the asked paths with their authors, and names each folder once on the command line.
