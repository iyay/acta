---
id: SPC-0090
created: "2026-10-06 06:28:06"
hash: olt3mpt
started: "2026-10-06 06:32:45"
finished: "2026-10-06 06:41:18"
---
# skill text: fix stale facts and rules that contradict each other

Status: Bounded, approved by the user in chat on 2026-10-06.

Why: a prompt audit of `plugin/skills/` (target Claude Opus 5.5, implementer prompt also read for Sonnet 5.5) found skill text that names a command acta never had, states one user's setup as a fact for every repo, and rules that contradict each other inside a skill or across skill and template. Agents follow the text, so each one sends them the wrong way.

Design, one edit per finding:
- `migrate/SKILL.md`: drop `acta migrate superpowers` (no such command; the CLI has only `migrate-root`). superpowers docs take the generic route with `git mv`, then `docs/superpowers` leaves `legacy`.
- `build/SKILL.md`: the house-rules copy names any untracked `CLAUDE.md` or `AGENTS.md`, not one symlink setup. Drop the "main branch with consent" exception. The sandbox fallback stops and tells the user; it never works in the main checkout. Step heading `### 5.` becomes `### 3.`.
- `build/implementer-prompt.md`: omp always uses `agent="task"`; `sonnet` only under `subagent_models: split`. `[BRIEF_FILE]` becomes Task N of `[PLAN_PATH]`, read that task only. "TDD if required" becomes "every change started from a failing test I watched fail".
- `slice/SKILL.md`: the full header is for `full` plans only. No "Announce at start". The overview names one reader who sees only their task, and one commit per task.
- `debug/SKILL.md`: Phase 1 instrumentation and the Phase 3 probe run in a scratch copy, never the checkout (matches "Phases 1 to 3 change no file").
- `tdd/SKILL.md`: config is not an exception. The "Existing code has no tests" row says test the behavior you change, the rest is a note for a later plan. Drop the "Final Rule" section (fourth copy of the Iron Law).
- `review/SKILL.md`: "Every polish commit gets the two reviewers". The "You're absolutely right!" reason no longer cites an instruction file.
- `review/code-reviewer.md`: the generic checklist becomes the skill's three questions plus the deep lens on trust boundary, auth, money, migration and delete paths. Drop "Acknowledge strengths".
- `shape/SKILL.md`: the user review gate is an instruction in the user's chat language, not a fixed English quote.
- `internal/plugincheck`: pins follow the new text, and `MustNot` pins keep each stale phrase from coming back.
- The last task adds 1 to the patch version in the three plugin files.

Out of scope: the Iron Law and rationalization tables in tdd, debug and land; the copied subagent-models paragraph; the other low audit flags. Byte caps stay as they are.

Tests: `scripts/test ./internal/plugincheck/` red on the new pins first, then green. At land: `scripts/test --full` and `scripts/eval`.
