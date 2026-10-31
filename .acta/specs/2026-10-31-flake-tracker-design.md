---
parent: scratch/2026-10-31-flake-tracker
closes: [SCR-0062, SCR-0079, DBT-0147.02, DBT-0145.03]
id: SPC-0185
created: "2026-10-31 17:56:44"
hash: v83vz2l
started: "2026-10-31 18:02:11"
finished: "2026-10-31 19:41:53"
---
# Flake tracker for tests and evals

## Problem

Some tests and eval cases fail now and then and pass when run again. Today a
red full suite at land means one more full run, and a red eval case means more
eval runs. Each costs minutes to an hour. The known flakes (the TUI click
test, `TestWheelNotchesBelongToTheScreenTheyWereGatheredOn`, the task-trailer
eval, the brainstorm evals, debug-race) are tracked only in debt lines, scratch
items and agent memory. Nothing counts how often each one fails, so nobody can
see which flake costs the most.

## Scope

Three plans under this spec, cut by file area so two of them can build at
once:

- Flake record: the `flaky_tests` field and `acta flake record`.
- Flake rerun: the automatic rerun in this repo's `scripts/test --full`.
- Land rule: the land, build and debug rules, the same rule for red eval
  cases at the land gate, and the seed bugs. It waits for the flake record
  plan.

The record and the land rule work in every repo that uses acta. The automatic
rerun is per repo: this repo's `scripts/test` is the first one. A repo with
pytest or npm can wrap its own test command the same way.

Out of scope: TUI changes, fixing the root cause of any flake, and the real
reds in SCR-0079 (`answers-appended` and `build-close-no-suite` are red on
main too, so they are not flakes). Those two become their own bug.

## Shared contract

A test command may print lines of this form:

```
FLAKY: <name> (failed <k> of <n>)
```

`<name>` is the test's full name: `<package path>.<TestName>` for Go, such as
`internal/tui.TestClickOnTheBarTakesTheHighlightOff`, and `eval:<case>` for an
eval case. A FLAKY line means the test failed, then passed when run alone.

## Section 1: detection in scripts/test --full

A small Go helper in this repo runs `go test -json` inside the run-one lock.
Its output reads like plain `go test` output, so logs and the land baseline
look the same as today.

- A failed test runs again alone, with `-run '^TestX$' -count=1` and the same
  flags as the first run (tags, `-short`), up to 2 more times. It stops at the
  first pass.
- A failed subtest reruns its top-level test.
- A pass on a rerun prints the FLAKY line. When every failed test passed on a
  rerun, the run exits 0.
- A test that fails all 3 times is a real red: the run exits 1 and names it.
- No rerun, exit 1 at once: a build failure, a panic or timeout with no test
  name, or more than 5 failed tests in one run.
- Runs without `--full` do not change.

## Section 2: the record

A flake is a bug. The bug frontmatter gets a `flaky_tests:` field: a list of
test names in the contract form. One bug may list several tests that share a
cause.

`acta flake record <name> [--plan <plan id>]`:

- Finds the open bug (no `fixed_in`) whose `flaky_tests` holds the name, and
  adds one line to its `## Flakes` section: the time and the plan.
- With no such bug, files a new one: title `Flaky test <name>`, priority
  medium, root cause "Not known yet.", `flaky_tests: [<name>]`, and the first
  `## Flakes` line.
- A bug that has `fixed_in` and flakes again is not reopened. A new bug is
  filed, and its body names the old bug, so the fix history stays true.
- Commits the bug file on its own. A repeat record of the same file folds
  into its last commit, like `acta scratch add`.
- Prints `BUG-nnnn <title> (<count> flakes)`, where count is the number of
  `## Flakes` lines.

## Section 3: the land rule and evals

- After a full suite that exits 0, land reads the log for FLAKY lines. For
  each one it runs `acta flake record <name> --plan <plan id>` before the
  merge, in the checkout where the suite ran: the worktree, or the main
  checkout for a step-6 rerun. The land report names each flake and its bug.
- A rerun that passes counts as green only through this record. Since the
  record files a bug when none exists, no flake passes silently.
- During build and debug, a failing test listed in an open flake bug runs
  again alone once before any debugging starts.
- A full suite that exits 1 stays as today: fix the cause through the narrow
  failing tests.
- A repo whose test command never prints FLAKY lines sees no change.
- Evals: a red eval case at the land gate runs again alone
  (`--case <name>`) up to 2 times. A pass runs
  `acta flake record eval:<case> --plan <plan id>`. Red all 3 times is a
  real red, handled as today.

## Seed and cleanup

The land rule plan files flake bugs for the known flakes, each with its
`flaky_tests` names: `TestClickOnTheBarTakesTheHighlightOff`,
`TestWheelNotchesBelongToTheScreenTheyWereGatheredOn`, and the task-trailer,
brainstorm and debug-race evals. It also files the bug
for the real SCR-0079 reds. The flake rule now kept in agent memory moves
into the land skill, and the memory line goes.

## Testing

- Helper: table tests over recorded `go test -json` streams for each case
  above (pass on rerun, 3 of 3 red, build failure, panic, more than 5).
- `acta flake record`: new bug, existing bug, fixed bug flaking again, fold of
  a repeat record, the count line.
- Land rule: an eval case whose scaffold suite prints a FLAKY line, graded on
  the `acta flake record` call before the merge.
