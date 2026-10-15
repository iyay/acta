---
id: SCR-0069
hash: jsnvflv
title: Notify when a new acta release exists
status: brainstorming
created: "2026-10-15 09:20:51"
schema: "1"
started: "2026-10-15 09:20:54"
finished: "2026-10-15 09:32:35"
---
# Notify when a new acta release exists

## Words

### 2026-10-15

acta checks for a new release every time it runs, TUI installed or not. Notify only: print that a newer version exists and to run acta update. Never install by itself.

## Context

## Log

### 2026-10-15

Q1 where: CLI stderr + TUI status bar badge + hook session note (option 3).

### 2026-10-15

Q2 how often: at most once a day, result cached in ~/.acta/, refresh runs in the background when cache is older than 24h, never blocks.

### 2026-10-15

Q3 opt-out: ACTA_NO_UPDATE_CHECK=1 env plus update_check: off in ~/.acta/config.yaml. Source builds never check.

### 2026-10-15

Approach: spawn detached child (acta update --check --cache) when cache stale; child writes ~/.acta/update.json; touch cache mtime before spawn as lock; notice shows from next run.

### 2026-10-15

Section 1 approved: cache ~/.acta/update.json {latest, checked}; skip on source build, env, config; stale >24h -> touch then spawn detached acta update --check --cache; child 10s timeout, atomic write, silent fail; update cmd itself does not spawn and sets cache after success.

### 2026-10-15

Section 2 approved: notice only when cached latest > current (semver). CLI stderr line after command, not for acta hook, --json, or non-tty stderr. TUI title badge from cache at open. SessionStart line, tell user once. Tests: stale/skip/compare, fake spawn, hook golden, TUI title.

## Open questions
