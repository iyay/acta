---
type: Gotcha
title: Skill line cap counts the whole folder
description: MaxLines in plugincheck counts every markdown file in a skill folder, while byte caps are per file
paths: [internal/plugincheck/, plugin/skills/]
timestamp: 2026-10-07T14:17:35Z
---

`MaxLines` in `internal/plugincheck/skill_*_test.go` counts lines across every `.md` file in the skill folder, not only `SKILL.md`. A skill with a second file (shape with `probe.md`, frame with `startup.md`) needs a cap that fits both. The error reads `N lines of markdown, cap is M`, where N is the folder total.

Byte caps in `internal/plugincheck/budget_test.go` work the other way: one cap per file. A file over its cap, a file with no cap, and a cap for a gone file all fail.
