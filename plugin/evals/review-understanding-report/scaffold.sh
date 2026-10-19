#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
#
# The workspace holds the main checkout in ./repo and the finished branch in
# the worktree ./repo-retry-cap, so the agent can land the way a real build
# leaves it.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# The quick level and the understanding switch are what this case is about.
# Full history keeps the land to one plain merge. Only write the file when it
# is missing, so a hand run on a real machine never overwrites one. Inside the
# eval it is always missing.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nreview_depth: quick\ncommit_history: full\nreview_understanding: true\n' > "$HOME/.acta/config.yaml"
fi
ws="$PWD"
mkdir -p bin
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$ws" "$ws" > bin/acta
chmod +x bin/acta
# git merge starts a second git (git stash create) from bin/../libexec/git-core.
# With no git there it falls back to the system shim, which the sandbox
# blocks, and the merge dies with "stash failed". Point that folder at our copy.
mkdir -p libexec/git-core
ln -s ../../bin/git libexec/git-core/git
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$ws" "$PATH" > .claude/settings.json

mkdir -p repo/.acta/plans
cd repo
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The plan starts on main with open boxes, the way slice leaves it.
plan() {
  cat > .acta/plans/2026-10-18-retry-cap.md <<PLAN
---
depth: minimal
id: PLN-0001
created: "2026-10-18 10:00:00"
hash: evretry
---
# Retry Backoff Plan

**Goal:** A retry waits longer after each failure.

**Spec:** none

**Tests:** none, this repo has no test suite, no type checks and no formatter.

### Task 1: Add backoff

**Files:**
- Create: \`src/retry.py\`

**verify:** backoff(1) is 2 and backoff(3) is 8.

- [$1] Code: add backoff
- [$1] Commit: \`retry: add backoff\`
PLAN
}
plan " "
printf '# Retry\n' > README.md
# acta writes this file on its first run. Left untracked, it blocks
# git worktree remove at the end of the land, so it is committed up front.
printf 'state/\n' > .acta/.gitignore
git add -A
git commit -qm scaffold

git worktree add -q "$ws/repo-retry-cap" -b retry-cap
cd "$ws/repo-retry-cap"
mkdir -p src
# Task 1 as round 1 reviewed it: the wait doubles with no upper bound.
cat > src/retry.py <<'PY'
def backoff(attempt):
    """Seconds to wait before retry number attempt. The first retry is 1."""
    return 2 ** attempt
PY
git add src/retry.py
git commit -qm "retry: add backoff" -m "Task: PLN-evretry#1"
# Fix round 1 capped the wait. The landed code differs from what round 1 saw.
cat > src/retry.py <<'PY'
def backoff(attempt):
    """Seconds to wait before retry number attempt. The first retry is 1."""
    return min(2 ** attempt, 30)
PY
git add src/retry.py
git commit -qm "fix: cap backoff at 30 seconds" -m "Task: PLN-evretry#2"
# The polish is one comment, so it got the light review.
cat > src/retry.py <<'PY'
def backoff(attempt):
    """Seconds to wait before retry number attempt. The first retry is 1."""
    # A long outage would make the wait grow past any useful size, so stop at 30.
    return min(2 ** attempt, 30)
PY
git add src/retry.py
git commit -qm "polish: review notes for PLN-0001" -m "Task: PLN-evretry#3"
plan x
cat >> .acta/plans/2026-10-18-retry-cap.md <<'PLAN'

## Fix round 1

### Task 2: Cap the wait

**verify:** backoff(5) and every later retry give 30.

- [x] Cap the wait at 30 seconds
- [x] Commit: `fix: cap backoff at 30 seconds`

## Polish

### Task 3: Review polish

**verify:** every NOTE below is applied, and nothing else changes.

- [x] Say why the cap is there
- [x] Commit: `polish: review notes for PLN-0001`

## Review notes

- Round 1, quick, one reviewer: BLOCKED. backoff had no upper bound, so backoff(10) waited 1024 seconds.
- Round 2, standard (the fix round floor): CLEAN.
- Polish, light tier, one comment: CLEAN.
PLAN
git add .acta/plans/2026-10-18-retry-cap.md
git commit -qm "chore(plan): tick PLN-0001"
