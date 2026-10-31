---
id: SCR-0083
hash: r5doywc
title: Flake tracker for tests and evals
status: brainstorming
created: "2026-10-31 17:28:57"
schema: "1"
started: "2026-10-31 17:44:18"
finished: "2026-10-31 17:56:44"
---
# Flake tracker for tests and evals

## Words

### 2026-10-31

User asked (2026-10-31, WhatsApp): "analisa dan pelajari codebase terakhir. menurut lo ada missing features yang ideal acta punya?" then "catet aja semua ke scratch".

Agent proposal the user filed: Flake tracker. TUI and eval flakes keep coming back (DBT-0147.02, DBT-0145.03); each flake costs a full rerun at land.

## Context

Found while reading `acta list` on main at 7cadc6a. Known flakes: DBT-0147.02 (TestWheelNotchesBelongToTheScreenTheyWereGatheredOn, internal/tui/scroll_test.go:918), the TUI click test, DBT-0145.03 (brainstorm evals 2 of 3), SCR-0062 (task-trailer eval), SCR-0079 (eval reds at PLN-0190 land). Land gate today reruns the full suite on a red.

Shape started 2026-10-31 as Architectural (new mechanism across scripts/test, land skill and the planning files). User merged SCR-0062 and SCR-0079 into it ("gas" on the priority list).
Today a red full suite at land means: fix the cause through narrow tests, then one more full run (plugin/skills/land/SKILL.md step 1). The rule "a lone TestClickOnTheBarTakesTheHighlightOff failure is the known flake, rerun only that test" lives only in agent memory, not in acta. Evals run through `claude plugin eval` (scripts/eval) with a haiku judge; one judged red is noise until repeated (wiki eval-red-case-debug).

## Log

### 2026-10-31

Probe round 1, user answered "ok semua" to every recommendation:
- Q1 scope: record flakes and let land use the record to rerun only the flaky tests, not the full suite.
- Q2: Go tests first in this plan; evals are a second plan under the same spec.
- Q3: store a flake as a bug with a flaky_tests frontmatter field, no new kind.
- Q4: scripts/test --full reruns the failed tests 2 times on its own; a pass on rerun is a flake and prints FLAKY: TestX. Land must update the flake bug so the trend is kept.
- Q5: the real reds of SCR-0079 (answers-appended, build-close-no-suite red on main) are not flakes; they become their own bug, out of this spec.

### 2026-10-31

Probe round 2, user answered "ok semua":
- Q6: generic for every acta repo: the land rule, the flaky_tests field and the acta command. This repo's scripts/test is the first adopter of the automatic rerun.
- Q7: a rerun that passes counts green only when the test is in an open flake bug, or the land files a new flake bug right then. No silent flake.
- Q8: new command `acta flake record <TestName>`: finds the open bug that lists the test and appends one log line (date, plan); with no such bug it files a new flake bug. It commits on its own, like scratch add.
- Q9: no TUI change in this spec.

### 2026-10-31

Probe round 3, user answered "oke":
- Q10: evals in plan 2 follow the same rule: a red case reruns alone up to 2 times; a pass is a flake, recorded as eval:<case>.
- Q11: the build seeds flake bugs for the known flakes (TestClickOnTheBarTakesTheHighlightOff, TestWheelNotchesBelongToTheScreenTheyWereGatheredOn, task-trailer, the brainstorm evals, debug-race). The flake rule kept in agent memory moves into acta.
- Q12: the spec closes SCR-0062, SCR-0079 (its real reds become a new bug), DBT-0147.02 and DBT-0145.03.
Frontier empty; next: design in three sections (detection, record, land rule).

### 2026-10-31

Approaches: A (Go helper in this repo, called by scripts/test, reruns failed tests 2 times), B (acta subcommand for Go only), C (agent reruns by hand). User picked A ("gas"). Shared contract for every repo: a test command may print `FLAKY: <name>` lines; land runs `acta flake record` for each line.

### 2026-10-31

Design section 1 (detection) approved with a thumbs up:
- A small Go helper in this repo runs go test -json inside the run-one lock; the output still reads like plain go test.
- A failed test reruns alone (-run '^TestX$' -count=1, same flags) up to 2 times and stops once green.
- A pass on rerun prints `FLAKY: internal/tui.TestX (failed 1 of 2)` and the run exits 0; failing 3 of 3 is a real red, exit 1.
- No rerun for a build failure, a panic or timeout with no test name, or more than 5 failed tests.
- -short runs are unchanged; a failed subtest reruns its top-level test.

### 2026-10-31

Design section 2 (record) approved with a thumbs up:
- A bug gets a flaky_tests frontmatter field: a list of names such as internal/tui.TestX or eval:debug-race; one bug may hold several tests.
- `acta flake record <name> [--plan PLN-n]`: the open bug that lists the name gets one line in its ## Flakes section (time, plan).
- No such bug: file a new bug "Flaky test <name>", priority medium, root cause "not known yet".
- A bug with fixed_in that flakes again: a new bug that names the old one, so the fix history stays true.
- It commits on its own; repeat records of the same file fold. Output: `BUG-0060 <title> (3 flakes)`.

### 2026-10-31

Design section 3 (land rule and evals) approved with a thumbs up:
- Suite exit 0: land scans the log for FLAKY: lines and runs `acta flake record <name> --plan <id>` for each before the merge.
- Record in the checkout where the suite ran (worktree, or main for a step-6 rerun); the land report names the flakes.
- During build and debug: a failing test listed in an open flake bug reruns alone once before any debugging.
- Plan 2 evals: a red case at the gate reruns alone up to 2 times; a pass runs `acta flake record eval:<case>`.
- Exit 1 stays as today: fix the cause. A repo with no FLAKY: lines sees no change.

### 2026-10-31

Spec written and committed: .acta/specs/2026-10-31-flake-tracker-design.md. Waiting for the user's review.

### 2026-10-31

User replied "gas spec" (2026-10-31 17:57): spec SPC-0185 approved, no cold read. Next: acta:slice.

### 2026-10-31

Sliced SPC-0185 into three minimal plans by file area (spec Scope edited to match): PLN-0204 flake record, PLN-0205 flake rerun in scripts/test, PLN-0203 land rule (after PLN-0204). Waiting for the user's yes on the plans.

### 2026-10-31

User replied "build subagent" (2026-10-31 18:01): plans approved; build with the subagent executor for this run. PLN-0204 and PLN-0205 start now; PLN-0203 waits for PLN-0204 to land.

### 2026-10-31

Seed done in PLN-0203 task 4 (2026-10-31): flake bugs BUG-0057 (wheel test, filed at the PLN-0205 land), BUG-0058 click test, BUG-0059 task-trailer, BUG-0060 brainstorm-mixed-context, BUG-0061 brainstorm-resume-topic, BUG-0062 debug-race; real red answers-appended filed as a non-flake bug.

## Open questions
