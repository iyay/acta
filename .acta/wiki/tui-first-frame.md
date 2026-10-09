---
type: Decision
title: The TUI's first frame reads only the main tree
description: Before the first frame runTUI only reads main tree files; git, worktrees, authors and the watcher load after it
paths: [internal/cli/cli.go, internal/tui/watch.go, internal/tui/model.go, internal/board/closed.go, internal/trees/]
timestamp: 2026-10-09T14:53:31Z
---

The first frame of `acta` must show in under 100 ms, however many commits, worktrees or branches the repo has (SPC-0119, BUG-0040).

So `tuiBoards` in `internal/cli/cli.go` gives `runTUI` the first board from `board.Load`: the files of the main tree and nothing else. Anything that runs git waits for the background load: the worktree and branch scan in `trees.Others`, the author lookup in `Board.FillAuthors`, and the watcher's folder list. `tui.Start` sets the watcher up off the first frame, then sends one full load from `trees.LoadWithAuthors`. That load runs after the watcher has its folders, so a change made during startup is not lost. `Model.Init` loads nothing, or startup would load twice.

Why: on this repo the old path took about 1.1 s. About 780 ms was a full-history `git log` per planning folder for the AUTHOR line alone, and about 220 ms was the same worktree scan run twice.

Authors stay out of `board.Load` and `trees.Load`, so CLI commands never pay for them. Only the TUI calls `FillAuthors`, and it asks git once per checkout. `gitc.Authors` names the folders of the files, not each file, so the command line stays short however many files there are.

`TestTUIBoardsFirstFrameIsMainOnly` fails if the first board or the later loader goes back to the wrong load. Check the timing with the startup probe in /tui-pty-probe.md: five runs, each first frame under 100 ms.
