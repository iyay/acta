#!/usr/bin/env bash
# A run starts in an empty folder, and a run that may use Bash runs sandboxed
# with the home directory unreadable, so the acta CLI and git have to sit
# inside the workspace before the agent starts. This script runs as the user,
# not in the sandbox, so it can still find them on PATH.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init,
# change git config and commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no voice file makes the session rules send the agent to
# /acta:setup before any other work, which spends turns this case needs.
# Only write it when it is missing, so a hand run never overwrites a real one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin src tests .acta/plans
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
# The two folders are empty packages, so the unittest command in the plan can
# import src.slug from the repo root.
: > src/__init__.py
: > tests/__init__.py
# The bug: each space becomes its own dash, so "Hello   World" gives three
# dashes where the plan wants one.
cat > src/slug.py <<'PY'
def slugify(text):
    return text.strip().lower().replace(" ", "-")
PY
# The unit already has a table test and one more test func. The case checks
# that the fix adds a row to the table and no third test func.
cat > tests/test_slug.py <<'PY'
import unittest

from src.slug import slugify


class SlugifyTest(unittest.TestCase):
    def test_slugify(self):
        cases = [
            ("Hello World", "hello-world"),
            ("  Trim Me  ", "trim-me"),
            ("already-slug", "already-slug"),
        ]
        for text, want in cases:
            with self.subTest(text=text):
                self.assertEqual(slugify(text), want)

    def test_empty_text_gives_empty_slug(self):
        self.assertEqual(slugify(""), "")


if __name__ == "__main__":
    unittest.main()
PY
# A grader cannot see the order of runs and edits in a long trace, so every
# python run logs its exit code and whether each file still matches the
# scaffold. A red run with the test changed and the code not yet changed is
# the failing row seen before the fix. A python under the home folder will not
# start in the sandbox, so the wrapper takes its own folder off PATH and runs
# python3 by name, the same one a plain python3 call would find.
cat > bin/python3 <<WRAP
#!/bin/sh
# If PATH names this folder another way, python3 below finds this script
# again. Stop then, so it never loops forever.
if [ -n "\${SLUG_PY_WRAPPED:-}" ]; then
  echo "bin/python3: found itself on PATH" >&2
  exit 127
fi
SLUG_PY_WRAPPED=1
export SLUG_PY_WRAPPED
PATH=\$(printf '%s' "\$PATH" | tr ':' '\n' | grep -vx "$PWD/bin" | paste -sd ':' -)
export PATH
python3 "\$@"
code=\$?
src=changed
cmp -s "$PWD/src/slug.py" "$PWD/.git/slug.orig" && src=orig
tests=changed
cmp -s "$PWD/tests/test_slug.py" "$PWD/.git/test_slug.orig" && tests=orig
echo "exit=\$code src=\$src tests=\$tests" >> "$PWD/.git/test-runs.log"
exit \$code
WRAP
chmod +x bin/python3
cp bin/python3 bin/python
# The task reads like a bug fix from a debug session: it asks for a
# regression test and names the input and the result, not where the test goes.
cat > .acta/plans/2026-11-23-slug.md <<'PLAN'
---
depth: minimal
id: PLN-0001
created: "2026-11-23 10:00:00"
hash: slugrow
---
# Slug Fix Plan

**Goal:** A run of spaces in a slug becomes one dash.

**Spec:** none

**Tests:** `./bin/python3 -m unittest discover -s tests -t .`

### Task 1: Fix a run of spaces giving many dashes

**Bug:** slugify("Hello   World") returns "hello---world".

**Files:**
- Modify: `src/slug.py`
- Test: `tests/test_slug.py`

**verify:** slugify turns a run of spaces inside the text into one dash, so slugify("Hello   World") is "hello-world".

- [ ] **Failing test:** a regression test that reproduces the bug: slugify("Hello   World") must return "hello-world".
- [ ] **Change:** `src/slug.py` turns each run of spaces into one dash.
- [ ] **Commit:** `fix(slug): one dash for a run of spaces`
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
# Copies of the two files as the scaffold leaves them, kept inside .git so no
# commit picks them up. The python wrapper compares against them.
cp src/slug.py .git/slug.orig
cp tests/test_slug.py .git/test_slug.orig
