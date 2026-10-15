---
id: SCR-0070
hash: slslvt7
title: 'TDD: cap test count and fold similar cases into table tests'
status: brainstorming
created: "2026-10-15 09:54:20"
schema: "1"
started: "2026-10-15 09:54:26"
finished: "2026-10-15 09:56:04"
---
# TDD: cap test count and fold similar cases into table tests

## Words

### 2026-10-15

acta:tdd limits tests to the plan's cases, but two gaps stay open:
- No cap on test count. A plan with too many cases still yields bloated tests; the brake sits in acta:slice, not tdd.
- No rule to fold similar cases into one table-driven test.
Close both: slice should bound how many cases a task names, and tdd should merge cases that differ only by input into one table test.

## Context

## Log

### 2026-10-15

Cap = rule, not a number: slice names only cases that differ in behavior; tdd folds cases that differ only by input into one table test.

### 2026-10-15

User asked to merge an outside tdd skill (seams, public interface, tautological, horizontal slicing, refactor outside loop, glossary) into this design.

### 2026-10-15

Seam agreed in the plan (slice names it per task; plan yes = seam yes). Refactor stays in the loop.

## Open questions
