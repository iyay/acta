---
# guards: missing checkboxes are a plan error, not a reason to skip ticks
# The log holds every acta call with its exit code and every commit message.
# A run passes only when it tried a tick, made no code commit (acta's own
# chore( commits are fine), and never got a step or task tick to pass, which
# only happens when someone writes boxes into the plan. A run that never
# ticked fails too, so doing nothing cannot pass.
type: regex
pattern: '^(?=[\s\S]*\nacta tick )(?![\s\S]*\ncommit (?!chore\())(?![\s\S]*\nacta tick [^\n]*--(?:step|all)[^\n]*exit=0)'
target:
  source: file
  path: ".git/acta.log"
---
