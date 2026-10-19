---
parent: specs/2026-10-19-build-coding-discipline-design
depth: minimal
after: [PLN-0170]
id: PLN-0171
created: "2026-10-19 11:33:25"
hash: s10ntrm
started: "2026-10-19 13:37:40"
finished: "2026-10-19 13:53:36"
---
# Build executor coding discipline implementation plan

**Goal:** Make caller tracing, pre-commit diff checks, and evidence-based failure triage explicit across build executors without extra workflow or redundant tests.
**Spec:** `.acta/specs/2026-10-19-build-coding-discipline-design.md`
**Tests:** Fast: `scripts/test`; full: `scripts/test --full` (land only); focused: `scripts/test ./internal/plugincheck/`; behavioral: `scripts/eval --case build-coding-discipline`, `scripts/eval --case build-inline-flow`, and `scripts/eval --case lean-tests`, with the branch binary on PATH.

## Global Constraints

- Follow acta:lean: understand the task and the code first; then skip it, reuse code here, stdlib, a native feature, an installed dependency, one line, the minimum; never cut validation, security or accessibility.
- Wait for PLN-0170 to land, then read the landed build instructions and matching wiki pages; do not change that plan or its in-flight branch, and do not restore its earlier structure.
- Build inline in a new isolated worktree. Preserve every executor's selection, gates, progress, commits, final review and land; never push.
- Put common coding guidance in the existing shared task loop; worker prompts reference that contract rather than repeat it. Reuse the existing self-check before commit, without per-task reviewers, extra reports or new fix rounds.
- Scope tracing to affected callers, shared code and existing tests. Escalate incompatible contracts and unapproved scope instead of expanding the implementation.
- Distinguish code defects, setup failures and unresolved plan decisions using observed evidence; use acta:debug for defects, fix routine setup only within existing authority, and escalate unresolved decisions. Do not repeat unchanged failed attempts, weaken checks, or count setup errors as TDD RED.
- Reuse coverage and small fixtures; each case protects distinct approved behavior or a concrete regression. Add no coverage, line-count or assertion quota, framework, ledger, CLI feature or model setting.
- Compare identical baseline/candidate requests in an isolated sandbox; inspect tools and artifacts, not self-reports. Keep reports outside the repo and pin HOME/TMPDIR to fresh temporary directories for scaffolds, write probes and mutants.

## Waves

- Wave 1: Task 1. One task owns the shared instructions and their behavioral proof; another split would duplicate setup and touch the same files.

### Task 1: Apply and verify a shared coding contract

**Files:** Modify `plugin/skills/build/SKILL.md` and `plugin/skills/build/implementer-prompt.md`; update affected checks in `internal/plugincheck/skill_build_test.go`, `budget_test.go` and `evals_test.go`; create `plugin/evals/build-coding-discipline/{case.yaml,prompt.md,scaffold.sh,graders/behavior.md}` only for behavior not covered after PLN-0170; reuse `plugin/evals/build-inline-flow/` and `plugin/evals/lean-tests/` unchanged unless their existing checks need a directly relevant extension.
**Seam:** The build skill running an approved task in the existing eval sandbox, observed through its tool trace, implementation, tests and commit; existing plugincheck checks guard packaging and workflow contracts.
**verify:** Every executor receives the same coding contract: relevant callers are traced before the shared implementation changes; the task diff is checked before commit for excess code, accidental contract changes, lost validation/error handling and redundant tests; implementation defects, environment failures and missing plan decisions take the correct debug/setup/escalation path without blind retries or false RED; list evidence and untested paths without treating static text checks as runtime proof.
- [x] Red: After the dependency lands, inspect and run existing relevant cases, then use the smallest missing behavioral scenario against the unchanged landed skill to observe a concrete failure in tracing, pre-commit checking or failure triage; retain passing coverage without duplicating cases or manufacturing a failure.
- [x] Change: Add the concise shared contract and align the worker self-check before commit, then pass `scripts/test ./internal/plugincheck/` and the affected eval commands above; compare baseline/candidate artifacts and traces, record uncovered paths and maintain risk-based test selection without duplicating debug's own tests.
- [x] Commit: `refactor(build): make executor coding checks explicit`, with the task trailer; use the existing wiki, deep review and land gates after the task checks pass.

## State

### Next

Task 1 committed. Next: wiki check, deep review over main..HEAD, land.
