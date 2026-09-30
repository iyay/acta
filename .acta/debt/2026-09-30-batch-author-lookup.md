---
id: DBT-0032
hash: gdjj0ws
parent: plans/2026-09-30-batch-author-lookup
---
# Batch author lookup review notes

- [ ] gitc.Authors runs git log without -z, so names with a quote, backslash, tab or newline come back C-quoted and fall back to user.name; acta slugs are [a-z0-9-] so only hand-named files hit it; fix with -z and NUL-split parsing (internal/gitc/gitc.go:111).
- [ ] No test pins --no-renames: removing it keeps every Author test green, yet it matters when a moved file's old name is written again and both names are asked in one call (internal/gitc/gitc_test.go, TestAuthorsAsksOnceForManyFiles).
- [ ] A commit with an empty %an skips its files in Authors, so a newer re-add can win; the old code fell back to user.name there; git normally refuses empty names (internal/gitc/gitc.go:117-123).
- [ ] Authors passes every path of a folder as one argument list; fine up to about 10k files per folder, then the whole folder silently falls back to user.name with no log line (internal/gitc/gitc.go:111).
- [ ] With many paths git log simplifies history differently: one file added on both sides of a merge can get the side-branch author instead of the first-parent one (rare, needs two authors).
- [ ] With log.follow=true in user config, the old one-path log followed renames and gave the first author; the new call gives the mover, which is what the plan says.
- [ ] acta show SCR-0013 measured 0.42 s on an idle run and 0.70-1.32 s under load average 28-40; the plan target "well under 0.5 s" holds only on the idle run; 11 author git calls remain, one per folder across 4 worktrees.
- [ ] acta tick --start/--all left an uncommitted started/finished stamp in the parent spec again; reverted before landing (known, see tick stamps).
