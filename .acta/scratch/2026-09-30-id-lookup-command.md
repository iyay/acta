---
id: SCR-0028
hash: iivf3gy
title: Default acta command to find any item by id
status: raw
created: "2026-09-30"
schema: "1"
---
# Default acta command to find any item by id

## Words

### 2026-09-30

User 2026-09-30: "buat file name lookup-nya sendiri gimana? seakrang kan gak ada ID atau hash difilename-nya" / "bukan hanya dari scratch item sih, nyari semua kind" / "catet, sama jadi command default acta buat nyari/lookup issue"

## Context

Came up while PLN-0040 (batch author lookup, SPC-0031) was being built through dispatch.

Measured 2026-09-30 on main: `grep -rlx 'id: <ID>' .acta` finds the file for any kind (SPC, PLN, BUG, DBT, SCR) in about 0.02 s over 137 files, with no git and no acta. Five ids in one `grep -rlxE` also took 0.017 s. Hashes work the same way: `grep -rlx 'hash: <hash>' .acta`. No skill teaches this today (no SKILL.md mentions a grep by id).

Idea: make this the default acta way to find an item. A command (name open, for example `acta find <id>` or `acta path <id>`) that prints the file path, and skills that point agents at it (or at the raw grep).

Limits found:
- Sub-items (task PLN-0040.01, debt line DBT-0031.01) are not written in any file; they are counted from the parent file's order. Look up the parent, then take the Nth task or line.
- Old ids (PLAN-30, SCRATCH-20) are no longer in frontmatter after the migration; only `acta show` maps them through canon aliases.
- The root can move (`.acta.yaml`, `ACTA_ROOT`, `--root`), so the path in the grep must follow it.
- Items that live only in a worktree or an unmerged branch are not in the main tree.
- .acta/specs/2026-09-26-review-debt-design.md still has `id: DEBT-3` (old format), so a new-format grep misses it.

Related: SCR-0020 (acta path idea, dropped from SPC-0031 because the slow part of acta show was fillAuthors, not the lookup). Keeping file names as they are (no id or hash in the name) was the recommendation: renames on every id --fix-duplicates and merge conflicts buy nothing when grep is this fast.

## Log

## Open questions
