---
id: SCR-0053
hash: hnfoyjm
title: 'acta update: check and install the latest version'
status: raw
created: "2026-10-07 09:40:36"
schema: "1"
---
# acta update: check and install the latest version

## Words

### 2026-10-07

User idea, 2026-10-07: acta needs an update mechanism. `acta update` checks for a newer version and, when there is one, installs the latest version on its own.

Depends on SCR-0052 (curl install script and release pipeline): there must be published releases to check against. Open: what gets updated (the binary, the plugin in each harness, or both), where the version check reads from, and whether a normal command should hint that an update exists.

## Context

Ruling 2026-10-07: the acta plugin folder is embedded in the binary (go:embed) and `acta setup` extracts it to ~/.acta/plugin, then installs into Claude Code and omp from there (PLN-0113 Task 04). So `acta update` = replace the binary, then extract the plugin again. A GitHub marketplace comes later, after the repo is public.

## Log

## Open questions
