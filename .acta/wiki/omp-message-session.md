---
type: Reference
title: omp message and session facts
description: CLI text goes straight to the model with no slash parsing, and /new, resume, fork emit session_switch
paths: [plugin/omp/]
timestamp: 2026-10-05T15:28:00Z
---

- `omp "/goal ..."` does not start goal mode: the CLI hands the text to the model, and slash commands parse only in the TUI input. A fresh process is already a new session, so drop `/new` at first launch; send `/goal` through a herdr prompt.
- omp fires `session_switch` (new, resume, fork), not `session_start`. The acta extension listens to it to reset the one-brainstorm block.
- A plain herdr prompt to a working omp agent steers its running goal; use it for late findings instead of a second `/goal`.
- `--plugin-dir` loads skills but never runs `package.json` extensions; install with `omp plugin link` instead.
