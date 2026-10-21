---
parent: specs/2026-10-21-wiki-check-one-git-pass-design
depth: minimal
id: PLN-0184
created: "2026-10-21 09:42:01"
hash: zhr3jcx
started: "2026-10-21 09:43:16"
finished: "2026-10-21 09:51:20"
---

# Wiki check reads git once per run

**Goal:** Make `acta wiki check` run a fixed number of git commands per run, whatever the number of pages, with the same messages as today.
**Spec:** .acta/specs/2026-10-21-wiki-check-one-git-pass-design.md
**Tests:** fast `scripts/test ./internal/gitc ./internal/wiki`; full `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Keep `--literal-pathspecs --full-history --no-merges` on every history read, so stale keeps its meaning (see `.acta/wiki/wiki-freshness-from-git.md`).
- `internal/doctor/wiki.go` still calls `gitc.CheckCommits` and `gitc.Unseen`; leave both functions and that caller as they are.
- Match page paths in Go with `covers`, the rule `Match` already uses. No second matcher.

## Waves

- Wave 1: Task 1.
- Wave 2: Task 2.

### Task 1: One git pass that returns the wiki history

**Files:** Modify `internal/gitc/gitc.go`; test in `internal/gitc/gitc_test.go`.
**Seam:** a new `gitc.WikiHistory(repo, wikiDir, checkedFile string, paths []string)` that returns the parent graph of HEAD, the files each non-merge commit touched under the wiki folder or `paths` (newest first, with short sha and subject), the check commits per page id from the checked file, and the dirty files under the wiki folder.
**verify:** For every history shape the current `CheckCommits` and `Unseen` tests cover (merge on either side, line moved by a union merge, line only removed, longer id with the same prefix, side branch), the check commits per id equal what `CheckCommits` gives for the checked file. The call runs at most four git commands, whatever the number of paths. List each shape checked.
- [x] Red: a table test builds each shape, calls `WikiHistory`, and compares its check map with `CheckCommits` on the checked file; it fails to compile because `WikiHistory` does not exist.
- [x] Change: add `WikiHistory` with `rev-list --parents HEAD`, one `log -p -U0` over the checked file reusing the parse in `CheckCommits` (move that parse into a shared helper), one `log --name-only` over the wiki folder plus `paths`, and one `status --porcelain` over the wiki folder.
- [x] Commit: `perf(gitc): read the wiki history in one pass`

### Task 2: Check every page from the one history

**Files:** Modify `internal/wiki/check.go`; test in `internal/wiki/check_test.go`.
**Seam:** `wiki.Check(repo, pages, rng)`.
**verify:** For every page shape the current `check_test.go` covers (fresh, stale, uncommitted edit, check by line, check by page edit, merge, side branch, range filter), `Check` returns the same problems in the same order as before. A run over 1 page and over 20 pages starts the same number of git processes, counted with a `git` wrapper put first on PATH in the test. Ancestry is computed from the parent graph; a check commit counts as its own ancestor. List each shape checked.
- [x] Red: a test with a counting `git` wrapper on PATH runs `Check` over 1 and 20 pages and wants equal counts; it fails because each page runs its own git commands.
- [x] Change: in `Check`, call `gitc.WikiHistory` once after the range filter, and make `staleMsgs` read the dirty set, the check commits and the newest unseen commit from it, walking parents in memory. A `WikiHistory` error gives the same `cannot read the git history: ...` message on each page that has usable paths.
- [x] Commit: `perf(wiki): check every page from one git pass`

## Fix round 1

### Task 3: Count a new page in an untracked folder as not committed

**Files:** Modify `internal/gitc/wikihistory.go`; test in `internal/gitc/wikihistory_test.go`.
**Seam:** `gitc.ReadWikiHistory` then `WikiHistory.Dirty`.
**verify:** Every file with edits not committed under the wiki folder is dirty on its own path: an edited page, a new page in a known folder, a new page in a folder git does not track yet. A clean page is not dirty.
- [x] Red: `TestWikiHistoryDirty` adds `wiki/sub/n.md` in a new folder; it fails because status lists only `wiki/sub/`.
- [x] Change: pass `--untracked-files=all` to the status call.
- [x] Commit: `fix(gitc): see new pages inside untracked wiki folders`

## Review notes

Deep review, full tier, two reviewers. Round 1: one BLOCKER (status without --untracked-files=all hid a new page in an untracked folder), fixed in fix round 1; round 2 CLEAN. No fact-check pass: the BLOCKER was proved by a red test first. Notes left: per-page WikiHistory.CheckCommits scans every commit's files (fine at 5.5 s on 20k commits); a work-tree-only rename from git add -N is read as one entry; --untracked-files=all lists every file of a big untracked folder under the wiki. Plan deviations accepted by the spec reviewer: no pathspec on the name-only log, ancestry walk lives in gitc, ReadWikiHistory takes no paths.
