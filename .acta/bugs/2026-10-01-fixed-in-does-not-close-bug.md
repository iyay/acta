---
id: BUG-0023
hash: ccjztcl
fixed_in: ccc03ef
finished: "2026-10-01 14:40:03"
---
# A bug with fixed_in stays open when no plan has it as parent

## Symptom
A bug whose fix has landed and whose file has `fixed_in: <merge sha>` still shows as `open` in `acta list` and in the TUI. On main b904b88 three bugs hang like this: BUG-0015 (fixed_in c7a9554), BUG-0020 (fixed_in b5f7571) and BUG-0002 (fixed_in 622a74a). Every fix commit is on main. BUG-0014 has the same fixed_in as BUG-0015 and shows `fixed`, because the plan that fixed both (PLN-0046) names only BUG-0014 as its parent.

## Root cause
A bug's status comes only from plans whose frontmatter says `parent: bugs/<stem>`. In `internal/board/board.go:621-626` a bug with no written status gets `parentStatus` (`internal/board/board.go:679`), which returns `open` when `plans == 0`. `fixed_in` is read into `Item.FixedIn` (`internal/board/board.go:413`) but never used for status. `closedByStatus` (`internal/board/closes.go:84`) only works for specs. A plan has one parent and `closes:` cannot name a bug, so a bug fixed as a side effect, or by a plan whose parent is a spec, can only be closed by `fixed_in`, and that does nothing. `acta:land` step 8 and `acta:bug` both tell agents to record `fixed_in`, so they expect it to close the bug.

## Repro
In an empty git repo:

    mkdir -p .acta/bugs
    printf -- '---\nid: BUG-0001\nfixed_in: abc1234\n---\n# Some bug\n' > .acta/bugs/2026-10-01-x.md
    git add -A && git commit -m x
    acta list -type bug -all

It prints `open ... BUG-0001`. Want: `fixed`.

## Found in
main b904b88, found by acta:debug while re-assessing the open bug list.
