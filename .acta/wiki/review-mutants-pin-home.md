---
type: Gotcha
title: Review mutants must pin HOME
description: Revert checks run tests with their safety guard removed, so any write under HOME hits the real home
paths: [internal/plugincheck/]
timestamp: 2026-10-07T22:17:31Z
---

A review mutant with a HOME guard removed overwrote the user's real voice config. Revert and mutation checks run tests whose safety line is the very line removed.

Fix: every reviewer brief that allows revert checks says to run mutants with HOME and TMPDIR set to fresh temp dirs. The same holds for write commands, which auto-commit: run them in a temp clone, never the real repo.
