---
id: SPC-0177
created: "2026-10-26 05:53:24"
hash: y106rl9
started: "2026-10-26 05:58:13"
finished: "2026-10-26 06:56:28"
---
# Optional cold read before planning

## Problem

Shape's self-review shares the conversation that produced the spec. It can miss
decisions explained in chat but absent from the document. A new reader should be
able to understand the spec without that conversation.

## Scope

This Bounded change adds an optional cold read to the existing Bounded and
Architectural shape flows. Spike keeps its current flow. Bootstrap is separate.

## Behavior

1. After writing and self-reviewing the spec, shape offers a cold read before
   requesting final spec approval. Ask once for that spec in the current shape
   flow. Skipping proceeds directly to the existing approval gate.
2. On acceptance, launch one read-only reviewer in a fresh context with no
   inherited conversation. Give it the reviewer instructions and spec path.
   It reads the full spec and follows its references when needed to check a
   claim. Do not supply chat summaries, prior design conclusions, or a plan.
3. The reviewer checks missing decisions, contradictions, and ambiguity that
   would make implementation depend on unstated knowledge. It reports at most
   five consequential findings, each with a file and line, the consequence,
   and the smallest clarification. It skips style nits and unrelated code
   defects. If it finds none, it says so. It never edits files or runs write
   commands.
4. The main agent presents the findings briefly in chat and discusses them
   with the user. Findings are advice, not an automatic gate. Apply only
   agreed spec changes, self-review them, and commit through `acta commit`.
   Then request approval of the current spec before invoking slice.
5. Run one pass. Another pass requires a user request. If fresh-context
   delegation is unavailable or fails, explain that the cold read did not
   complete and offer retry or continuation without it. Never substitute a
   same-context review or continue through the user's decision automatically.

The offer and discussion use the configured chat language. Reviewer instructions
and repo artifacts use English. The review does not approve the spec or plan.
Existing implementation review and landing gates remain unchanged.

## Files and verification

Update `plugin/skills/shape/SKILL.md` and add
`plugin/skills/shape/cold-read.md` for the reviewer instructions. Extend focused
plugincheck tests and greenfield behavior evals to cover acceptance, skipping,
fresh-context dispatch, finding triage, unavailable delegation, and final spec
approval before planning. Keep skill text within the existing folder and file
budgets; reference the companion instructions only when cold read is chosen.

No new command, setting, dependency, report directory, or review loop is needed.
