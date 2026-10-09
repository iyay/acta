---
id: DBT-0108
hash: u3i234c
parent: plans/2026-10-09-detail-box-padding
---
# Review NOTEs: Detail box padding, commit rows in the detail, no diff line numbers

- [ ] (medium) Detail padding is checked on screen only for a plan: no test draws a task, spec, bug, debt file, debt note, idea, empty board or a 3 to 4 line box through paneView and checks the D column and the blank top and bottom lines.
- [ ] (low) TestColorDiffHasNoGutter checks only the first character of each line, so a space gutter before a context line, or text dropped after the first character, would pass; the old exact-text check at full width is gone.
