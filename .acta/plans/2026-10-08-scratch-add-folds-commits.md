---
parent: specs/2026-10-08-scratch-add-folds-commits
depth: minimal
id: PLN-0119
created: "2026-10-08 08:25:17"
hash: jmy430y
started: "2026-10-08 08:26:49"
---
# acta scratch add folds into its own last commit

**Goal:** Many `acta scratch add` calls on one item in a row leave one commit, not one per call, and never rewrite a commit that anything else already holds.

**Spec:** .acta/specs/2026-10-08-scratch-add-folds-commits.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Only `scratch add` folds. Every other write command keeps one commit per call.
- A failed fold never loses text: any failed check or failed amend ends in the plain commit `gitc.Commit` makes today.
- Tests use temp repos only (`setupRepo` in `internal/gitc/gitc_test.go`, the temp config in `internal/write/scratch_test.go`). acta write commands auto-commit, so never run them in this repo.

## Waves

- Wave 1: Task 01, Task 03
- Wave 2: Task 02

### Task 01: gitc.CommitOrFold

**Files:**
- Modify: `internal/gitc/gitc.go`
- Test: `internal/gitc/gitc_test.go`

**verify:** `CommitOrFold` amends HEAD only when all five fold rules in the spec hold, and on every other path makes the same commit or skip that `Commit` makes. List each path checked (each rule broken on its own, `wasDirty`, busy repo, detached HEAD, amend error) and what it returns.

- [x] Failing test: `TestCommitOrFold` in `internal/gitc/gitc_test.go`, table-driven over temp repos from `setupRepo`. Case "folds": commit `a.md` with message M through `CommitOrFold`, change `a.md`, call again with M; HEAD count stays the same, `Result{Committed: true, Folded: true}`, and HEAD holds both changes. One case per broken rule, each expecting a new commit and `Folded: false`: subject differs; HEAD also changes `b.md`; HEAD is a merge commit; a bare remote fetched so `origin/main` contains HEAD; another local branch at HEAD; a linked worktree on a new branch from HEAD; HEAD author name differs from `git config user.name`. Plus: `wasDirty` true gives the same skip as `Commit`. Run `scripts/test ./internal/gitc/ -run TestCommitOrFold` and see it fail.
- [x] Code: add `Folded bool` to `Result`. Add `CommitOrFold(repo, path, msg string, wasDirty bool) Result` next to `Commit`: same first checks as `Commit` (git repo, `wasDirty`, `busy`, which already refuses a detached HEAD), then a small `canFold(repo, path, msg) bool` that checks the five rules with `run` (`log -1 --format=%s%x00%an%x00%P`, `diff-tree --no-commit-id --name-only -r HEAD`, `branch -r --contains HEAD`, `for-each-ref --contains HEAD refs/heads`, and `Worktrees` plus `IsAncestor` for other checkouts), then `git add -- path` and `git commit --amend --only -m msg -- path`. Any error inside `canFold` means no fold. An amend error falls back to `Commit`. Comments in plain English say why each rule is there.
- [x] Run `scripts/test ./internal/gitc/` passes, `go vet ./internal/gitc/` and `gofmt -l internal/gitc` clean, commit.

### Task 02: scratch add uses CommitOrFold

**Files:**
- Modify: `internal/write/scratch.go`, `internal/write/ops.go`, `.acta/wiki/commit-subject-form.md`
- Test: `internal/write/scratch_test.go`

**verify:** Any run of `scratch add` calls on one item with nothing in between leaves exactly one `add to <stem>` commit holding every text, and every other write command (`scratch new`, `SetValue`, bugs, debt) still makes one commit per call. List each command checked and how many commits it made.

- [ ] Failing test: in `internal/write/scratch_test.go`, three `AddScratch` calls on one item in a row make one `chore(scratch): add to <stem>` commit whose file holds all three texts; a `SetValue` between two adds makes the second add a new commit. Run `scripts/test ./internal/write/ -run Scratch` and see it fail.
- [ ] Code: add `finishFold` next to `finish` in `internal/write/ops.go`, the same as `finish` but calling `gitc.CommitOrFold`; `internal/write/scratch.go:128` calls `finishFold`. Fix any older scratch test that counted one commit per add. Add one line to `.acta/wiki/commit-subject-form.md` (bump its `timestamp`): `scratch add` amends its own last commit when that commit is the same item's unshared add, so a brainstorm leaves one commit per item.
- [ ] Run `scripts/test ./internal/write/` passes, vet and gofmt clean, commit.

### Task 03: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.40`, so `0.1.41`. List each file and its version.

- [x] Failing test: none new; `internal/plugincheck` already checks the three files agree.
- [x] Code: set `"version": "0.1.41"` in the three files.
- [x] Run `scripts/test ./internal/plugincheck/` passes, commit.
