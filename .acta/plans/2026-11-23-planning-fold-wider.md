---
parent: specs/2026-11-23-fewer-planning-commits-design
depth: minimal
id: PLN-0251
created: "2026-11-23 20:40:14"
hash: ofwrqfn
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
- [ ] Red: new rows in the `TestCommitOrFoldInSameKind` table (mixed kinds, `docs(spec)` HEAD, kind-less `chore:` HEAD, two-path write) fail because `canFold` needs a `chore(` HEAD and one shared kind.
- [ ] Change: in `canFold` skip the `chore(` prefix and same-kind checks when HEAD's files and every new path sit under root; build the `chore(acta)` subject in `subject.go` for mixed or kind-less folds; add that form to the wiki page.
- [ ] Commit: feat(gitc): fold any unpushed planning commit under the planning root

### Task 2: spec-dates and mark writes take the fold

**Files:** `internal/cli/specdates.go`, `internal/cli/specdates_test.go`, `internal/write/mark.go`, `internal/write/mark_test.go`
**Seam:** `cmdSpecDates` and `write.MarkItem`
**verify:** Wiring only. `acta spec-dates` right after a planning commit folds into it; `MarkItem` passes the planning root so its write folds the same way. One success path each, and one case each where HEAD is a code commit and a new commit is made.
- [ ] Red: a row in `TestSpecDatesCommitsAndLeavesNoChange` expects one commit after a planning HEAD and fails because `cmdSpecDates` calls `gitc.Commit`.
- [ ] Change: `cmdSpecDates` calls `gitc.CommitOrFoldIn(cfg.RepoRoot, cfg.Root, ...)`; `MarkItem` calls `gitc.CommitPathsIn(cfg.RepoRoot, cfg.Root, ...)`.
- [ ] Commit: feat(cli): spec dates and marks fold into the planning commit
