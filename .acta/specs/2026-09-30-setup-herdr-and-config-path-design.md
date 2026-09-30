---
created: "2026-09-30 22:47:34"
parent: debt/2026-09-30-build-owns-dispatch
id: SPC-0050
hash: dyyudug
---
# Setup offers dispatch only inside herdr, and config show names the file it read

Status: design approved by the user in chat on 2026-09-30 ("fix ini"). Bounded: both flows exist, `plugin/skills/setup/SKILL.md` and `acta config show` in `internal/cli/config_cmd.go`.

## Why

1. Since PLN-0058, build runs `dispatch` only when `HERDR_ENV=1` is set (this session is inside a herdr pane). `plugin/skills/setup/SKILL.md:29` still offers `dispatch` when `herdr` is on PATH alone. A user in a plain terminal can pick `dispatch` in setup, and every build then quietly runs as `subagent`. This is DBT-0051.15.
2. `acta config show` prints `file: ~/.acta/config.yaml (exists: true)` even when that file is missing and the values came from an older file. `internal/cli/config_cmd.go` prints `config.UserPath()`, but `config.ResolveUser()` can read `~/.acta/voice.yaml` (when the move fails) or `~/.pm/voice.yaml`. Seen on 2026-09-30: the user moved `~/.acta/config.yaml` away, and `show` still named it and printed the values of `~/.pm/voice.yaml`.

## Design

1. **Setup.** The executor question offers `dispatch` only when `HERDR_ENV=1` is in the environment. `herdr` on PATH alone is not enough, because dispatch needs this session's own pane. Without `HERDR_ENV=1`, setup does not offer `dispatch` at all.
2. **Config path.** A new `config.ResolveUserFile() (User, string, bool, error)` does what `ResolveUser` does and also returns the path of the file it read. With no file at all, the path is `UserPath()` and exists is false. `ResolveUser` calls it and drops the path, so its other callers do not change.
3. **Show.** `acta config show` prints the path `ResolveUserFile` returns. When that path is not `UserPath()`, the line also says where writes go: `file: <read path> (exists: true, old file; config set writes <UserPath>)`. `--json` puts the read path in `path` and adds `writes` with `UserPath()`. Nothing in the repo reads these fields today (checked 2026-09-30).
4. `acta config set` does not change: it still writes to `UserPath()` only.

## Testing

- `internal/plugincheck/skill_setup_test.go`: the new setup sentence is required; the old `or `herdr` on PATH` wording is refused.
- `internal/config`: `ResolveUserFile` returns the right path for each case: config.yaml, voice.yaml moved, voice.yaml that cannot move, `~/.pm/voice.yaml`, no file, `PM_VOICE_FILE`.
- `internal/cli`: with only `~/.pm/voice.yaml` under a temp HOME, `config show` names that file and the `writes` path, in text and in JSON.
