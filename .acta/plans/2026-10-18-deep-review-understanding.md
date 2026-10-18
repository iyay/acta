---
parent: specs/2026-10-18-deep-review-understanding-design
depth: minimal
id: PLN-0158
created: "2026-10-18 19:24:42"
hash: jwjcbtm
---
# Deep Review Understanding Plan

**Goal:** A successful landing after deep review ends with a short explanation and one optional behavior question, without delaying the merge.

**Spec:** `.acta/specs/2026-10-18-deep-review-understanding-design.md`

**Tests:** `scripts/test` (fast), `scripts/test --full` (full, at land only); task loop: `scripts/test ./internal/plugincheck/`; behavior: `scripts/eval --case 'deep-understanding-*'`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build is not authorized by this plan-writing turn; stop after saving and committing the plan, even though its depth is minimal.
- When build is requested, use an isolated worktree and the inline executor; begin with observed failing checks, review once at the close, and never push.
- Keep review, fact-check, fix, polish, gates and auto-land intact; use the existing selected review depth and review notes, with no setting, state, provider, extra reviewer or question engine.
- Preserve the selected closing review depth through polish and landing, including a one-run override; never infer it from the deep lens or reselect it from a default setting.
- Read the applicable wiki before edits: `plugin-loads-from-repo`, `plugin-ships-in-binary`, `plugincheck-caps-per-folder`, `skill-text-fits-everyone`, `eval-sandbox-greenfield` and `test-locked-runs`.
- Write repo content in English; keep skill text plain and usable across users and harnesses, and adjust only the text budgets that need it.
- At land, run `scripts/test --full` and the required `scripts/eval` gate, plus the existing wiki and state checks; do not substitute the selected eval cases for the landing gate.

## Waves

- Wave 1: Task 1 (inline).

### Task 1: Add the optional question to the deep landing report

**Files:** Modify `plugin/skills/review/SKILL.md`, `plugin/skills/review/depth.md`, `plugin/skills/land/SKILL.md`, `internal/plugincheck/skill_review_test.go`, `internal/plugincheck/skill_land_test.go`, and, only if needed, `internal/plugincheck/budget_test.go`; create the following eval files:

- `plugin/evals/deep-understanding-report/case.yaml`, `plugin/evals/deep-understanding-report/prompt.md`, `plugin/evals/deep-understanding-report/scaffold.sh`, `plugin/evals/deep-understanding-report/graders/behavior.md`.
- `plugin/evals/deep-understanding-standard-lens/case.yaml`, `plugin/evals/deep-understanding-standard-lens/prompt.md`, `plugin/evals/deep-understanding-standard-lens/scaffold.sh`, `plugin/evals/deep-understanding-standard-lens/graders/behavior.md`.
- `plugin/evals/deep-understanding-followup/case.yaml`, `plugin/evals/deep-understanding-followup/prompt.md`, `plugin/evals/deep-understanding-followup/scaffold.sh`, `plugin/evals/deep-understanding-followup/graders/behavior.md`.

**Seam:** Existing `acta:review` close and `acta:land` report, exercised by plugin contract checks and greenfield eval prompts and traces; follow the existing `review-depth-floor` and `polish-light-review` fixture and grader formats without touching the user's home or Git config.

**verify:** Every successful deep landing keeps its required evidence and report fields, explains the main behavior and its meaningful risk or boundary and test limits using inspected final landed code, and ends with exactly one optional outcome-and-reason question; no answer or question precedes or gates the merge. Contract checks own the exclusions: quick, standard, blocked or failed landings, no new setting/status/reviewer, and skip, silence or a new request never blocking or restarting completed landing. The report eval owns a deep one-run override against a standard default and a final landed revision that differs from the earlier reviewed snapshot, including a light polish; its trace must prove landing completes before the question and its explanation must match the final code. The standard-lens eval owns successful standard review of a sensitive path with no understanding question. The follow-up eval owns a short, differently worded prediction with a missed concept: explain the actual cause against source, resolve ambiguity without grading wording, and offer at most one optional further scenario. List the paths checked; inspect eval traces as well as grader results.

- [ ] Red: Add focused contract checks and the three bounded greenfield cases, run `scripts/test ./internal/plugincheck/ -run 'Test(ReviewUnderstanding|LandUnderstanding)'` and `scripts/eval --case 'deep-understanding-*'` before changing skills, and observe the missing deep-report contract fail.
- [ ] Change: Carry the selected depth into landing, put the explanation, final optional question and follow-up rules in the existing skills, fit only affected budgets, then make the narrow checks, all plugin checks and selected evals green while retaining all existing landing requirements.
- [ ] Commit: `feat(review): explain deep changes with an optional question after land`.
