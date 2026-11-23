---
id: DBT-0173
hash: qbcuv0p
parent: plans/2026-11-23-planning-fold-wider
---
# Review NOTEs: Wider fold for planning commits on main

- [ ] (medium) internal/gitc/gitc.go is 1140 lines against the 800 cap and foldArgs is 58 lines; move actaLines, foldStems and the fold subject code into their own file.
- [ ] (medium) A planning write after a land can fold into the commit refs/acta/last-land points at, leaving last-land on an orphan so the next tidy folds no earlier chores; the fold should treat the last-land commit as shared.
