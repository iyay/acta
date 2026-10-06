---
id: DBT-0014
hash: rhb4l5a
parent: plans/2026-09-28-tui-top-tabs
---
# Review NOTEs: TUI Top Tabs Per Kind Implementation Plan

- [x] Clicking a tab name in the top bar (line 0) does nothing; at the base a click on a pane-title kind switched it. model.go:410-411 comment still says a click on a tab name switches tab.
- [ ] (low) Comments made wrong by the change: view.go:207 and :230 ("sidebar pane"), frame.go:89 ("the pane number, then the tabs"), frame.go:150 and scroll.go:44 ("sidebar pane"), model.go:19-20 points at "the table in sidebar.go"; sidebar.go keeps its name with no sidebar left.
- [x] enter on a plan row from a branch that is not checked out now only folds the plan; it no longer warns how to check it out (e still warns).
- [x] Help says "enter  focus the detail on the row" (view.go:90) while enter on a Plans task row does nothing.
- [ ] (low) Empty list pane says "nothing here" (scroll.go:132) while the detail says "No items" (detail.go:31).
- [ ] (low) TestEOpensTheEditorFromEveryPane skips a pane with nothing selected via continue, so an empty fixture pane asserts nothing.
- [ ] (low) Done sub-tab wrap-around (] on the last goes to the first) is no longer tested; swapping % n at model.go:296 for a clamp stays green.
- [x] sidebar_test.go:257 names a variable ids_, not Go style.
- [ ] (low) A resize re-clamps only the open tab; a saved tab's Done cursor can come back off-screen until that pane gets focus.
- [ ] (low) helpLines is 14 lines, so the help popup is cut below about 17 rows.
- [ ] (low) Return visit keeping the Done sub-tab has no test; dropping m.done in openTab (sidebar.go:99) stays green.
- [ ] (low) view_test.go and scroll_test.go are over the plan's 800-line limit (already over at the base).
- [x] TestViewShowsTheSidebarAndTheDetail was renamed TestViewShowsTheTabBarAndTheDetail though the plan keeps the old name. (stale)
- [ ] (low) The search query is shared by all tabs and the Done lists ignore it (already so at the base).
- [x] The spec layout sketch shows a key-hint bottom line; the status line still says "? help".
- [ ] (low) TestTabBarDropsTheNameFarthestFromTheOpenTab copies the production drop loop, so a shared mistake stays green; hard-coded expected bars would pin it.
- [ ] (low) widthOfBar and the w >= len(open)+1 check count bytes, not cells; right only while tab names are ASCII.
- [x] Plan Task 3 verify says tabBar drops like dropOrder, but titlePieces drops strictly from the right; title.go:68 comment also says "farthest from the open one first".
- [ ] (low) Drop-farthest-first can leave spare room (Debts at 25..31 shows 21 cells where 25 would fit); as the plan asks.
- [x] The implementer did not report the width-band list the Task 3 verify line asks for. (stale)
