---
id: BUG-0041
hash: oizbc71
priority: low
---
# acta commit on a file with nothing new fails with an empty reason

## Symptom
`acta commit <path> -m "<message>"` on a planning file that has no new changes prints `written, not committed: commit failed: git commit --only -m <message> -- <path>: exit status 1: ` and exits 2. The text after the last colon is empty, so the user cannot tell why it failed. It happens in the usual shape flow: `acta id` already commits the new spec, so the `acta commit` that follows has nothing left to commit.

## Root cause
`gitc.Commit` (`internal/gitc/gitc.go:47`) runs `git commit --only` even when the file has no changes. Git exits 1 and writes its reason ("nothing to commit" and the status) to stdout, not stderr. `run` (`internal/gitc/gitc.go:477`) puts only stderr into the error, so the reason is empty.

## Repro
1. Write a new spec in `.acta/specs/`.
2. Run `acta id`. It commits the spec as `chore(spec): assign short ids`.
3. Run `acta commit .acta/specs/<file>.md -m "chore(spec): x"`. It prints the empty reason and exits 2.

Seen on main at 5012b47 with `.acta/specs/2026-10-09-instant-tui-startup.md`.

## Found in
main, while reading the output of the shape flow for SPC-0119.
