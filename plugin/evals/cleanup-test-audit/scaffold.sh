#!/usr/bin/env bash
# Builds a tiny Python repo whose tests hold one text test, one near duplicate
# pair and one sound behavior test. The case asks for a test audit and checks
# that the report names the first two, leaves the sound one out, and changes
# nothing.
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
mkdir -p bin src tests .acta/scratch .acta/debt
# The system git may be a stub that the sandbox cannot run, so the real binary
# is copied. acta starts git by name, so ./bin/acta is a small script that
# puts ./bin first and then starts the real binary.
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
cp "$(command -v acta)" bin/acta.real
printf '#!/bin/sh\nPATH="%s/bin:$PATH"\nexport PATH\nexec "%s/bin/acta.real" "$@"\n' "$PWD" "$PWD" > bin/acta
chmod +x bin/acta
: > src/__init__.py
: > tests/__init__.py
cat > src/cli.py <<'PY'
import sys


def main(argv):
    if not argv:
        print("usage: cli NAME")
        return 2
    print("hello " + argv[0])
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
PY
cat > src/price.py <<'PY'
def total(prices, tax):
    return round(sum(prices) * (1 + tax), 2)
PY
cat > src/slug.py <<'PY'
def slugify(text):
    return "-".join(text.lower().split())
PY
# The text test: it reads the source and asserts it holds a line, so it
# passes whatever main does.
cat > tests/test_cli.py <<'PY'
import unittest
from pathlib import Path


class CliSourceTest(unittest.TestCase):
    def test_cli_prints_usage(self):
        source = Path("src/cli.py").read_text()
        self.assertIn('print("usage: cli NAME")', source)


if __name__ == "__main__":
    unittest.main()
PY
# The near duplicate pair: two names for the same input and the same assert.
cat > tests/test_price.py <<'PY'
import unittest

from src.price import total


class TotalTest(unittest.TestCase):
    def test_total_with_tax(self):
        self.assertEqual(total([10, 20], tax=0.1), 33.0)

    def test_total_includes_tax(self):
        self.assertEqual(total([10, 20], tax=0.1), 33.0)


if __name__ == "__main__":
    unittest.main()
PY
# The sound test: one table test that calls the code and checks its output.
cat > tests/test_slug.py <<'PY'
import unittest

from src.slug import slugify


class SlugifyTest(unittest.TestCase):
    def test_slugify(self):
        cases = [
            ("Hello World", "hello-world"),
            ("  Trim Me  ", "trim-me"),
        ]
        for text, want in cases:
            with self.subTest(text=text):
                self.assertEqual(slugify(text), want)


if __name__ == "__main__":
    unittest.main()
PY
git init -q -b main .
git config user.email eval@example.invalid
git config user.name eval
# The agent finds ./bin on PATH through the project settings. The two git
# lines keep git from reading config files that the sandbox may lock. Python
# writes no __pycache__, so a run that tidies one up never looks like a delete.
mkdir -p .claude
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1","PYTHONDONTWRITEBYTECODE":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git add -A
git commit -qm scaffold
