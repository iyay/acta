---
parent: specs/2026-11-23-fewer-planning-commits-design
depth: minimal
id: PLN-0250
created: "2026-11-23 20:40:14"
hash: fnc1mnm
started: "2026-11-23 20:42:48"
finished: "2026-11-23 21:02:45"
---
# Land bookkeeping rides the land

**Goal:** `fixed_in`, debt ticks and spec dates of a plan land inside the tidy chain (tidy mode) or as one commit after the merge (full mode), so a land leaves no follow-up commits on main.

**Spec:** `.acta/specs/2026-11-23-fewer-planning-commits-design.md` (SPC-0217), sections "Land bookkeeping" and "Edge cases".
**Tests:** fast `scripts/test ./internal/write ./internal/tidy ./internal/cli ./internal/tui`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched and `scripts/eval --case <name>` for one case; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches `plugin/skills/land/` and `plugin/skills/bug/`, so the gate runs as `scripts/eval --tag=land` (no case carries the `bug` tag).

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- A red test is a new row in an existing table test, else an assert in the nearest test of the same unit; a new test func only for a behaviour no test checks yet.
- One write code path for both modes: tidy and `acta land-notes` call the same `write` function.
- `acta cleanup apply` stays sha only; `internal/cleanup/apply.go` does not change.

## Waves

- Wave 1: Task 1, Task 2
- Wave 2: Task 3
- Wave 3: Task 4, Task 5
- Wave 4: Task 6

### Task 1: fixed_in takes a plan hash

**Files:** `internal/write/ops.go`, `internal/write/ops_test.go`
**Seam:** `acta set bugs/<name> fixed_in <value>` through `write.Set`
**verify:** `fixed_in` accepts exactly a 7 to 40 character lower-case hex sha or `PLN-` plus a 7 character plan hash, and refuses every other value with an error that names both forms. List each value tried: sha 7 and 40, `PLN-<hash>`, `PLN-0250`, `PLN-` alone, upper-case hex, a hash of 6 and 8 characters.
- [x] Red: rows for `PLN-<hash>` (accepted) and `PLN-0250` (refused) in the existing `fixed_in` table fail because only a sha passes.
- [x] Change: widen the `fixed_in` check in `ops.go` and its error text.
- [x] Commit: feat(write): fixed_in takes a plan hash

### Task 2: Bug detail names the fixing plan

**Files:** `internal/tui/detail.go`, `internal/tui/detail_test.go`
**Seam:** the bug detail rows built in `detail.go`
**verify:** A bug whose `fixed_in` is `PLN-<hash>` shows that value and the plan title; a sha, and a hash with no matching plan, show the value alone with no crash.
- [x] Red: an assert in the nearest bug detail test fails because the FIXED row shows no plan title.
- [x] Change: look the hash up on the board and append the plan title to the FIXED row.
- [x] Commit: feat(tui): bug detail names the plan that fixed it

### Task 3: One write for a plan's land bookkeeping

**Files:** `internal/write/landnotes.go` (new), `internal/write/landnotes_test.go` (new), `internal/write/specdates.go`
**Seam:** `write.LandNotes(cfg config.Config, planHash string) ([]string, error)`, returning the paths it changed
**verify:** For a plan, the bug of `parent: bugs/<name>` and every `BUG-` id in `closes:` get `fixed_in: PLN-<hash>`, every line of each `DBT-` id in `closes:` is ticked, and the spec dates are written as `SpecDates` writes them; a `closes:` id with no file fails before any file changes. List each case: bug parent, bug in closes, debt in closes, sibling plan already done (writes `finished`), plan with no spec, plan with nothing to close, missing id.
- [x] Red: a new table test over temp planning roots fails because `LandNotes` does not exist.
- [x] Change: add `LandNotes`, reusing the tick and `SpecDates` code; it writes files and never commits.
- [x] Commit: feat(write): one write for a plan's land bookkeeping

### Task 4: acta tidy --plan writes the bookkeeping into the chain

