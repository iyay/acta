---
parent: debt/2026-09-28-skill-rules
depth: minimal
closes: [SPC-0093, DBT-0011.16, DBT-0011.24, DBT-0011.30, DBT-0011.31]
id: PLN-0102
created: "2026-10-06 08:47:02"
hash: hgvqno6
started: "2026-10-06 08:49:40"
finished: "2026-10-06 09:00:55"
---
# doctor --fix refusals Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** `acta doctor --fix` never writes into `.git`, through a symlinked root, or onto a `.gitignore` that is not a regular file, and the doctor report always says the same thing the disk shows.

**Spec:** .acta/specs/2026-10-06-doctor-fix-refusals.md

**Tests:** `scripts/test ./internal/doctor/ ./internal/cli/` (fast), `scripts/test --full` (full)

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- One function, `refusal(e Env) string` in `internal/doctor/doctor.go`, decides; `Fix` and `checkRepo` both call it and never repeat its rules.
- When `refusal` gives a reason: `Fix` returns `nil, nil` and writes nothing; `checkRepo` returns `Fail` with that reason as `Msg` and a manual `Fix` line that never says `acta doctor --fix`.
- Every test builds its repo under `t.TempDir()`; never touch the real repo or HOME.
- Run tests with `scripts/test`, never bare `go test`.

## Waves

- Wave 1: Task 01, Task 04
- Wave 2: Task 02
- Wave 3: Task 03
- Wave 4: Task 05

Tasks 01 to 03 all edit `internal/doctor/doctor.go` and `doctor_test.go`, so they run one after another. Task 04 touches only `internal/cli/`.

### Task 01: refusal() and the .git case

**Files:** `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`

**verify:** No acta root that is the repo's `.git` folder or inside it can be written by `Fix`, and `checkRepo` never reports ok or suggests `--fix` for one. The outside-the-repo refusal keeps working through the same function. List every root shape tested (`.git`, `.git/sub`, outside, normal).

- [x] Red: table test with roots `<repo>/.git` and `<repo>/.git/acta`: after `Fix`, no file under the temp repo changed (compare a file listing with sizes before and after), and `checkRepo` is `Fail` with a `Fix` line that does not contain `acta doctor --fix`.
- [x] Green: add `refusal(e Env) string`; move the outside-the-repo rule and the `.gitignore`-is-a-link rule into it; add the `.git` rule (real path of the root equal to or under the real path of `<RepoRoot>/.git`); `Fix` and `checkRepo` call it first.
- [x] Commit: `fix(doctor): --fix never writes inside .git (DBT-0011.30)`

### Task 02: no symlink in the root's path

**Files:** `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`

**verify:** No acta root whose path below the repo root passes through a symlink, whether the link's target is inside the repo, outside it or missing, can be written by `Fix`, and `checkRepo` never suggests `--fix` for one. List every link shape tested.

- [x] Red: cases `.acta` as a link to another in-repo folder that holds a tracked `.gitignore`, `.acta` as a dangling link, and `sub/.acta` where `sub` is the link; each must leave every file unchanged and give `Fail` with no `acta doctor --fix`.
- [x] Green: in `refusal`, walk each part of the root's path from the repo root down with `os.Lstat` and refuse on the first symlink, naming it ("<path> is a symlink; replace it with a real folder").
- [x] Commit: `fix(doctor): --fix refuses an acta root that goes through a symlink (DBT-0011.31, DBT-0011.24)`

### Task 03: .gitignore must be a regular file

**Files:** `internal/doctor/doctor.go`, `internal/doctor/doctor_test.go`

**verify:** `Fix` writes the `.gitignore` only when it is missing or a regular file; for any other kind (folder, link) it writes nothing and `checkRepo` fails with no `--fix` hint. List every file kind tested.

- [x] Red: `.acta/.gitignore` as a folder; `Fix` must return `nil, nil` with nothing changed (today it returns an error), and `checkRepo` must fail with no `acta doctor --fix`.
- [x] Green: in `refusal`, `os.Lstat` the `.gitignore`; when it exists and `!fi.Mode().IsRegular()`, refuse ("<path> is not a regular file").
- [x] Commit: `fix(doctor): --fix refuses a .gitignore that is not a regular file (DBT-0011.24)`

### Task 04: unreadable git status means no commit

**Files:** `internal/cli/doctor.go`, `internal/cli/doctor_test.go`

**verify:** `acta doctor --fix` never commits when it cannot read the file's git status, and always says why. List every path through the commit decision (auto_commit off, dirty, status error, clean).

- [x] Red: make `gitc.IsDirty` fail for a fixable repo (for example a `.git` file that is not a real git dir, or a broken index), run `cmdDoctor([]string{"--fix"}, ...)` from that folder, and assert stderr contains `fixed, not committed: cannot read git status` and no commit was made.
- [x] Green: keep the `IsDirty` error; pass it into `skipReason` (or check it right before) so it gives `cannot read git status: <err>`.
- [x] Commit: `fix(doctor): --fix does not commit when git status cannot be read (DBT-0011.16)`

### Task 05: Version 0.1.25

**Files:** `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`

**verify:** The three files carry the same version, 0.1.25, and `internal/plugincheck` passes. List each file and the version it holds.

- [x] Red: none needed; `scripts/test ./internal/plugincheck/` guards that the three agree.
- [x] Green: bump the patch from 0.1.24 to 0.1.25 in all three files.
- [x] Commit: `chore: version 0.1.25`

## Polish

### Task 06: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] `internal/doctor/doctor.go` checkRepo: a `.gitignore` that exists but is not a regular file (folder, fifo) gets its own fix line, "replace <path> with a regular file", not the "replace the .gitignore link" line; add a test assert on that fix text in the existing folder test.
- [x] Commit: `polish: review notes for PLN-0102`

## Review notes

- A .acta link that points out of the repo is reported as outside the repo, with "fix root in .acta.yaml" rather than "replace the link"; both sides still refuse.
- The non-regular .gitignore fix line rebuilds the same filepath.Join that refusal does; matches the block's style.
