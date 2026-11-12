---
parent: debt/2026-11-13-review-depth-auto
depth: minimal
closes: [DBT-0166.01, DBT-0166.02]
after: [PLN-0238]
id: PLN-0239
created: "2026-11-13 06:13:09"
hash: lsr8dl1
---
# Review evals count reviewer roles

**Goal:** Review-depth evals distinguish reviewer identities and retries from helper calls and fix the polish case depth.

**Spec:** `.acta/specs/2026-11-13-workflow-debt.md` (SPC-0208)
**Tests:** fast `scripts/test ./internal/plugincheck`; full `scripts/test --full`.
**Gates:** eval: `scripts/eval`; runs when the diff touches `plugin/skills/` or `plugin/hooks/`; narrow form `scripts/eval --tag=<skill folder>` for each skill folder touched; whole when the diff touches `plugin/hooks/` or `plugin/references/`.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Build inline, serially, in a worktree under `~/.wwt/acta/`; never push.
- Keep source text generic across repos and harnesses; adjust byte caps only where needed, without cutting safety instructions.
- Validate changed graders with passing and failing samples before scoring a skill; use `evalomp.LoadCases` and `evalomp.Grade` in plugincheck for fixed graders, and measure judge rubrics on planted traces with the existing eval runner.
- For every skill-text task, run each named case once per call with `scripts/eval --case <name> --runs 10 -j 8` on old and new text under identical corrected graders and scaffolds; use an isolated copy for the old arm and the branch binary on PATH, never replace the installed binary.
- Record raw results with `acta state set`; isolate and exclude any regressing text change and keep its debt open, removing that item from `closes:` before land.
- Full suites run only at land; focused checks run after each change. Run scaffolds and mutation checks with fresh HOME and TMPDIR outside the repo.

## Build brief

- Land order: start only after PLN-0238 lands on main; these serial plans share plugincheck and eval files.

## Waves

- Wave 1: Task 1

### Task 1: Make reviewer-count evidence match the chosen depth

**Files:** `plugin/evals/review-auto-safe-one-reviewer/{prompt.md,scaffold.sh,graders/one-reviewer.md}`, `plugin/evals/polish-full-review/{prompt.md,scaffold.sh,graders/reviewer-dispatched.md,graders/standards-reviewer-dispatched.md}`, `internal/plugincheck/workflow_debt_evals_test.go`, `internal/plugincheck/evals_test.go`
**Seam:** existing review eval grader interface, with trace-focused judge rubrics where distinct reviewer identity requires judgment.
**verify:** The safe auto case explicitly records an inline build and accepts one combined reviewer, including a repeat request to that same reviewer for Coverage; a helper alone and two distinct reviewers fail. The standard polish case explicitly sets its depth and requires separate Spec and Standards reviewers; a combined brief containing both words cannot pass. Evaluate positive and negative traces for identities, resumed calls and helper roles, and retain literal dispatch checks only where they prove an actual role.
- [ ] Red: add `TestWorkflowDebtReviewerRoles` samples and judge-trace controls, then run `scripts/test ./internal/plugincheck -run 'TestWorkflowDebtReviewerRoles|TestEvalCases|TestEvalLLM'` to expose the current raw-call maximum and incidental-word checks.
- [ ] Change: declare the executor and standard polish depth in their scaffolds/prompts, grade distinct reviewer roles with existing grader types, update the narrow judge allowlist and samples, then run both named cases for 10 runs each and keep the corrected results separate from prior skill-effect claims.
- [ ] Commit: test(review): count distinct reviewer roles and allow Coverage retries

