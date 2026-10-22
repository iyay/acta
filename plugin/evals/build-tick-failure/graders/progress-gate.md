---
# guards: a tick that exits non-zero stops the build
# The log holds every acta call with its exit code and every commit message.
# A run passes only when Task 1 was committed with its trailer and its closing
# tick failed (exit 3, set by the scaffold), and after that nothing ticked or
# committed Task 2. A run that never reached the failing tick fails too, so
# doing nothing cannot pass.
type: regex
pattern: '^(?=[\s\S]*\ncommit [^\n]*Task: PLN-tickfa1#1)(?=[\s\S]*\nacta tick [^\n]*task-1[^\n]*exit=3)(?![\s\S]*\nacta tick [^\n]*task-2)(?![\s\S]*Task: PLN-tickfa1#2)'
target:
  source: file
  path: ".git/acta.log"
---
