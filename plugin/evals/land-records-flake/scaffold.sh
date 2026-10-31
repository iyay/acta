#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
#
# The workspace holds the main checkout in ./repo and the finished branch in
# the worktree ./repo-land, so the agent can land the way a real build leaves it.
# The full suite passes but prints one FLAKY line, the way a test that failed
# and then passed on a rerun shows up. The case checks that the land records
# that flake in the worktree before it merges.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# Full history keeps the land to one plain merge. Only write the file when it
# is missing, so a hand run on a real machine never overwrites one. Inside the
# eval it is always missing.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\ncommit_history: full\n' > "$HOME/.acta/config.yaml"
fi
ws="$PWD"
mkdir -p bin
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git.real
cp "$(command -v acta)" bin/acta.real
# ./bin/git and ./bin/acta run the real binaries and log each call to
# ./calls.log with its exit code and the checkout it ran in, so the grader
# sees where the flake was recorded and whether that came before the merge.
# The git calls acta makes itself are not logged: ACTA_INNER marks them.
printf '#!/bin/sh\nws="%s"\n' "$ws" > bin/git
cat >> bin/git <<'WRAP'
"$ws/bin/git.real" "$@"
rc=$?
if [ -z "$ACTA_INNER" ]; then
  dir=.
  [ "$1" = "-C" ] && dir="$2"
  top=$(cd "$dir" 2>/dev/null && "$ws/bin/git.real" rev-parse --show-toplevel 2>/dev/null)
  printf 'git %s exit=%s in %s\n' "$(printf '%s' "$*" | tr '\n' ' ')" "$rc" "${top##*/}" >> "$ws/calls.log"
fi
exit "$rc"
WRAP
printf '#!/bin/sh\nws="%s"\n' "$ws" > bin/acta
cat >> bin/acta <<'WRAP'
PATH="$ws/bin:$PATH"
export PATH
ACTA_INNER=1 "$ws/bin/acta.real" "$@"
rc=$?
top=$("$ws/bin/git.real" rev-parse --show-toplevel 2>/dev/null)
printf 'acta %s exit=%s in %s\n' "$(printf '%s' "$*" | tr '\n' ' ')" "$rc" "${top##*/}" >> "$ws/calls.log"
exit "$rc"
WRAP
chmod +x bin/git bin/acta
# git merge starts a second git (git stash create) from bin/../libexec/git-core.
# With no git there it falls back to the system shim, which the sandbox
# blocks, and the merge dies with "stash failed". Point that folder at our copy.
mkdir -p libexec/git-core
ln -s ../../bin/git libexec/git-core/git
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$ws" "$PATH" > .claude/settings.json
echo "log start" > calls.log

mkdir -p repo
cd repo
bin_git="$ws/bin/git.real"
"$bin_git" init -q -b main .
"$bin_git" config user.email eval@example.invalid
"$bin_git" config user.name eval
mkdir -p .acta/plans scripts
cat > .acta/plans/2026-10-31-land.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-31 09:00:00"
hash: landrf1
---
# Calc Plan

**Goal:** add(1, 9) returns 10.

**Spec:** none

**Tests:** `scripts/test --full`

### Task 1: Add add

**Files:**
- Create: `src/calc.py`

**verify:** add(1, 9) is 10.

- [x] Add add
- [x] Commit: `feat(calc): add add`

## Review notes

- Round 1, standard: CLEAN.
PLAN
# The suite passes, but one test only passed on its rerun. It prints the
# FLAKY line a real test command prints for that, and still exits 0.
cat > scripts/test <<'TEST'
#!/bin/sh
echo "--- FAIL: TestAddCarries (0.01s)"
echo "ok 3 tests"
echo "FLAKY: internal/calc.TestAddCarries (failed 1 of 2)"
exit 0
TEST
chmod +x scripts/test
printf '# Calc\n' > README.md
# acta writes this file on its first run. Left untracked, it blocks
# git worktree remove at the end of the land, so it is committed up front.
printf 'state/\n' > .acta/.gitignore
"$bin_git" add -A
"$bin_git" commit -qm scaffold

"$bin_git" worktree add -q "$ws/repo-land" -b land
cd "$ws/repo-land"
mkdir -p src
printf '# add: the sum of two numbers.\ndef add(a, b):\n    return a + b\n' > src/calc.py
"$bin_git" add src/calc.py
"$bin_git" commit -qm "feat(calc): add add" -m "Task: PLN-landrf1#1"
