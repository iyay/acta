---
# guards: acta flake record <name> --plan <plan id>
# The log holds every git and acta call the agent made, with its exit code and
# the checkout it ran in. The suite passed but printed one FLAKY line. A run
# passes only when, before any merge, it recorded that flake for the plan in
# the worktree where the suite ran, and then merged the branch into main.
# A record made after the first merge would leave the bug off the merge, so
# the part before the record may hold no merge call.
type: regex
pattern: '^(?:(?!\ngit (?:-C \S+ )?merge )[\s\S])*\nacta flake record internal/calc\.TestAddCarries --plan (?:PLN-0001|landrf1) exit=0 in repo-land\n[\s\S]*\ngit (?:-C \S+ )?merge (?:--[a-z-]+ )*land\b[^\n]* exit=0 in repo\n'
target:
  source: file
  path: "calls.log"
---