**Files:** `internal/tidy/tidy.go`, `internal/tidy/remap.go`, `internal/tidy/tidy_test.go`, `internal/cli/tidy.go`
**Seam:** `tidy.Run(repo, tidy.Options{..., Plan: hash})` and `acta tidy --plan <hash> <parent> <branch>`
**verify:** With `--plan`, the tip tree equals the merge of parent and branch plus the `LandNotes` writes plus today's hash swaps, the commit count is the same as without `--plan`, and the tip message has one `Land:` body line; `proof` refuses any other planning change, and a failure writes no ref. List each case: bug-fix plan, main changed the spec frontmatter after the fork, a sibling plan landed on main after the fork, missing `closes:` id, a tampered tree.
- [x] Red: rows in `TestKeepRuleAndFolds` style tests with `Plan` set fail because `Options` has no `Plan` and the tip has no `fixed_in`.
- [x] Change: copy the target tree's planning files to a temp dir, run `LandNotes` there, write the changed blobs into the last node's tree, add the `Land:` body line, and let `proof` accept exactly those writes; add the `--plan` flag.
- [x] Commit: feat(tidy): write the land bookkeeping into the tidy chain

### Task 5: acta land-notes for full mode

**Files:** `internal/cli/landnotes.go` (new), `internal/cli/landnotes_test.go` (new), `internal/cli/cli.go`, `internal/cli/help.go`
**Seam:** `acta land-notes <plan hash>`
**verify:** Wiring only. On main after a merge, one run makes exactly one commit `chore(acta): land of PLN-<hash>` holding every file `LandNotes` changed; an error from `LandNotes` exits non-zero and commits nothing.
- [x] Red: a CLI test that runs `land-notes` in a temp repo fails because the command does not exist.
- [x] Change: add the command, call `write.LandNotes`, commit the paths with `gitc.CommitPathsIn`, and add it to help.
- [x] Commit: feat(cli): acta land-notes writes a plan's bookkeeping in one commit

### Task 6: Land and bug skills use the new bookkeeping

**Files:** `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, `plugin/skills/bug/SKILL.md`, `plugin/evals/land-tidy-progress/` (scaffold and a new grader), `plugin/evals/land-full-progress/` (a new grader), `.acta/wiki/tick-stamps-spec.md` (no cap change: the edits remove more text than they add)
**Seam:** the land and bug skill text, proven by `scripts/eval --case land-tidy-progress` and `scripts/eval --case land-full-progress`
**verify:** No land step writes `fixed_in`, ticks debt or runs `spec-dates` on main after the parent moves in tidy mode; full mode runs `acta land-notes` once after the merge and nothing else; the bug skill says land writes `fixed_in: PLN-<hash>` and a sha is only for a fix outside a plan. List every land line checked: SKILL.md steps 5, 8, 9; tidy.md steps 1, 4, 5, 8, 9.
- [x] Red: the new graders fail on the current skill text: `scripts/eval --case land-tidy-progress` finds no `fixed_in: PLN-` in the tidy tip, and `scripts/eval --case land-full-progress` finds `acta spec-dates` after the merge.
- [x] Change: edit the skill lines per the spec, add a bug and a debt item to the tidy case scaffold, add both graders, update the wiki page.
- [x] Commit: feat(land): bookkeeping rides the tidy chain or one land-notes commit

## Fix round 1

### Task 7: Telemetry, tidy rerun line, lost-note test, wiki facts

**verify:** Every help command is a known telemetry name; a land in either mode records plan_landed once; a tidy rerun keeps --plan; proof refuses a tree that drops any note; both wiki pages keep every fact they held at BASE that is still true, under 250 words. List each path checked.

- [x] BLOCKER: internal/telemetry/store.go:139 commandKeys lacks "land-notes", so scripts/test ./internal/cli fails TestEveryHelpCommandIsAKnownTelemetryName ("command \"land-notes\" counts as other"); add it to the list.
- [x] [fix] plan_landed telemetry is only recorded in internal/cli/specdates.go:50, which land no longer runs; record it in cmdLandNotes after its commit and in cmdTidy when --plan succeeds.
- [x] [fix] plugin/skills/land/tidy.md step 3, FLAKY line: "run acta tidy again" drops --plan; say to run the step 1 command again.
- [x] [fix] internal/tidy/tidy.go proof: no test covers errLostNotes (a tree that drops a note file passes when the found-count check is disabled); add a row to TestPlanNotesRideTheTip for a tampered tree that loses one note.
- [x] [fix] .acta/wiki/land-tidy-history.md and tick-stamps-spec.md rewrites dropped true facts: a docs commit outside the planning root stays; the full fallback order of the fold subject; wiki checks ride in .acta/wiki/.checked; brief gates allow the parent file. Put them back within the 250 word limit.
- [x] Commit with the trailer `Task: PLN-fnc1mnm#7`

