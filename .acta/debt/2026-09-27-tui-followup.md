---
id: DBT-0006
hash: gjfgz59
parent: plans/2026-09-27-tui-followup
---
# Review NOTEs: TUI follow-up Implementation Plan

- [x] view.go is now 903 lines, further over the 800-line cap; statusLineBoxes/linkSegments/stripCodes are still not moved out.
- [x] RecordAgent still lets a --start with no name after a named tick overwrite and wipe the stored agent name (recs[taskID] always takes the new empty Agent).
- [x] The (devel)/empty version-string check that normalizeVersion already does is still duplicated, now in internal/cli/cli.go instead of cmd/pmb/main.go; the debug.ReadBuildInfo path is still untested. (stale)
- [ ] (low) open(url) in title.go still calls exec.Command(...).Run() as part of Update, still blocking the UI while the browser opener runs, and still drops the error silently.
- [ ] (low) Search results still skip the in-progress split; enter on a not-checked-out item still shows the "not checked out" message and focuses [3].
- [ ] (low) Version/Feedback text is still cut mid-word at widths 30-39.
- [ ] (low) httpLink still accepts "https://:80".
- [ ] (low) Click tests still reuse the code's own linkSegments parser instead of an independent check.
