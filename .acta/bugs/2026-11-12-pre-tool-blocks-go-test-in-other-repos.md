---
id: BUG-0088
hash: hw1kefl
priority: low
---
# pre-tool blocks go test in a repo that has no scripts/test

## Symptom
From an acta session, `cd ~/.wwt/skills/<worktree>/walkie/meow && go test ./...` is blocked with "acta: run tests with scripts/test, not go test", although the walkie repo has no scripts/test. The same block fires on any command whose text only mentions those words, such as an `acta bug new` heredoc or a `walkie send` message.

## Root cause
`internal/hook/gotest.go` `GoTestBlock(dir, command)` checks for scripts/test in the hook's working folder (the session's repo), not in the folder the command `cd`s into first. It also matches the words inside heredocs and quoted message text.

## Repro
In an acta session whose cwd is the acta repo, run a Bash command that starts with `cd <another repo without scripts/test> && go test ./...`.

## Found in
Land gate of walkie PLN-0285 on 2026-11-12, run from the acta session.
