---
# guards: the final plan check runs after tidy moves the parent
# The log holds every git and acta call the agent made, with its exit code and
# the checkout it ran in. A run passes only when, after the parent moved in
# ./repo and before any worktree remove, it read the plan by hash with acta
# show --json in ./repo and proved a task commit with a passing
# git merge-base --is-ancestor there, and then cleaned up.
type: regex
pattern: '^(?=[\s\S]*\ngit (?:-C \S+ )?(?:merge --ff-only|reset --keep) [^\n]*exit=0 in repo(?:(?!\ngit (?:-C \S+ )?worktree remove )[\s\S])*\nacta show (?=[^\n]*landtp1)(?=[^\n]*--json)[^\n]* exit=0 in repo\n)(?=[\s\S]*\ngit (?:-C \S+ )?(?:merge --ff-only|reset --keep) [^\n]*exit=0 in repo(?:(?!\ngit (?:-C \S+ )?worktree remove )[\s\S])*\ngit (?:-C \S+ )?merge-base --is-ancestor \S+ (?:HEAD|main) exit=0 in repo\n)(?=[\s\S]*\ngit (?:-C \S+ )?worktree remove )'
target:
  source: file
  path: "calls.log"
---
