---
id: BUG-0029
hash: jpbscq6
---
# Implementer subagents cannot write the report file the build template asks for

## Symptom
In Claude Code, an implementer subagent that follows the build hand-off tries to write its report to the file the hand-off names. The Write tool refuses with "Subagents should return findings as text, not write report files". The report file never exists. The subagent then puts the whole report in its reply, which is long and fills the orchestrator's context.

## Root cause
plugin/skills/build/implementer-prompt.md:141 tells the implementer "Write your full report to [REPORT_FILE]". Claude Code blocks subagents from writing report files with the Write tool, so this step cannot succeed there.

## Repro
1. Run acta:build with the subagent executor in Claude Code, with a hand-off built from implementer-prompt.md that names a [REPORT_FILE] path.
2. The implementer's Write call to that path fails with the message above.

## Found in
main. Seen during the PLN-0081 build in ../acta-wiki on 2026-10-04: the implementers of Task 1 and Task 7 both hit it. Found while orchestrating, reported to the user.
