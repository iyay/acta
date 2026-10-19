---
parent: specs/2026-10-19-drop-repo-language-and-subagent-models-design
depth: minimal
id: PLN-0175
created: "2026-10-19 20:02:37"
hash: tg80o89
started: "2026-10-19 20:10:26"
finished: "2026-10-19 20:16:26"
---

# Drop repo language and subagent models: Go side

**Goal:** Remove the `repo_language` and `subagent_models` settings from the acta binary, ignore them in old files, and drop them on every write.

**Spec:** .acta/specs/2026-10-19-drop-repo-language-and-subagent-models-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Repo files are always English. No setting picks another language.
- Old files holding `repo_language` or `subagent_models` load with no error and no message.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2, Task 3

### Task 1: Config drops both keys on read and write

**Files:** `internal/config/user.go`, `internal/config/repo_user.go`, `internal/config/user_test.go`, `internal/config/repo_user_test.go`, `internal/plugincheck/no_old_names_test.go`, `internal/plugincheck/style_test.go`
**Seam:** `config.LoadUserFile`, `config.SaveUserFile`, `config.SaveRepoUser`, `config.User`
**verify:** No load or save path keeps either old key. List each reader and writer of the global and repo files checked; for each, a file holding both keys plus others loads with no error, and after save the old keys are gone while every other key and value is unchanged. `SaveRepoUser` removes a `repo_language` node even when the call sets another key. `--unset repo_language` and setting `repo_language` per repo are rejected like any unknown repo key.
**Commit:** `refactor(config): drop repo_language and subagent_models`

- [x] Red: failing test for the verify line, watched failing
- [x] Change: minimal code, focused tests pass
- [x] Commit with the plan task trailer

### Task 2: Hook note always names English and has no model line

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/cli/hook_repo_test.go`
**Seam:** `hook.Build` session note text (the function `hook_test.go` calls today)
**verify:** No session note any config can produce names a repo language other than English or holds a `Subagent models` line. List every note variant checked (each style, harness, and a config file still holding the old keys).
**Commit:** `refactor(hook): repo files always English, no subagent model line`

- [x] Red: failing test for the verify line, watched failing
- [x] Change: minimal code, focused tests pass
- [x] Commit with the plan task trailer

### Task 3: CLI and setup drop the flags and questions

**Files:** `internal/cli/config_cmd.go`, `internal/cli/help.go`, `internal/cli/cli_test.go`, `internal/cli/config_repo_test.go`, `internal/cli/notice_test.go`, `internal/cli/setup_cmd_test.go`, `internal/setup/form.go`, `internal/setup/form_test.go`, `internal/setup/plan_test.go`, `internal/setup/run_test.go`
**Seam:** `cli.Run` with `config show`, `config show --json`, `config set`; `setup` form fields
**verify:** No CLI or setup surface names either setting: list each checked (`config show`, `--json`, `config set` flags, `--repo` flags, help text, setup form). `--repo-language`, `--subagent-models` and `--clear-subagent-models` fail as unknown flags. A config file still holding both keys prints neither in `config show` or `--json` (wiring only; the ignore rule lives in Task 1).
**Commit:** `refactor(cli): drop repo language and subagent models flags`

- [x] Red: failing test for the verify line, watched failing
- [x] Change: minimal code, focused tests pass
- [x] Commit with the plan task trailer

## State

### Findings

Removing the User fields breaks hook, cli and setup at compile time, so task 1's commit carries every source change; tasks 2 and 3 add only their tests.
MergeRepo refused subagent_models as a personal key; it is now dropped, so old .acta.yaml files load.
Minimal plan had no checkboxes, so tick --all failed; boxes were added in the worktree copy.

## Fix round 1

### Task 4: Keep comments and drop old keys on every repo write

**Files:** `internal/config/repo_user.go`, `internal/config/legacy_keys_test.go`, `internal/hook/hook_test.go`
**Seam:** `config.SaveRepoUser`, `config.UnsetRepoUser`
**verify:** No write to `.acta.yaml` loses a user comment or keeps an old key. Both writers drop `repo_language` and `subagent_models`; a comment above a dropped key moves to the next kept key, or closes the file when none follows. List each writer and comment place checked (top, middle, last key).
**Commit:** `fix(config): keep comments when old repo keys are dropped`

- [x] Red: failing tests for comment carry in SaveRepoUser and old-key drop in UnsetRepoUser
- [x] Change: one shared key-removal helper used by both writers; stale hook test comments fixed
- [x] Commit with the plan task trailer

## Polish

### Task 5: Review polish

**verify:** Each NOTE below applied, and nothing else changes.

- [x] SaveRepoUser test case where the dropped key ends the file, so the end-of-file comment branch is covered
- [x] Comment above that branch says a leftover comment goes under the last key, not that the set always adds one
- [x] `var carried string` in place of `carried := ""` plus a second assignment
- [x] Commit: `polish: review notes PLN-0175`

## Review notes

Deep review, 2 rounds plus polish: round 1 BLOCKER (SaveRepoUser lost comments above dropped keys) fixed in Task 4; low notes left as is (test helper ignores a read error, nested old maps lose inner comments by design).
