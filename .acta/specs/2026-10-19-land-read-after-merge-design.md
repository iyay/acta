---
parent: bugs/2026-10-19-land-reads-code-before-merge
id: SPC-0151
created: "2026-10-19 06:13:55"
hash: fcd1lcr
---
# Land reads the landed code after the merge

## Problem

BUG-0047: the understanding part of the landing report says "read the final
landed code" but not when or from where. About 1 run in 3 of the eval case
`review-understanding-report` reads `src/retry.py` in the worktree before the
merge and fails.

## Design

Bounded change to `plugin/skills/land/SKILL.md`, the first bullet under
`### Understanding, after the land`:

- After the merge (or the tidy move of the parent) has succeeded, read the
  landed files from the parent checkout, or from the landed commit with
  `git show <landed-sha>:<path>` (from any folder). A read of the old
  worktree's working files, or any read before the merge, does not count.
- Then explain the final landed code (fixes, polish and the merge included),
  not the review notes or an earlier diff.

Out of scope: the failed run also amended the merge commit. No root cause yet.

## Files

- `plugin/skills/land/SKILL.md`
- `internal/plugincheck/skill_land_test.go` (the `Must` phrase), and the byte
  cap in `internal/plugincheck/budget_test.go` if the file grows past it

## Testing

- `scripts/test ./internal/plugincheck/` goes red on the new phrase, then green.
- `scripts/eval --case review-understanding-report --runs 3` passes 3 of 3
  (it was 2 of 3 before).
- Coverage limit: that case pins `commit_history: full`, so it proves the
  plain merge path only. The tidy path shares the same skill sentence but has
  no eval; it is checked by reading `tidy.md`, which hands the report back to
  `SKILL.md`. A tidy eval case is not part of this fix.
