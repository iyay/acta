---
id: DBT-0040
hash: sjf3pma
parent: plans/2026-09-30-title-from-frontmatter
---
# Review NOTEs: Item title falls back to the frontmatter title Implementation Plan

- [ ] (low) field uses fmt.Sprint, so a non-string title: (a yaml date, a number, a list, a multi-line block) shows as raw text; no tracked file has one today (board.go:395).
- [ ] (low) TestFileItemTitleFallback loops without t.Run subtests; the message names the case, style only (board_test.go:670).
