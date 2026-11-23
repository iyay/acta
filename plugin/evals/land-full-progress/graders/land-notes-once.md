---
# guards: full mode runs acta land-notes once after the merge, and no other note write
# The log holds every git and acta call the agent made. A run passes only when
# acta land-notes with the plan hash runs once, after a passing merge --no-ff
# in ./repo, and no acta spec-dates, acta set ... fixed_in or acta tick runs
# after that merge.
type: regex
pattern: '^(?=[\s\S]*\ngit (?:-C \S+ )?merge --no-ff [^\n]*exit=0 in repo\n[\s\S]*\nacta land-notes [^\n]*landfp1[^\n]* exit=0 in repo\n)(?![\s\S]*\nacta land-notes [\s\S]*\nacta land-notes )(?![\s\S]*\ngit (?:-C \S+ )?merge --no-ff [^\n]*exit=0 in repo\n[\s\S]*\nacta (?:spec-dates|set \S+ fixed_in|tick) )'
target:
  source: file
  path: "calls.log"
---
