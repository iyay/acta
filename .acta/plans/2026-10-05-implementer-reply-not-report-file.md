---
parent: bugs/2026-10-04-subagent-report-file-refused
depth: minimal
id: PLN-0094
created: "2026-10-05 20:55:01"
hash: mqu4822
started: "2026-10-05 21:00:30"
finished: "2026-10-05 21:03:23"
---
# Implementer reply instead of report file Implementation Plan

> **For agentic workers:** run this plan with acta:build, task by task. Steps use checkbox (`- [ ]`) syntax, and acta reads those boxes as task progress.

**Goal:** Implementers report in one reply of at most 25 lines and never write a report file.

**Spec:** `.acta/specs/2026-10-05-implementer-reply-not-report-file.md`

**Tests:** `scripts/test`, `scripts/test --full`

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Skill text uses short plain words and fits every user and harness.
- Comments in plain English a 10-year-old reads back without stopping.

## Waves

- Wave 1: Task 1
- Wave 2: Task 2

### Task 1: template and guard

**Files:** Modify `plugin/skills/build/implementer-prompt.md`, any other file under `plugin/skills/build/` that names an implementer report file, a test file under `internal/plugincheck/`, and `internal/plugincheck/budget_test.go` only if a cap needs it.
**verify:** No file under `plugin/skills/build/` asks an implementer to write a report file, the template asks for one reply of at most 25 lines with the fields the spec names, and plugincheck fails when `[REPORT_FILE]` or a report-file step comes back. List each file and line checked.
- [x] Failing test: a plugincheck test that reads implementer-prompt.md and fails on `[REPORT_FILE]`; it fails on today's text.
- [x] Code: rewrite the Report Format section per the spec; drop other report-file mentions in the build skill files.
- [x] Commit: `skills: implementers reply with a capped report, no report file`.

### Task 2: version bump

**Files:** Modify `plugin/.claude-plugin/plugin.json`, `plugin/.claude-plugin/marketplace.json`, `plugin/package.json`.
**verify:** The three files agree on one `x.y.z`, one patch above the version on the parent branch at the time this task runs; `internal/plugincheck` passes.
- [x] Failing test: none new; run `scripts/test ./internal/plugincheck` before and after the bump to watch it stay green.
- [x] Code: add 1 to the patch in all three files.
- [x] Commit: `plugin: bump patch version`.

## State

### Next

plan task-2 unblocked: capped-reply template landed in 7c79240
guard TestImplementerPromptNoReportFile fails on REPORT_FILE + report-file step
next: version patch bump task

## Review notes

- The RED and GREEN checks in the new test match any all-caps RED or GREEN in the file.
- Line 141 says "no report file" in passing; the test bans only the old phrases.
