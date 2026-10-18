#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
#
# The workspace holds the main checkout in ./repo and the finished branch in
# the worktree ./repo-slug, so the agent can land the way a real build leaves
# it.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# The deep level with the understanding switch set to false is what this case
# is about: deep must not turn the question on. Full history keeps the land to
# one plain merge. Only write the file when it is missing, so a hand run on a
# real machine never overwrites one. Inside the eval it is always missing.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nrepo_language: English\nreview_depth: deep\ncommit_history: full\nreview_understanding: false\n' > "$HOME/.acta/config.yaml"
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
  cat > .acta/plans/2026-10-18-slug.md <<PLAN
---
depth: minimal
id: PLN-0001
created: "2026-10-18 10:00:00"
hash: evslugx
---
# Slug Plan

**Goal:** Turn a title into a lower-case slug joined by dashes.

**Spec:** none

**Tests:** none, this repo has no test suite, no type checks and no formatter.

### Task 1: Add slug

**Files:**
- Create: \`src/slug.py\`

**verify:** slug("Hello  World") is "hello-world".

- [$1] Code: add slug
- [$1] Commit: \`slug: add the helper\`
PLAN
}
plan " "
printf '# Slug\n' > README.md
# acta writes this file on its first run. Left untracked, it blocks
# git worktree remove at the end of the land, so it is committed up front.
printf 'state/\n' > .acta/.gitignore
git add -A
git commit -qm scaffold

git worktree add -q "$ws/repo-slug" -b slug
cd "$ws/repo-slug"
mkdir -p src
cat > src/slug.py <<'PY'
def slug(title):
    # split() with no argument drops runs of spaces, so no dash doubles up.
    return "-".join(title.lower().split())
PY
git add src/slug.py
git commit -qm "slug: add the helper" -m "Task: PLN-evslugx#1"
plan x
cat >> .acta/plans/2026-10-18-slug.md <<'PLAN'

## Review notes

- Round 1, deep, two reviewers, no BLOCKER so no fact-check: CLEAN.
PLAN
git add .acta/plans/2026-10-18-slug.md
git commit -qm "chore(plan): tick PLN-0001"
