---
parent: debt/2026-10-06-setup-block-refresh
depth: minimal
closes: [SPC-0096, DBT-0082.01, DBT-0077.01, DBT-0077.02]
id: PLN-0105
created: "2026-10-06 11:59:31"
hash: r7t1ld5
---
# acta state set and round order Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta state set` never empties a State part unless `--clear` asks for it, and `acta state` shows the review round that is really newest.

**Spec:** .acta/specs/2026-10-06-state-set-and-round-order.md

**Tests:** `scripts/test ./internal/write/ ./internal/cli/ ./internal/board/ ./internal/gitc/ ./internal/hook/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Tests build temp repos under `t.TempDir()` and never touch this checkout or HOME; acta write commands auto-commit, so tests run them only in temp repos.
- Run tests with `scripts/test`, never bare `go test`.

## Waves

- Wave 1: Task 01, Task 02
- Wave 2: Task 03

Task 01 and Task 02 touch different files and can run together.

### Task 01: state set refuses an empty stdin without --clear

**Files:** `internal/write/state.go`, `internal/write/state_test.go`, `internal/cli/state.go`, `internal/cli/state_test.go`, `internal/hook/hook.go`, `internal/hook/hook_test.go`

**verify:** No call of `acta state set` empties a State part or commits unless `--clear` is given and stdin is blank; every other blank-stdin path refuses with nothing written. List every stdin and flag combination tested (empty, blank lines only, text, `--clear` alone, `--clear` with text).

- [ ] Red: in `internal/write/state_test.go`, a plan with a filled `next`: an empty body and a blank-lines body without clear return an error naming `--clear`, and the plan file is byte for byte the same with no new commit; clear with an empty body empties `next`; clear with text is refused. In `internal/cli/state_test.go`, `acta state set <plan> next --clear` with no stdin empties it.
- [ ] Green: give the write function a `clear bool` (keep its other parameters); refuse blank body without it with `nothing on stdin; pipe the lines in, or use --clear to empty <part>`; refuse text with it; the CLI parses `--clear` and the usage line reads `usage: acta state set <plan> next|findings|rulings [--clear] < text.md`; `runningRule` in `internal/hook/hook.go` says `run acta state set <plan> next with the lines on stdin`, and its hook test expectation follows.
- [ ] Commit: `fix(state): state set refuses an empty stdin without --clear (DBT-0082.01)`

### Task 02: newest round by git ancestry

**Files:** `internal/gitc/gitc.go`, `internal/gitc/gitc_test.go`, `internal/board/state.go`, `internal/board/state_test.go`

**verify:** Whenever both round subjects are on the branch, `acta state` shows the round whose commit is the descendant; git errors fall back to the first found without failing. List every order tested (new subject newer, old subject newer, only one present, git error).

- [ ] Red: in `internal/gitc/gitc_test.go`, `IsAncestor` on a two-commit repo says true for parent-of-child and false the other way. In `internal/board/state_test.go`, rewrite `TestRunningRoundPrefersNewSubjectWithoutOrder` so the old-subject (`acta: tick fix round`) commit is newer and must be the one shown; add the case where the new subject is newer.
- [ ] Green: add `IsAncestor(repo, a, b string) (bool, error)` running `git merge-base --is-ancestor a b` (exit 1 means false, other failures are errors); `newestRound` returns the second found when the first is its ancestor, else the first; drop the `FirstSeen` calls and fix the doc comment.
- [ ] Commit: `fix(state): newest round follows git ancestry (DBT-0077.01, DBT-0077.02)`

### Task 03: Version 0.1.28

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.28, and `internal/plugincheck` passes. List each file and the version it holds.

- [ ] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [ ] Green: bump the patch from 0.1.27 to 0.1.28 in all three files.
- [ ] Commit: `chore: version 0.1.28`
