---
parent: specs/2026-11-23-fewer-planning-commits-design
depth: minimal
id: PLN-0250
created: "2026-11-23 20:40:14"
hash: fnc1mnm
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
- [ ] Red: rows for `PLN-<hash>` (accepted) and `PLN-0250` (refused) in the existing `fixed_in` table fail because only a sha passes.
- [ ] Change: widen the `fixed_in` check in `ops.go` and its error text.
- [ ] Commit: feat(write): fixed_in takes a plan hash

### Task 2: Bug detail names the fixing plan

**Files:** `internal/tui/detail.go`, `internal/tui/detail_test.go`
**Seam:** the bug detail rows built in `detail.go`
**verify:** A bug whose `fixed_in` is `PLN-<hash>` shows that value and the plan title; a sha, and a hash with no matching plan, show the value alone with no crash.
- [ ] Red: an assert in the nearest bug detail test fails because the FIXED row shows no plan title.
- [ ] Change: look the hash up on the board and append the plan title to the FIXED row.
- [ ] Commit: feat(tui): bug detail names the plan that fixed it

### Task 3: One write for a plan's land bookkeeping

**Files:** `internal/write/landnotes.go` (new), `internal/write/landnotes_test.go` (new), `internal/write/specdates.go`
**Seam:** `write.LandNotes(cfg config.Config, planHash string) ([]string, error)`, returning the paths it changed
**verify:** For a plan, the bug of `parent: bugs/<name>` and every `BUG-` id in `closes:` get `fixed_in: PLN-<hash>`, every line of each `DBT-` id in `closes:` is ticked, and the spec dates are written as `SpecDates` writes them; a `closes:` id with no file fails before any file changes. List each case: bug parent, bug in closes, debt in closes, sibling plan already done (writes `finished`), plan with no spec, plan with nothing to close, missing id.
- [ ] Red: a new table test over temp planning roots fails because `LandNotes` does not exist.
- [ ] Change: add `LandNotes`, reusing the tick and `SpecDates` code; it writes files and never commits.
- [ ] Commit: feat(write): one write for a plan's land bookkeeping

### Task 4: acta tidy --plan writes the bookkeeping into the chain

**Files:** `internal/tidy/tidy.go`, `internal/tidy/remap.go`, `internal/tidy/tidy_test.go`, `internal/cli/tidy.go`
**Seam:** `tidy.Run(repo, tidy.Options{..., Plan: hash})` and `acta tidy --plan <hash> <parent> <branch>`
**verify:** With `--plan`, the tip tree equals the merge of parent and branch plus the `LandNotes` writes plus today's hash swaps, the commit count is the same as without `--plan`, and the tip message has one `Land:` body line; `proof` refuses any other planning change, and a failure writes no ref. List each case: bug-fix plan, main changed the spec frontmatter after the fork, a sibling plan landed on main after the fork, missing `closes:` id, a tampered tree.
- [ ] Red: rows in `TestKeepRuleAndFolds` style tests with `Plan` set fail because `Options` has no `Plan` and the tip has no `fixed_in`.
- [ ] Change: copy the target tree's planning files to a temp dir, run `LandNotes` there, write the changed blobs into the last node's tree, add the `Land:` body line, and let `proof` accept exactly those writes; add the `--plan` flag.
- [ ] Commit: feat(tidy): write the land bookkeeping into the tidy chain

### Task 5: acta land-notes for full mode

**Files:** `internal/cli/landnotes.go` (new), `internal/cli/landnotes_test.go` (new), `internal/cli/cli.go`, `internal/cli/help.go`
**Seam:** `acta land-notes <plan hash>`
**verify:** Wiring only. On main after a merge, one run makes exactly one commit `chore(acta): land of PLN-<hash>` holding every file `LandNotes` changed; an error from `LandNotes` exits non-zero and commits nothing.
- [ ] Red: a CLI test that runs `land-notes` in a temp repo fails because the command does not exist.
- [ ] Change: add the command, call `write.LandNotes`, commit the paths with `gitc.CommitPathsIn`, and add it to help.
- [ ] Commit: feat(cli): acta land-notes writes a plan's bookkeeping in one commit

### Task 6: Land and bug skills use the new bookkeeping

**Files:** `plugin/skills/land/SKILL.md`, `plugin/skills/land/tidy.md`, `plugin/skills/bug/SKILL.md`, `plugin/evals/land-tidy-progress/` (scaffold and a new grader), `plugin/evals/land-full-progress/` (a new grader), `.acta/wiki/tick-stamps-spec.md` (no cap change: the edits remove more text than they add)
**Seam:** the land and bug skill text, proven by `scripts/eval --case land-tidy-progress` and `scripts/eval --case land-full-progress`
**verify:** No land step writes `fixed_in`, ticks debt or runs `spec-dates` on main after the parent moves in tidy mode; full mode runs `acta land-notes` once after the merge and nothing else; the bug skill says land writes `fixed_in: PLN-<hash>` and a sha is only for a fix outside a plan. List every land line checked: SKILL.md steps 5, 8, 9; tidy.md steps 1, 4, 5, 8, 9.
- [ ] Red: the new graders fail on the current skill text: `scripts/eval --case land-tidy-progress` finds no `fixed_in: PLN-` in the tidy tip, and `scripts/eval --case land-full-progress` finds `acta spec-dates` after the merge.
- [ ] Change: edit the skill lines per the spec, add a bug and a debt item to the tidy case scaffold, add both graders, update the wiki page.
- [ ] Commit: feat(land): bookkeeping rides the tidy chain or one land-notes commit
