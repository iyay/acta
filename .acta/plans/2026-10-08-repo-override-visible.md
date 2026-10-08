---
parent: specs/2026-10-08-repo-override-visible-design
closes: [SCR-0055]
depth: minimal
id: PLN-0118
created: "2026-10-08 07:44:25"
hash: rnplm6t
started: "2026-10-08 07:45:45"
finished: "2026-10-08 07:52:35"
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

- [x] Failing test: add `--unset` and `overrides` to the setup skill's required strings and the `repo-override-ask` row to `evalCases` (phrase from the new skill text); run `scripts/test ./internal/plugincheck/` and see it fail.
- [x] Code: the skill text from spec section 4; raise `MaxLines` and the byte cap only as far as the new text needs; the eval case copied in shape from an existing scaffolded case, `.acta.yaml` holding `build_executor: dispatch`, graders for the reply naming `.acta.yaml`, no Bash call matching `config set`, and `.acta.yaml` still holding `build_executor: dispatch`.
- [x] Run `scripts/test ./internal/plugincheck/` passes, commit.

### Task 06: Bump the plugin patch version

**Files:**
- Modify: `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** All three files carry the same version, one patch above `0.1.39`, so `0.1.40`. List each file and its version.

- [x] Failing test: change only `plugin.json` to `0.1.40`; `scripts/test ./internal/plugincheck/ -run TestManifests` fails.
- [x] Code: set `0.1.40` in the other two files.
- [x] Run `scripts/test ./internal/plugincheck/ -run TestManifests` passes, commit.

## Polish

### Task 07: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] `UnsetRepoUser` in `internal/config/repo_user.go` keeps a comment that yaml ties to a removed key: move it to the next key left (or to the document when none is left), with a test for a header comment right above the removed first key and for a line comment on it.
- [x] `internal/cli/config_repo_test.go`: a test that `--repo --executor inline --unset style` exits with bad input and leaves `.acta.yaml` unchanged, so the CLI's own non-repo-key check is locked in.
- [x] `plugin/skills/setup/SKILL.md`: "something else than" becomes "something other than".
- [x] Commit: `polish: review notes for PLN-0118`

## Review notes

- Round 1 CLEAN on both axes, no BLOCKER; polish review CLEAN on both axes.
- Spec section 2 says the override block shows in the per-prompt text too; the plugin-conflict block it copies is session start only, so the block is session start only (omp gets it through hook session-start).
- The wizard has no separate summary screen before writing; each override answer is echoed as a line before Apply, then Apply prints one line per write and the closing commit line.
- Orchestrator rulings beyond the spec: a user value unset with no default shows as "yours: not set", and the wizard leaves out the "yours" choice for it; both are pinned by tests.
- The 7 rewritten "(repo)" assertions all compare a repo value that differs from the user's effective value, so they are stricter, not weaker.
- yaml.Marshal turns 2-space nested indents into 4 spaces in both SaveRepoUser and UnsetRepoUser; older than this plan.
- The no-config-set grader errs strict: a `config set --help` call would count as a write.

## Fix round 1

### Task 08: The rule to ask before touching .acta.yaml reaches the agent without the setup skill

**Files:**
- Modify: `internal/hook/hook.go`, `plugin/skills/setup/SKILL.md`
- Test: `internal/hook/hook_test.go`, `internal/plugincheck/skill_setup_test.go` (only if a required string moves)

**verify:** An agent asked to change a key the repo file overrides gets the ask-first rule from text it always loads (the session note), whether or not it opens the setup skill, and the setup skill is picked for a one-setting change. The `repo-override-ask` eval case passes. List the text each surface now carries.

Eval evidence (round 1, `scripts/eval --case repo-override-ask --keep-temp`): the agent never opened the setup skill; it ran `acta config show`, saw `build_executor: dispatch (repo; yours: not set)`, said "Needs changing" and ran `acta config set --repo --executor subagent`. Graders `no-config-set` and `repo-file-untouched` failed.

- [x] Failing test: `internal/hook/hook_test.go` asserts the override block also says to ask the user before changing any of those keys, because `.acta.yaml` is committed and shared with everyone who clones the repo, and to write nothing until the user picks keep, write or remove; run `scripts/test ./internal/hook/` and see it fail.
- [x] Code: extend the block sentence in `internal/hook/hook.go`; widen the setup skill `description` so it also covers a request to change one setting (stay under the description cap in `internal/plugincheck/budget_test.go`).
- [x] Run `scripts/test ./internal/hook/ ./internal/plugincheck/` passes, commit; the orchestrator reruns the eval case.

### Task 09: Review polish, round 2

**verify:** the NOTE below is applied, and nothing else changes.

- [x] `plugin/skills/setup/SKILL.md` description names the old trigger words again within the cap: "acta: Use on first run and when the user asks to change setup or one setting (chat language, style, tone, build executor, plan depth): runs acta doctor and writes the optional acta block in CLAUDE.md or AGENTS.md."
- [x] Commit: `polish: review notes for PLN-0118 round 2`
- Fix round 1 (Task 08): eval case repo-override-ask was red (0.33, twice) because the agent never opened the setup skill and wrote .acta.yaml; the ask-first rule moved into the session-note override block, and the case then scored 1.00 three times out of three. Round 2 CLEAN on both axes.
- Spec section 2 does not yet mention the ask-first sentence Task 08 added to the block; the code and this note are the record.
- Polish round 2 (Task 09) put the old trigger words back in the setup description. One polish reviewer marked the dropped "subagent models" as a BLOCKER; it gives no reproducible wrong result, so it was moved to debt instead of reverting, since the revert would drop five other trigger words.
- Land gates: scripts/test --full 20 ok, go vet and gofmt clean, scripts/eval 13 of 13 at 1.00.
