---
id: DBT-0026
hash: hxt9i9m
parent: plans/2026-09-29-fast-tests
---
# Review NOTEs: Faster Test Suite Implementation Plan

- [ ] Spec asks for common sizes 80x24 and 120x40 under -short; plan edgeWidths leaves out 80 and 120 (internal/tui/view_test.go edgeWidths).
- [ ] edgeWidths misses the help popup clamp ends from clamp(m.width*2/3, 24, 60) in view.go (widths 35/36, 89/90/91); only the full run covers them.
- [ ] internal/tui/view_test.go comment on TestSweep says "the slow tests above it" but they sit below.
- [ ] internal/tui/view_test.go sweep helper has no doc comment, unlike other helpers in the file.
- [ ] internal/tui/watch_test.go:60 reads loads without mu (pre-existing); parallel load makes a race report more likely; not reproduced in 30 race runs.
- [ ] cmd/acta and internal/trees got no parallel tests because their helpers use t.Setenv; the planned speedup there did not happen.
- [ ] Race run -count=3 -race -short took 4:27; three separate runs were not shown by the implementer, one reviewer ran count=3 green.
