---
id: SCR-0092
hash: avi1n2x
title: debug-four-line-report
status: raw
created: "2026-11-07 17:56:08"
schema: "1"
---
# debug-four-line-report

## Words

### 2026-11-07

Give every Phase 3 debug report four fixed lines: Cause (with file:line, or not proven plus what was ruled out), Repro (a command the user can paste), Evidence, Unverified. Aimed at the debug-order-dependent eval, whose failures left the paste-ready failing-order command out of the final reply (SPC-0157 runs). Split out of SPC-0191 on 2026-11-07, which was closed because its eval gate showed the independent check adds nothing. Measure on debug-order-dependent alone, old vs new, before shipping.

## Context

## Log

## Open questions
