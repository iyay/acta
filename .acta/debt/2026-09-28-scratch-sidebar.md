---
id: DEBT-10
hash: u2v4
parent: plans/2026-09-28-scratch-sidebar
---
# Review NOTEs: Scratchpad Kind and Five-Pane Sidebar Implementation Plan

- [ ] Success output of acta scratch new/add prints a path like ../../private/tmp/... when the repo sits behind a symlinked dir (macOS /tmp); pre-existing relative-path logic in report.
- [ ] acta list gives no sign of a spec whose parent points at nothing; only acta show prints the problem (same as debt today).
- [ ] internal/tui/view.go is 701 lines and model.go 657; split them before the next TUI feature so they stay under 800.
- [ ] go test ./... takes over 2 minutes; the tui package alone takes about 63s.
