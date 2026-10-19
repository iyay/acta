---
status: draft
id: SPC-0163
created: "2026-10-19 19:49:39"
hash: vhf4qdx
---

# Drop repo language and subagent models settings

## Why

Two user settings add choices nobody needs. Files in the repo are always
English. Each harness already picks its own subagent model, so acta should
not pick one.

## Design

1. Remove `repo_language`. Everything written to the repo is English.
   - Drop `RepoLanguage` from `internal/config/user.go`, `RepoKeys` in
     `internal/config/repo_user.go`, the `--repo-language` flag in
     `internal/cli/config_cmd.go`, and the field in `internal/setup/form.go`.
   - `internal/hook/hook.go` writes "English" in the two voice lines.
2. Remove `subagent_models`. No `split` mode.
   - Drop `SubagentModels`, its check, the `--subagent-models` and
     `--clear-subagent-models` flags, the setup form question, and
     `modelLines` plus the `Subagent models` session note line in `hook.go`.
3. Skills (`build`, `build/implementer-prompt.md`, `debug`, `review`,
   `review/depth.md`, `shape`, `shape/probe.md`, `slice`, `setup`) and
   `plugin/README.md`: the model rule becomes "name no model; the harness
   default picks it". This covers every subagent: workers, the quick
   reviewer, both deep reviewers, and the fact-check (which today follows
   the reviewer model). No skill keeps "own model, never lower". The
   report's first line still names its model.
4. Old config files that still hold either key load with no error and no
   message. The keys are ignored and never show in `acta config show` or
   its `--json`. Both writers drop them:
   - `SaveUser` / `SaveUserFile` (global `~/.acta/config.yaml`) drop
     `repo_language` and `subagent_models`.
   - `SaveRepoUser` (`.acta.yaml`) edits YAML nodes and keeps unknown
     keys, so removing `repo_language` from `RepoKeys` is not enough. It
     must delete a `repo_language` node on every write.
   Every other key and value in either file stays as it was.
5. Help text in `internal/cli/help.go` loses both settings. Eval scaffolds
   that write `repo_language` stop writing it.

## Testing

- Red test: a global config with `repo_language` and `subagent_models`
  loads clean, and neither key shows in `acta config show` or `--json`.
  This must fail today, since both keys show now.
- Red test: after `SaveUserFile`, the global file has neither key and every
  other setting is unchanged.
- Red test: a `.acta.yaml` with `repo_language` and other repo keys, after
  `SaveRepoUser` sets one key, has no `repo_language` and the other keys
  unchanged.
- Red test: the session note has no `Subagent models` line and names English
  for repo files.
- Existing tests that set either key are updated or removed.
- `internal/plugincheck` stays green after the skill edits.

## Out of scope

The user's own CLAUDE.md model rules. acta only stops adding its own.
