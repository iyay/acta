#!/usr/bin/env bash
# Two approved plans from one split spec, both ready now. The case checks what
# slice does after the yes: ask subagent or inline, offer the default, or go
# straight to the parallel build, by what the user config says.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init and
# commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# The user config is the one thing the four parallel cases change. The eval
# home is a sealed throwaway, so only write it when it is missing: a hand run
# on a real machine keeps its own file.
CONFIG_EXTRA='build_executor: inline
subagent_picks: 3'
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\nplan_depth: minimal\nplan_split: on\n%s\n' "$CONFIG_EXTRA" > "$HOME/.acta/config.yaml"
fi
mkdir -p bin .acta/specs .acta/plans
# The system git does not work in the Bash sandbox, so the real binary is
# copied in, and ./bin/acta puts ./bin first on PATH before it starts acta.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
cat > .acta/specs/2026-11-04-text-tools-design.md <<'SPEC'
---
id: SPC-0001
created: "2026-11-04 09:00:00"
hash: textsp1
---
# Text tools

Status: approved by the user in chat on 2026-11-04.

Two helpers that do not depend on each other. `src/slug.py` holds slug(text):
lower case, spaces become dashes. `src/title.py` holds title(text): every word
starts with a capital letter. The tests use unittest from the standard library.
SPEC
cat > .acta/plans/2026-11-04-slug.md <<'PLAN'
---
parent: specs/2026-11-04-text-tools-design
depth: minimal
id: PLN-0001
created: "2026-11-04 09:10:00"
hash: slugpl1
---
# Slug helper

**Goal:** slug(text) returns the text in lower case with spaces as dashes.
**Spec:** `.acta/specs/2026-11-04-text-tools-design.md`
**Tests:** fast `python3 -m unittest discover -s tests`

## Waves

- Wave 1: Task 1

### Task 1: slug helper

**Files:** `src/slug.py`, `tests/test_slug.py`
**Seam:** `slug(text)`
**verify:** slug("Hello World") returns "hello-world".
- [ ] Red: the test fails because `src/slug.py` does not exist.
- [ ] Change: `src/slug.py` holds slug(text).
- [ ] Commit: feat(slug): add the slug helper
PLAN
cat > .acta/plans/2026-11-04-title.md <<'PLAN'
---
parent: specs/2026-11-04-text-tools-design
depth: minimal
id: PLN-0002
created: "2026-11-04 09:11:00"
hash: titlpl1
---
# Title helper

**Goal:** title(text) starts every word with a capital letter.
**Spec:** `.acta/specs/2026-11-04-text-tools-design.md`
**Tests:** fast `python3 -m unittest discover -s tests`

## Waves

- Wave 1: Task 1

### Task 1: title helper

**Files:** `src/title.py`, `tests/test_title.py`
**Seam:** `title(text)`
**verify:** title("hello world") returns "Hello World".
- [ ] Red: the test fails because `src/title.py` does not exist.
- [ ] Change: `src/title.py` holds title(text).
- [ ] Commit: feat(title): add the title helper
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
