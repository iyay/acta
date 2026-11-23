#!/usr/bin/env bash
# Builds a tiny Python slug helper where a run of spaces gives a run of dashes.
# The unit already has a table test, so the regression test the diagnosis
# proposes fits as one more row in it. The file also has single-case test
# funcs, as real repos often do, so a new func looks just as natural.
set -euo pipefail
# The eval runs this in an empty folder. Anywhere else it would git init and
# commit files that are not ours, so stop first.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no config file sends the agent to setup first, which
# spends turns. Only write it when it is missing, so a hand run is safe.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
# The sandbox may not run the system git stub, so the real binary is copied.
mkdir -p bin src tests
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
: > src/__init__.py
: > tests/__init__.py
# The bug: each space becomes its own dash.
cat > src/slug.py <<'PY'
def slugify(text):
    return text.strip().lower().replace(" ", "-")
PY
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

    def test_digits_are_kept(self):
        self.assertEqual(slugify("Route 66"), "route-66")

    def test_dash_in_text_is_kept(self):
        self.assertEqual(slugify("Spider-Man"), "spider-man")


if __name__ == "__main__":
    unittest.main()
PY
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings. The two git
# lines keep git from reading config files that the sandbox may lock.
# Running the app would leave __pycache__ folders in the repo, and a run that
# cleans them up looks like it deleted repo files, so Python writes none.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1","PYTHONDONTWRITEBYTECODE":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
