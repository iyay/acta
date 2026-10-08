---
parent: specs/2026-10-08-repo-override-visible-design
closes: [SCR-0055]
depth: minimal
id: PLN-0118
created: "2026-10-08 07:44:25"
hash: rnplm6t
started: "2026-10-08 07:45:45"
---
# Repo overrides are visible and changeable

**Goal:** A repo value in `.acta.yaml` that differs from the user's own value is named in `acta config show`, the session note, the setup wizard and the setup skill, and can be kept, written or removed.

**Spec:** .acta/specs/2026-10-08-repo-override-visible-design.md

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Repo keys are exactly `repo_language`, `build_executor`, `plan_depth`, `commit_history`, `coding_guide` (`config.RepoKeys`). Personal keys never go into `.acta.yaml`.
- Nothing commits `.acta.yaml`, and nothing writes it without the user choosing to.
- Tests use temp repos and a temp `HOME`; never this repo's `.acta.yaml` or the real home. acta write commands auto-commit, so tests run them only in temp clones.
- Text the agent or the user reads in the wizard, the hook and the skill is plain English with no repo paths and no user names.

## Waves

- Wave 1: Task 01, Task 06
- Wave 2: Task 02, Task 03, Task 04
- Wave 3: Task 05

### Task 01: Overrides and UnsetRepoUser in internal/config

**Files:**
- Modify: `internal/config/repo_user.go`
- Test: `internal/config/repo_user_test.go`

**verify:** `Overrides` lists a repo key exactly when its `.acta.yaml` value differs from the user's effective value (defaults included), on every input: same, different, user unset, no file, unreadable file, personal key in the file. `UnsetRepoUser` removes only the named keys, keeps every other line and order, and refuses any key outside `RepoKeys`. List each input checked and its result.

**Interfaces:**
- Produces: `type Override struct{ Key, Repo, Yours string }`; `func Overrides(user User, repoRoot string) ([]Override, error)` (same errors as `MergeRepo`, list in `RepoKeys` order); `func UnsetRepoUser(repoRoot string, keys []string) (string, error)` (returns the file path; `ErrBadUser` for a non-repo key).

- [x] Failing test: tables for `Overrides` and `UnsetRepoUser` over the inputs in the verify line, in `t.TempDir()` repos; run `scripts/test ./internal/config/` and see them fail (functions missing).
- [x] Code: add both functions next to `MergeRepo` and `SaveRepoUser`, reusing their file reading, defaults and yaml node walk.
- [x] Run `scripts/test ./internal/config/` passes, vet and gofmt clean, commit.

### Task 02: acta config set --repo --unset and config show overrides

**Files:**
- Modify: `internal/cli/config_cmd.go`
- Test: `internal/cli/config_repo_test.go`

**verify:** `--unset` only ever changes `.acta.yaml`, only with `--repo`, never for a key also being set in the same call, and never commits; `config show` names the user's value next to every differing repo key in text and JSON, and prints the old `(repo)` mark for an equal one. List every flag mix tried and its exit code and file result.

**Interfaces:**
- Consumes: `config.Overrides`, `config.UnsetRepoUser`, `config.Override` from Task 01.

- [x] Failing test: `--unset` removes a key, repeats, mixes with a set, refuses same key set and unset, refuses without `--repo`, refuses a non-repo key, makes no commit; `config show` prints `subagent (repo; yours: dispatch)` and JSON `overrides`, `from_repo` kept; run `scripts/test ./internal/cli/ -run TestConfig` and see them fail.
- [x] Code: a repeatable `--unset` flag on `config set`, wired to `UnsetRepoUser` inside `setRepo`; `config show` uses `Overrides` for the mark and the JSON field; update `configUsage`.
- [x] Run `scripts/test ./internal/cli/ -run TestConfig` passes, vet and gofmt clean, commit.

### Task 03: Session note names repo overrides

**Files:**
- Modify: `internal/hook/hook.go`, `internal/cli/hook.go`
- Test: `internal/hook/hook_test.go`, `internal/cli/hook_test.go` (whichever already covers the `RepoErr` line)

