---
parent: specs/2026-10-12-acta-help-and-man-page-design
depth: minimal
closes: [DBT-0034.07, DBT-0044.05]
id: PLN-0138
created: "2026-10-12 09:47:15"
hash: gbwihvg
started: "2026-10-12 09:49:31"
finished: "2026-10-12 10:00:50"
---

# acta Help, Version and Man Page Implementation Plan

**Goal:** `acta help`, `acta <cmd> --help`, `acta --version` and `man acta` all work, from one command table.

**Spec:** .acta/specs/2026-10-12-acta-help-and-man-page-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- No new dependency. Roff is written by hand from the table with `strings.Builder`.
- `acta` with no arguments still opens the TUI, and `acta -root X` still works.
- `acta run-one -- <cmd> --help` passes `--help` to the child command: nothing after `--` is ever read as a help flag.
- Help goes to stdout with exit 0. Errors go to stderr with exit 1 (`exitBadInput`).
- The version string comes from `plugin.Version()`; the source build mark comes from `update.IsRelease()`.

## Waves

- Wave 1: Task 01
- Wave 2: Task 02
- Wave 3: Task 03

### Task 01: Command table, help and version

**Files:**
- Create: `internal/cli/help.go`, `internal/cli/help_test.go`
- Modify: `internal/cli/cli.go` (`Run`, the `default:` case, `flags`)

One table in `help.go` holds every command (name, group, summary, usage, flags, example) in the spec's four groups with `voice` left out. `Run` answers `help`, `help <cmd>`, `-h`, `--help`, `<cmd> -h`, `<cmd> --help` (only before a `--`) and `--version` from it, and the unknown-command error becomes `unknown command "x"; run acta help`.

**verify:** Every way of asking for help (`help`, `help <cmd>`, `-h`, `--help`, `<cmd> -h`, `<cmd> --help`, `-root X --help`) prints table text to stdout with exit 0, and none prints the Go flag default `Usage of`; `help <unknown>` and an unknown command go to stderr with exit 1. Every `case` string in `Run`'s switch has a table entry, checked by a test that reads `cli.go` with `go/ast`, so a new command without an entry fails. `--version` prints `acta <version>` for a release build and `acta <version> (source build)` otherwise. List every entry path checked, and show that a `--help` after `--` in `run-one` still reaches the child.

- [x] **Step 1:** Write failing tests in `internal/cli/help_test.go` for every help path, `--version` and the unknown-command hint; run `scripts/test ./internal/cli -run 'TestHelp|TestVersion|TestUnknown'` and watch them fail.
- [x] **Step 2:** Add the table and routing in `help.go` and `cli.go`; run the same command and watch it pass.
- [x] **Step 3:** Run `scripts/test ./internal/cli`, `go vet ./internal/cli`, `gofmt -l internal/cli`, then commit with trailer `Task: PLN-gbwihvg#1`.

### Task 02: Man page written by setup and update

**Files:**
- Create: `internal/cli/man.go`, `internal/cli/man_test.go`
- Modify: `internal/cli/setup_cmd.go` (`cmdSetup`), `internal/cli/update_cmd.go` (`cmdRefreshPlugin` or `refreshPlugin`)

`man.go` renders `acta.1` in roff from the Task 01 table (NAME, SYNOPSIS, one section per group, `--version`) and writes it to `<dir of exePath()>/../share/man/man1/acta.1`, making the folders. `cmdSetup` and the plugin refresh in `acta update` call it; a failure prints one warning line and changes neither the flow nor the exit code.

**verify:** The man page lists every command and flag that `acta help` lists, and no other, checked by a test that walks the table; roff special characters in table text (a leading `.` or `'`, a backslash) are escaped. The file lands at the path derived from the binary location for a binary under a temp `bin/`. List every failure path (unwritable folder, a file where the folder should be, `exePath` error) and show that each only warns while setup and update keep their exit code.

