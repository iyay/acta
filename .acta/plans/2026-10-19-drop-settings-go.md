---
parent: specs/2026-10-19-drop-repo-language-and-subagent-models-design
depth: minimal
id: PLN-0175
created: "2026-10-19 20:02:37"
hash: tg80o89
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

### Task 2: Hook note always names English and has no model line

**Files:** `internal/hook/hook.go`, `internal/hook/hook_test.go`, `internal/cli/hook_repo_test.go`
**Seam:** `hook.Build` session note text (the function `hook_test.go` calls today)
**verify:** No session note any config can produce names a repo language other than English or holds a `Subagent models` line. List every note variant checked (each style, harness, and a config file still holding the old keys).
**Commit:** `refactor(hook): repo files always English, no subagent model line`

### Task 3: CLI and setup drop the flags and questions

**Files:** `internal/cli/config_cmd.go`, `internal/cli/help.go`, `internal/cli/cli_test.go`, `internal/cli/config_repo_test.go`, `internal/cli/notice_test.go`, `internal/cli/setup_cmd_test.go`, `internal/setup/form.go`, `internal/setup/form_test.go`, `internal/setup/plan_test.go`, `internal/setup/run_test.go`
**Seam:** `cli.Run` with `config show`, `config show --json`, `config set`; `setup` form fields
**verify:** No CLI or setup surface names either setting: list each checked (`config show`, `--json`, `config set` flags, `--repo` flags, help text, setup form). `--repo-language`, `--subagent-models` and `--clear-subagent-models` fail as unknown flags. A config file still holding both keys prints neither in `config show` or `--json` (wiring only; the ignore rule lives in Task 1).
**Commit:** `refactor(cli): drop repo language and subagent models flags`
