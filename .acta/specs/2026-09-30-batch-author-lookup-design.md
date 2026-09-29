---
parent: scratch/2026-09-29-fast-path-lookup
created: "2026-09-30"
id: SPC-0031
hash: psimmjv
---
# Batch author lookup

Status: design approved by the user in chat on 2026-09-30. Bounded (changes the existing author lookup in `internal/board` and `internal/gitc`).

## Why

SCR-0020 asked for `acta path <id>` because `acta show` was slow. A measure on main bdc5f52 found a different cause:

- `acta show SCR-0013` takes about 2.9 s, every run.
- This repo has one worktree and no unmerged branches, so loading other trees is not the cost.
- `acta show` starts `git` 141 times. 137 of those are `git log --diff-filter=A --format=%an -- <file>`, one per planning file, from `fillAuthors` in `internal/board/closed.go`.
- The 137 calls take 3.2 s. One `git log --diff-filter=A --name-only --format=%an` over the same files takes 0.07 s.
- The author is only shown on the AUTHOR line of the TUI detail. Every board load pays for it: `list`, `show`, `set`, `tick` and the TUI.

So the fix is to ask git once per folder, not once per file. `acta path` is not needed and is dropped.

## Design

1. **`gitc.Authors(repo string, paths []string) map[string]string`** replaces `gitc.Author`. It runs one `git log --diff-filter=A --name-only --format=%an -- <paths>` in `repo` and maps each path to the name on the oldest commit that added it. Git lists newest first, so a later line wins. A path git never saw is left out of the map. A folder outside git gives an empty map without running git, the same as `Author` today.
2. **`fillAuthors`** groups the items on disk by folder and calls `Authors` once per folder. A path missing from the map falls back to `gitUserName(root)`, asked at most once, the same as today.
3. The `gitAuthor` test variable becomes `gitAuthors`, so tests can still count calls without a real repo.

Author names and fallbacks stay exactly as they are now. Only the number of git calls changes.

## Out of scope

- `acta path <id>`. Measure `acta show` after this lands; file a new scratch item only if it is still slow.
- The stale `status: brainstorming` line in the SCR-0013 frontmatter.

## Testing

- A board with many files in one folder makes one `gitAuthors` call for that folder, not one per file. Counted through the test variable.
- In a temp repo: user A adds a file and user B edits it later. The author is A.
- A file that is not committed gets the `UserName` name, and `UserName` is asked once for many such files.
- A folder outside git gives an empty author and runs no git.
- Two files added by different people in one commit range each get their own author.
