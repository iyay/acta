---
id: DBT-0076
hash: k5c1c7q
parent: plans/2026-10-05-omp-session-switch-rules
---
# Review NOTEs: omp session switch rules and doctor files check Implementation Plan

- [ ] (low) checkFiles in internal/doctor/doctor.go reads each spec and plan with os.ReadFile, which follows symlinks: a plans/x.md link to a FIFO hangs doctor, and a link to a big file outside the repo is read whole. Lstat and skip non-regular files, as EnsureGitignore does.
- [ ] (low) checkFiles skips a file it cannot read, or a Glob error, with no word in the report, so an unreadable spec or plan is never flagged.
