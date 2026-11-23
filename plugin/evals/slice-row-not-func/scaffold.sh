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
mkdir -p bin src tests .acta/specs .acta/plans
# The repo file pins the depth, so the run does not depend on the user's own
# setting, and a minimal plan keeps the run short.
printf 'plan_depth: minimal\n' > .acta.yaml
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
: > src/__init__.py
: > tests/__init__.py
cat > src/slug.py <<'PY'
def slugify(text):
    return "-".join(text.lower().split())
PY
# The unit already has a table test. The case checks that the plan's task
# names this test and a new row in it, not a new test func.
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
# The spec says it is approved, which is what the slice skill needs first.
cat > .acta/specs/2026-11-23-slug-punct-design.md <<'SPEC'
---
id: SPC-0001
created: "2026-11-23 09:00:00"
hash: slugpnc
---
# Slugs drop punctuation

Status: Bounded, approved by the user in chat on 2026-11-23.

slugify(text) in `src/slug.py` drops punctuation, so slugify("Hello, World!")
is "hello-world". Letters, digits and spaces work as today. The tests use
unittest from the standard library: `python3 -m unittest discover -s tests -t .`.
SPEC
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings. The two git lines
# keep git from reading config files that the sandbox may lock.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