**verify:** The note names every differing key with its repo and user value whenever overrides exist and `RepoErr` is nil, and adds no override text in any other case. List every input combination checked.

**Interfaces:**
- Consumes: `config.Overrides`, `config.Override` from Task 01.

- [x] Failing test: an input with overrides gives one block naming key, repo value, user value and `acta config set --repo --unset <key>`; no overrides gives no block; `RepoErr` set gives no block; run `scripts/test ./internal/hook/ ./internal/cli/ -run Hook` and see them fail.
- [x] Code: `Overrides []config.Override` on the hook input, filled in `internal/cli/hook.go` next to `MergeRepo`; the block written next to the plugin-conflict block; check any hook text size cap in `internal/plugincheck` still holds.
- [x] Run `scripts/test ./internal/hook/ ./internal/cli/ -run Hook` passes, vet and gofmt clean, commit.

### Task 04: Setup wizard asks about each repo override

**Files:**
- Modify: `internal/setup/form.go`, `internal/setup/plan.go`, `internal/setup/run.go`
- Test: `internal/setup/form_test.go`, `internal/setup/plan_test.go`, `internal/setup/run_test.go`

**verify:** The wizard writes `.acta.yaml` only for a key the user answered yours or remove, never on the default answer, never outside a git repo, never when `.acta.yaml` cannot be read, and never commits. List every path checked and what lands on disk.

**Interfaces:**
- Consumes: `config.Overrides`, `config.SaveRepoUser`, `config.UnsetRepoUser` from Task 01.

- [x] Failing test: no difference gives no question; outside a repo gives no question; keep, yours and remove give no action, a set-repo action and an unset-repo action; the default is keep; an unreadable `.acta.yaml` gives one line and the other steps go on; the closing line shows only when `.acta.yaml` changed; run `scripts/test ./internal/setup/` and see them fail.
- [x] Code: overrides computed from the values just chosen; one select per key with keep as default; two new action kinds shown in the summary and run through `SaveRepoUser` / `UnsetRepoUser`.
- [x] Run `scripts/test ./internal/setup/` passes, vet and gofmt clean, commit.

### Task 05: Setup skill handles repo overrides, with an eval case

**Files:**
- Modify: `plugin/skills/setup/SKILL.md`, `internal/plugincheck/skill_setup_test.go`, `internal/plugincheck/budget_test.go`, `internal/plugincheck/evals_test.go`
- Create: `plugin/evals/repo-override-ask/` (`case.yaml`, `prompt.md`, `scaffold.sh`, `graders/`)

**verify:** No route through the setup skill writes `.acta.yaml` before the user picks write or remove, and the skill names the shared, committed file every time it asks. The eval case fails when the agent writes the repo file or runs `config set` without asking. List each grader and the behaviour it rejects.

**Interfaces:**
- Consumes: `acta config show --json` field `overrides` and `acta config set --repo --unset <key>` from Task 02.

- [ ] Failing test: add `--unset` and `overrides` to the setup skill's required strings and the `repo-override-ask` row to `evalCases` (phrase from the new skill text); run `scripts/test ./internal/plugincheck/` and see it fail.
- [ ] Code: the skill text from spec section 4; raise `MaxLines` and the byte cap only as far as the new text needs; the eval case copied in shape from an existing scaffolded case, `.acta.yaml` holding `build_executor: dispatch`, graders for the reply naming `.acta.yaml`, no Bash call matching `config set`, and `.acta.yaml` still holding `build_executor: dispatch`.
- [ ] Run `scripts/test ./internal/plugincheck/` passes, commit.

### Task 06: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.39`, so `0.1.40`. List each file and its version.

- [x] Failing test: change only `plugin.json` to `0.1.40`; `scripts/test ./internal/plugincheck/ -run TestManifests` fails.
- [x] Code: set `0.1.40` in the other two files.
- [x] Run `scripts/test ./internal/plugincheck/ -run TestManifests` passes, commit.
