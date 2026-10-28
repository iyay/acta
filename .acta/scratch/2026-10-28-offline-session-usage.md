---
id: SCR-0080
hash: iomxx0b
title: Offline session usage
status: brainstorming
created: "2026-10-28 22:48:40"
schema: "1"
started: "2026-10-28 22:48:49"
finished: "2026-10-28 23:01:08"
---
# Offline session usage

## Words

### 2026-10-28

Add an offline usage reader to acta, following the TokenEconomizer research and approved measurement spike.

The user requested detailed analysis of Microsoft TokenEconomizer, approved a read-only measurement spike, then asked to continue with the offline usage design.

The spike inspected 61 historical Codex logs for this repository, grouped into 22 root sessions, covering 2,318 unique model responses from 2026-10-15 through 2026-10-26. Recorded usage was 227,518,220 input tokens, including 221,152,000 cached input tokens, and 684,695 output tokens. Per-response sums matched final thread counters in all logs. These are recorded usage counts, not invoice amounts or proven savings.

The proposed first increment reads existing transcripts without new runtime hooks or LLM calls, groups root sessions with child agents, and separates fresh input, cached input and output by model. Monetary estimates require explicit pricing and must remain distinct from billed amounts. Stage attribution and successful-task outcomes are not yet established by the spike.

Decide initial harness support and whether explicit-price cost estimates belong in the first increment. Preserve existing workflow gates. Do not implement before design and plan approval.

## Context

## Log

### 2026-10-28

Scope answers approved by the user: Codex only for the first increment; token usage only, with fresh input, cached input and output separated per model. No monetary estimates in this increment.

### 2026-10-28

The user approved the proposed design: add `acta usage [--json]` as a stateless Go CLI command that reads existing Codex transcripts; scope it to the current repository and identifiable worktrees; group root sessions with child agents; count each response once; report fresh input, cached input, output and response counts by root session and model; report coverage and partial data; tolerate an unfinished final line in an active transcript. Use standard-library Go with focused fixtures. No monetary estimates, runtime hooks, database, LLM calls, task-success claims or savings claims in this increment.

## Open questions
