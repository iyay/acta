---
type: Runbook
title: Plugin changes need a restart, not a reinstall
description: Claude Code loads the acta plugin in place from the repo's plugin/ folder; the plugin cache copy is stale and unused
paths: [plugin/]
timestamp: 2026-10-07T03:03:04Z
---

The local marketplace is a `directory` source pointing at this repo's `plugin/` folder, so skills load from there in place. The copy under `~/.claude/plugins/cache/` is old and unused.

After a plan that changes `plugin/` lands on main, restart Claude Code (skills and output styles are read at start). Run `go install ./cmd/acta` too when the CLI or hook text changed. No plugin reinstall or `/plugin update`.

Check again if the marketplace source changes.
