---
id: SCR-0074
hash: ilg5eeb
title: Split one spec into several parallel plans
status: brainstorming
created: "2026-10-18 18:44:45"
schema: "1"
started: "2026-10-18 18:44:47"
---
# Split one spec into several parallel plans

## Words

### 2026-10-18

Let one spec split into several plans, so several sessions can build the same spec at once. Useful when the build executor is inline: each session runs one plan in its own worktree. User wants it as an option in acta setup.

Today: plans link to a spec through parent:, and nothing stops two plans per spec, but slice never offers a split, there is no rule for which plans can run together, no file-overlap check, and spec status may not know about partial progress.

## Context

## Log

### 2026-10-18

2026-10-18: user split the inline-default ask out; it ships first as its own Bounded spec.

## Open questions
