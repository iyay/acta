---
id: DBT-0056
hash: f6lnyde
parent: plans/2026-10-01-show-path-lookup
---
# Review NOTEs: acta show --path Implementation Plan

- [ ] (low) internal/cli/cli.go namesID: an empty want matches a frontmatter id: or hash: with no value, so `acta show "" --path` or `.x --path` can print a file where plain show says unknown id; guard want != "".
- [ ] (low) internal/cli/cli.go namesID: indentation is trimmed, so a nested key like `meta:\n  id: X` in frontmatter also matches; no real file has one today.
- [ ] (low) internal/cli/cli.go scanPath: a sub-item that does not exist (PLN-0001.99, SPC-0001.01) still prints the parent path, where plain show says unknown id; spec item 3 allows it.
- [ ] (low) internal/cli/cli.go scanPath: accepts a frontmatter id the board would reject (wrong kind prefix or bad format), so --path can find a file plain show reports as broken.
- [ ] (low) With duplicate ids the fast path and the board both take the first match, but could disagree across worktrees; run acta id --fix-duplicates after merges.
- [ ] (low) The fast path reads only the main tree; an item that also lives in a worktree could print a different path than plain show.
- [ ] (low) Unit tests call scanPath directly, so a revert shows as a compile failure before any behaviour check.
- [ ] (low) internal/cli/cli_test.go: "io" sits in its own import group after time.
