---
id: SCR-0071
hash: plyi001
title: Stop the debt pile and drop line-by-line review notes
status: brainstorming
created: "2026-10-15 10:15:11"
schema: "1"
started: "2026-10-15 10:15:14"
finished: "2026-10-15 10:22:39"
---
# Stop the debt pile and drop line-by-line review notes

## Words

### 2026-10-15

Review [note] NOTEs are written line by line into each plan's ## Review notes (438 lines in 69 plans); no commit ever acted on one. Debt has 616 open items vs 158 closed and keeps growing. Ask: keep only a one-line round summary for notes; triage the debt pile once; add a guard so debt stops growing faster than it closes.

## Context

## Log

### 2026-10-15

Q1 low debt: A. A (low) NOTE never becomes debt; it drops to [note] and is discarded. Debt takes only (medium) and (high).

### 2026-10-15

Q2 existing low items: A. Close all 542 open (low) items in one commit as dropped, reason 'low, under new debt bar'. Git keeps the history.

### 2026-10-15

Q3 guard: A. acta debt new accepts only lines starting (medium) or (high); any other line is an error. Untagged counts as low. The 5 untagged open items close with the low ones.

### 2026-10-15

Design approved: 1 review skill (debt only medium/high, notes one summary line), 2 debt new guard rejects whole batch, 3 one-time close of 547 low/untagged items. Out: TUI, cap, doctor warning.

## Open questions
