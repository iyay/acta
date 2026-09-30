---
id: BUG-0022
hash: irsa4qh
---
# config set tells the user to fix the wrong file when an old config file is broken

## Symptom
The setting comes from an old file (~/.pm/voice.yaml, or ~/.acta/voice.yaml that could not be moved) and that file does not parse. `acta config set --language Korean` stops with the parse error and "fix or delete ~/.acta/config.yaml first". ~/.acta/config.yaml does not exist; the broken file is the old one. The parse error itself names the right file, but the next line points the user somewhere else.

## Root cause
internal/cli/config_cmd.go:92 prints `path`, which is config.UserPath(), not the file that failed. `config.ResolveUserFile` (added in PLN-0059) now returns the path it read and could give the right one.

## Repro
1. HOME=$(mktemp -d) PM_VOICE_FILE= ; mkdir -p $HOME/.pm
2. printf 'chat_language: [\n' > $HOME/.pm/voice.yaml
3. HOME=$HOME PM_VOICE_FILE= acta config set --language Korean
4. stderr ends with "fix or delete $HOME/.acta/config.yaml first".

## Found in
main, by the Standards reviewer in review round 1 of PLN-0059 (setup-herdr-config-path); code already on main before that plan.
