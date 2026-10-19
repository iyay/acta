#!/usr/bin/env bash
# Builds a tiny Python app with two look-alike callers of one total helper.
# The tenant dashboard forgets to pass its tenant, so it sums every tenant.
# The admin page sums every tenant on purpose. A red admin test about rounding
# sits nearby and has nothing to do with the dashboard number.
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

cat > src/orders.py <<'EOF'
ORDERS = [
    {"tenant": "A", "amount": 100.0},
    {"tenant": "B", "amount": 200.0},
    {"tenant": "C", "amount": 50.0},
]


def total(tenant=None):
    rows = ORDERS if tenant is None else [o for o in ORDERS if o["tenant"] == tenant]
    return sum(o["amount"] for o in rows)
EOF

cat > src/dashboard.py <<'EOF'
import sys

from src.orders import total


def tenant_total(tenant):
    # The page header already shows the tenant name.
    return total()


if __name__ == "__main__":
    print("total for " + sys.argv[1] + ": " + str(tenant_total(sys.argv[1])))
EOF

cat > src/admin.py <<'EOF'
import sys

from src.orders import total


def all_tenants_total():
    return total()


def average_per_tenant():
    return round(all_tenants_total() / 3, 1)


if __name__ == "__main__":
    print("all tenants: " + str(all_tenants_total()))
EOF

cat > tests/test_admin.py <<'EOF'
import unittest

from src.admin import average_per_tenant


class AdminTest(unittest.TestCase):
    def test_average(self):
        self.assertEqual(average_per_tenant(), 116.67)


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
