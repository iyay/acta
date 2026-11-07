---
parent: debt/2026-11-07-debug-temp-copy-check
closes: [DBT-0159.01, DBT-0159.02, DBT-0159.03]
depth: minimal
id: PLN-0220
created: "2026-11-07 20:10:37"
hash: dj0dlyj
---
# Probe copy hardening

**Goal:** `probe-copy.sh` refuses a planted fallback folder, copies the whole work tree from any subfolder, and gives the copy its own git so git there never reaches another repo.

**Spec:** `.acta/specs/2026-11-07-probe-copy-hardening-design.md` (SPC-0194)

**Tests:** fast `scripts/test ./internal/plugincheck/`; full `scripts/test --full` (land only).

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- `plugin/skills/debug/SKILL.md` does not change, so no eval runs for this plan.
- Never `rm -rf`. Tests write only under `t.TempDir()`.
- Comments in plain English, short words, saying why.

## Waves

- Wave 1: Task 1.

### Task 1: Harden probe-copy.sh

**Files:** modify `plugin/skills/debug/probe-copy.sh`, `internal/plugincheck/probe_copy_test.go`.
**Seam:** `scripts/test ./internal/plugincheck/ -run TestProbeCopy`.
**verify:** No run of the script can copy into a folder it did not just make, copy less than the whole work tree it was run inside, or leave a copy where git finds any repo other than the copy's own; list each path (fallback name planted as a folder or a symlink, run from root or subfolder, source from `git rev-parse` or `pwd -P`, TMPDIR plain or inside another repo) and what the script does. Every failure still prints nothing and exits non-zero.
- [ ] Red: `TestProbeCopy` fails because the three new cases (planted fallback folder, run from a subfolder, TMPDIR inside another repo) and the changed `.git` rule do not hold for the current script.
- [ ] Change: use `mkdir -m 700` without `-p` for the fallback, take the source from `git rev-parse --show-toplevel` else `pwd -P`, and run `git init -q` in the copy only when the source came from git, with the header comment updated.
- [ ] Commit: `fix(debug): harden probe-copy.sh fallback, source and git`.
