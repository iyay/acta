---
parent: debt/2026-11-07-stray-worktrees-command
closes: [DBT-0156.01, DBT-0156.02, DBT-0156.03, DBT-0156.04, DBT-0156.05]
depth: minimal
id: PLN-0218
created: "2026-11-07 19:53:22"
hash: ne34gjc
started: "2026-11-07 19:57:16"
---
# Worktree debt: Go code

**Goal:** Close the five DBT-0156 items: unknown ages and deleted refs stay protected, doctor tells clean from dirty strays, setup skips the worktree check, and acta state finds the plan being built.

**Spec:** `.acta/specs/2026-11-07-worktree-debt-design.md` (items 1 to 5)

**Tests:** fast `scripts/test ./internal/worktrees ./internal/doctor ./internal/cli ./internal/board`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- No git call this plan adds may force anything: no `--force`, `-f` on `worktree remove`, or `-D` on `branch`.
- Every `git status` the new code runs uses `--no-optional-locks`.
- Only `acta state <plan>` and `acta state set` change how they find a plan. The other callers of `board.PlanStates` keep today's behaviour.

## Waves

- Wave 1: Task 1, Task 2, Task 4
- Wave 2: Task 3

### Task 1: Unknown ages and deleted refs stay protected

**Files:** `internal/worktrees/worktrees.go`, `internal/worktrees/worktrees_test.go`
**Seam:** `worktrees.List(repo, plans, now)` and `worktrees.Clean(repo, plans, now)`
**verify:** No worktree whose age or commits cannot be read is ever a stray. List each case: admin folder present with no HEAD, index or logs/HEAD (protected, "cannot read activity times"); missing worktree whose branch ref was deleted and whose last logs/HEAD commit no branch, tag or remote holds (protected); the same with that commit held by main (missing, removed by Clean); the same with logs/HEAD unreadable (protected, "cannot read its commits"); missing worktree whose branch ref exists (missing, as today).
- [x] Red: table cases for each path fail because the empty admin folder reads idle and the deleted-ref worktree reads missing.
- [x] Change: treat a zero activity time as unknown, and for a missing worktree whose ref `git show-ref --verify` cannot find, check the commit from the last line of `<admin>/logs/HEAD` like a detached one.
- [x] Commit: fix(worktrees): keep worktrees whose age or commit is unknown

### Task 2: Doctor tells clean from dirty strays

**Files:** `internal/doctor/worktrees.go`, `internal/doctor/worktrees_test.go`, `internal/doctor/doctor.go`, `internal/cli/doctor.go`, `internal/cli/worktrees_test.go`
**Seam:** `doctor.Run(doctor.Env)` with the strays split by state in `Env`, filled by `internal/cli/doctor.go`.
**verify:** The doctor line never offers a fix that cannot fix what it counts. List each case: only clean or missing strays (count, fix names `acta worktrees --clean`); only dirty strays (count, no command in the fix, says they need the user); both (both counts, fix says `--clean` takes the clean ones and the dirty ones need the user); none (OK). The CLI wiring passes one success path through `strayWorktrees`.
- [x] Red: doctor tests with dirty-only and mixed strays fail because the message counts them as one and always names `--clean`.
- [x] Change: split the Env field into clean-or-missing and dirty lists and word the message and fix from them.
- [x] Commit: fix(doctor): say which stray worktrees --clean can remove

### Task 3: Setup skips the worktree check

**Files:** `internal/cli/setup_cmd.go`, `internal/cli/doctor.go`, `internal/cli/setup_cmd_test.go`
**Seam:** `doctorEnv` as `acta setup` and `acta doctor` call it.
**verify:** `acta setup` never runs `git status` in a linked worktree, and `acta doctor` still runs the worktree check. Wiring only: one setup run in a repo with a stray shows no worktrees line, and one doctor run still does.
- [ ] Red: a setup test in a repo with an idle stray fails because the summary holds a worktrees line.
- [ ] Change: let setup build its doctor Env without the stray lists, and let doctor keep filling them.
- [ ] Commit: perf(setup): skip the stray worktree check

### Task 4: acta state finds the plan being built

**Files:** `internal/cli/state.go`, `internal/cli/state_test.go`
**Seam:** `cli.Run([]string{"state", <plan>})` and `cli.Run([]string{"state", "set", <plan>, <part>})`
**verify:** A plan named on the command line is found in the linked worktree the command runs in when `git status` there shows its file modified or untracked, before any branch commit touched it. List each case: plan file modified and uncommitted in the current linked worktree (found); the same plan file unchanged in another linked worktree (not found, as today); run from the main checkout with the file changed there (not found, as today); a plan the branch committed (found, as today).
- [x] Red: a state test in a fresh worktree with an uncommitted tick fails with "no running plan".
- [x] Change: when the named plan is not in `board.PlanStates`, look it up in the current linked worktree's board if git status shows its file changed there.
- [x] Commit: fix(state): read the plan being built before its first commit
