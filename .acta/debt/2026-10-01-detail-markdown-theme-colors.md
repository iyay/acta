---
id: DBT-0068
hash: mhdk6tl
parent: plans/2026-10-01-detail-markdown-theme-colors
---
# Review NOTEs: Detail Markdown Theme Colors Implementation Plan

- [ ] (low) chromaLock in internal/tui/markdown.go guards only acta's own check-and-Register of a theme's chroma style; glamour and chroma read styles.Registry through styles.Get with no lock, so two first draws on different goroutines would race on the map. Fine today because the TUI renders on the tea goroutine only; breaks the day rendering moves off it.
