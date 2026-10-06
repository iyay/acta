---
parent: debt/2026-09-29-locks
depth: minimal
closes: [SPC-0095, DBT-0021.03, DBT-0021.05, DBT-0008.02]
id: PLN-0104
created: "2026-10-06 10:48:28"
hash: ffaorfe
started: "2026-10-06 10:52:59"
finished: "2026-10-06 11:17:36"
---
# Debt write safety Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta debt new` never loses notes to a race and never commits someone else's tick, and a debt tick never marks the wrong box.

**Spec:** .acta/specs/2026-10-06-debt-write-safety.md

**Tests:** `scripts/test ./internal/write/ ./internal/cli/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `lock(path)` in `internal/write/tick.go` is a `flock`; the same process taking it twice on one file blocks, so any function called while the lock is held must not take it again.
- Tests run on temp repos under `t.TempDir()` and keep the lock folder pointed at a temp folder (`lockRoot`), as the existing write tests do; never touch the real repo, cache folder or HOME.
- Run tests with `scripts/test`, never bare `go test`.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02
- Wave 3: Task 03

Task 02 changes `TickLine`'s signature, which `internal/write/debt_test.go` calls, and Task 01 adds tests to that file, so they run one after another.

### Task 01: NewDebt holds one lock through the commit

**Files:** `internal/write/ops.go`, `internal/write/debt_test.go`

**verify:** No two `NewDebt` calls on the same debt file, whether the file exists yet or not, can lose a note, and no other write to that file can land between `NewDebt`'s write and its commit. List every path through `NewDebt` (create, append with new lines, append with nothing new) and where each takes and releases the lock.

- [x] Red: two goroutines call `NewDebt` for the same plan with different notes and no debt file yet; after both, the file holds both notes (today one is lost). A second test: a `TickLine` started while `NewDebt` holds the file must wait, so the "new debt" commit holds no tick.
- [x] Green: `NewDebt` takes `lock(path)` before reading the file and defers the unlock; it picks `createDebt` or the append path under the lock; `addDebtLines` drops its own `lock`; `finish` runs inside the lock.
- [x] Commit: `fix(write): new debt holds one lock through its commit (DBT-0021.03, DBT-0021.05)`

### Task 02: TickLine checks its line by text

**Files:** `internal/write/tick.go`, `internal/write/mark.go`, `internal/cli/tick.go`, `internal/write/tick_test.go`, `internal/write/debt_test.go`

**verify:** `TickLine` never changes a box whose text (priority tag removed) is not the item's text; a stale line number either finds the one matching line or refuses with nothing written. List every case tested (line matches, line moved, text twice, text missing, tagged line).

- [x] Red: in `tick_test.go`, a debt file where a line was inserted above the item after the board was read: `TickLine` with the old line number and the item's text ticks the right box; a file with the text twice and a file without it both return an error and leave the file byte for byte the same; a `(high) ` tagged line matches its untagged text.
- [x] Green: `TickLine(path string, line int, text string, state byte) error`; compare the line's text without the priority tag (reuse the tag reader the board uses) to `text`; on mismatch scan for exactly one matching checklist line; update the callers in `internal/cli/tick.go` (two places) and `internal/write/mark.go` to pass `it.Title`, and the existing calls in `debt_test.go`.
- [x] Commit: `fix(write): a debt tick checks its line by text (DBT-0008.02)`

### Task 03: Version 0.1.27

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.27, and `internal/plugincheck` passes. List each file and the version it holds.

- [x] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [x] Green: bump the patch from 0.1.26 to 0.1.27 in all three files.
- [x] Commit: `chore: version 0.1.27`

## State

### Next

task-03 done. Commit 5f71d36 bumped 0.1.27 all 3. plugincheck pass. Review/land next. debt DBT-0021 started-line left uncommitted in worktree.

## Fix round 1

### Task 04: Tick text matches the board title exactly

**Files:** `internal/write/tick.go`, `internal/write/tick_test.go`, `internal/write/ops.go`, `internal/write/debt_test.go`

**verify:** For every debt line the board parses, the text `lineText` gives equals the title the board gives, so `TickLine` ticks every item it ticked before this plan. List every spacing case tested (trailing space, two spaces after the box, two spaces after the tag, plain).

- [ ] Red: in `tick_test.go`, parse a debt file holding `- [ ] foo ` and `- [ ]  (high) bar` with `board.Parse`, then `TickLine` each item with its parsed title; both must tick (today: "no checklist line with text").
- [ ] Green: `lineText` takes the text the same way `itemRe` group 2 does (everything after `] `, no `TrimSpace`), then strips the priority tag the same way the board does for its title.
- [ ] Drop the `// ponytail:` line above `findLine`; say in plain English why it returns -1 and -2.
- [ ] Remove `appendDebt` from `ops.go` (no production caller left); point its tests at `NewDebt`.
- [ ] Commit: `fix(write): debt tick text matches the board title exactly`
