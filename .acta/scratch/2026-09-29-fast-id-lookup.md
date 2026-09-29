---
id: SCR-0013
hash: oxoqf86
title: Cheapest and fastest way to find an item by id
status: brainstorming
created: "2026-09-29"
---
User 2026-09-29, in chat: asked what acta reads to find "SCRATCH-1" when the user says "brainstorm scratch-1", then: "mendingan dari sekarang dipikirin", "tw nyari cara yang paling murah dan cepet".

Today: the id lives only in frontmatter; file names hold date + slug. `acta show <id>` loads the whole board (every .md frontmatter, every kind) and, through loadAllTrees, every worktree and unmerged branch via git, then picks one id.

Options ranked in chat (cost per lookup):
1. Id in the file name (SCRATCH-1-newest-first-sort.md): one readdir, no file opened; agents can `ls .acta/scratch/SCRATCH-1-*` without acta. Costs: file contract change, one-time rename of every file, rename again when acta id --fix-duplicates renumbers (or use the stable hash in the name, but users say the number).
2. Fast path in acta (recommended first): when given an id, read only the folder of its prefix, only the frontmatter up to the second ---, no git calls and no other trees unless the id is not in the main tree. No contract change, no renames, no merge conflicts; a fraction of a millisecond.
3. Index file (.acta/.index.json): about as fast as 1, but conflicts on nearly every merge and goes stale when files are edited by hand. Least suitable.

Recommendation: 2 now, 1 only if 2 is measured slow at thousands of files. Brainstorm in a new session; measure alongside SCRATCH-10 (scroll performance).

User 2026-09-29 (brainstorm): TUI list rows do not line up and SCRATCH-XX is too long; same for every kind. Wants rows aligned like "ID0000  PLN-0001 Title". Prefix becomes 3 letters (PLN, SPC, BUG, ...). Hash should grow from 4 to 5 characters. Path upgraded Bounded to Architectural: id format change touches every file and the id contract.

User 2026-09-29: list shows the id, not the hash. Id = 3-letter prefix + 4 digits, like PLN-0001, BUG-0001, SPC-0001, DBT-0001. Goal: agents match and find an item fast, with few tokens. User floated TSK (task), STK (subtask), SRT (scratch).

User 2026-09-29: hash stays. Length follows git short commit hash (7). Agent view: hash is random (randHash), letter first then a-z0-9 (IsHash in internal/board/ids.go); freeHash only checks taken hashes in one tree, so a longer hash cuts cross-branch clashes.

User 2026-09-29: approved migration: keep old numbers (PLAN-30 becomes PLN-0030), extend old hashes to 7 with random chars, accept old ids (PLAN-30) as aliases on lookup.

User 2026-09-29: task number is 2 digits (PLN-0030.03) so list rows line up.

User 2026-09-29: tasks never reach 3 digits; the plan skill must split big work into smaller plans. Proposal: acta flags a plan with 100+ tasks as a problem; plan skill tells the agent to split instead.

User 2026-09-29: debt items also 2 digits (DBT-0022.04). Approach 1 chosen: one-time migration (acta id --migrate rewrites id and hash in frontmatter, one commit) plus old ids accepted as aliases.

Section 1 approved 2026-09-29: IsID 4 digits, IsHash 7; canon(id) lowercases, maps old prefix to new, drops leading zeros; alias keys stored canon; hash lookup by unique prefix like git; list shows id + title only, hash in detail and acta show.

Section 2 approved 2026-09-29: no new command; acta id normalizes old ids on every run (first migration and files brought in by merges). Rewrites id, hash (extend to 7 via freeHash) and closes: ids; parent: is a path so untouched; bodies, commits, memory untouched (canon handles them). Board reads both formats. Note in plan: run first migration when worktrees are quiet.

Section 3 approved 2026-09-29: skills land/brainstorm/plan/scratch updated to new examples; plan skill caps 99 tasks per plan; tests for IsID, IsHash, canon, hash prefix, acta id normalize (idempotent), 100-task and 10000 problems, TUI fixed-width row; fixtures moved to new format with some kept old to test aliases. Out of scope: acta path fast lookup, slug shown as title.
