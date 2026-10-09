---
id: BUG-0039
hash: q24c7b4
status: wontfix
finished: "2026-10-09 21:25:54"
---
# acta commit fails when the file has nothing new to commit

## Symptom
`acta commit <path> -m "..."` on a planning file with no uncommitted changes prints `written, not committed: commit failed: git commit --only ... exit status 1:` with an empty reason and exits 2. Seen right after `acta id`, which already commits the new spec under `chore: assign short ids`; the skills then tell the agent to run `acta commit`, which fails.

## Root cause
internal/gitc/gitc.go:47: `Commit` runs `git commit --only -m <msg> -- <path>` even when the path matches HEAD. git exits 1 with "nothing to commit", and the error text is empty in the Reason. `CommitOrFold` (gitc.go:59) does not fold here because HEAD is a `chore:` id commit, so it falls through to `Commit`. Nothing checks whether the path is clean before committing. A clean path should report "nothing to commit" and exit 0.

## Repro
In a temp clone of the repo: `acta commit .acta/specs/<any committed spec>.md -m "chore(spec): x"; echo $?` prints the error above and `2`.

## Found in
main, while shaping SCR-0048 (SPC-0118): `acta id` committed the spec, then `acta commit` failed.
