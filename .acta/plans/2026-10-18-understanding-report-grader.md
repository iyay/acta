---
parent: specs/2026-10-18-understanding-report-grader-design
depth: minimal
id: PLN-0160
created: "2026-10-18 20:34:07"
hash: d56wio4
---
# Complete Understanding Report Grading

**Goal:** The existing report grader rejects a missing or invented boundary, incomplete landing report fields, or missing evidence of reading final code after merge.

**Spec:** `.acta/specs/2026-10-18-understanding-report-grader-design.md`

**Tests:** `scripts/test` (fast), `scripts/test --full` (full, at land only); narrow checks: `scripts/test ./internal/plugincheck/`; behavior: `scripts/eval --case review-understanding-report --keep-temp`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Stop after writing and committing this plan; the user has not authorized build, despite minimal depth.
- When build is requested, use an isolated worktree and the inline executor, observe the failing checks before the grader edit, review at the close, and never push.
- Implementation changes only the existing grader; keep criteria 1–6, `type: llm`, `focus: trace`, the guards comment, landing skill, prompt, scaffold, other evals and runner unchanged.
- Read `eval-sandbox-greenfield` and `eval-merge-needs-libexec` before work; keep replay traces, judge prompts and output in a temporary directory outside the repo, never write a runner or add dependencies.
- Use the existing Claude judge CLI for bounded rubric replays: put the grader body, one synthetic trace and the instruction to return PASS or FAIL with a short reason into each prompt, then run `claude -p --model haiku --tools '' --no-session-persistence < "$grader_replay_dir/<case>.prompt"`; `grader_replay_dir` is a directory from `mktemp -d`, and case names are listed in the task.
- Repo content stays English; judge equivalent wording by meaning rather than literal headings. At land run the existing full test, wiki and state gates; run `scripts/eval` when required by the landing skill, without replacing that gate with the narrow case.
- Keep the four candidate traces fixed between red and green; replace only their rubric, retain each judge reason, and report inconsistent grading instead of rerunning until a desired verdict appears.

## Waves

- Wave 1: Task 1 (inline).

### Task 1: Require complete evidence and report content

**Files:** Modify `plugin/evals/review-understanding-report/graders/behavior.md` only; temporary replay artifacts are not committed.

**Seam:** The existing trace-focused LLM rubric, exercised through bounded judge CLI replays and `scripts/eval --case review-understanding-report --keep-temp`; use the current retry fixture and its final `min(2 ** attempt, 30)` implementation.

**verify:** PASS always requires all nine criteria: the existing six plus a source-supported risk/boundary or justified no-visible-risk statement, all required gate/cleanup/wiki/next-action report fields, and a successful source read after merge from the final parent checkout or exact landed SHA. Use `complete` and three negative replay cases: `missing-boundary`, `missing-report-fields`, and `missing-final-read`, with each negative breaking only its named new criterion; the complete trace must pass and each negative must fail. Preserve valid no-tests and no-wiki statements, optional-question behavior and semantic wording; accept Read, cat and git show equivalents while rejecting stale-revision, pre-merge, failed-read and merely-claimed-read evidence. The current cap example already establishes a boundary, so the missing-boundary case must retain the original six criteria while adding an unsupported risk claim that violates criterion 7, rather than treating the accurate cap itself as absent. List the replay verdicts and inspect the real eval trace for merge/read/report order; do not infer behavior from the rubric text alone.

- [ ] Red: Prepare the four replay prompts outside the repo, run each against the original rubric with the judge CLI, and observe at least one negative incorrectly receive PASS while the complete control passes; stop and report if the old rubric does not reproduce the gap.
- [ ] Change: Append criteria 7–9 with the approved evidence and wording rules, rerun the same four judge inputs with only the rubric replaced, then pass `scripts/test ./internal/plugincheck/` and `scripts/eval --case review-understanding-report --keep-temp`, inspecting the saved trace and keeping all artifacts outside the repo.
- [ ] Commit: `test(evals): require complete understanding report evidence`.
