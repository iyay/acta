---
id: DBT-0049
hash: gevkmls
parent: plans/2026-09-30-tui-tab-click-sidebar-hints
---
# Review NOTEs: Clickable Top Tabs, Wider Sidebar and Pane Key Hints Implementation Plan

- [ ] No test checks that statusLine calls hintLine; swapping the call for a no-op keeps the suite green.
- [ ] The left column has no upper limit now (100 columns on a 300-column screen); the plan asks for this.
- [ ] hints() is rebuilt on every frame and on every statusLineBoxes call; cheap today.
- [ ] Fold: space shows only on a tree head row, stricter than the spec text but matches what space does.
- [ ] Detail: enter shows on a plan head row, where enter folds the row first.
- [ ] The spec says the popup refuses debt lines, but openPopup only refuses tasks, so Status: s is hidden on a debt line where it works.
- [ ] No test covers a debt line row, nor Tick hidden on a worktree task row.
- [ ] The own comment in hints.go still says those rows get only the keys that work anywhere, but Edit and Copy now show on worktree and legacy rows.
- [ ] No test covers a legacy row, nor a worktree or not-on-disk row in the detail pane.
- [ ] The detail pane never shows Type: t, but t works there on a plain row the user owns.
- [ ] No test covers a press on a tab name while a drag highlight is showing.
