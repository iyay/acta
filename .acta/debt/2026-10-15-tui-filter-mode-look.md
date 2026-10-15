---
id: DBT-0121
hash: mvro76y
parent: plans/2026-10-15-tui-filter-mode-look
---
# Review NOTEs: TUI filter mode look

- [ ] (medium) On the terminal theme the dim band equals the live band except for bold, since dimBand cannot mix ANSI slot numbers; a filter being typed barely shows on the picked row there (internal/tui/styles.go dimBand).
- [ ] (medium) The kept-filter status line says "esc clears", but on the Commits screen esc closes the screen and in the detail box esc moves the focus back (internal/tui/view.go statusLine).
