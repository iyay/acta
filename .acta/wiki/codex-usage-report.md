---
type: Decision
title: The usage report counts recorded Codex tokens once
description: acta usage reads Codex transcripts offline; count per-response usage, never cumulative counters
paths: [internal/usage/, internal/cli/usage.go]
---

`acta usage [--json]` (SPC-0182, PLN-0199) reads the Codex transcripts under `$CODEX_HOME/sessions`, or `~/.codex/sessions`, and counts recorded model usage per root session and model. It is offline on purpose: no update probe, no cache, no writes, no network.

What it counts, and the traps:

- The per-response `usage` block only. `turn_token_usage`, `thread_token_usage` and the `event_msg` `token_count` stream are cumulative; adding them counts the same requests twice.
- Cached input sits inside `input_tokens`, so fresh input is `input - cached`. Reasoning tokens sit inside `output_tokens`.
- One count per `response_id`; copies that disagree are all dropped, and scan order must not pick a winner.
- A session belongs where its `session_meta.cwd` started, matched against the main checkout plus its existing worktrees. A nested repository belongs to itself. A child joins the root session that matched, even when it started elsewhere; a child whose root log is gone needs its own matching start and marks its group partial.
- The report describes discovered logs, not history: a session whose only checkout was deleted or moved can be missing. Help and the report say so.

Exit codes: 0 complete or no data, 2 partial (diagnostics on stderr, or in the JSON), 3 operational failure with no report printed, 1 outside a repository or a bad call.

Verified once against an independent sum over the closed 2026-10-15..26 logs: 22 root sessions, 2,318 responses, 6,366,220 fresh, 221,152,000 cached, 684,695 output, matching the spike in the spec.
