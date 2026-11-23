---
# guards: tidy writes the land notes into its tip, and nothing writes them on main after
# The log holds every git and acta call, and right after the parent moves in
# ./repo the bug and debt files the parent then holds, as tip: lines. A run
# passes only when those lines show fixed_in: PLN-landtp1 and the ticked debt
# line, and no acta spec-dates, acta set ... fixed_in or acta tick runs after
# the parent moved.
type: regex
pattern: '^(?=[\s\S]*\ngit (?:-C \S+ )?(?:merge --ff-only|reset --keep) [^\n]*exit=0 in repo\n(?=(?:tip: [^\n]*\n)*tip: fixed_in: PLN-landtp1\n)(?:tip: [^\n]*\n)*tip: - \[x\] \(medium\) parse has no comment)(?![\s\S]*\ngit (?:-C \S+ )?(?:merge --ff-only|reset --keep) [^\n]*exit=0 in repo\n[\s\S]*\nacta (?:spec-dates|set \S+ fixed_in|tick) )'
target:
  source: file
  path: "calls.log"
---
