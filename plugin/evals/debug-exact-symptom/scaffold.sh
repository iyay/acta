#!/usr/bin/env bash
# Builds a tiny Python app where an explicit "verbose: false" in the settings
# file is lost and the default wins. The process still exits 0, so only the
# printed output shows the bug.
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

cat > settings.json <<'EOF'
{"name": "nightly", "verbose": false}
EOF

cat > src/settings.py <<'EOF'
import json

DEFAULTS = {"name": "job", "verbose": True}


def load(path):
    with open(path) as f:
        raw = json.load(f)
    return {key: raw.get(key) or DEFAULTS[key] for key in DEFAULTS}
EOF

cat > src/cli.py <<'EOF'
import sys

from src.settings import load


def main(argv):
    cfg = load(argv[0] if argv else "settings.json")
    if cfg["verbose"]:
        print("debug: loaded settings for " + cfg["name"])
    print("ran " + cfg["name"])
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1:]))
EOF

cat > tests/test_cli.py <<'EOF'
import unittest

from src.cli import main


class CliTest(unittest.TestCase):
    def test_runs(self):
        self.assertEqual(main(["settings.json"]), 0)


if __name__ == "__main__":
    unittest.main()
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
