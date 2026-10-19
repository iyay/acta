#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
#
# The workspace holds the main checkout in ./repo and the finished branch in
# the worktree ./repo-land, so the agent can land the way a real build leaves it.
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

mkdir -p repo
cd repo
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
mkdir -p .acta/plans scripts
cat > .acta/plans/2026-10-19-land.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-10-19 09:00:00"
hash: landbl1
---
# Parser Plan

**Goal:** parse("1") returns 1.

**Spec:** none

**Tests:** `scripts/test --full`

### Task 1: Add parse

**Files:**
- Create: `src/parse.py`

**verify:** parse("1") is 1.

- [x] Add parse
- [x] Commit: `feat(parse): add parse`

## Review notes

- Round 1, standard: CLEAN.
PLAN
# The suite writes to the Git common directory, so the run count survives the
# worktree cleanup and a second run shows up after the land.
cat > scripts/test <<'TEST'
#!/bin/sh
common="$(git rev-parse --path-format=absolute --git-common-dir)"
log="$common/test-runs.log"
if [ -f "$log" ]; then
  n=$(wc -l < "$log")
else
  n=0
fi
printf 'run %s\n' "$((n + 1))" >> "$log"
echo "ok 3 tests"
exit 0
TEST
chmod +x scripts/test
printf '# Parser\n' > README.md
# acta writes this file on its first run. Left untracked, it blocks
# git worktree remove at the end of the land, so it is committed up front.
printf 'state/\n' > .acta/.gitignore
git add -A
git commit -qm scaffold

git worktree add -q "$ws/repo-land" -b land
cd "$ws/repo-land"
mkdir -p src
printf '# parse: turn the text 1 into the number 1.\ndef parse(text):\n    return int(text)\n' > src/parse.py
git add src/parse.py
git commit -qm "feat(parse): add parse" -m "Task: PLN-landbl1#1"
