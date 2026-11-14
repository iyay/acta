#!/usr/bin/env bash
set -euo pipefail
if [ -n "$(ls -A)" ]; then
  echo "scaffold.sh: run this in an empty folder" >&2
  exit 1
fi
mkdir -p "$HOME/.acta" .acta/specs reports refs bin .claude
if [ ! -f "$HOME/.acta/config.yaml" ]; then
  printf 'chat_language: English\nstyle: adhd\n' > "$HOME/.acta/config.yaml"
fi
cp "$(xcrun --find git 2>/dev/null || command -v git)" bin/git
export PATH="$PWD/bin:$PATH"
export GIT_CONFIG_GLOBAL=/dev/null GIT_CONFIG_NOSYSTEM=1
printf '{"env":{"PATH":"%s/bin:%s","GIT_CONFIG_GLOBAL":"/dev/null","GIT_CONFIG_NOSYSTEM":"1"}}\n' "$PWD" "$PATH" > .claude/settings.json
python3 - <<'FIXTURE'
import json
from pathlib import Path
files = json.loads(r'''{"refs/contract.md": "# Contract\nThe implementation follows the approved archive rule.\n", ".acta/specs/evidence-valid.md": "---\nid: SPC-0001\nhash: read001\n---\n# Evidence valid\n## Goal\nExport an archive.\n## Behavior\nFollow [the contract](../../refs/contract.md).\n", "reports/evidence-valid.txt": "No consequential findings.\n\n### Read\nSpec sections:\n- Goal\n- Behavior\nReference paths:\n- refs/contract.md\n", ".acta/specs/evidence-missing.md": "---\nid: SPC-0002\nhash: read002\n---\n# Evidence missing\n## Goal\nExport an archive.\n## Behavior\nFollow [the contract](../../refs/contract.md).\n", "reports/evidence-missing.txt": "No consequential findings.\n", ".acta/specs/evidence-empty.md": "---\nid: SPC-0003\nhash: read003\n---\n# Evidence empty\n## Goal\nExport an archive.\n## Behavior\nFollow [the contract](../../refs/contract.md).\n", "reports/evidence-empty.txt": "No consequential findings.\n\n### Read\nSpec sections:\n\nReference paths:\n", ".acta/specs/evidence-generic.md": "---\nid: SPC-0004\nhash: read004\n---\n# Evidence generic\n## Goal\nExport an archive.\n## Behavior\nFollow [the contract](../../refs/contract.md).\n", "reports/evidence-generic.txt": "No consequential findings.\n\n### Read\nAll of it. References: none.\n"}''')
for name, text in files.items():
    Path(name).write_text(text)
FIXTURE
git init -q -b main
git config user.email eval@example.invalid
git config user.name eval
git add .acta/specs reports refs
git commit -qm 'spec: record completed cold-read inputs'
