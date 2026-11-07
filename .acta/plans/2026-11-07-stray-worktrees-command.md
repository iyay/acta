---
parent: specs/2026-11-07-stray-worktrees-design
depth: minimal
id: PLN-0215
created: "2026-11-07 17:36:17"
hash: z7jtm2h
started: "2026-11-07 17:43:59"
---
# Stray worktrees: the acta worktrees command

**Goal:** Add `acta worktrees` (list, `--clean`, `--json`) and an `acta doctor` warning for stray worktrees, as the spec defines them.

**Spec:** `.acta/specs/2026-11-07-stray-worktrees-design.md`

**Tests:** fast `scripts/test ./internal/worktrees ./internal/cli ./internal/doctor`; full `scripts/test --full`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- No git call this plan adds may pass `--force` or `-f` to `worktree remove`, or `-D` to `branch`. The pre-tool hook blocks them and the spec forbids them.
- Every `git status` the new code runs uses `--no-optional-locks`, and the 24-hour activity times are read before any git command in that worktree.
- Leave `gitc.Worktrees` as it is: the board and `internal/trees` depend on it skipping prunable worktrees. The new package parses `git worktree list --porcelain` itself, locked and prunable lines included.
- Tests make their own temp repos, set the author inside each repo, and never touch the real home.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2
- Wave 3: Task 3, Task 4

### Task 1: Classify worktrees

**Files:** `internal/worktrees/worktrees.go`, `internal/worktrees/worktrees_test.go`
**Seam:** `worktrees.List(repo string, plans []Plan, now time.Time) ([]Entry, error)`, where `Plan` holds a slug and a status, and `Entry` has the spec's `--json` fields.
**verify:** Every linked worktree gets exactly one state from the spec (`protected` with a reason, `clean`, `dirty`, `missing`), the main checkout and the calling worktree never appear as strays, and no rule lets a worktree that holds unmerged commits, a not-done and not-dropped plan's slug, a lock, or activity within 24 hours become a stray. List each case checked: fresh plan worktree by slug, unmerged branch, fully ticked plan waiting to land, dropped plan with no commits, locked, active within 24 hours, clean, dirty (status capped at 10 lines plus `... and N more`), only gitignored files, detached, missing folder, harness worktree inside the repo, and a listing that leaves every `index` time unchanged.
- [x] Red: a table test over temp repos fails because `internal/worktrees` does not exist.
- [x] Change: add the package that parses `git worktree list --porcelain`, reads activity times first, and applies the spec's rules in its order.
- [x] Commit: feat(worktrees): classify stray worktrees

### Task 2: Clean stray worktrees

**Files:** `internal/worktrees/worktrees.go`, `internal/worktrees/worktrees_test.go`
**Seam:** `worktrees.Clean(repo string, plans []Plan, now time.Time) ([]Entry, error)`
**verify:** `Clean` removes only `clean` and `missing` strays, each with plain `git worktree remove <path>` and then `git branch -d`, and no path through it ever runs a forced remove or a forced branch delete. List each path: removed with branch deleted, removed with `-d` refused (`branch_kept` true and the refusal in `reason`), detached removed with no branch call, a remove git refuses (`kept` with git's error text and no error returned), and dirty or protected left as they were. A `List` right before `Clean` does not stop it from removing an idle clean stray.
- [ ] Red: tests for each path above fail because `Clean` does not exist.
- [ ] Change: add `Clean` on top of `List`, recording each git call so a test can assert no `--force`, `-f` or `-D` appears.
- [ ] Commit: feat(worktrees): remove clean stray worktrees without force

### Task 3: acta worktrees command

**Files:** `internal/cli/worktrees.go`, `internal/cli/worktrees_test.go`, `internal/cli/cli.go`, `internal/cli/help.go`
**Seam:** `cli.Run([]string{"worktrees", ...}, stdout, stderr)`
**verify:** Wiring only. The command loads the board, passes each plan's slug and status to `internal/worktrees`, and prints the text list, or one JSON object per line with `--json`. One success path (`--clean` in a repo with one clean stray removes it and exits 0) and one error passed up (run outside a git repo exits non-zero with the error on stderr). Bad flags print the usage line, and `acta help` lists the command.
- [ ] Red: a CLI test fails because `worktrees` is an unknown command.
- [ ] Change: add `cmdWorktrees`, its case in `cli.go` and its help entry.
- [ ] Commit: feat(cli): add acta worktrees

### Task 4: Doctor warns on stray worktrees

**Files:** `internal/doctor/worktrees.go`, `internal/doctor/worktrees_test.go`, `internal/doctor/doctor.go`, `internal/cli/doctor.go`
**Seam:** `doctor.Run(doctor.Env)` with a new `Env.StrayWorktrees []string`, filled by `internal/cli/doctor.go` from `worktrees.List`.
**verify:** Wiring only. The `worktrees` check is OK with no strays, and warns with the count and the fix `acta worktrees --clean` when there is at least one. It never fails the doctor run, and outside a git repo it prints no worktrees line.
- [ ] Red: a doctor test with one stray in `Env` fails because no `worktrees` result exists.
- [ ] Change: add `checkWorktrees` to `Run`, and fill the new field in the CLI.
- [ ] Commit: feat(doctor): warn on stray worktrees

## State

### Next

Task 2: add Clean on top of List in internal/worktrees; runGit var lets tests record calls.
