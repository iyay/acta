---
id: SPC-0074
created: "2026-10-05 10:12:19"
hash: b6uqifd
---
# Clearer names for the TUI top tabs

Status: Bounded, approved by the user in chat on 2026-10-05.

Why: three top tab names read wrong in English. "Scratches" means marks on a surface; the place is called Scratchpad elsewhere in acta. "Debts" reads as money owed; tech debt is uncountable. "Activities" reads as a list of things to do; an event feed is "Activity".

Design:

- In `topTabs` (`internal/tui/sidebar.go`): `Scratches` becomes `Scratchpad`, `Debts` becomes `Debt`, `Activities` becomes `Activity`. Bugs, Specs and Plans stay.
- Every other place in the TUI that shows these names (`model.go`, `styles.go`, help text) follows the table, so no old name is left on screen.
- Tab order, keys and kinds do not change. `Scratchpad` is one cell wider than `Scratches`; the tab bar must still fit at the narrowest size the tests already check.
- Tests that name the old labels change to the new ones.

Out of scope: the CLI words (`acta scratch`, `acta debt`) and the folder names.
