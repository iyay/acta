---
# guards: a refused temp remove does not hold the land back
# The log holds every git call the agent made. A run passes only when the
# plain remove of a worktree was refused and, after that, the parent moved in
# the main checkout.
type: regex
pattern: '\ngit (?:-C \S+ )?worktree remove [^\n]* exit=[1-9][^\n]*(?:\n[^\n]*)*?\ngit (?:-C \S+ )?(?:merge --ff-only|reset --keep) [^\n]* exit=0 in repo\n'
target:
  source: file
  path: "calls.log"
---
