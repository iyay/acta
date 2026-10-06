---
id: DBT-0036
hash: ff6jovz
parent: plans/2026-09-30-scratch-add-old-item
---
# Review NOTEs: Scratch Add Old Item Implementation Plan

- [ ] (low) internal/write/scratch_test.go:434: the no-section subtest is named "section " and prints as section_; a label like "section none" reads better in failure output.
- [ ] (low) acta scratch add usage and --section help list four section names, but an old item ignores which one was picked; the help does not say so.
