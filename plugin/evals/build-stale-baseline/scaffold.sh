#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI has to sit inside the
# workspace before the agent starts. This script runs as the user, not in the
# sandbox, so it can still find the CLI on PATH.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no voice file makes the session rules send the agent to
# /acta:setup before any other work, which spends turns this case needs.
# The eval home is a sealed throwaway, but a hand run in a normal shell would
# overwrite the voice file on a real machine, so only write it when it is
# missing. Inside the eval it is always missing, so the run still gets one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin src tests .acta/specs .acta/plans
# The build runs git, and the system git does not work in the Bash sandbox. It is
# a stub that writes a cache file to a folder the sandbox locks. The real binary
# does not need that folder. A link would not do, because the sandbox cannot read
# the developer tools folder that the link points into. So the real git is copied.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
# acta starts git by name, and the system git may come first on PATH. So ./bin/acta
# is a small script that puts ./bin first and then starts the real binary. It holds
# the full folder name, because it cannot count on where it is run from.
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
# This suite must never run before the first task, so it fails loudly if it
# does. The task's own test is python3 -m unittest, not this file.
mkdir -p scripts
cat > scripts/test <<'TEST'
#!/bin/sh
echo "the saved suite belongs to landing, not build setup" >&2
exit 1
TEST
chmod +x scripts/test
# The spec says it is approved, which is what the build skill checks first.
cat > .acta/specs/2026-10-19-stale-design.md <<'SPEC'
---
id: SPC-0001
created: "2026-10-19 09:00:00"
hash: stalesp1
---
# Slug helper

Status: Bounded, approved by the user in chat on 2026-10-19.

A function slug(text) returns the text in lower case with spaces turned into
dashes.
SPEC
printf '# The suite is scripts/test. It belongs to landing.\n' > requirements.txt
# One task, depth minimal. The plan hash is stalepl1: the trailer must carry it.
cat > .acta/plans/2026-10-19-stale.md <<'PLAN'
---
parent: specs/2026-10-19-stale-design
depth: minimal
id: PLN-0001
created: "2026-10-19 09:10:00"
hash: stalepl1
---
# Slug Implementation Plan

**Goal:** slug(text) returns the text in lower case with spaces as dashes.

**Spec:** `.acta/specs/2026-10-19-stale-design.md`

**Tests:** fast `python3 -m unittest discover -s tests`; full `scripts/test`

### Task 1: Add slug

**Files:**
- Create: `src/slug.py`
- Test: `tests/test_slug.py`

**verify:** slug("Hello World") returns "hello-world".

- [ ] **Failing test:** `tests/test_slug.py` asserts slug("Hello World") == "hello-world"; fails because the module does not exist.
- [ ] **Change:** `src/slug.py` holds slug(text).
- [ ] **Commit:** `feat(slug): add the slug helper`
PLAN
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# An older landing saved this record for a tree that is not this one. It stays
# untracked in the Git common directory, the way a real record does, and it
# must not make setup run a suite or wait for an answer.
cat > .git/acta-baseline.md <<'BASE'
commit: 0000000000000000000000000000000000000000
tree: 0000000000000000000000000000000000000000
command: scripts/test
when: 2026-10-18T09:00:00Z
runtime: python3 3.12
exit: 0
log: /tmp/acta-baseline-old.log
summary: 1 old test passed
BASE
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
# The branch is the worktree of the build. It starts at main with no work done.
git checkout -q -b stale
