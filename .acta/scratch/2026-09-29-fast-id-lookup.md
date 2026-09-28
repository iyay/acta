---
id: SCRATCH-13
hash: oxoq
title: Cheapest and fastest way to find an item by id
status: raw
created: "2026-09-29"
---
User 2026-09-29, in chat: asked what acta reads to find "SCRATCH-1" when the user says "brainstorm scratch-1", then: "mendingan dari sekarang dipikirin", "tw nyari cara yang paling murah dan cepet".

Today: the id lives only in frontmatter; file names hold date + slug. `acta show <id>` loads the whole board (every .md frontmatter, every kind) and, through loadAllTrees, every worktree and unmerged branch via git, then picks one id.

Options ranked in chat (cost per lookup):
1. Id in the file name (SCRATCH-1-newest-first-sort.md): one readdir, no file opened; agents can `ls .acta/scratch/SCRATCH-1-*` without acta. Costs: file contract change, one-time rename of every file, rename again when acta id --fix-duplicates renumbers (or use the stable hash in the name, but users say the number).
2. Fast path in acta (recommended first): when given an id, read only the folder of its prefix, only the frontmatter up to the second ---, no git calls and no other trees unless the id is not in the main tree. No contract change, no renames, no merge conflicts; a fraction of a millisecond.
3. Index file (.acta/.index.json): about as fast as 1, but conflicts on nearly every merge and goes stale when files are edited by hand. Least suitable.

Recommendation: 2 now, 1 only if 2 is measured slow at thousands of files. Brainstorm in a new session; measure alongside SCRATCH-10 (scroll performance).
