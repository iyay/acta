---
parent: bugs/2026-10-04-subagent-report-file-refused
id: SPC-0084
created: "2026-10-05 20:55:01"
hash: heddmzy
---
# Implementers reply with a capped report instead of a report file

Status: Bounded, approved by the user in chat on 2026-10-05. Fixes BUG-0029.

Why: `plugin/skills/build/implementer-prompt.md` tells the implementer to write its full report to `[REPORT_FILE]`. Claude Code refuses that write from a subagent ("Subagents should return findings as text, not write report files"), so the implementer puts the whole report in its reply and fills the orchestrator's context.

Design:
- The template drops `[REPORT_FILE]` and the report-file step. The implementer sends one reply of at most 25 lines: status; commits (short SHA and subject); TDD evidence as one RED line (command and the key failing line) and one GREEN line; files changed; self-review findings and concerns.
- BLOCKED or NEEDS_CONTEXT still put the specifics in the reply.
- Every other line in the build skill files that names a report file for the implementer goes too.
- `internal/plugincheck` fails when the implementer template names `[REPORT_FILE]` or a report file.
- The last task adds 1 to the patch version in the three plugin files.

Tests: the plugincheck rule; plugincheck passes on the new text.
