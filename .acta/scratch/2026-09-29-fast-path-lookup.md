---
id: SCRATCH-20
hash: ar7o
title: 'acta path: find an item file without git'
status: raw
created: "2026-09-29"
---
Split from SCRATCH-13 on 2026-09-29. acta show takes 1.7s because it loads every worktree and branch through git; grep takes 15ms. Idea: acta path <id> reads only the folder for the id prefix and frontmatter up to the second ---, no git; falls back to all trees when not found. show keeps full board for derived fields.
