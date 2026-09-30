---
created: "2026-09-30 23:21:23"
parent: bugs/2026-09-30-config-set-names-wrong-broken-file
id: SPC-0051
hash: xzu47ti
---
# config set names the broken file it read

Status: design approved by the user in chat on 2026-09-30. Bounded: the `set` case of `acta config` in `internal/cli/config_cmd.go` exists.

## Why

When the setting comes from an old file (`~/.pm/voice.yaml`, or a `~/.acta/voice.yaml` that could not be moved) and that file does not parse, `acta config set` prints the parse error and then `fix or delete ~/.acta/config.yaml first`. That file does not exist. The line prints `config.UserPath()`, not the file that failed. Reproduced on 2026-09-30 with a temp HOME.

## Design

1. The `set` case calls `config.ResolveUserFile()` instead of `config.ResolveUser()`.
2. On a read error, the `fix or delete <path> first` line names the path `ResolveUserFile` returned: the file that failed to parse. When that file is `~/.acta/config.yaml`, the output is the same as today.
3. Nothing else changes: `set` still writes only to `UserPath()`, and still never overwrites a broken file.

## Testing

`internal/cli/cli_test.go`: under a temp HOME with `PM_VOICE_FILE=""` and a broken `~/.pm/voice.yaml`, `config set --language Korean` exits 1, stderr names `~/.pm/voice.yaml` in the fix line and never names `~/.acta/config.yaml`, and `~/.acta/config.yaml` is still not created.
