---
parent: specs/2026-10-31-flake-tracker-design
depth: minimal
id: PLN-0204
created: "2026-10-31 17:59:57"
hash: de70hm8
started: "2026-10-31 18:02:11"
finished: "2026-10-31 18:13:51"
---
# Flake record implementation plan

**Goal:** Add the `flaky_tests` bug field and `acta flake record <name> [--plan <plan id>]`, which logs a flake on its open bug or files a new flake bug.
**Spec:** `.acta/specs/2026-10-31-flake-tracker-design.md` (Section 2)
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Test names use the contract form: `<package path>.<TestName>` for Go, `eval:<case>` for evals. Matching is exact.
A new flake bug: title `Flaky test <name>`, priority medium, `## Root cause` "Not known yet.", `flaky_tests: [<name>]`, first `## Flakes` line.
Writes go through the existing `internal/write` helpers (`NewBug`, the fold in `finish`), never a bare `os.WriteFile` of a new path.

## Waves

Wave 1: Task 1. Wave 2: Task 2. Wave 3: Task 3.

### Task 1: board reads flaky_tests

**Files:** Modify `internal/board/board.go`, `internal/board/board_test.go`.
**Seam:** `board.Load` and the `Item.FlakyTests []string` field it fills.
**verify:** A bug's `flaky_tests` list loads in file order in both YAML forms (`[a, b]` and a block list), through the same `listField` reader as `closes:`; a bug without the field, and every other kind, has an empty list.
- [x] Red: a table test over a flow list, a block list and a missing field fails because `Item` has no `FlakyTests`.
- [x] Change: Add `FlakyTests` to `Item` and fill it with `listField(doc.Front, "flaky_tests")` for bugs, then run `scripts/test ./internal/board/` to green.
- [x] Commit: `feat(flake): bugs read flaky_tests`.

### Task 2: RecordFlake writes the record

**Files:** Create `internal/write/flake.go`, `internal/write/flake_test.go`.
**Seam:** `write.RecordFlake(cfg config.Config, name, planID string) (Outcome, int, error)`, returning the bug's outcome and its `## Flakes` line count.
**verify:** Every call ends with exactly one bug holding the name and one new `## Flakes` line, never two bugs open for one name. Cases: an open bug listing the name gets the line; no bug files a new flake bug per the Global Constraints; a bug with `fixed_in` gets a new bug whose body names the old id, and the old file stays unchanged; a repeat record of the same file folds into its last commit; an empty name or a name with a newline is refused with nothing written; an unknown `--plan` id is refused.
- [x] Red: the table test fails because `RecordFlake` does not exist.
- [x] Change: Find the bug through `board.Load` and `FlakyTests`, append the `## Flakes` line (`- <time> <plan id or "no plan">`) or build the body for `NewBug`, set `flaky_tests` with `SetField`, then run `scripts/test ./internal/write/ -run TestRecordFlake` to green.
- [x] Commit: `feat(flake): record a flake on its bug`.

### Task 3: acta flake record command

**Files:** Create `internal/cli/flake.go`, `internal/cli/flake_test.go`; modify `internal/cli/cli.go`, `internal/cli/help.go`.
**Seam:** `cli.Run` with `flake record <name> [--plan <plan id>]`.
**verify:** Wiring only: one success path prints `BUG-nnnn <title> (<count> flakes)` and exits 0; one refusal from `RecordFlake` reaches stderr with the bad-input exit; a missing name or unknown subcommand prints the usage line; `acta help flake` shows the entry in the Write group.
- [x] Red: the CLI test fails because `flake` is an unknown command.
- [x] Change: Add the `flake` case to `cli.go`, `cmdFlakeRecord` in `flake.go` and the help entry, then run `scripts/test ./internal/cli/ -run 'TestFlake|TestHelp'` to green.
- [x] Commit: `feat(flake): acta flake record command`.

## Build brief

Land order: PLN-0204 and PLN-0205 land on main in either order; they share no file. PLN-0203 lands after PLN-0204.

## Polish

### Task 4: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] internal/write/flake.go: trim spaces from the name before every check and use, so 'trail.TestY ' and 'trail.TestY' are one name and a second record never files a second open bug.
- [x] internal/write/flake.go: build the slug as "flaky-" plus the cleaned name with its own dashes trimmed, so a name like "[x]" or "../x" gives flaky-x, not a refused flaky--x.
- [x] internal/write/flake.go: cap the slug part at 60 characters (dashes trimmed after the cut), so a long subtest name never fails with "file name too long".
- [x] internal/write/flake.go: write the plan's short id (PLN-nnnn) on the ## Flakes line when the plan has one, not the text as typed.
- [x] internal/cli/flake_test.go: add a case that records onto an existing bug whose title is not "Flaky test <name>" and checks the printed title is the bug's own.
- [x] Commit: `polish: review notes for PLN-0204`

## Review notes

- Review: deep level, no floor raised; round 1 CLEAN, polish review CLEAN. Notes kept, not tracked: a CRLF bug file gets an LF line; YAML non-string names (null, 0x10) do not round-trip; a ### under ## Flakes stays in the list; Legacy bugs are not refused; the CLI loads the board twice for the title; a trailing newline in a name is trimmed; the TrimRight after the slug cut is redundant; the quiet NewBug is not pinned by a test.

## State

### Next

All three tasks done; Task 3 landed in e8efe77 (internal/cli/flake.go, acta flake record).
Next: commit the plan ticks, then acta:review over the branch.

### Findings

acta state set refuses to commit while the plan file holds uncommitted task ticks (exit 2, 'file had other uncommitted changes'); the text is still written. Workers tick, then state set fails every task until the orchestrator commits the plan. Seen on PLN-0204 task 1.
