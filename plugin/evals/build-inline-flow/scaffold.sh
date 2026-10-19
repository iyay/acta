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
# The spec says it is approved, which is what the build skill checks first.
cat > .acta/specs/2026-10-19-slug-design.md <<'SPEC'
---
id: SPC-0001
created: "2026-10-19 09:00:00"
hash: slugsp1
---
# Slug helper

Status: Bounded, approved by the user in chat on 2026-10-19.

A function slug(text) returns the text in lower case with spaces turned into
dashes.
SPEC
# The project needs no third-party package: its tests use the standard library.
# The file still exists, so a build that installs by habit shows up in the trace.
printf '# No third-party packages. Tests use unittest from the standard library.\n' > requirements.txt
# One task, depth minimal. The plan hash is slugpl1: the trailer must carry it.
cat > .acta/plans/2026-10-19-slug.md <<'PLAN'
---
parent: specs/2026-10-19-slug-design
depth: minimal
id: PLN-0001
created: "2026-10-19 09:10:00"
hash: slugpl1
---
# Slug Implementation Plan

**Goal:** slug(text) returns the text in lower case with spaces as dashes.

**Spec:** `.acta/specs/2026-10-19-slug-design.md`

**Tests:** fast `python3 -m unittest discover -s tests`

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
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
# The branch is the worktree of the build. It starts at main with no work done.
git checkout -q -b slug
