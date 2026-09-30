---
id: DBT-0056
hash: oaj6wom
parent: plans/2026-10-01-priority-and-build-worktree
---
# Review NOTEs: Priority for Bugs and Debt Items, and Build Always Uses git worktree add Implementation Plan

- [ ] The spec puts the debt line priority helper next to MarkItem in internal/write/mark.go; the plan and the code put it in the new internal/write/priority.go.
- [ ] The H/M/L row tag comes from rowText, so it also shows in the Done pane, search and Activities rows, not only the Bugs and Debts open lists; the spec does not rule on this.
- [ ] No test checks that paintID colors the priority letter (bright red, yellow, dim); only the plain row text is tested.
- [ ] setBugPriority writes the bug file in place, while setLinePriority takes the lock and writes a temp file then renames it.
- [ ] A bad priority typed in the editor during bug new is not refused; it only shows up later as a board Problem, unlike --priority.
- [ ] addDebtLines (internal/write/ops.go) compares raw line text, so after a debt line gets a tag, running acta debt new again with the same note adds it a second time; comparing through board.SplitPriority would fix it.
- [ ] acta list --json and acta show --json carry no priority field, and the tag is gone from the debt Title, so an agent using the CLI cannot see any priority.
- [ ] setPriority wraps every error in bad(), including read, write and lock failures, so those exit 1 (bad input) instead of 3.
- [ ] setLinePriority returns a plain fmt.Errorf for "not a checklist box" where TickLine uses bad(); same behaviour, different style.
- [ ] The new priority map in styles.go sits before cyan, so gofmt re-aligned six existing lines and five field comments.
- [ ] paintID finds the tag with HasPrefix on the text after the id, which only holds while rowText is the only caller that adds the tag.
