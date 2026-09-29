---
id: BUG-0003
hash: q0zm021
fixed_in: 739178e
---
# Session start writes a file outside the repo when .acta/.gitignore is a symlink

## Symptom
A repo that commits `.acta/.gitignore` as a symlink makes acta write `.agents.json` into the link target on every session start and on `acta tick`. The target can be outside the repo (a new file is created when it is missing) or a tracked file inside it, such as `.git/config`. The error is dropped, so nothing is shown.

## Root cause
`internal/hook/gitignore.go:16-35` (`EnsureGitignore`) reads and then calls `os.WriteFile` on `<root>/.gitignore` without an `os.Lstat` check, so it follows the link. Callers `internal/cli/hook.go:45` and `internal/cli/tick.go:115` ignore its error. The fix belongs in `EnsureGitignore` itself so every caller gets it. `acta doctor` already refuses links (`internal/doctor/doctor.go` `isLink`, PLAN-18 Task 10).

## Repro
In a temp git repo: `mkdir outside .acta && ln -s ../../outside/new.conf .acta/.gitignore && git add -A && git commit -m x`, then `acta hook session-start`. `outside/new.conf` now exists with `.agents.json`.

## Found in
Branch skill-rules, PLAN-18 review round 3 and the extra fix round (both reviewers). The code was already on main before PLAN-18.
