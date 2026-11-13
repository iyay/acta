#!/usr/bin/env bash
set -euo pipefail
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
mkdir -p "$HOME/.acta" bin baseline/src baseline/data .claude
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1","PYTHONDONTWRITEBYTECODE":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
: > baseline/src/__init__.py
printf '{"acme":{"region":"ex ","subtotal":100.0}}\n' > baseline/data/customers.json
cat > baseline/src/tax.py <<'PYTHON'
RATES = {"EX": 0.0, "STD": 0.08}


def lookup(region):
    return RATES.get(region.strip().upper(), 0.08)
PYTHON
cat > baseline/src/report.py <<'PYTHON'
import json
import sys
from src.tax import lookup

if __name__ == "__main__":
    with open("data/customers.json") as f:
        customer = json.load(f)[sys.argv[1]]
    print("total: " + str(round(customer["subtotal"] * (1 + lookup(customer["region"])), 2)))
PYTHON
git -C baseline init -q -b main
git -C baseline config user.email eval@example.invalid
git -C baseline config user.name eval
git -C baseline add src data
git -C baseline commit -qm scaffold
git -C baseline worktree add -q -b report-lookup ../checkout
# Only the investigated worktree has this exact-match regression.
python3 - <<'PYTHON'
from pathlib import Path
p = Path('checkout/src/tax.py')
p.write_text(p.read_text().replace('region.strip().upper()', 'region'))
PYTHON
git -C checkout add src/tax.py
git -C checkout commit -qm 'tax: use region key directly'
