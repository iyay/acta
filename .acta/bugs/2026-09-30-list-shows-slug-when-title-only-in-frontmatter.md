---
id: BUG-0018
hash: j4wi74l
started: "2026-09-30"
fixed_in: 2c4324a
finished: "2026-09-30"
---
# List shows the slug instead of the title for old scratch items

## Symptom
Some scratch rows in the TUI list, and `acta show`, show the file slug (for example `bug-debt-severity`) instead of the title the item was made with. `acta show SCR-0011` prints `bug-debt-severity` as the title. 16 files under `.acta/` are hit, all scratch items made before 2026-09-29.

## Root cause
The board takes an item's title only from the first `# ` heading in the body (`internal/board/parse.go:99`). When there is no heading, it falls back to the slug (`internal/board/board.go:392`). It never reads the `title:` field in the frontmatter.
Before commit 25d71a2 (2026-09-29), `acta scratch new --title` wrote the title only into the frontmatter (`internal/write/scratch.go:42`) and put no `# ` heading in the body. So those old items have a title the board cannot see. New items get both, so only old files are hit.

## Repro
1. `acta show SCR-0011`: the title shown is `bug-debt-severity`.
2. `grep '^title:' .acta/scratch/2026-09-29-bug-debt-severity.md` prints `title: Severity or priority for bugs and debt`.
3. The same file has no line that starts with `# `.

## Found in
main at e48bce4, while checking scratch SCR-0021 (acta:debug, reading).