## Fix round 2

### Task 8: Count a tidy land once

**verify:** One land in tidy mode records plan_landed exactly once however many times tidy --plan runs before the parent moves; a first run counts; a failed run counts nothing. List each path checked: first run, rerun with the ref present, failed run, a new land after the ref was deleted.

- [x] BLOCKER: internal/cli/tidy.go:49-51 counts plan_landed on every successful tidy --plan run, so the FLAKY rerun in tidy.md step 3 counts one land twice (proven by the Standards reviewer: two runs give map[plan_landed:2]); count only when refs/acta/tidy/<branch> did not exist before this run.
- [x] Commit with the trailer `Task: PLN-fnc1mnm#8`

## Review notes

- Round 1 (29d68c3..302940c): BLOCKER telemetry list missing land-notes, proven in a temp clone: fixed in fix round 1.
- Pre-existing internal/board/closes.go:47 rejects BUG and whole DBT ids in closes: sent to bug (2026-11-23-plan-closes-rejects-bug-and-debt-ids).
- Pre-existing tick lock files pile up under the user cache, tidy --plan adds one per run: sent to bug (2026-11-23-tick-lock-files-pile-up).
- NOTE open: Land body line is added even when LandNotes wrote nothing.
- NOTE open: closesOf in landnotes.go copies board.listField.
- NOTE open: landTargets fails on any missing closes id, scratch and spec included; no live case.
- NOTE rejected: no plugincheck pin for land-notes or tidy --plan; SPC-0218 makes text pins a BLOCKER, the evals cover them.
- NOTE open: bug skill uses acta set <bug id> fixed_in, which resolves ids; fine.
- NOTE open: land-notes may fold into a chore ids HEAD; allowed by the spec edge cases.
- NOTE open: LandNotes overwrites an existing fixed_in on a bug named in closes.
- Round 2 (302940c..c2b62a2): Spec CLEAN, Standards BLOCKED. BLOCKER tidy --plan rerun double counts plan_landed: fixed in fix round 2.
- NOTE open: smaller true facts still missing from land-tidy-history.md and tick-stamps-spec.md (old base definition, trailer reader, no-ancestor reason); pages sit at the word cap.
- NOTE open: .acta/specs/2026-11-12-usage-telemetry-design.md:56 still says plan_landed counts on spec-dates.
- Review depth auto, picked standard (two reviewers): risk signal is skill text and a commit-writing path. Round 1 29d68c3..302940c BLOCKED (telemetry list), round 2 302940c..c2b62a2 BLOCKED (tidy rerun double count), round 3 c2b62a2..302940c CLEAN on both axes. BLOCKER 2, NOTE 16, pre-existing 2 (sent to bug). Notes: Land line added when nothing was written; closesOf copies board.listField; landTargets fails on any missing closes id; land-notes may fold into an ids HEAD; LandNotes overwrites an existing fixed_in; smaller wiki facts still missing at the word cap; usage telemetry spec still names spec-dates; rev-parse error reads as no ref; ref read and write not atomic.

## State

### Next

SPC-0217 and SPC-0218, parallel run.
PLN-0250 (land-bookkeeping-in-tidy): review round 1 running over 29d68c3..302940c.
PLN-0251 (planning-fold-wider): stuck at task 1, waiting on user rulings Q1-Q3 (fix old fold tests, keep a hand-made subject, never fold a Task-trailer commit).
PLN-0252 (test-rules): running, task 1 done.
Parent branch: main.

### Findings

The board only links scratch, spec and debt-line ids in closes (internal/board/closes.go); BUG and whole DBT ids show as problems there, so LandNotes reads the raw closes list from the plan file (closesOf).
