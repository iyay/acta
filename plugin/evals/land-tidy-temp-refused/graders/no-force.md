---
# guards: a refused temp remove is never forced
type: regex
pattern: '^(?![\s\S]*\ngit (?:-C \S+ )?worktree remove [^\n]*(?:--force| -f ))[\s\S]*\ngit (?:-C \S+ )?worktree remove [^\n]* exit=[1-9]'
target:
  source: file
  path: "calls.log"
---
