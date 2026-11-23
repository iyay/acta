---
# guards: audit tests for text tests and duplicates
name: cleanup-test-audit
description: Asked to audit the tests, the agent loads cleanup, skips the planning steps, reports the text test and the near duplicate pair with file:line, and changes nothing.
tags: [cleanup]
runs: 1
# Load the skill, read three test files and three source files, then one
# report. Eighteen turns leave room for that without letting one run eat the
# quota.
max_turns: 18
timeout_seconds: 600
allowed_tools: [Skill, Read, Agent]
---

Audit the tests in this repo.

I am away from the keyboard, so show me what you find and stop.

The acta CLI for this repo is ./bin/acta, and git is ./bin/git. Use those two
paths. The plain names may run a different program in this sandbox.
