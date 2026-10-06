---
parent: debt/2026-10-06-skill-prompt-audit-fixes
closes: [DBT-0084.01, DBT-0084.02]
id: SPC-0091
created: "2026-10-06 07:09:08"
hash: atoye2r
started: "2026-10-06 07:12:34"
finished: "2026-10-06 07:21:02"
---
# skill text: audit follow-ups (implementer questions, build report, debug probes)

Status: Bounded, approved by the user in chat on 2026-10-06.

Why: three leftovers from the skill prompt audit (SPC-0090). The implementer prompt tells a subagent to "ask" when it cannot hold a chat. Build quotes fixed English report lines while chat goes in the user's language. A debug reference still sends the agent to add instrumentation to real code, and the CI probe example has no answer for probes that only run in CI.

Design:
- `build/implementer-prompt.md`: "Ask them now" and "ask questions… Don't guess" become: stop and report NEEDS_CONTEXT with the question; don't guess.
- `build/SKILL.md`: the two quoted "Already in isolated workspace at…" lines become an instruction: tell the user, in their chat language, the worktree path and branch, or that HEAD is detached and a branch is needed at finish time.
- `debug/root-cause-tracing.md`: "When you can't trace manually, add instrumentation:" puts the instrumentation in a throwaway clone outside the repo.
- `debug/SKILL.md`, by the CI example: a probe that can only run in CI gets its steps written out and the user asked to run them; never push.
- `defense-in-depth.md` stays: its Layer 4 logging is part of the fix, after the root cause.
- `internal/plugincheck`: `MustNot` pins for "Ask them now", "Already in isolated workspace", "When you can't trace manually, add instrumentation:"; red before the text change.
- The last task adds 1 to the patch version in the three plugin files.

Tests: `scripts/test ./internal/plugincheck/` red then green. At land: `scripts/test --full` and `scripts/eval`.
