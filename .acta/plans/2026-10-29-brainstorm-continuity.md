---
parent: specs/2026-10-29-brainstorm-continuity-design
depth: minimal
id: PLN-0200
created: "2026-10-29 06:33:47"
hash: zh8r9ld
---
# Brainstorm continuity implementation plan

**Goal:** Allow multiple Architectural brainstorms in one session and guide continuation or handoff by the discussion scenario.
**Spec:** `.acta/specs/2026-10-29-brainstorm-continuity-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full`; task checks name only their owning packages.

## Global Constraints

Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
Keep Architectural scratch items, answer logs, classification, design approval, and plan approval; keep Frame items raw.
Continue in the current session by default; only the user chooses a handoff, using methods available in the current harness.
Leave old session state files untouched and ignored; add no command, setting, dependency, or Walkie-specific behavior.
Keep this as one plan because runtime enforcement, session instructions, and eval expectations implement the same rule and must land together.

## Waves

Wave 1: Task 1. Wave 2: Task 2. Wave 3: Task 3. Run inline in order; tasks share verification surfaces.

### Task 1: Remove brainstorm blocking and its private bookkeeping

**Files:** Modify `internal/hook/session.go`, `internal/hook/session_test.go`, `internal/hook/gotest_test.go`, `internal/cli/hook.go`, `internal/cli/hook_session_test.go`, `internal/cli/hook_test.go`, and `plugin/hooks/pre-tool`.
**Seam:** `cli.Run` for `hook pre-tool`, `hook post-tool`, and `hook prompt`; shared hook event parsing and test-runner guard.
**verify:** Starting different scratch topics in one session never blocks or emits a session-limit reminder, with absent, valid old, or malformed old state; post-tool writes no brainstorm state; existing CLI tests retain coverage for wiki hints, patch input, invalid payloads, and the independent test-runner block.
- [ ] Red: Add CLI regressions for multiple topics and old state, run `scripts/test ./internal/cli/ -run 'TestHookSession'`, and observe the existing block or reminder fail the new expectations.
- [ ] Change: Remove brainstorm-only command matching, state tracking, blocking, and reminders while retaining event parsing and shared helpers still used elsewhere; update affected hook and CLI tests and run `scripts/test ./internal/hook/ ./internal/cli/ -run 'Test(ParseEvent|Hook|GoTest|BareGoTest)'` to green.
- [ ] Commit: `refactor(hook): remove per-session brainstorm limit`.

### Task 2: Replace active session-limit instructions with scenario guidance

**Files:** Modify `plugin/skills/shape/SKILL.md`, `plugin/skills/frame/SKILL.md`, `internal/hook/hook.go`, `internal/hook/hook_test.go`, `plugin/hooks/default-rules.md`, `internal/cli/hook_test.go`, `internal/plugincheck/skill_shape_test.go`, `internal/plugincheck/skill_frame_test.go`, `plugin/omp/FACTS.md`, and `.acta/wiki/codex-harness.md`; create `internal/plugincheck/brainstorm_continuity_test.go`.
**Seam:** `hook.SessionStart`, the generated default-rules fixture, and plugincheck reads of active skill and harness instructions.
**verify:** Active instructions permit continuing in-session, distinguish related, unrelated, resumed, and mixed discussions, preserve scratch and approval gates, and require user choice for handoff; no active surface reinstates a count-based limit or assumes a Claude-only handoff under another harness; Frame stays raw because design has not begun.
- [ ] Red: Add scenario and instruction-consistency assertions, replace obsolete required-limit assertions, run `scripts/test ./internal/plugincheck/ ./internal/hook/ -run 'Test(BrainstormContinuity|SkillShape|SkillFrame|Frame|SessionStart)'`, and observe old instructions fail.
- [ ] Change: Write the approved scenario guidance, synchronize generated session rules and conditional Herdr text, update Frame rationale and active harness documentation, and run the same checks plus `scripts/test ./internal/plugincheck/ -run 'Test.*(Budget|Shape|Frame|Continuity)'` to green without raising budgets unless the reviewed text cannot fit.
- [ ] Commit: `feat(shape): guide brainstorm continuity by scenario`.

### Task 3: Prove continuation and voluntary handoff with behavior evals

**Files:** Modify `internal/plugincheck/evals_test.go`, `plugin/evals/second-brainstorm-choices/prompt.md`, `plugin/evals/second-brainstorm-choices/graders/choices.md`, `plugin/evals/one-file-fix-no-brainstorm/prompt.md`, `plugin/evals-routing/routing-second-01/prompt.md`, `plugin/evals-routing/routing-second-01/graders/choices-offered.md`, `plugin/evals-routing/routing-second-02/prompt.md`, `plugin/evals-routing/routing-second-02/graders/choices-offered.md`, `plugin/evals-routing/routing-arch-02/prompt.md`, and `plugin/evals-routing/routing-arch-04/prompt.md`; create `prompt.md`, `scaffold.sh`, and `graders/continuity.md` in each of `plugin/evals/brainstorm-related-topic/`, `plugin/evals/brainstorm-resume-topic/`, `plugin/evals/brainstorm-mixed-context/`, and `plugin/evals/brainstorm-user-handoff/`.
**Seam:** Greenfield eval prompts, scaffolded scratch records, and graders loaded by the existing eval runners.
**verify:** Related topics reuse only applicable decisions; unrelated topics keep scope separate and can stay in-session; resumed topics read recorded decisions and check changed assumptions; mixed discussions get a scope recap and a voluntary session offer; an explicit handoff preserves the scratch identifier and uses a supported method without launching anything unasked; scratch and approval requirements remain enforced.
- [ ] Red: Add `TestEvalBrainstormScenarios` for scenario case coverage and guards, run `scripts/test ./internal/plugincheck/ -run TestEvalBrainstormScenarios`, and observe missing cases or obsolete mandatory-handoff expectations fail.
- [ ] Change: Add the four scenario cases, rewrite obsolete limit-based graders and routing assumptions, register guarded phrases and judgement graders, run `scripts/test ./internal/plugincheck/ -run 'Test(Eval|Scaffolds)'`, then run `scripts/eval --case 'second-brainstorm-choices|brainstorm-related-topic|brainstorm-resume-topic|brainstorm-mixed-context|brainstorm-user-handoff|one-file-fix-no-brainstorm'` and the four touched routing cases with `--eval-dir evals-routing` using the runner's supported case filters.
- [ ] Commit: `test(shape): cover brainstorm continuation and handoff scenarios`.
