#!/usr/bin/env bash
# Builds a report that reads and parses the whole price file once for every
# item. The work grows with the item count, so a big report is slow. A
# nearby string join looks wasteful but costs almost nothing.
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

python3 - <<'EOF'
import json
prices = {"sku" + str(i): round(1 + i * 0.01, 2) for i in range(5000)}
with open("prices.json", "w") as f:
    json.dump(prices, f)
EOF

cat > src/catalog.py <<'EOF'
import json


def price(sku, path="prices.json"):
    with open(path) as f:
        return json.load(f)[sku]
EOF

cat > src/report.py <<'EOF'
import sys

from src.catalog import price


def build(skus):
    lines = []
    for sku in skus:
        lines.append(sku + "," + str(price(sku)))
    out = ""
    for line in lines:
        out = out + line + "\n"
    return out


if __name__ == "__main__":
    n = int(sys.argv[1])
    print(len(build(["sku" + str(i) for i in range(n)])))
EOF

cat > tests/test_report.py <<'EOF'
import unittest

from src.report import build


class ReportTest(unittest.TestCase):
    def test_two(self):
        self.assertEqual(build(["sku0", "sku1"]), "sku0,1.0\nsku1,1.01\n")


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
