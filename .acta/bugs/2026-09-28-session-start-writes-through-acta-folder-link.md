---
id: BUG-0004
hash: a6sr3z9
fixed_in: e7507a8
---
# Session start writes a .gitignore outside the repo when the .acta folder is a symlink

## Symptom
A repo that commits `.acta` as a symlink to a folder outside the repo (for example `.acta -> ../outside`) makes `acta hook session-start` and `acta tick` create or append `outside/.gitignore` with `.agents.json`. Nothing is shown. Only a file named `.gitignore` can be written, never an arbitrary file.

## Root cause
`internal/hook/gitignore.go` (`EnsureGitignore`): `os.Stat(root)` follows the link, and `inGitRepo` walks up the unresolved path, so it finds the repo's `.git` and treats the outside folder as inside. The PLAN-20 `Lstat` check only covers the last path part (`.gitignore`). `acta doctor` already guards this with `inRepoPath` in `internal/doctor/doctor.go`; the hook does not.

## Repro
In a temp git repo: `mkdir ../outside && ln -s ../outside .acta && git add -A && git commit -m x`, then `acta hook session-start`. `../outside/.gitignore` now holds `.agents.json`.

## Found in
Branch gitignore-link, PLAN-20 review round 1 (both reviewers). The code was already on main before PLAN-20.
