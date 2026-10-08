---
parent: scratch/2026-10-07-repo-override-invisible
id: SPC-0109
created: "2026-10-08 07:41:02"
hash: n8isf8t
started: "2026-10-08 07:45:45"
---
# Repo overrides are visible and changeable from setup, the session note and chat

Status: Architectural, every section approved by the user in chat on 2026-10-08 (log in SCR-0055).

## Why

A repo's `.acta.yaml` can override five of the user's own settings (`repo_language`, `build_executor`, `plan_depth`, `commit_history`, `coding_guide`). Today only `acta config show` hints at it, with a bare `(repo)` mark. The setup wizard never reads `.acta.yaml`, the setup skill never mentions it, and there is no way to remove a key from it. On 2026-10-07 a user picked `subagent` in setup while the repo said `dispatch`; the build went to omp and the user could not tell why.

## Goal

Wherever a user sets or reads a setting, a repo value that differs from the user's own value is named next to it, and the user can keep it, write their value into the repo, or remove the repo key. `.acta.yaml` is shared with everyone who clones the repo, so nothing writes it without the user choosing to, and nothing commits it.

## Design

### 1. Config and CLI

- `internal/config` gets `Overrides(user User, repoRoot string) ([]Override, error)`, next to `MergeRepo`. `Override` is `{Key, Repo, Yours string}`. It lists only repo keys whose `.acta.yaml` value differs from the user's effective value, where the effective value includes the default when the user never set the key (an unset `plan_depth` is `full`). It reads the same file the same way as `MergeRepo` and returns the same errors.
- `internal/config` gets `UnsetRepoUser(repoRoot string, keys []string) (string, error)`. It removes the keys from `.acta.yaml`, keeps every other line and the key order the way `SaveRepoUser` does, and leaves an empty file when nothing is left. A key that is not in the file is no error. A key that is not one of the five repo keys is refused with an error naming the five.
- `acta config set --repo --unset <key>` calls it. The flag repeats, and can sit next to set flags in one call. The same key set and unset in one call is refused. `--unset` without `--repo` is refused. It prints the path and does not commit, like `set --repo` today.
- `acta config show` prints `<repo value> (repo; yours: <user value>)` for a key in the override list, and `(repo)` as today for a repo key equal to the user's value. `--json` adds `"overrides": [{"key", "repo", "yours"}]` and keeps `from_repo`.

### 2. Session note

- `internal/hook`'s input gets `Overrides []config.Override`, filled from `config.Overrides` where `RepoErr` is filled today.
- When the list is not empty, the note adds one short English block: it names each key, the repo value and the user's value in one sentence, tells the agent to tell the user once at the start, and names `acta config set --repo --unset <key>` as the way back to the user's value.
- An empty list adds nothing. The block shows wherever the plugin-conflict block shows (session start and the per-prompt text), so omp gets it too.
- When `RepoErr` is set, the override block is left out; the existing `RepoErr` line covers that case.

### 3. Setup wizard

- After the normal questions, the wizard runs `config.Overrides` with the values the user just chose, not the old file.
- For each key in the list it asks one select, in English like the rest of the wizard: keep (the repo keeps its value), yours (write the user's value into `.acta.yaml`), remove (drop the key, the user's value applies). The default is keep.
- No override question when the list is empty or when the wizard runs outside a git repo (`Env.RepoRoot` empty).
- Two new plan actions, set repo key and unset repo key, call `SaveRepoUser` and `UnsetRepoUser`. They show in the summary before anything is written. Nothing is committed; when `.acta.yaml` changed, the wizard ends with one line saying so and that the user commits it when the team should get it.
- A `.acta.yaml` that cannot be read skips the override step with one line naming the error; the other setup steps go on.

### 4. Setup skill

- `plugin/skills/setup/SKILL.md` gains plain text, with no repo paths and no user names: on an ask to change a setting, run `acta config show --json` and look at `overrides`. A key with no override goes to `acta config set` as today. A key with an override gets one question that says `.acta.yaml` is committed and shared with everyone who clones the repo, with three choices: keep, write the new value with `acta config set --repo --<flag>`, or remove the key with `acta config set --repo --unset <key>`. After a write, say the file changed and is not committed.
- On a question like "why did the build go to X", answer from the `(repo; yours: ...)` line of `acta config show`.
- `internal/plugincheck` caps for the setup skill (`MaxLines` in `skill_setup_test.go`, the byte cap in `budget_test.go`) are raised in the same plan when the text needs it, and the skill's required strings gain `--unset` and `overrides`.
- New eval case `repo-override-ask`, listed in `evalCases` against the new skill phrase. Its scaffold makes a repo whose `.acta.yaml` holds `build_executor: dispatch`, and the prompt asks to use subagent for builds from now on. Graders: the reply names `.acta.yaml`; no Bash call matching `config set` runs; `.acta.yaml` still holds `build_executor: dispatch`.

## Out of scope

- Auto-committing `.acta.yaml`.
- `--unset` for the user's own file.
- Any new repo key, or moving a personal key into the repo file.

## Tests

- `Overrides`: a table over same value, different value, user unset (default counts), no file, unreadable file.
- `UnsetRepoUser` and `--unset`: removes the key and keeps the other lines, missing key is fine, non-repo key refused, mixed with a set, same key set and unset refused, refused without `--repo`.
- `config show`: the text line and the `--json` `overrides` field, for a differing and an equal key.
- Hook: the block names key, repo value and user value; nothing without overrides; no block when `RepoErr` is set.
- Wizard: no question with no difference or outside a repo; keep, yours and remove each give the right action; the default changes nothing.
- All of these run in temp repos and a temp home, never this repo's `.acta.yaml` or the real home.
- The new eval case runs in the land gate, since the skill changes.

The last task adds 1 to the patch version in the three plugin files.
