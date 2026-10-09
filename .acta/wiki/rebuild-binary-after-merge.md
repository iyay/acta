---
type: Gotcha
title: Rebuild the acta binary after merge
description: Hooks and TUI use the acta on PATH, so rebuild it after any merge touching cmd/ or internal/
paths: [cmd/acta/, internal/hook/, internal/tui/]
timestamp: 2026-10-09T14:31:19Z
---

The binary on PATH (`go install ./cmd/acta`) does not rebuild itself. After a merge touching `cmd/` or `internal/`, hooks call old code and the TUI draws old screens. On 2026-09-26 a spec-line fix looked broken for this reason.

Fix: run `go install ./cmd/acta` and restart the TUI. Eval runs need the branch binary first on PATH too, or cases that read hook text or config paths go red with the old binary.
