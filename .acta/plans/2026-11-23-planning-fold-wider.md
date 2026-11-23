---
parent: specs/2026-11-23-fewer-planning-commits-design
depth: minimal
id: PLN-0251
created: "2026-11-23 20:40:14"
hash: ofwrqfn
started: "2026-11-23 20:42:42"
finished: "2026-11-24 04:55:24"
---
# Wider fold for planning commits on main

**Goal:** A planning write on main folds into an unpushed planning HEAD whatever its subject or kind, so one step makes one planning commit.

**Spec:** `.acta/specs/2026-11-23-fewer-planning-commits-design.md` (SPC-0217), section "Wider fold on main".
**Tests:** fast `scripts/test ./internal/gitc ./internal/cli ./internal/write`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched and `scripts/eval --case <name>` for one case; whole when the diff touches `plugin/hooks/` or `plugin/references/`. This plan touches neither, so the gate does not run.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- A red test is a new row in an existing table test, else an assert in the nearest test of the same unit; a new test func only for a behaviour no test checks yet.
- A pushed, tagged, merge or code commit is never amended. A write with any path outside the planning root keeps today's rules exactly.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: Fold any planning HEAD under the root

**Files:** `internal/gitc/gitc.go`, `internal/gitc/subject.go`, `internal/gitc/gitc_test.go`, `internal/gitc/subject_test.go`, `.acta/wiki/commit-subject-form.md`
**Seam:** `gitc.CommitOrFoldIn(repo, root, path, msg, wasDirty)` and `gitc.CommitPathsIn(repo, root, paths, msg)`
**verify:** A write whose paths all sit under root folds into HEAD exactly when HEAD touches only root files, is unpushed, untagged, not a merge, has the user as author, and no other worktree holds it; every other case makes a plain commit and never amends. List each case tried: scratch then spec then plan, `docs(spec)` HEAD, `chore: assign short ids` HEAD, two-path ids write, a write with one path outside root, pushed HEAD, tagged HEAD, code HEAD, worktree on HEAD. The mixed-kind subject is `chore(acta): <kinds> of <stems>` with every earlier subject kept in the body.
- [x] Red: new rows in the `TestCommitOrFoldInSameKind` table (mixed kinds, `docs(spec)` HEAD, kind-less `chore:` HEAD, two-path write) fail because `canFold` needs a `chore(` HEAD and one shared kind.
- [x] Change: in `canFold` skip the `chore(` prefix and same-kind checks when HEAD's files and every new path sit under root; build the `chore(acta)` subject in `subject.go` for mixed or kind-less folds; add that form to the wiki page.
- [x] Commit: feat(gitc): fold any unpushed planning commit under the planning root

### Task 2: spec-dates and mark writes take the fold

**Files:** `internal/cli/specdates.go`, `internal/cli/specdates_test.go`, `internal/write/mark.go`, `internal/write/mark_test.go`
**Seam:** `cmdSpecDates` and `write.MarkItem`
**verify:** Wiring only. `acta spec-dates` right after a planning commit folds into it; `MarkItem` passes the planning root so its write folds the same way. One success path each, and one case each where HEAD is a code commit and a new commit is made.
- [x] Red: a row in `TestSpecDatesCommitsAndLeavesNoChange` expects one commit after a planning HEAD and fails because `cmdSpecDates` calls `gitc.Commit`.
- [x] Change: `cmdSpecDates` calls `gitc.CommitOrFoldIn(cfg.RepoRoot, cfg.Root, ...)`; `MarkItem` calls `gitc.CommitPathsIn(cfg.RepoRoot, cfg.Root, ...)`.
- [x] Commit: feat(cli): spec dates and marks fold into the planning commit

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] internal/gitc/gitc.go inRoot: when the planning root is the repo top-level (root: . or ACTA_ROOT=.), every file counts as planning, so a planning write folds into a code commit (repro: feat HEAD on src/x.go, root = repo, folded=true); return false when the resolved root equals --show-toplevel, with a fold-table row.
- [x] internal/gitc/gitc.go foldStems: a symlinked root with resolved paths leaks ".." as a kind (subject "chore(acta): scratch, spec, .. of a, s"); resolve root the way inRoot does, with a fold-table row.
- [x] .acta/wiki/commit-subject-form.md: say a hand-written HEAD keeps its subject even when a new file comes in, and put back that outside the root acta id may fold into a non-chore HEAD that added exactly the files it writes; stay under 250 words.
- [x] Commit: `polish: review notes for PLN-0251`

## Review notes

- Review depth auto, picked standard (two reviewers, deep lens on the Standards axis): risk signal is a path that amends git commits. Round 1 7ff9c0f..6d629c8 CLEAN on both axes; polish 9ad1ea4..4dac864 CLEAN. BLOCKER 0, NOTE 11, pre-existing 2. Notes: acta commit with a docs(spec) message folds under a chore HEAD and takes the chore(acta) subject; subject_test.go not touched; spec-dates could amend a planning-only tidy tip (no real case); inRoot measures unresolved paths against a resolved root (fails safe); busy() misses a stopped multi-commit cherry-pick (pre-existing); trailers read from the last paragraph only; case-insensitive root spelling fails safe; polish subject wording differs from the plan.

## State

### Next

Both tasks committed (18a1ad4 Task 1, Task 2 on top). Next: acta:review over the branch, then land.
PATH acta still has the old fold; rebuild it after land so the wider fold takes effect.

### Open rulings

Answered 2026-11-23: write rows join Task 1; a hand-written planning HEAD keeps its subject; a Task trailer HEAD never folds.
