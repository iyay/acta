---
parent: debt/2026-09-29-locks
closes: [DBT-0021.03, DBT-0021.05, DBT-0008.02]
id: SPC-0095
created: "2026-10-06 10:43:34"
hash: zdq6rla
---
Status: Bounded, approved by the user in chat on 2026-10-06. Written down because both changes guard planning data against loss or a wrong tick.
Why: two first-time `acta debt new` calls for the same plan can both see no file and one overwrites the other's notes. The "new debt" commit runs after the lock is released, so a tick landing in between rides along in it. `acta tick` on a debt item trusts the line number from when the board was loaded, so a hand edit in between can tick the wrong box.

# Debt writes: one lock for the whole new-debt step, and ticks that check their line

## Changes

1. **`NewDebt` holds one lock (DBT-0021.03, DBT-0021.05).** In `internal/write/ops.go`, `NewDebt` takes `lock(path)` once, before it reads the file, and releases it after the commit. Under the lock it reads the file again: missing means `createDebt`, present means appending the new lines. `addDebtLines` no longer takes the lock itself, because a second `flock` on the same file from the same process would block. `finish` runs before the unlock, so no other write to that file can land between our write and our commit.
2. **`TickLine` checks its line (DBT-0008.02).** In `internal/write/tick.go`, `TickLine(path, line, text, state)` takes the item's text (its title, which never holds the priority tag). It ticks the box on `line` when that line's text, with any priority tag removed, equals `text`. Otherwise it looks for the one checklist line in the file whose text equals `text` and ticks that. No such line, or more than one, is an error that names the text, and nothing is written. The callers pass `it.Title`: `internal/cli/tick.go` (two places) and `internal/write/mark.go`.

## Out of scope

DBT-0021.01 (safeDir checks only the last folder) and DBT-0008.06 (debt file named by today) stay open as debt.

## Testing

Failing test first for each case, in `internal/write`, on temp repos:
- two goroutines run `NewDebt` for the same plan with different notes and no file yet; both notes end up in the file;
- a `TickLine` that runs while `NewDebt` is between write and commit is not in the "new debt" commit;
- a line inserted above a debt item before the tick: the right box is ticked;
- two lines with the same text, or no line with the text: an error, file unchanged.

Gate: `scripts/test --full`.

## Close

Last task bumps the patch version to 0.1.27 in `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json` and `plugin/package.json`.
