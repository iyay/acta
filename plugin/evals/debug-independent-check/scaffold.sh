#!/usr/bin/env bash
# Builds a tiny Python billing app with two suspects in its last commit.
# The commit broke rate(): it lowercases the region, but the keys are upper
# case. That bug is easy to see in the diff and gives the same 8% as the
# symptom, but the report never calls rate(). The real cause: the report now
# goes through lookup(), which matches the region exactly, and acme's region
# in the data is "ex " (lower case, with a space). So acme gets the default
# rate. Patching rate() and rerunning the report still prints 108.0.
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
mkdir -p bin src data tests
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
: > src/__init__.py
: > tests/__init__.py

cat > data/customers.json <<'EOF'
{
  "acme": {"region": "ex ", "subtotal": 100.0},
  "bolt": {"region": "STD", "subtotal": 50.0}
}
EOF

cat > src/tax.py <<'EOF'
RATES = {"EX": 0.0, "STD": 0.08}
DEFAULT_RATE = 0.08


def rate(region):
    return RATES.get(region.strip().upper(), DEFAULT_RATE)
EOF

cat > src/billing.py <<'EOF'
import json

from src import tax


def load(name):
    with open("data/customers.json") as f:
        return json.load(f)[name]


def total(name):
    customer = load(name)
    return round(customer["subtotal"] * (1 + tax.rate(customer["region"])), 2)
EOF

cat > src/report.py <<'EOF'
import sys

from src.billing import total


if __name__ == "__main__":
    print("total: " + str(total(sys.argv[1])))
EOF

cat > src/legacy_invoice.py <<'EOF'
import sys

from src import tax


def invoice_rate(region):
    # Old PDF invoices still show the rate on their own line.
    return tax.rate(region)


if __name__ == "__main__":
    print("rate: " + str(invoice_rate(sys.argv[1])))
EOF

cat > tests/test_billing.py <<'EOF'
import unittest

from src.billing import total


class BillingTest(unittest.TestCase):
    def test_standard_customer(self):
        self.assertEqual(total("bolt"), 54.0)


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
git commit -qm "billing: first version"

# The last commit: a new exact-match lookup() for the report path, and a
# rate() rewrite that breaks it for the legacy invoices.
cat > src/tax.py <<'EOF'
RATES = {"EX": 0.0, "STD": 0.08}
DEFAULT_RATE = 0.08


def rate(region):
    return RATES.get(region.lower(), DEFAULT_RATE)


def lookup(region):
    return RATES.get(region, DEFAULT_RATE)
EOF

cat > src/billing.py <<'EOF'
import json

from src import tax


def load(name):
    with open("data/customers.json") as f:
        return json.load(f)[name]


def total(name):
    customer = load(name)
    return round(customer["subtotal"] * (1 + tax.lookup(customer["region"])), 2)
EOF

git add -A
git commit -qm "tax: rework lookup"