- [x] **Step 1:** Write failing tests in `internal/cli/man_test.go` for the roff text, the escape rules, the install path and each failure path; run `scripts/test ./internal/cli -run TestMan` and watch them fail.
- [x] **Step 2:** Add `man.go` and the calls in `cmdSetup` and the plugin refresh; run the same command and watch it pass.
- [x] **Step 3:** Run `scripts/test ./internal/cli`, `go vet ./internal/cli`, `gofmt -l internal/cli`, then commit with trailer `Task: PLN-gbwihvg#2`.

### Task 03: Usage strings read from the table

**Files:**
- Modify: `internal/cli/cli.go`, `internal/cli/commit.go`, `internal/cli/commits.go`, `internal/cli/config_cmd.go`, `internal/cli/doctor.go`, `internal/cli/eval_omp.go`, `internal/cli/hook.go`, `internal/cli/id.go`, `internal/cli/migrate.go`, `internal/cli/runone.go`, `internal/cli/setup_cmd.go`, `internal/cli/state.go`, `internal/cli/tick.go`, `internal/cli/tidy.go`, `internal/cli/update_cmd.go`, `internal/cli/wiki.go`
- Test: `internal/cli/help_test.go`

Every `usage: acta ...` string and `*Usage` constant in `internal/cli` is replaced by a lookup into the Task 01 table, so the short usage line on a bad call and `acta help <cmd>` come from one place.

**verify:** No `usage: acta` literal is left in a non-test Go file of `internal/cli` outside `help.go`, checked by a test that scans the package source. Each bad-call path that printed a usage line before still prints the same usage text to stderr with the same exit code. List every command checked.

- [x] **Step 1:** Write the failing source-scan test in `internal/cli/help_test.go`; run `scripts/test ./internal/cli -run TestNoUsageLiteral` and watch it fail.
- [x] **Step 2:** Replace each usage string and `*Usage` constant with a table lookup; run the same command and watch it pass.
- [x] **Step 3:** Run `scripts/test ./internal/cli`, `go vet ./internal/cli`, `gofmt -l internal/cli`, then commit with trailer `Task: PLN-gbwihvg#3`.

## Polish

### Task 04: Review polish

**verify:** Each NOTE below is applied, and nothing else changes.

- [x] `internal/cli/update_cmd_test.go`: the four `TestUpdateRefreshPlugin*` tests stub `exePath` to a binary under `t.TempDir()` (the `useExe` helper from `man_test.go`), so no run writes `acta.1` outside the test folder.
- [x] `internal/cli/man.go`: the escape comment says a backslash becomes `\e`, as the code does.
- [x] `internal/cli/man.go`: the install-path comment says links are kept on macOS but followed on Linux, where `os.Executable` reads `/proc/self/exe`.
- [x] `internal/cli/man.go`: every `-` in rendered roff text becomes `\-`, not only a leading one, so `-h, --help` and the usage lines copy as plain ASCII dashes; a test pins `\-\-help` after the comma.
- [x] `internal/cli/help.go`: `tidy`, `migrate-root`, `doctor`, `setup` and `update` get an example, the `tick` usage line names `--agent NAME`, and the writing-commands test list covers the five new names.
- [x] Commit: `polish: review notes PLN-0138`

## Review notes

- Review: deep level, full tier, round 1 CLEAN from both reviewers (no BLOCKER, so no fact-check); polish commit e666ee9 reviewed by both reviewers, CLEAN. Debt filed for three NOTEs.
- `acta help help` prints `unknown command "help"; run acta help`; exit code is right, the wording reads oddly.
- Help detection treats a flag value that is literally `-h` or `--help` as a help request (`acta commit p -m -h` prints help and commits nothing); no agent passes such values.
- `--h` and `-version` are not caught by the help and version routes and fall back to Go's flag output.
- The `.TH` line has no date, so `mandoc -T lint` warns; long summary lines give STYLE warnings only.
- Task 03 changed the bad-call usage text of `acta dispatch` and `acta wiki` to repeat the command before each subcommand; accepted, nothing parses it.
- Task 01 edited four existing tests outside its file list, because the spec moved help to stdout; accepted.
- The `tidy` example names this repo's branch `acta-help-man`; a neutral name would age better.
- No test pins `--agent NAME` in the `tick` usage line or the `useExe` stubs in the `TestUpdateRefreshPlugin*` tests.

