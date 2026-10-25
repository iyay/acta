---
parent: specs/2026-10-26-build-close-no-suite-eval
closes: [DBT-0142.02]
depth: minimal
id: PLN-0196
created: "2026-10-26 07:16:27"
hash: r1kijqe
---
# Make the build close eval trustworthy

**Goal:** Make the close eval prove that the agent reaches review handoff without invoking review.
**Spec:** `.acta/specs/2026-10-26-build-close-no-suite-eval.md`
**Tests:** Fast `scripts/test ./internal/plugincheck -run TestBuildCloseNoSuiteReviewHandoff`; full `scripts/test --full`.

## Global Constraints
- Follow acta:lean: understand task and code first; skip it, reuse code here, stdlib, native feature, installed dependency, one line, minimum; never cut validation, security or accessibility.
- Keep the existing type-check and broad-suite marker graders. Do not change build instructions or unrelated behavior.
- Main baseline passed three eval runs. Keep candidate traces and record tool calls and skill reads from three runs in the worktree checkout.

## Files
- `internal/plugincheck/evals_test.go` checks the close handoff grader contract.
- `plugin/evals/build-close-no-suite/graders/review-handoff.md` checks the final handoff and tool trace.

## Waves
- Wave 1: Task 1.

### Task 1: Grade review handoff from the trace

**Files:** Modify `internal/plugincheck/evals_test.go` and `plugin/evals/build-close-no-suite/graders/review-handoff.md`.

**Seam:** The `build-close-no-suite` eval grader reads the final reply and agent tool trace.

**verify:** No run with a `Skill` call for `acta:review` can pass the handoff grader, while a run that reaches handoff without that call can pass; list the final reply, tool calls, and skill reads checked.

- [ ] Red: Add `TestBuildCloseNoSuiteReviewHandoff` to require a trace-focused LLM grader that rejects a `Skill` call for `acta:review`; run the focused test and observe failure because the current grader checks only the final reply.
- [ ] Change: Make the grader check the final handoff and fail if the trace contains a `Skill` call for `acta:review`; pass the focused test and `scripts/eval --case build-close-no-suite --runs 3 --scaffold --no-publish --keep-temp --trust-plugin`.
- [ ] Commit: `test(eval): verify build close review handoff`
