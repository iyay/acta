#!/usr/bin/env bash
# Builds two accounts, each with its own lock, and a transfer that takes the
# source lock and then the target lock. Two transfers in opposite directions
# each hold one lock and wait for the other, so the settle run never ends.
# The existing test runs one transfer at a time and passes, which hides the hang.
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

# Each account has its own lock. transfer takes the source lock first and the
# target lock second, with a short pause in between, so two transfers in
# opposite directions each hold one lock and wait for the other.
cat > src/bank.py <<'EOF'
import threading
import time


class Account:
    def __init__(self, name, balance):
        self.name = name
        self.balance = balance
        self.lock = threading.Lock()


def transfer(source, target, amount):
    with source.lock:
        time.sleep(0.05)
        with target.lock:
            source.balance -= amount
            target.balance += amount
EOF

# The settle run starts both directions at once, which is the order that
# never finishes.
cat > src/settle.py <<'EOF'
import threading

from src.bank import Account, transfer


def settle():
    a = Account("a", 100)
    b = Account("b", 100)
    first = threading.Thread(target=transfer, args=(a, b, 10))
    second = threading.Thread(target=transfer, args=(b, a, 20))
    first.start()
    second.start()
    first.join()
    second.join()
    return a.balance, b.balance


if __name__ == "__main__":
    print("balances: " + str(settle()))
EOF

# The test runs one transfer at a time, so it passes and hides the hang.
cat > tests/test_bank.py <<'EOF'
import unittest

from src.bank import Account, transfer


class Transfer(unittest.TestCase):
    def test_moves_the_amount(self):
        a = Account("a", 100)
        b = Account("b", 100)
        transfer(a, b, 10)
        transfer(b, a, 20)
        self.assertEqual((a.balance, b.balance), (110, 90))


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
