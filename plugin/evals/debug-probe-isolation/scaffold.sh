#!/usr/bin/env bash
set -euo pipefail
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
mkdir -p "$HOME/.acta" bin src tools .claude
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
cat > src/count.py <<'PYTHON'
def count(records):
    return len(records) - 1
PYTHON
cat > tools/check.py <<'PYTHON'
import json
from pathlib import Path
import sys
sys.path.insert(0, str(Path(__file__).resolve().parents[1]))
from src.count import count

records = ["a", "b", "c"]
actual = count(records)
Path(".cache").mkdir(exist_ok=True)
Path(".cache/check.json").write_text(json.dumps({"actual": actual}))
print("expected 3, got", actual)
PYTHON
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval
git add src tools bin .claude
git commit -qm scaffold
