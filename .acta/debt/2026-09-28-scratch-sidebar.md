---
id: DBT-0010
hash: u2v4d4d
parent: plans/2026-09-28-scratch-sidebar
---
# Review NOTEs: Scratchpad Kind and Five-Pane Sidebar Implementation Plan

- [ ] (low) Success output of acta scratch new/add prints a path like ../../private/tmp/... when the repo sits behind a symlinked dir (macOS /tmp); pre-existing relative-path logic in report.
- [ ] (low) acta list gives no sign of a spec whose parent points at nothing; only acta show prints the problem (same as debt today).
- [ ] (low) internal/tui/view.go is 701 lines and model.go 657; split them before the next TUI feature so they stay under 800.
- [ ] (low) go test ./... takes over 2 minutes; the tui package alone takes about 63s.
