#!/usr/bin/env bash
# A repo with three items cleanup lists: a raw idea that a commit already
# built, an open debt line about code that is gone, and an open bug whose fix
# commit is in the log. The acta CLI and git sit inside the workspace, since a
# sandboxed run cannot read the home directory.
set -euo pipefail
# Fixtures must never replace a developer's files.
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
# A first run with no voice file sends the agent to setup first. Only write it
# when it is missing, so a hand run never overwrites a real one.
mkdir -p "$HOME/.acta"
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
mkdir -p bin .claude .acta/scratch .acta/debt .acta/bugs src
cp "$(command -v acta)" bin/acta
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
export PATH="$PWD/bin:$PATH"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval

cat > .acta/scratch/2026-09-01-csv-export.md <<'MD'
---
id: SCR-0001
status: raw
---
# Export the report as CSV
MD
cat > .acta/debt/2026-09-02-report.md <<'MD'
---
id: DBT-0001
---
# Report debt

- [ ] (medium) legacyTotals in src/report.py adds the tax twice
MD
cat > .acta/bugs/2026-09-03-empty-name-crash.md <<'MD'
---
id: BUG-0001
---
# Greeting crashes on an empty name

## Symptom
greet("") raises IndexError.
MD
cat > src/report.py <<'PY'
def legacy_totals(rows):
    return sum(r["net"] + r["tax"] + r["tax"] for r in rows)
PY
cat > src/greet.py <<'PY'
def greet(name):
    return "Hello, " + name[0].upper() + name[1:]
PY
git add -A
git commit -qm "chore: start"

cat > src/export.py <<'PY'
import csv


def export_csv(rows, path):
    with open(path, "w", newline="") as f:
        csv.writer(f).writerows(rows)
PY
git add src/export.py
git commit -qm "feat(report): export the report as CSV"

cat > src/report.py <<'PY'
def totals(rows):
    return sum(r["net"] + r["tax"] for r in rows)
PY
git add src/report.py
git commit -qm "refactor(report): replace legacy_totals with totals"

cat > src/greet.py <<'PY'
def greet(name):
    if not name:
        return "Hello"
    return "Hello, " + name[0].upper() + name[1:]
PY
git add src/greet.py
git commit -qm "fix(greet): an empty name no longer crashes"
