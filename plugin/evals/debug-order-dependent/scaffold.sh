#!/usr/bin/env bash
# Builds a small suite where one test changes a shared module setting and
# never puts it back. A later test then fails, but only when both run in that
# order. Each test passes alone.
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
  printf 'chat_language: English\nstyle: adhd\nrepo_language: English\n' > "$HOME/.acta/config.yaml"
fi
# The sandbox may not run the system git stub, so the real binary is copied.
mkdir -p bin src tests
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
: > src/__init__.py
: > tests/__init__.py

cat > src/money.py <<'EOF'
CONFIG = {"currency": "EUR", "places": 2}


def fmt(amount):
    return "{:.{p}f} {c}".format(amount, p=CONFIG["places"], c=CONFIG["currency"])
EOF

cat > tests/test_a_parse.py <<'EOF'
import unittest


class ParseTest(unittest.TestCase):
    def test_strip(self):
        self.assertEqual(" 3 ".strip(), "3")
EOF

cat > tests/test_b_yen.py <<'EOF'
import unittest

from src import money


class YenTest(unittest.TestCase):
    def test_yen_has_no_places(self):
        money.CONFIG["currency"] = "JPY"
        money.CONFIG["places"] = 0
        self.assertEqual(money.fmt(5), "5 JPY")
EOF

cat > tests/test_c_fmt.py <<'EOF'
import unittest

from src.money import fmt


class FmtTest(unittest.TestCase):
    def test_euro(self):
        self.assertEqual(fmt(5), "5.00 EUR")
EOF

cat > tests/test_d_sum.py <<'EOF'
import unittest


class SumTest(unittest.TestCase):
    def test_sum(self):
        self.assertEqual(sum([1, 2]), 3)
EOF

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
