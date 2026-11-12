---
id: SCR-0096
hash: qbozp23
title: Inline level for review depth auto
status: raw
created: "2026-11-13 00:55:25"
schema: "1"
---
# Inline level for review depth auto

## Words

### 2026-11-13

Bring back an inline level for review_depth auto: the orchestrator reviews a safe diff itself, with no reviewer subagent.

PLN-0235 took it out on 2026-11-13 because its eval case missed the bar: 3 of 10 runs wrote the `### Coverage` heading the spec asked for, against a bar of 7. In all 10 runs the agent picked inline correctly and showed its evidence; 7 wrote it as a bold label instead of a heading. So the pick works and the report shape does not hold.

Idea: give the inline review a fixed reply template to copy, measure with the strict heading grader first, and only then put the level back. The three inline conditions were: no risk signal; nothing changed that code outside the diff uses; every changed logic line covered by a test the task record shows failing first.

## Context

## Log

## Open questions
