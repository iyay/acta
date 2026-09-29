---
id: DEBT-15
hash: lctk
parent: plans/2026-09-29-pane-sort-toggle
---
# Review NOTEs: Pane Sort Toggle Implementation Plan

- [ ] On a very narrow terminal the sort word is cut before the tab names (frame.go paneTop), so the title can hide the current sort.
- [ ] Search results ignore the pane sort (searchRows unchanged); spec does not cover search.
- [ ] Help test checks only the text oldest / newest, not the o key itself.
- [ ] order.go: items with same date and same file keep input order on both flips; harmless for tasks, two same-ID items from different trees would not swap.
- [ ] Sort word is drawn inline in paneTop, not through titlePieces; if it becomes clickable, tabX must learn about it.
- [ ] TestTitleShowsTheSortOfEachPane counts oldest/newest across the whole view; a fixture title with those words would break it.
- [x] The sort flag is per box, shared across tabs: o on Specs also flips Plans and Bugs (model.go newest).
- [ ] While a search is open the title still says oldest/newest, but search rows keep board order.
- [ ] o does not call clampOff/moveTo, so the selected item can scroll off screen until the next move.
- [ ] TestViewShowsTheTabBarAndTheDetail dropped its j, so its detail now shows a different plan.
- [ ] Merge result has 692 tests vs 693 on main's landing; the exact delta is not fully explained.
