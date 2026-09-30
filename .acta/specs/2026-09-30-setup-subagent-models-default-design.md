---
parent: bugs/2026-09-30-setup-asks-subagent-models-every-run
created: "2026-09-30"
id: SPC-0045
hash: v5iresj
started: "2026-09-30"
finished: "2026-09-30"
---
# Setup remembers a no to split subagent models; user config moves into internal/config

Status: design approved by the user in chat on 2026-09-30. Bounded (the existing `subagent_models` setting, its check, the `config set` flag, one line of the setup skill, and a package move with no change in behaviour).

## Why

`acta:setup` asks "split subagent models?" on every run. A no saves nothing, so the field stays empty. Empty also means "never asked", and the first-run rule asks for every empty field. So the question comes back each time.

The user setting already lives in `~/.acta/config.yaml` and the command is `acta config`, but the Go package is still `internal/voice`. The user asked to fold it into `internal/config`, which today holds the repo setting (`.acta.yaml`, planning dirs, links).

## Part 1: package move

- `internal/voice/voice.go` becomes `internal/config/user.go`. Its tests become `internal/config/user_test.go` and `internal/config/user_acta_test.go`. `internal/voice/` is removed.
- Names that would clash with `internal/config`, or read badly there, are renamed:

  | old | new |
  |---|---|
  | `voice.Voice` | `config.User` |
  | `voice.Default` | `config.UserDefault` |
  | `voice.Path` | `config.UserPath` |
  | `voice.Resolve` | `config.ResolveUser` |
  | `voice.SaveResolved` | `config.SaveUser` |
  | `voice.Load` | `config.LoadUser` |
  | `voice.Save` | `config.SaveUserFile` |
  | `voice.ErrBad` | `config.ErrBadUser` |

  The `Validate` method and the unexported helpers (`voicePath`, `oldPath`, `fill`) keep their names; none clash.
- Every importer switches to `internal/config`: `internal/cli` (`cli.go`, `config_cmd.go`, `doctor.go`, `hook.go`), `internal/doctor` (`doctor.go`, `doctor_test.go`), `internal/hook` (`hook.go`, `hook_test.go`), `internal/plugincheck/no_old_names_test.go`.
- No behaviour change: same file paths, same `PM_VOICE_FILE` variable, same YAML keys, same error text.
- `CLAUDE.md` Architecture: drop the `internal/voice` line, and say `internal/config` holds both the repo setting and the user setting.

## Part 2: bug fix

- `subagent_models` takes a second value, `default`. It means "pick no model; the user's own config wins". It acts the same as empty, but it records that the user answered.
- `internal/config/user.go`: the check accepts `""`, `split` and `default`. Any other value is still an error, and the message names both values.
- `internal/cli/config_cmd.go`: the flag help and `configUsage` say `--subagent-models split|default`. `config show` prints `subagent_models: default` like any other set value.
- `plugin/skills/setup/SKILL.md`: a no saves `acta config set --subagent-models default`. The question is asked only while the field is empty.
- The other skills (brainstorm, build, plan, review, debug, dispatch) stay as they are. They act only on `split`, so `default` already falls through to the user's own config.

Part 1 goes first, so Part 2 edits the file only in its new place.

## Tests

- Package move: the existing voice tests pass unchanged apart from names; `go build ./...` and `go vet ./...` are clean; no `internal/voice` import is left.
- user config: `default` passes the check; a value like `mixed` fails.
- cli: `acta config set --subagent-models default` writes the value, and `acta config show` prints it.
- plugincheck: the setup skill names `--subagent-models default` for the no answer.

## Out of scope

A user whose field is empty today gets asked once more. After that answer the question stops. No migration. The `PM_VOICE_FILE` variable keeps its name.
