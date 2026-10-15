---
id: DBT-0118
hash: bc7y8ol
parent: plans/2026-10-15-tui-search-active-pane
---
# Review NOTEs: TUI search filters the active pane only

- [ ] (medium) Enter on a plan or Activity head during a search does nothing: toggleRow returns true with no fold change, so the detail never opens (internal/tui/model.go:695).
- [ ] (medium) checkSearchRows only checks that a kept head has a deeper row after it, not that the deeper row matches; a third tree level would let a stray head pass (internal/tui/sidebar_test.go).
